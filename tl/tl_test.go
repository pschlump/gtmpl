package tl

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

type ConfigFile struct {
	Name   string `json:"name"`
	DbConn string `json:"dbconn" default:"user=postgres dbname=postgres port=5432 host=127.0.0.1 sslmode=disable"`
	DbName string `json:"dbname" default:"postgres"`
}

var gCfg ConfigFile

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

	if _, ok := mdata["test1"]; !ok {
		t.Errorf("Expected some data back, did not get any")
	}

}

/*
Expected Data
	{
		"test1": [
			{
				"blockerr": null,
				"blockhash": null,
				"blockno": null,
				"created": "2020-01-30T05:03:56.37554+0000",
				"document_file_name": null,
				"document_hash": null,
				"ethstatus": null,
				"file_name": "./www/files/3c0da6cd8cddf7e9d6c2ae649f5fd5ab5271cfa1cc67a6f238cd001f21728839.xls",
				"hash": null,
				"id": "17da237a-2ff2-41db-51c9-93932181bd5b",
				"note": null,
				"orig_file_extension": ".xls",
				"orig_file_name": "post-tx.xls",
				"signature": null,
				"txid": null,
				"updated": "2020-05-02T20:14:35.72913+0000",
				"url_file_name": "/files/3c0da6cd8cddf7e9d6c2ae649f5fd5ab5271cfa1cc67a6f238cd001f21728839.xls",
				"user_id": "52bc4522-bed8-4ee4-73b3-be0ed73d7f1f"
			},
			{
				"blockerr": null,
				"blockhash": null,
				"blockno": null,
				"created": "2020-01-31T15:25:19.13184+0000",
				"document_file_name": null,
				"document_hash": null,
				"ethstatus": null,
				"file_name": "./www/files/3c0da6cd8cddf7e9d6c2ae649f5fd5ab5271cfa1cc67a6f238cd001f21728839.xls",
				"hash": null,
				"id": "f98c32dd-00c2-4080-5af1-debe903d8a48",
				"note": null,
				"orig_file_extension": ".xls",
				"orig_file_name": "post-tx.xls",
				"signature": null,
				"txid": null,
				"updated": "2020-05-02T20:14:35.72913+0000",
				"url_file_name": "/files/3c0da6cd8cddf7e9d6c2ae649f5fd5ab5271cfa1cc67a6f238cd001f21728839.xls",
				"user_id": "52bc4522-bed8-4ee4-73b3-be0ed73d7f1f"
			},
			{
				"blockerr": null,
				"blockhash": null,
				"blockno": null,
				"created": "2020-05-07T19:48:10.77581+0000",
				"document_file_name": null,
				"document_hash": null,
				"ethstatus": null,
				"file_name": "./www/files/cbafec6c72cc6689c18d65835324b10ce3637ce5a8da5c4115c8d52013c9dcd3.xlsx",
				"hash": null,
				"id": "473e1222-d5e3-484f-5ff2-6477216cefc0",
				"note": null,
				"orig_file_extension": ".xlsx",
				"orig_file_name": "post-tx.xlsx",
				"signature": null,
				"txid": null,
				"updated": "2020-05-07T19:48:11.83381+0000",
				"url_file_name": "/files/cbafec6c72cc6689c18d65835324b10ce3637ce5a8da5c4115c8d52013c9dcd3.xlsx",
				"user_id": "52bc4522-bed8-4ee4-73b3-be0ed73d7f1f"
			}
		]
	}
*/

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

func Test_ReadJsonTemplateConfigFile(t *testing.T) {

	DbOn["db4a"] = true

	// func ReadJsonTemplateConfigFile(fn string) (ds JsonTemplateRunnerType, err error) {
	ds, err := ReadJsonTemplateConfigFile("./testdata/testTemplateConfig1.json")
	if db823 {
		fmt.Printf("%s\n", godebug.SVarI(ds))
	}
	if err != nil {
		t.Errorf("Error error: %s\n", err)
	}

	expect := `{
	"Template": [
		"base-table.html",
		"lof.html"
	],
	"JsonLayout": null,
	"TemplateSet": null,
	"SelectData": [
		{
			"To": "test1",
			"Stmt": "select * from t_ymux_documents where user_id = $1",
			"Bind": {
				"$1": "user_id"
			},
			"ErrOn": ""
		}
	],
	"Test": null
}`
	got := godebug.SVarI(ds)
	if got != expect {
		ioutil.WriteFile(",c", []byte(expect), 0644)
		ioutil.WriteFile(",d", []byte(got), 0644)
		t.Errorf("Error Unexpected Results got ->%s<- expected ->%s<-\n", got, expect)
	}

	ds, err = ReadJsonTemplateConfigFile("./testdata/page-cfg.json")
	if err != nil {
		t.Errorf("Error error: %s\n", err)
	}
}

// ==============================================================================================================================

//func TmplProcess(
//	item string, //  "page_name", "partial" etc.
//	tmpl_name string, // .html/.tmpl file or .json file with data+selects+templates
//	dataFunc func(name string) string,
//) (tmpl_rendered string, status int, err error) {
func Test_TmplProcess(t *testing.T) {

	DbOn["db4a"] = true

	data := map[string]string{
		"user_id": "52bc4522-bed8-4ee4-73b3-be0ed73d7f1f",
	}
	fx := func(s string) string {
		return data[s]
	}
	tmpl_rendered, status, err := TmplProcess("page", "page-cfg.json", fx)
	fmt.Printf("->%s<- %d %s\n", tmpl_rendered, status, err)
}

// ==============================================================================================================================

var db821 = false
var db822 = false
var db823 = false
var db824 = false
