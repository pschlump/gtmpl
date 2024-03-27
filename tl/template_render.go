package tl

// TODO - set the path / gPath
// 		- test at top level
// 		- fix TemplateSet to work.
// 		- test a "set" of templates based on request type

// TODO - Implement the "error" stuff
// TODO - Implement the "test" stuff

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pschlump/dbgo"
	template "github.com/pschlump/extend"
	extendsprig "github.com/pschlump/extend/extendsprig"
	"github.com/pschlump/filelib"
	ymux "github.com/pschlump/gtmpl/sizlib"
)

// "git.q8s.co/pschlump/piserver/ymux"

// "github.com/pschlump/sprig"
// "text/template"

/*

	1. Run multiple selects

		select: [
			{
				"to": "name"
				"stmt": "select..."
				"errror_on": 0 rows etc.
			}
			, {
				"to": "name...x"
				"stmt": "select..."
				"errror_on": 0 rows etc.
				"bind": [
					"$1": "{{.name}}"
				]
			}
		]

	Bind values by name from GET/POST and previous queries

	Run Template at the end

		"template": [ "base.html", "tmpl1.html" ... ]

	Set of request for template

		"TemplateSet": {
			"page_name":
				{ "template": [ "full_page.html", "extend1.tmpl" ] }
			, "partial":
				{ "template":  [ "section.html", "extend1.tmpl" ]
				, "target": "body"
				}
		}

	Layout Info
		"jsonLayout": {
			-- data for layout / style of layout
		}

	Testing
		"test": [
			{
				"data": { "user_id": "123" }
			,	"expect": ...
			}
		]

*/

type DataType struct {
	To    string // Name of data item to place data in.
	Stmt  string // SQL ro run
	Bind  map[string]string
	ErrOn string // Xyzzy - needs work
}

type LayoutItemType struct {
	MatchTo     string `json:"match"`
	TemplateFor string `json:"for"`
}

type TemplateSetType struct {
	TemplateList []string `json:"Template"`
	Target       string   `json:"Target"`
}

type JsonTemplateRunnerType struct {
	TemplateList []string                   `json:"Template"`
	JsonLayout   []LayoutItemType           `json:"JsonLayout"`
	TemplateSet  map[string]TemplateSetType `json:"TemplateSet"`
	Data         []DataType                 `json:"SelectData"`
	Test         map[string]interface{}     // Xyzzy - TBD
}

var gPath string

func init() {
	gPath = "./testdata;./testdata/list-of-files" // Xyzzy - pull from config / gConfig
}

func TmplProcess(
	item string, //  "page_name", "partial" etc.
	tmpl_name string, // .html/.tmpl file or .json file with data+selects+templates
	dataFunc func(name string) string,
	DB *sql.DB,
) (tmpl_rendered string, status int, err error) {
	_, _, tmpl_rendered, status, err = tmplProcessInternal(item, tmpl_name, gPath, dataFunc, DB)
	if status == 200 || err == nil {
		return
	}
	if db115 {
		fmt.Printf("Error On Template: status=%d error:%s\n", status, err)
	}
	return
}

func tmplProcessInternal(
	item string, //  "page_name", "partial" etc.
	tmpl_name string, // .html/.tmpl file or .json file with data+selects+templates
	path string,
	dataFunc func(s string) string,
	DB *sql.DB,
) (body, data, tmpl_rendered string, status int, err error) {

	// 1. Find/Clasify the 'tmpl_name' - .html / .tmpl, or .json file.
	full_path, file_type, err := PathFind(path, tmpl_name)

	// 2. if .html/.tmpl - just process template and return
	if file_type == ".html" || file_type == ".tmpl" {
		mdata := map[string]interface{}{}
		tmpl_rendered, err = RenderTemplate(DB, mdata, full_path)
		return
	}

	if file_type != ".json" {
		err = fmt.Errorf("Invalid file type, msut be .html, .tmpl, or .json")
		return
	}

	// 3. if .json
	//		a. Read/Deciperh .json file
	ds, err := ReadJsonTemplateConfigFile(full_path)
	if err != nil {
		fmt.Printf("File: %s Error: %s at %s\n", full_path, err, dbgo.LF())
		return
	}

	mdata := make(map[string]interface{})

	//		c. Run the .SQL section to collect the data
	tdata, err := ProcessSQL(DB, &ds, dataFunc)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	mdata["data"] = tdata

	tmp := make(map[string]string)
	s := dbgo.SVar(ds.JsonLayout)
	json.Unmarshal([]byte(s), &tmp)
	mdata["jsonLayout"] = tmp

	//		b. Read set of f temlates for "item" (do a parse on a list of items)
	var templateList []string
	if item == "" {
		templateList = ds.TemplateList
	} else {
		aa, ok := ds.TemplateSet[item]
		if !ok {
			err = fmt.Errorf("Invalid/Missing item name >%s< ds->%s<- at:%s\n", item, dbgo.SVarI(ds), dbgo.LF())
			return
		}
		templateList = aa.TemplateList
	}

	for ii, tmpl_name := range templateList {
		full_path, _, err := PathFind(path, tmpl_name)
		if err != nil {
			fmt.Printf("Missing file ->%s<-\n", tmpl_name)
			templateList[ii] = "./testdata/empty.html"
			continue
		}
		templateList[ii] = full_path
	}

	//    	d. Run the template with the data
	tmpl_rendered, err = RenderTemplate(DB, mdata, tmpl_name, templateList...)

	//		e. Return results if successful.
	return
}

// full_path, file_type, err := PathFind(path, tmpl_name)
// PathFind searches each location in the path for the specified file.  It returns the first full file name
// that is found or an error of "Not Found"
// path - top level directory to search in.
// fn - file name or pattern to search for.
func PathFind(path, fn string) (full_path, file_type string, err error) {
	err = fmt.Errorf("Not Found")
	ss := strings.Split(path, ";")
	for _, pp := range ss {
		full_path = filepath.Join(pp, fn)
		if filelib.Exists(full_path) {
			file_type = filepath.Ext(fn)
			err = nil
			return
		}
	}
	return
}

// RenderTemplate will take the data in mdata and combine the templates and render the resulting template.
// The template that is rendered is "render".
// Each of the files, fns, is a full path to a file to render.
// Example Call: tmpl_rendered, err = RenderTemplate(mdata, full_path)
func RenderTemplate(DB *sql.DB, mdata map[string]interface{}, tmplList string, fns ...string) (tmpl_rendered string, err error) {

	if DbOn["db4a"] {
		fmt.Printf("Top of RenderTemplate fns=%s AT: %s\n", dbgo.SVarI(fns), dbgo.LF())
	}
	//create a new template with some name
	name := fmt.Sprintf("tmpl_%s", tmplList)
	tmpl := template.New(name).Funcs(extendsprig.TxtFuncMap())
	// 		.Option("missingkey=zero","missingvalue=empty")
	//		.Funcs(extendsprig.TxtFuncMap()) // .Funcs(sprig.TxtFuncMap())
	tmpl, e0 := tmpl.ParseFiles(fns...)
	if e0 != nil {
		err = fmt.Errorf("Parse: error %s on %s, at:%s\n", e0, tmplList, dbgo.LF())
		return
	}

	var buffer bytes.Buffer
	foo := bufio.NewWriter(&buffer)

	e0 = tmpl.ExecuteTemplate(foo, "render", mdata)
	if e0 != nil {
		err = fmt.Errorf("Execute Error: %s\n", e0)
		return
	}

	foo.Flush()
	tmpl_rendered = buffer.String()

	return
}

// mdata, err := ProcessSQL(ds)
func ProcessSQL(DB *sql.DB, ds *JsonTemplateRunnerType, getDataForSQL func(name string) string) (mdata map[string]interface{}, err error) {

	/*

	   type DataType struct {
	   	To    string // Name of data item to place data in.
	   	Stmt  string // SQL ro run
	   	Bind  map[string]string
	   	ErrOn string
	   }

	   type LayoutItemType struct {
	   	MatchTo     string `json:"match"`
	   	TemplateFor string `json:"for"`
	   }

	   type TemplateSetType struct {
	   	TemplateList []string `json:"template"`
	   	Target       string   `json:"target"`
	   }

	   type JsonTemplateRunnerType struct {
	   	TemplateList []string                   `json:"template"`
	   	JsonLayout   []LayoutItemType           `json:"jsonLayout"`
	   	TemplateSet  map[string]TemplateSetType `json:"TempalteSet"`
	   	Data         []DataType                 `json:"SelectData"`
	   	Test         map[string]interface{}
	   }

	*/
	// goal mdata["data"] with all the data from each of the items in ds.Data

	mdata = make(map[string]interface{})

	for _, dd := range ds.Data {
		to := dd.To
		stmt := dd.Stmt

		maxpos := 0
		for key := range dd.Bind {
			pos := getPos(key)
			if pos < 0 {
			}
			if pos > maxpos {
				maxpos = pos
			}
		}
		if db114 {
			fmt.Printf("AT:%s maxpos=%d\n", dbgo.LF(), maxpos)
		}
		indata := make([]string, maxpos, maxpos)
		for jj := 0; jj < len(dd.Bind); jj++ {
			indata[jj] = ""
		}

		for key, vv := range dd.Bind {
			pos := getPos(key) - 1
			indata[pos] = getDataForSQL(vv)
		}
		if db114 {
			fmt.Printf("AT:%s indata=%s\n", dbgo.LF(), dbgo.SVar(indata))
		}

		// Type convert from string to interface{} in slice, so
		// go from []string, to []interface{}
		indata2 := make([]interface{}, maxpos, maxpos)
		for rr := range indata {
			indata2[rr] = indata[rr]
		}

		if db114 {
			fmt.Printf("AT:%s stmt=%s\n", dbgo.LF(), stmt)
		}
		rows, err := ymux.SQLQuery(DB, stmt, indata2...)

		if db114 {
			fmt.Printf("AT:%s err=%s\n", dbgo.LF(), err)
		}
		if err != nil {
			fmt.Printf("Error: %s error:%s data:%s\n", stmt, err, dbgo.SVar(indata))
			mdata[to] = fmt.Sprintf("Error: %s error:%s data:%s\n", stmt, err, dbgo.SVar(indata))
		} else {
			data, _, _ := RowsToInterface(rows)
			mdata[to] = data
		}
	}

	if db114 {
		fmt.Printf("Results of SQL, mdata = %s\n", dbgo.SVarI(mdata))
	}

	return
}

func getPos(s string) (n int) {
	nn, _ := strconv.ParseInt(s[1:], 10, 64)
	n = int(nn)
	if n < 1 {
		n = 1
	}
	return
}

// ds, err := ReadJsonTemplateConfigFile(full_paty)
func ReadJsonTemplateConfigFile(fn string) (ds JsonTemplateRunnerType, err error) {

	// type JsonTemplateRunnerType struct {
	buf, e0 := ioutil.ReadFile(fn)
	if e0 != nil {
		err = e0
		return
	}

	err = json.Unmarshal(buf, &ds)
	if err != nil {
		return
	}

	return
}

// data, mdata, err := processSQL(ds)

// xyzzy - implement
// xyzzy - test
func TmplTest(
	item string, //  "page_name", "partial" etc.
	tmpl string, // .tmpl file or .json file with data+selects+templates
	test_name string, // A test to run
	DB *sql.DB,
) {
	fx := func(s string) string {
		return "" // xyzzy - implement!
	}
	body, data, tmpl_rendered, status, err := tmplProcessInternal(item, tmpl, gPath, fx, DB)
	// now check error/status, status 200 => err == nil
	_, _, _, _, _ = body, data, tmpl_rendered, status, err

	// xyzzy - take [data] and process the "test" seciton of it!

	return
}

var db114 = false
var db115 = false