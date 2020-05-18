package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"gitlab.com/pschlump/PureImaginationServer/ReadConfig"
	"gitlab.com/pschlump/PureImaginationServer/ymux"
)

// func PathFind(path, fn string) (full_path, file_type string, err error) {

func Test_FindPath(t *testing.T) {

	fp, ex, err := PathFind("tmpl1;./testdata", "base-table.html")

	if err != nil {
		t.Errorf("Error Got error %s when expecting success", err)
	}
	if ex != ".html" {
		t.Errorf("Expected .html got ->%s<-\n", ex)
	}
	if fp != "testdata/base-table.html" {
		t.Errorf("Expected ->testdata/base-table.html<- got ->%s<-\n", fp)
	}
}

// func RenderTemplate(mdata map[string]interface{}, fns ...string) (tmpl_rendered string, err error) {
func Test_RenderTemlate(t *testing.T) {

	var mdata map[string]interface{}

	json.Unmarshal([]byte(`{
"data": [
	{
		  "id": "111222333444"
		, "original_file_name": "abc-def.xls"
	}
]
}`), &mdata)

	expect := `
	 
	<table class="table">
		<thead>
			<tr>
				
			<th>
				Original File Name
			</th>
			<th>
				Action
			</th>

			<tr>
		</thead>
		<tbody>
				
	
		<th>
			<td>
				abc-def.xls
			<td>
			<td>
				<button class="bind-this" data-id="111222333444" data-click-run="run-form">Run</button>
			<td>
		</th>
	

		</tbody>
	</table>
	
	<button>Upload New File</button>

`

	tv, err := RenderTemplate(mdata, "testdata/base-table.html", "testdata/list-of-files/lof.html")

	if db821 {
		fmt.Printf("AT: %s Template ->%s<- error:%s\n", godebug.LF(), tv, err)
	}

	if err != nil {
		t.Errorf("Got error from render : %s\n", err)
	}

	if tv != expect {
		ioutil.WriteFile(",a", []byte(expect), 0644)
		ioutil.WriteFile(",b", []byte(tv), 0644)
		t.Errorf("Expected ->%s<- got ->%s<-\n", expect, tv)
	}
}

// func ProcessSQL(ds JsonTemplateRunnerType) (mdata map[string]interface{}, err error) {
func Test_ProcessSQL(t *testing.T) {
	var ds JsonTemplateRunnerType

	SetupDatabase()

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
	ds.Data = []DataType{
		{
			To:   "test1",
			Stmt: "select * from t_ymux_documents where user_id = $1",
			Bind: map[string]string{
				"$1": "user_id",
			},
		},
	}

	data := map[string]string{
		"user_id": "52bc4522-bed8-4ee4-73b3-be0ed73d7f1f",
	}
	fx := func(s string) string {
		return data[s]
	}

	mdata, err := ProcessSQL(&ds, fx)

	if db822 {
		fmt.Printf("at:%s err:%s mdata=%s\n", godebug.LF(), err, godebug.SVarI(mdata))
	}

}

var dbInit = false

// DB is the connection info to the database.  It must be external to be used.
var DB *sql.DB

func SetupDatabase() {

	if !dbInit {
		dbInit = true

		err := ReadConfig.ReadFile("./cfg.json", &gCfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%sFailed to read config file%s\n", MiscLib.ColorRed, MiscLib.ColorReset)
		}
		db_x := ConnectToAnyDb("postgres", gCfg.DbConn, gCfg.DbName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%sFailed to connect to database: %s%s\n", MiscLib.ColorRed, err, MiscLib.ColorReset)
			os.Exit(1)
		}

		ymux.DB = db_x.Db // data, err := SelData2(db_x.Db, *optQuery)
		DB = db_x.Db      // data, err := SelData2(db_x.Db, *optQuery)
	}
}

var db821 = false
var db822 = true
