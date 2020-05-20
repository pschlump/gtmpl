package main

/*
gtmpl - a template processor with merge of multiple data sources

	--cli 'JSON'				Command line JSON data
	--data file					file containing JSON data
	--tmpl [file/dir]
	--out [file/dir]

	--debug flag

TODO: Need a way to modify output file name to SampleCorpToken.sol for contract that
	modifies name.
	Some sort of "fileNameTemplate.tmpl" that gets used with data.

--- Later ----------------------------------------------------------------------------

TODO: modifiers on data (pipes) |UC |ContractName etc.
	Built in functions that can be piped to.

TODO: connect to PG and pull data from pg
	1. Use cfg.json to connect to PG / Redis
	2. Pull in data from each as necessary
	3. Add that to theData
	--pg "name: select ..."
	--pg "table: name" => "select * from name"
	--redis "name: key" -> data
	--redis "name: set-key" -> data

TODO: some CLI processing is not yet done. (xyzzy 3)

*/

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pschlump/MiscLib"
	template "github.com/pschlump/extend"
	sprig "github.com/pschlump/extend/extendsprig"
	"github.com/pschlump/filelib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/gtmpl/tl"
	"github.com/pschlump/ms"
	"gitlab.com/pschlump/PureImaginationServer/ReadConfig"
)

// "gitlab.com/pschlump/PureImaginationServer/ReadConfig"

//
//1. gtmpl -cli {data} -data file.json -tmpl Temlate.tmpl -out fn.out --inputDataMerged merged.data.json --tmplDir ./dir/
//	1. Run each of --cli --data in order
//	2. Run a single template --tmpl - or all the tempaltes in a directory
//	3. Out specifies path/or file
//	4. --inputDataMerged - is the merged JSON data
//

type ConfigFile struct {
	Name   string `json:"name"`
	DbConn string `json:"dbconn" default:"user=postgres dbname=postgres port=5432 host=127.0.0.1 sslmode=disable"`
	DbName string `json:"dbname" default:"postgres"`
}

var Cli = ""
var DbOn map[string]bool
var tmplOpt = ""
var tmplIsDir = false
var out *os.File = os.Stdout

func init() {
	DbOn = make(map[string]bool)
}

var optCfg = flag.String("cfg", "cfg.json", "Global Configuration File.")           // 1
var optVersion = flag.Bool("version", false, "Display version of this program.")    // 2
var optHelp = flag.Bool("help", false, "Display usage/help information.")           // 3
var optCli = flag.String("cli", "", "Provide data on command line in JSON format.") // 4
var optData = flag.String("data", "", "Provide data in a JSON or XML file.")        // 5
var optTmpl = flag.String("tmpl", "", "Template to process.")                       // 6
var optOut = flag.String("out", "", "Destination to send output to.")               // 7
var optDebug = flag.String("debug", "", "Comma seperated list of debug flags.")     // 8
var optTmplList = flag.String("tmpl-list", "", "Template list to parse.")           // 9
var optExtend = flag.String("tmpl-extend", "", "Template to process with extend.")  // 10

var optDbConn = flag.String("conn", "", "Database (PostgreSQL) connection string.")
var optDbName = flag.String("dbname", "", "Database (PostgreSQL) name.")
var optQuery = flag.String("sql", "", "Database (PostgreSQL) select to get data.")
var optUseSubData = flag.Bool("sub-data", false, "use .data as a field for array of data.")

func init() {
	//	flag.StringVar(optCfg, "C", "", "Global Configuration File.")                   // 1
	//	flag.BoolVar(optVersion, "V", false, "Display version of this program.")        // 2
	//	flag.BoolVar(optHelp, "H", false, "Display usage/help information.")            // 3
	//	flag.StringVar(optCli, "c", "", "Provide data on command line in JSON format.") // 4
	//	flag.StringVar(optData, "d", "", "Provide data in a JSON or XML file.")         // 5
	//	flag.StringVar(optTmpl, "t", "", "Template to process.")                        // 6
	//	flag.StringVar(optOut, "o", "", "Destination to send output to.")               // 7
	//	flag.StringVar(optDebug, "D", "", "Comma seperated list of debug flags.")       // 8
}

var gCfg ConfigFile

func main() {

	flag.Parse()

	fns := flag.Args()
	_ = fns

	theData := make(map[string]interface{})
	tmplList := make([]string, 0, 25)

	mergeData := func(data []byte) {
		tD := make(map[string]interface{})
		err := json.Unmarshal(data, &tD)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing json, error=%s, json=%s\n", err, data)
			os.Exit(1)
		}
		for key, val := range tD {
			theData[key] = val
		}
	}

	if *optDebug != "" {
		ss := strings.Split(*optDebug, ",")
		for _, s := range ss {
			DbOn[s] = true
		}
		tl.SetDbOn(DbOn)
	}

	if *optVersion {
		fmt.Printf("gtmpl version v1.0.4\n")
		os.Exit(0)
	}
	if *optHelp {
		Usage()
		os.Exit(0)
	}

	// fmt.Printf("AT: %s\n", godebug.LF())
	if *optCfg != "" {
		err := ReadConfig.ReadFile(*optCfg, &gCfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	}

	if *optDbConn != "" {
		gCfg.DbConn = *optDbConn
	}
	if *optDbName != "" {
		gCfg.DbName = *optDbName
	}

	// fmt.Printf("optQuery == ->%s<- AT: %s\n", *optQuery, godebug.LF())
	if *optQuery != "" {
		// fmt.Printf("AT: %s\n", godebug.LF())
		db_x := tl.ConnectToAnyDb("postgres", gCfg.DbConn, gCfg.DbName)
		if db_x == nil {
			fmt.Fprintf(os.Stderr, "%sUnable to connection to database: %s\n", MiscLib.ColorRed, MiscLib.ColorReset)
			os.Exit(1)
		}
		data, err := tl.SelData2(db_x.Db, *optQuery)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%sUnable to connection to database/failed on table select: %v%s\n", MiscLib.ColorRed, err, MiscLib.ColorReset)
			os.Exit(1)
		}
		if DbOn["query"] {
			fmt.Printf("Data=%s\n", godebug.SVarI(data))
		}

		if *optUseSubData {
			theData = map[string]interface{}{
				"data": data,
			}
		} else if len(data) == 1 {
			theData = data[0]
		} else if len(data) > 1 {
			fmt.Printf("Warning - %d rows returend from %s, using 0th row\n", len(data), *optQuery)
			theData = data[0]
		} else if len(data) == 0 {
			fmt.Printf("Warning - 0 rows returend from %s\n", *optQuery)
		}
	}

	// fmt.Printf("AT: %s\n", godebug.LF())
	if *optCli != "" {
		dt := *optCli
		mergeData([]byte(dt))
	}
	if *optData != "" {
		fn := *optData
		dt, err := ioutil.ReadFile(fn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s, error=%s\n", fn, err)
			os.Exit(1)
		}
		mergeData(dt)
	}

	// fmt.Printf("AT: %s\n", godebug.LF())
	if *optTmpl != "" {
		tmplOpt = *optTmpl
		// if dir - find all ./*.tmpl files and put those in list, if file just add to processing list.
		if filelib.ExistsIsDir(tmplOpt) {
			// check if --tmpl:fn is a file or dir.
			fns, dirs := filelib.GetFilenames(tmplOpt)
			// fmt.Printf("AT: %s, fns=%s\n", godebug.LF(), fns)
			if len(dirs) > 0 {
				fmt.Fprintf(os.Stderr, "Warning: not performaing recursive directory search on %s - sub-directories %s skipped\n", tmplOpt, dirs)
			}
			tmplList = append(tmplList, fns...)
			tmplIsDir = true
		} else if filelib.Exists(tmplOpt) {
			tmplList = append(tmplList, tmplOpt)
		} else {
			fmt.Fprintf(os.Stderr, "`--tmpl %s` must be a file or a directory containing template files\n", tmplOpt)
			os.Exit(1)
		}
	}

	// fmt.Printf("AT: %s\n", godebug.LF())
	// if --tmpl is a directory then --out must be a directory -check-
	if tmplIsDir {
		if !filelib.ExistsIsDir(*optOut) {
			fmt.Fprintf(os.Stderr, "if tempalte input is a directory the --out must also specify a directory, out=%s\n", *optOut)
			os.Exit(3)
		}
	}

	if DbOn["echo_input"] {
		fmt.Printf("AT: %s\n", godebug.LF())
		fmt.Printf("Data: %s\n", godebug.SVarI(theData))
		fmt.Printf("gCfg: %s\n", godebug.SVarI(gCfg))
		fmt.Printf("TMPL files tmplList: %s\n", godebug.SVarI(tmplList))
	}

	if DbOn["db4"] {
		fmt.Printf("AT: %s\n", godebug.LF())
	}
	if *optExtend != "" {
		// *optExtend is a template name that will have an "extend" in it.
		if !filelib.Exists(*optExtend) {
			fmt.Printf("Missing File ->%s<-\n", *optExtend)
		} else {

			name := fmt.Sprintf("derived_%s", *optExtend)
			tmpl := template.New(name)

			rtFuncMap := template.FuncMap{
				"Center":      ms.CenterStr,   //
				"PadR":        ms.PadOnRight,  //
				"PadL":        ms.PadOnLeft,   //
				"PicTime":     ms.PicTime,     //
				"FTime":       ms.StrFTime,    //
				"PicFloat":    ms.PicFloat,    //
				"nvl":         ms.Nvl,         //
				"Concat":      ms.Concat,      //
				"title":       strings.Title,  // The name "title" is what the function will be called in the template text.
				"ifDef":       ms.IfDef,       //
				"ifIsDef":     ms.IfIsDef,     //
				"ifIsNotNull": ms.IfIsNotNull, //
				// From: https://stackoverflow.com/questions/21482948/how-to-print-json-on-golang-template/21483211
				// "marshal": func(v interface{}) template.JS {
				"marshal": func(v interface{}) string {
					a, _ := json.Marshal(v)
					// return template.JS(a)
					return string(a)
				},
				"emptyList": func(v []string) bool {
					// fmt.Fprintf(os.Stderr, "%s v=%s %s\n", MiscLib.ColorRed, godebug.SVarI(v), MiscLib.ColorReset)
					// if len(v) == 0 {
					// 	return true
					// } else {
					// 	return false
					// }
					return len(v) == 0
				},

				// "import": ...			// Import file at run time.

				// I think that the binding time is wrong on this.  We need to change the template and pull in the base
				// at "Parse" time not at "Execute" time.
				"extend": func(vv string) string {
					fmt.Fprintf(os.Stderr, "Extend Called: %s at:%s\n", vv, godebug.LF())
					var baseTmpl *template.Template
					baseName := filepath.Base(vv)
					baseTmpl = template.New(baseName)

					b, err := ioutil.ReadFile(vv)
					if err != nil {
						fmt.Printf("AT: %s error: %s\n", godebug.LF(), err)
					}
					s := string(b)

					baseTmpl, err = baseTmpl.Parse(s)
					if err != nil {
						fmt.Printf("AT: %s error: %s\n", godebug.LF(), err)
					}

					// monkey patch in all of 'tmpl'(closure) into baseTmpl
					ts := tmpl.Templates()
					for _, tt := range ts {
						baseTmpl.AddParseTree(tt.Name(), tt.Tree)
					}

					tmpl = baseTmpl // replace 'tmpl' with baseTmpl

					return ""
				},
			}

			tmpl = tmpl.Funcs(rtFuncMap)
		}

	} else if *optTmplList != "" {
		var fp *os.File
		if DbOn["db4"] {
			fmt.Printf("AT: %s\n", godebug.LF())
		}
		//create a new template with some name
		name := fmt.Sprintf("tmpl_%s", *optTmplList)
		tmpl := template.New(name).Funcs(sprig.TxtFuncMap())
		fns := strings.Split(*optTmplList, ",")
		for ii, fn := range fns {
			if !filelib.Exists(fn) {
				fmt.Printf("Missing File %d, ->%s<-\n", ii, fn)
			}
		}
		if DbOn["db4"] {
			fmt.Printf("AT: %s - fns = %s\n", godebug.LF(), godebug.SVar(fns))
		}
		tmpl, err := tmpl.ParseFiles(fns...)
		if err != nil {
			fmt.Printf("Parse: error %s on %s, at:%s\n", err, *optTmplList, godebug.LF())
			goto done
		}
		fp, err = filelib.Fopen(*optOut, "w")
		if err != nil {
			fmt.Printf("Unable to open %s for output, error: %s ", *optOut, err)
			goto done
		}
		defer fp.Close()
		if DbOn["db4"] {
			fmt.Printf("%sAT: %s - defined = %s%s\n", MiscLib.ColorCyan, godebug.LF(), tmpl.DefinedTemplates(), MiscLib.ColorReset)
		}
		err = tmpl.ExecuteTemplate(fp, "foo", theData)
		if err != nil {
			fmt.Printf("Execute: %s\n", err)
			goto done
		}
	done:
	} else {
		for tn, tf := range tmplList {
			// fmt.Printf("AT: %s\n", godebug.LF())

			//create a new template with some name
			tmpl := template.New(fmt.Sprintf("tmpl_%d", tn)).Funcs(sprig.TxtFuncMap())

			if DbOn["proc_file"] {
				fmt.Printf("%sprocessing [%v]%s\n", MiscLib.ColorGreen, tn, MiscLib.ColorReset)
			}

			// read in template, parse it
			tmplFn := ""
			if tmplIsDir {
				tmplFn = tmplOpt + "/" + tf
			} else {
				tmplFn = tf
			}
			if DbOn["file_name"] {
				fmt.Printf("Template file name with path [%s]\n", tmplFn)
			}
			body, err := ioutil.ReadFile(tmplFn)
			if err != nil {
				fmt.Printf("Unable to open: %s error %s ", tmplFn, err)
				break
			}

			//parse some content and generate a template
			tmpl, err = tmpl.Parse(string(body))
			if err != nil {
				fmt.Printf("Parse: error %s on %s ", err, tmplFn)
				break
			}

			// generate output file name
			ofn := ""
			if tmplIsDir {
				bn := filelib.RmExt(tf) // strip off .tmpl - leaving basename
				// TODO - xyzzy - if not .tmpl on end - then ERROR --------------------------------------- <<<<<<<<<<<<<<<<<<<<<<<<<<<<
				bn = filepath.Base(bn) // just the name
				ofn = *optOut + "/" + bn
			} else {
				ofn = *optOut
			}
			if DbOn["file_name"] {
				fmt.Printf("Output file name with path [%s]\n", ofn)
			}

			fp, err := filelib.Fopen(ofn, "w")
			if err != nil {
				fmt.Printf("Unable to open %s for output, error: %s ", ofn, err)
				break
			}
			defer fp.Close()

			//merge template 'tmpl' with content of 's'
			// use output file
			// err = tmpl.Execute(os.Stdout, theData)
			err = tmpl.Execute(fp, theData)
			if err != nil {
				fmt.Printf("Execute: %s\n", err)
				return
			}
		}
	}
}

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage: %s\n", os.Args[0])
	flag.PrintDefaults()
}
