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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
var dbFlag map[string]bool
var tmplOpt = ""
var tmplIsDir = false
var outOpt = ""

func init() {
	dbFlag = make(map[string]bool)
	dbFlag["setup"] = false
}

func main() {

	theData := make(map[string]interface{})
	tmplList := make([]string, 0, 25)

	inRange := func(name string, pos int) string {
		if pos < len(os.Args) {
			return os.Args[pos]
		}
		fmt.Fprintf(os.Stderr, "Usage: Invalid option %s - argument required, position=%d\n", name, pos)
		os.Exit(1)
		return ""
	}

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

	// xyzzy - add in --help -h -?

	for ii := 1; ii < len(os.Args); ii++ {
		arg := os.Args[ii]
		if arg == "--cfg" || arg == "-C" {
			Cfg = inRange(arg, ii+1)
			ii++
		} else if arg == "--version" || arg == "version" {
			fmt.Printf("gtmpl version 0.0.2 - from /Users/corwin/go/src/www.2c-why.com/Corp-Reg/gtmpl\n")
			os.Exit(0)
		} else if arg == "--help" {
			fmt.Printf(`gtmpl version 0.0.3 

--cfg | -C <fn>				Config file, cfg.json for example.
--cli | -c "JSON-data"		Data in JSON format to use to substitute into template.
--data | -d <fn>			Data in JSON format in a file.

`)
		} else if arg == "--cli" || arg == "-c" {
			if dbFlag["setup"] {
				fmt.Printf("got a --cli at %d\n", ii)
			}
			dt := inRange(arg, ii+1)
			ii++
			mergeData([]byte(dt))
		} else if arg == "--data" || arg == "-d" {
			if dbFlag["setup"] {
				fmt.Printf("got a --data at %d\n", ii)
			}
			fn := inRange(arg, ii+1)
			ii++
			dt, err := ioutil.ReadFile(fn)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading %s, error=%s\n", fn, err)
				os.Exit(1)
			}
			mergeData(dt)
		} else if arg == "--tmpl" || arg == "-t" {
			if dbFlag["setup"] {
				fmt.Printf("got a --tmpl at %d\n", ii)
			}
			tmplOpt = inRange(arg, ii+1)
			ii++
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
				fmt.Fprintf(os.Stderr, "%s %s must be a file or a directory containing template files\n", arg, tmplOpt)
				os.Exit(1)
			}
		} else if arg == "--out" || arg == "-o" {
			if dbFlag["setup"] {
				fmt.Printf("got a --out at %d\n", ii)
			}
			outOpt = inRange(arg, ii+1)
			ii++
			// xyzzy -only 1 of these
		} else if arg == "--merged" || arg == "-M" {
			if dbFlag["setup"] {
				fmt.Printf("got a --merged at %d\n", ii)
			}
			// xyzzy - merged data flag -and- dump, where?
			ii++
			// xyzzy -only 1 of these
		} else if arg == "--debug" {
			if dbFlag["setup"] {
				fmt.Printf("got a --debug at %d\n", ii)
			}
			debugFlag := inRange(arg, ii+1)
			dbFlag[debugFlag] = true
			if dbFlag["setup"] {
				fmt.Printf("Debug flag %s enabled\n", debugFlag)
			}
			ii++
		} else {
			fmt.Fprintf(os.Stderr, "Usage: invalid option, %s\n", arg)
			os.Exit(2)
		}
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

	if dbFlag["echo_input"] {
		fmt.Printf("Data: %s\n", godebug.SVarI(theData))
		fmt.Printf("gCfg: %s\n", godebug.SVarI(gCfg))
		fmt.Printf("TMPL files tmplList: %s\n", godebug.SVarI(tmplList))
	}

	for tn, tf := range tmplList {
		//create a new template with some name
		tmpl := template.New(fmt.Sprintf("tmpl_%d", tn)).Funcs(sprig.TxtFuncMap())

		if dbFlag["proc_file"] {
			fmt.Printf("%sprocessing [%s]%s\n", MiscLib.ColorGreen, tmpl, MiscLib.ColorReset)
		}

		// read in template, parse it
		tmplFn := ""
		if tmplIsDir {
			tmplFn = tmplOpt + "/" + tf
		} else {
			tmplFn = tf
		}
		if dbFlag["file_name"] {
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
		if dbFlag["file_name"] {
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
