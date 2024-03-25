package main

// --old--
// --old-- import (
// --old-- 	"database/sql"
// --old-- 	"encoding/json"
// --old-- 	"fmt"
// --old-- 	"io/ioutil"
// --old-- 	"os"
// --old-- 	"testing"
// --old--
// --old-- 	"github.com/pschlump/MiscLib"
// --old-- 	"github.com/pschlump/dbgo"
// --old-- 	"git.q8s.co/pschlump/ReadConfig"
// --old-- 	"git.q8s.co/pschlump/piserver/ymux"
// --old-- )
// --old--
// --old-- // func PathFind(path, fn string) (full_path, file_type string, err error) {
// --old--
// --old-- func Test_FindPath(t *testing.T) {
// --old--
// --old-- 	fp, ex, err := PathFind("tmpl1;./testdata", "base-table.html")
// --old--
// --old-- 	if err != nil {
// --old-- 		t.Errorf("Error Got error %s when expecting success", err)
// --old-- 	}
// --old-- 	if ex != ".html" {
// --old-- 		t.Errorf("Expected .html got ->%s<-\n", ex)
// --old-- 	}
// --old-- 	if fp != "testdata/base-table.html" {
// --old-- 		t.Errorf("Expected ->testdata/base-table.html<- got ->%s<-\n", fp)
// --old-- 	}
// --old-- }
// --old--
// --old-- // func RenderTemplate(mdata map[string]interface{}, fns ...string) (tmpl_rendered string, err error) {
// --old-- func Test_RenderTemlate(t *testing.T) {
// --old--
// --old-- 	var mdata map[string]interface{}
// --old--
// --old-- 	json.Unmarshal([]byte(`{
// --old-- "data": [
// --old-- 	{
// --old-- 		  "id": "111222333444"
// --old-- 		, "original_file_name": "abc-def.xls"
// --old-- 	}
// --old-- ]
// --old-- }`), &mdata)
// --old--
// --old-- 	expect := `
// --old--
// --old-- 	<table class="table">
// --old-- 		<thead>
// --old-- 			<tr>
// --old--
// --old-- 			<th>
// --old-- 				Original File Name
// --old-- 			</th>
// --old-- 			<th>
// --old-- 				Action
// --old-- 			</th>
// --old--
// --old-- 			<tr>
// --old-- 		</thead>
// --old-- 		<tbody>
// --old--
// --old--
// --old-- 		<th>
// --old-- 			<td>
// --old-- 				abc-def.xls
// --old-- 			<td>
// --old-- 			<td>
// --old-- 				<button class="bind-this" data-id="111222333444" data-click-run="run-form">Run</button>
// --old-- 			<td>
// --old-- 		</th>
// --old--
// --old--
// --old-- 		</tbody>
// --old-- 	</table>
// --old--
// --old-- 	<button>Upload New File</button>
// --old--
// --old-- `
// --old--
// --old-- 	tv, err := RenderTemplate(mdata, "testdata/base-table.html", "testdata/list-of-files/lof.html")
// --old--
// --old-- 	if db821 {
// --old-- 		fmt.Printf("AT: %s Template ->%s<- error:%s\n", dbgo.LF(), tv, err)
// --old-- 	}
// --old--
// --old-- 	if err != nil {
// --old-- 		t.Errorf("Got error from render : %s\n", err)
// --old-- 	}
// --old--
// --old-- 	if tv != expect {
// --old-- 		ioutil.WriteFile(",a", []byte(expect), 0644)
// --old-- 		ioutil.WriteFile(",b", []byte(tv), 0644)
// --old-- 		t.Errorf("Expected ->%s<- got ->%s<-\n", expect, tv)
// --old-- 	}
// --old-- }
// --old--
// --old-- // func ProcessSQL(ds JsonTemplateRunnerType) (mdata map[string]interface{}, err error) {
// --old-- func Test_ProcessSQL(t *testing.T) {
// --old-- 	var ds JsonTemplateRunnerType
// --old--
// --old-- 	SetupDatabase()
// --old--
// --old-- 	/*
// --old-- 	   type DataType struct {
// --old-- 	   	To    string // Name of data item to place data in.
// --old-- 	   	Stmt  string // SQL ro run
// --old-- 	   	Bind  map[string]string
// --old-- 	   	ErrOn string // xyzzy - needs work
// --old-- 	   }
// --old--
// --old-- 	   type LayoutItemType struct {
// --old-- 	   	MatchTo     string `json:"match"`
// --old-- 	   	TemplateFor string `json:"for"`
// --old-- 	   }
// --old--
// --old-- 	   type TemplateSetType struct {
// --old-- 	   	TemplateList []string `json:"template"`
// --old-- 	   	Target       string   `json:"target"`
// --old-- 	   }
// --old--
// --old-- 	   type JsonTemplateRunnerType struct {
// --old-- 	   	TemplateList []string                   `json:"template"`
// --old-- 	   	JsonLayout   []LayoutItemType           `json:"jsonLayout"`
// --old-- 	   	TemplateSet  map[string]TemplateSetType `json:"TempalteSet"`
// --old-- 	   	Data         []DataType                 `json:"SelectData"`
// --old-- 	   	Test         map[string]interface{}     // xyzzy - TBD
// --old-- 	   }
// --old-- 	*/
// --old-- 	ds.Data = []DataType{
// --old-- 		{
// --old-- 			To:   "test1",
// --old-- 			Stmt: "select * from t_ymux_documents where user_id = $1",
// --old-- 			Bind: map[string]string{
// --old-- 				"$1": "user_id",
// --old-- 			},
// --old-- 		},
// --old-- 	}
// --old--
// --old-- 	data := map[string]string{
// --old-- 		"user_id": "52bc4522-bed8-4ee4-73b3-be0ed73d7f1f",
// --old-- 	}
// --old-- 	fx := func(s string) string {
// --old-- 		return data[s]
// --old-- 	}
// --old--
// --old-- 	mdata, err := ProcessSQL(&ds, fx)
// --old--
// --old-- 	if db822 {
// --old-- 		fmt.Printf("at:%s err:%s mdata=%s\n", dbgo.LF(), err, dbgo.SVarI(mdata))
// --old-- 	}
// --old--
// --old-- 	if _, ok := mdata["test1"]; !ok {
// --old-- 		t.Errorf("Expected some data back, did not get any")
// --old-- 	}
// --old--
// --old-- }
// --old--
// --old-- /*
// --old-- Expected Data
// --old-- 	{
// --old-- 		"test1": [
// --old-- 			{
// --old-- 				"blockerr": null,
// --old-- 				"blockhash": null,
// --old-- 				"blockno": null,
// --old-- 				"created": "2020-01-30T05:03:56.37554+0000",
// --old-- 				"document_file_name": null,
// --old-- 				"document_hash": null,
// --old-- 				"ethstatus": null,
// --old-- 				"file_name": "./www/files/3c0da6cd8cddf7e9d6c2ae649f5fd5ab5271cfa1cc67a6f238cd001f21728839.xls",
// --old-- 				"hash": null,
// --old-- 				"id": "17da237a-2ff2-41db-51c9-93932181bd5b",
// --old-- 				"note": null,
// --old-- 				"orig_file_extension": ".xls",
// --old-- 				"orig_file_name": "post-tx.xls",
// --old-- 				"signature": null,
// --old-- 				"txid": null,
// --old-- 				"updated": "2020-05-02T20:14:35.72913+0000",
// --old-- 				"url_file_name": "/files/3c0da6cd8cddf7e9d6c2ae649f5fd5ab5271cfa1cc67a6f238cd001f21728839.xls",
// --old-- 				"user_id": "52bc4522-bed8-4ee4-73b3-be0ed73d7f1f"
// --old-- 			},
// --old-- 			{
// --old-- 				"blockerr": null,
// --old-- 				"blockhash": null,
// --old-- 				"blockno": null,
// --old-- 				"created": "2020-01-31T15:25:19.13184+0000",
// --old-- 				"document_file_name": null,
// --old-- 				"document_hash": null,
// --old-- 				"ethstatus": null,
// --old-- 				"file_name": "./www/files/3c0da6cd8cddf7e9d6c2ae649f5fd5ab5271cfa1cc67a6f238cd001f21728839.xls",
// --old-- 				"hash": null,
// --old-- 				"id": "f98c32dd-00c2-4080-5af1-debe903d8a48",
// --old-- 				"note": null,
// --old-- 				"orig_file_extension": ".xls",
// --old-- 				"orig_file_name": "post-tx.xls",
// --old-- 				"signature": null,
// --old-- 				"txid": null,
// --old-- 				"updated": "2020-05-02T20:14:35.72913+0000",
// --old-- 				"url_file_name": "/files/3c0da6cd8cddf7e9d6c2ae649f5fd5ab5271cfa1cc67a6f238cd001f21728839.xls",
// --old-- 				"user_id": "52bc4522-bed8-4ee4-73b3-be0ed73d7f1f"
// --old-- 			},
// --old-- 			{
// --old-- 				"blockerr": null,
// --old-- 				"blockhash": null,
// --old-- 				"blockno": null,
// --old-- 				"created": "2020-05-07T19:48:10.77581+0000",
// --old-- 				"document_file_name": null,
// --old-- 				"document_hash": null,
// --old-- 				"ethstatus": null,
// --old-- 				"file_name": "./www/files/cbafec6c72cc6689c18d65835324b10ce3637ce5a8da5c4115c8d52013c9dcd3.xlsx",
// --old-- 				"hash": null,
// --old-- 				"id": "473e1222-d5e3-484f-5ff2-6477216cefc0",
// --old-- 				"note": null,
// --old-- 				"orig_file_extension": ".xlsx",
// --old-- 				"orig_file_name": "post-tx.xlsx",
// --old-- 				"signature": null,
// --old-- 				"txid": null,
// --old-- 				"updated": "2020-05-07T19:48:11.83381+0000",
// --old-- 				"url_file_name": "/files/cbafec6c72cc6689c18d65835324b10ce3637ce5a8da5c4115c8d52013c9dcd3.xlsx",
// --old-- 				"user_id": "52bc4522-bed8-4ee4-73b3-be0ed73d7f1f"
// --old-- 			}
// --old-- 		]
// --old-- 	}
// --old-- */
// --old--
// --old-- var dbInit = false
// --old--
// --old-- // DB is the connection info to the database.  It must be external to be used.
// --old-- var DB *sql.DB
// --old--
// --old-- func SetupDatabase() {
// --old--
// --old-- 	if !dbInit {
// --old-- 		dbInit = true
// --old--
// --old-- 		err := ReadConfig.ReadFile("./cfg.json", &gCfg)
// --old-- 		if err != nil {
// --old-- 			fmt.Fprintf(os.Stderr, "%sFailed to read config file%s\n", MiscLib.ColorRed, MiscLib.ColorReset)
// --old-- 		}
// --old-- 		db_x := ConnectToAnyDb("postgres", gCfg.DbConn, gCfg.DbName)
// --old-- 		if err != nil {
// --old-- 			fmt.Fprintf(os.Stderr, "%sFailed to connect to database: %s%s\n", MiscLib.ColorRed, err, MiscLib.ColorReset)
// --old-- 			os.Exit(1)
// --old-- 		}
// --old--
// --old-- 		ymux.DB = db_x.Db // data, err := SelData2(db_x.Db, *optQuery)
// --old-- 		DB = db_x.Db      // data, err := SelData2(db_x.Db, *optQuery)
// --old-- 	}
// --old-- }
// --old--
// --old-- func Test_ReadJsonTemplateConfigFile(t *testing.T) {
// --old--
// --old-- 	DbOn["db4a"] = true
// --old--
// --old-- 	// func ReadJsonTemplateConfigFile(fn string) (ds JsonTemplateRunnerType, err error) {
// --old-- 	ds, err := ReadJsonTemplateConfigFile("./testdata/testTemplateConfig1.json")
// --old-- 	if db823 {
// --old-- 		fmt.Printf("%s\n", dbgo.SVarI(ds))
// --old-- 	}
// --old-- 	if err != nil {
// --old-- 		t.Errorf("Error error: %s\n", err)
// --old-- 	}
// --old--
// --old-- 	expect := `{
// --old-- 	"Template": [
// --old-- 		"base-table.html",
// --old-- 		"lof.html"
// --old-- 	],
// --old-- 	"JsonLayout": null,
// --old-- 	"TemplateSet": null,
// --old-- 	"SelectData": [
// --old-- 		{
// --old-- 			"To": "test1",
// --old-- 			"Stmt": "select * from t_ymux_documents where user_id = $1",
// --old-- 			"Bind": {
// --old-- 				"$1": "user_id"
// --old-- 			},
// --old-- 			"ErrOn": ""
// --old-- 		}
// --old-- 	],
// --old-- 	"Test": null
// --old-- }`
// --old-- 	got := dbgo.SVarI(ds)
// --old-- 	if got != expect {
// --old-- 		ioutil.WriteFile(",c", []byte(expect), 0644)
// --old-- 		ioutil.WriteFile(",d", []byte(got), 0644)
// --old-- 		t.Errorf("Error Unexpected Results got ->%s<- expected ->%s<-\n", got, expect)
// --old-- 	}
// --old--
// --old-- 	ds, err = ReadJsonTemplateConfigFile("./testdata/page-cfg.json")
// --old-- 	if err != nil {
// --old-- 		t.Errorf("Error error: %s\n", err)
// --old-- 	}
// --old-- }
// --old--
// --old-- // ==============================================================================================================================
// --old--
// --old-- //func TmplProcess(
// --old-- //	item string, //  "page_name", "partial" etc.
// --old-- //	tmpl_name string, // .html/.tmpl file or .json file with data+selects+templates
// --old-- //	dataFunc func(name string) string,
// --old-- //) (tmpl_rendered string, status int, err error) {
// --old-- func Test_TmplProcess(t *testing.T) {
// --old--
// --old-- 	DbOn["db4a"] = true
// --old--
// --old-- 	data := map[string]string{
// --old-- 		"user_id": "52bc4522-bed8-4ee4-73b3-be0ed73d7f1f",
// --old-- 	}
// --old-- 	fx := func(s string) string {
// --old-- 		return data[s]
// --old-- 	}
// --old-- 	tmpl_rendered, status, err := TmplProcess("page", "page-cfg.json", fx)
// --old-- 	fmt.Printf("->%s<- %d %s\n", tmpl_rendered, status, err)
// --old-- }
// --old--
// --old-- // ==============================================================================================================================
// --old--
// --old-- var db821 = false
// --old-- var db822 = false
// --old-- var db823 = false
// --old-- var db824 = false
