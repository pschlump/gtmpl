package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pschlump/Go-FTL/server/sizlib"
	"github.com/pschlump/filelib"
	"github.com/pschlump/godebug"
	"gitlab.com/pschlump/PureImaginationServer/ymux"
)

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

		"template_set": {
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
	ErrOn string // xyzzy - needs work
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
	Test         map[string]interface{}     // xyzzy - TBD
}

var gPath string

func init() {
	gPath = "./testdata;./testdata/list-of-fiels" // xyzzy - pull from config / gConfig

}

func TmplProcess(
	item string, //  "page_name", "partial" etc.
	tmpl_name string, // .html/.tmpl file or .json file with data+selects+templates
) (tmpl_rendered string, status int, err error) {
	_, _, tmpl_rendered, status, err = tmplProcessInternal(item, tmpl_name, gPath)
	if status == 200 || err == nil {
		return
	}
	fmt.Printf("Error On Template: status=%d error:%s\n", status, err)
	return
}

func tmplProcessInternal(
	item string, //  "page_name", "partial" etc.
	tmpl_name string, // .html/.tmpl file or .json file with data+selects+templates
	path string,
) (body, data, tmpl_rendered string, status int, err error) {

	// 1. Find/Clasify the 'tmpl_name' - .html / .tmpl, or .json file.
	full_path, file_type, err := PathFind(path, tmpl_name)

	// 2. if .html/.tmpl - just process template and return
	if file_type == ".html" || file_type == ".tmpl" {
		mdata := map[string]interface{}{}
		tmpl_rendered, err = renderTemplate(mdata, full_path)
		return
	}

	if file_type != ".json" {
		err = fmt.Errorf("Invalid file type, msut be .html, .tmpl, or .json")
		return
	}

	// 3. if .json
	//		a. Read/Deciperh .json file
	ds, err := readJsonTemplateConfigFile(full_path)
	if err != nil {
		// xyzzy
	}

	mdata := make(map[string]interface{})

	//		c. Run the .SQL section to collect the data
	tdata, err := ProcessSQL(ds)
	if err != nil {
		// xyzzy
	}
	mdata["data"] = tdata

	tmp := make(map[string]string)
	s := godebug.SVar(ds.JsonLayout)
	json.Unmarshal([]byte(s), &tmp)
	mdata["jsonLayout"] = tmp

	//		b. Read set of f temlates for "item" (do a parse on a list of items)
	//    	d. Run the template with the data
	var templateList []string
	if item == "" {
		templateList = ds.TemplateList
	} else {
		aa, ok := ds.TemplateSet[item]
		if !ok {
			err = fmt.Errorf("Invalid/Missing item name >%s<\n", item)
			return
		}
		templateList = aa.TemplateList
	}
	tmpl_rendered, err = renderTemplate(mdata, templateList...)

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

// tmpl_rendered, err = renderTemplate(mdata, full_path)
func renderTemplate(mdata map[string]interface{}, fn ...string) (tmpl_rendered string, err error) {

	// xyzzy

	return
}

/*
type DataType struct {
	To    string // Name of data item to place data in.
	Stmt  string // SQL ro run
	Bind  map[string]string
	ErrOn string // xyzzy - needs work
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
	Test         map[string]interface{}     // xyzzy - TBD
}
*/

// mdata, err := ProcessSQL(ds)
func ProcessSQL(ds JsonTemplateRunnerType) (mdata map[string]interface{}, err error) {

	// xyzzy
	// goal mdata["data"] with all the data from each of the items in ds.Data

	mdata = make(map[string]interface{})
	mdata["data"] = make(map[string]interface{})

	for _, dd := range ds.Data {
		to := dd.To
		stmt := dd.Stmt
		indata := make([]string, len(dd.Bind), len(dd.Bind))
		for jj := 0; jj < len(dd.Bind); jj++ {
			indata[jj] = ""
		}
		for key, vv := range dd.Bind {
			pos := getPos(key)
			if pos >= len(indata) {
				for jj := len(indata); jj < pos; jj++ {
					indata[jj] = ""
				}
			}
			indata[pos] = vv
		}
		indata2 := make([]interface{}, len(indata), len(indata))
		for rr := range indata {
			indata2[rr] = indata[rr]
		}
		rows, err := ymux.SQLQuery(stmt, indata2...)
		if err != nil {
			fmt.Printf("Error: %s error:%s data:%s\n", stmt, err, godebug.SVar(indata))
			mdata[to] = "error"
		} else {
			data, _, _ := sizlib.RowsToInterface(rows)
			mdata[to] = data
		}
	}

	if db114 {
		fmt.Printf("Results of SQL : %s\n", godebug.SVarI(mdata))
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

// ds, err := readJsonTemplateConfigFile(full_paty)
func readJsonTemplateConfigFile(fn string) (ds JsonTemplateRunnerType, err error) {

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

func TmplTest(
	item string, //  "page_name", "partial" etc.
	tmpl string, // .tmpl file or .json file with data+selects+templates
	test_name string, // A test to run
) {
	body, data, tmpl_rendered, status, err := tmplProcessInternal(item, tmpl, gPath)
	// now check error/status, status 200 => err == nil
	_, _, _, _, _ = body, data, tmpl_rendered, status, err

	// xyzzy - take [data] and process the "test" seciton of it!

	return
}

var db114 = true
