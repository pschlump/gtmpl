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
	"text/template"

	"github.com/pschlump/MiscLib"
	"github.com/pschlump/filelib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/sprig"
)

//
//1. gtmpl -cli {data} -data file.json -tmpl Temlate.tmpl -out fn.out --inputDataMerged merged.data.json --tmplDir ./dir/
//	1. Run each of --cli --data in order
//	2. Run a single template --tmpl - or all the tempaltes in a directory
//	3. Out specifies path/or file
//	4. --inputDataMerged - is the merged JSON data
//

type ConfigFile struct {
	Name string `json:"name"`
}

var Cfg = "cfg.json"
var Cli = ""
var DbOn map[string]bool
var tmplOpt = ""
var tmplIsDir = false
var outOpt = ""

func init() {
	DbOn = make(map[string]bool)
}

var optCfg = flag.String("cfg", "", "xyzzy.")
var optVersion = flag.Bool("version", false, "xyzzy.")
var optHelp = flag.Bool("version", false, "xyzzy.")
var optCli = flag.String("cli", "", "xyzzy.")
var optData = flag.String("data", "", "xyzzy.")
var optTmpl = flag.String("tmpl", "", "xyzzy.")
var optOut = flag.String("out", "", "xyzzy.")
var optDebug = flag.String("Debug", "", "xyzzy.")

var optDbConn = flag.String("conn", "", "Database (PostgreSQL) connection string.")
var optDbName = flag.String("dbname", "", "Database (PostgreSQL) name.")
var optQuery = flag.String("sql", "", "Database (PostgreSQL) select to get data.")
var optUseSubData = flag.Bool("sub-data", false, "use .data as a field for array of data.")

func init() {
	flag.StringVar(optCfg, "C", "", "xyzzy.")
	flag.BoolVar(optVersion, "V", false, "xyzzy.")
	flag.BoolVar(optHelp, "H", false, "xyzzy.")
	flag.StringVar(optCli, "c", "", "xyzzy.")
	flag.StringVar(optData, "d", "", "xyzzy.")
	flag.StringVar(optTmpl, "t", "", "xyzzy.")
	flag.StringVar(optOut, "o", "", "xyzzy.")
	flag.StringVar(optDebug, "D", "", "xyzzy.")
}

func main() {

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
	}

	if *optVersion {
		fmt.Printf("gtmpl version 0.0.4\n")
		os.Exit(0)
	}
	if *optHelp {
		usage()
		os.Exit(0)
	}
	if *optCfg != "" {
		Cfg = *optCfg
	}

	if *optDbConn != "" {
		db_x := ConnectToAnyDb("postgres", *optDbConn, *optDbName)
		if db_x == nil {
			fmt.Fprintf(os.Stderr, "%sUnable to connection to database: s\n", MiscLib.ColorRed, MiscLib.ColorReset)
			os.Exit(1)
		}
		data, err := SelData2(db_x.Db, *optQuery)
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
	if *optOut != "" {
		outOpt = *optOut
	}
	if outOpt == "" {
		fmt.Fprintf(os.Stderr, "Usage: --out must be specified\n")
		os.Exit(1)
	}

	// if --tmpl is a directory then --out must be a directory -check-
	if tmplIsDir {
		if !filelib.ExistsIsDir(outOpt) {
			fmt.Fprintf(os.Stderr, "if tempalte input is a directory the --out must also specify a directory, out=%s\n", outOpt)
			os.Exit(3)
		}
	}

	var gCfg ConfigFile
	if Cfg != "" {
		gCfg = ReadConfig(Cfg)
	}

	if DbOn["echo_input"] {
		fmt.Printf("Data: %s\n", godebug.SVarI(theData))
		fmt.Printf("gCfg: %s\n", godebug.SVarI(gCfg))
		fmt.Printf("TMPL files tmplList: %s\n", godebug.SVarI(tmplList))
	}

	for tn, tf := range tmplList {
		//create a new template with some name
		tmpl := template.New(fmt.Sprintf("tmpl_%d", tn)).Funcs(sprig.TxtFuncMap())

		if DbOn["proc_file"] {
			fmt.Printf("%sprocessing [%s]%s\n", MiscLib.ColorGreen, tmpl, MiscLib.ColorReset)
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
			ofn = outOpt + "/" + bn
		} else {
			ofn = outOpt
		}
		if DbOn["file_name"] {
			fmt.Printf("Output file name with path [%s]\n", ofn)
		}

		fp, err := filelib.Fopen(ofn, "w")
		if err != nil {
			fmt.Printf("Unable to open %s for output, error: %s ", ofn, err)
			break
		}

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

// -------------------------------------------------------------------------------------------------
func ReadConfig(fn string) (rv ConfigFile) {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Must supply config file %s, errror=%s\n", fn, err)
		os.Exit(1)
	}

	err = json.Unmarshal(data, &rv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing %s, errror=%s\n", fn, err)
		os.Exit(1)
	}
	return
}

func usage() {
	fmt.Printf(`gtmpl version 0.0.4 

--cfg | -C <fn>				Config file, cfg.json for example.
--cli | -c "JSON-data"		Data in JSON format to use to substitute into template.
--data | -d <fn>			Data in JSON format in a file.

`)
}
