package main

import (
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"time"

	_ "github.com/lib/pq"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/godebug"
	"github.com/pschlump/uuid"
)

type MyDb struct {
	Db     *sql.DB
	DbType string
}

var DbBeginQuote = `"`
var DbEndQuote = `"`

func ConnectToAnyDb(db_type string, auth string, dbName string) *MyDb {
	mm := &MyDb{DbType: db_type}

	switch db_type {
	case "postgres":
		DbBeginQuote = `"`
		DbEndQuote = `"`
	case "oracle":
		os.Setenv("NLS_LANG", "")
		DbBeginQuote = `"`
		DbEndQuote = `"`
		db_type = "oci8"
	case "odbc":
		DbBeginQuote = `[`
		DbEndQuote = `]`
	default:
		panic("Invalid database type.")
	}

	db, err := sql.Open(db_type, auth)

	//db, err := sql.Open("odbc", "DSN=T1; UID=sa; PWD=f1ref0x12" )	// ODBC to Microsoft SQL Server
	//db, err := sql.Open("mymysql", "test/philip/f1ref0x12")		// mySQL
	//db, err := sql.Open("oci8", "scott/tiger@//192.168.0.101:1521/orcl")

	if err != nil {
		panic(err)
	}

	mm.Db = db

	switch db_type {
	case "postgres":
		db.SetMaxIdleConns(5)
		// SET SCHEMA 'database_name'; -- Postgres way to set sechema to ...

	case "oci8":
		// set a default schema?? - or just use schema connected to?
		// No activity for now.

	case "odbc":
		err := Run1(db, "use "+dbName)
		if err != nil {
			fmt.Printf("Unable to set database, to %s, %s\n", dbName, err)
		}
	}

	return mm
}

// -------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------
func SelData2(db *sql.DB, q string, data ...interface{}) ([]map[string]interface{}, error) {
	// 1 use "sel" to do the query
	Rows, err := SelQ(db, q, data...)

	if err != nil {
		fmt.Printf("Params: %s\n", godebug.SVar(data))
		// godebug.IAmAt2( fmt.Sprintf ( "Error (%s)", err ) )
		return make([]map[string]interface{}, 0, 1), err
	}

	rv, _, n := RowsToInterface(Rows)

	_ = n
	// tr.TraceDbEnd("SelData", q, n)
	return rv, err
}

// -------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------
func SelQ(db *sql.DB, q string, data ...interface{}) (Rows *sql.Rows, err error) {
	//godebug.TraceDb2("SelQ", q, data...)
	//godebug.TrIAmAt2(fmt.Sprintf("Query (%s) with data:", q))
	//godebug.DumpVar(data)
	if len(data) == 0 {
		Rows, err = db.Query(q)
	} else {
		Rows, err = db.Query(q, data...)
	}
	if err != nil {
		// tr.TraceDbError2("SelQ", q, err)
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("Database error (%v) at %s:%d, query=%s\n", err, file, line, q)
	}
	return
}

// -------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------
func RowsToInterface(rows *sql.Rows) ([]map[string]interface{}, string, int) {

	var finalResult []map[string]interface{}
	var oneRow map[string]interface{}
	var id string

	id = ""

	if rows == nil {
		return nil, "", 0
	}

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error())
	}
	length := len(columns)

	// Make a slice for the values
	values := make([]interface{}, length)

	// rows.Scan wants '[]interface{}' as an argument, so we must copy the
	// references into such a slice
	// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
	scanArgs := make([]interface{}, length)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	// Fetch rows
	j := 0
	for rows.Next() {
		oneRow = make(map[string]interface{}, length)
		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}

		// Print data
		for i, value := range values {
			// fmt.Printf ( "at top i=%d %T\n", i, value )
			switch value.(type) {
			case nil:
				// fmt.Println("n, %s", columns[i], ": NULL", godebug.LF())
				oneRow[columns[i]] = nil

			case []byte:
				// fmt.Printf("[]byte, len = %d, %s\n", len(value.([]byte)), godebug.LF())
				// if len==16 && odbc - then - convert from UniversalIdentifier to string (UUID convert?)
				if len(value.([]byte)) == 16 {
					// var u *uuid.UUID
					//
					if uuid.IsUUID(fmt.Sprintf("%s", value.([]byte))) {
						u, err := uuid.Parse(value.([]byte))
						if err != nil {
							// fmt.Printf("Error: Invalid UUID parse, %s\n", godebug.LF())
							oneRow[columns[i]] = string(value.([]byte))
							if columns[i] == "id" && j == 0 {
								id = fmt.Sprintf("%s", value)
							}
						} else {
							if columns[i] == "id" && j == 0 {
								id = u.String()
							}
							oneRow[columns[i]] = u.String()
							// fmt.Printf(">>>>>>>>>>>>>>>>>> %s, %s\n", value, godebug.LF())
						}
					} else {
						if columns[i] == "id" && j == 0 {
							id = fmt.Sprintf("%s", value)
						}
						oneRow[columns[i]] = string(value.([]byte))
						// fmt.Printf(">>>>> 2 >>>>>>>>>>>>> %s, %s\n", value, godebug.LF())
					}
				} else {
					// Floats seem to end up at this point - xyzzy - instead of float64 -- so....  Need to check our column type info and see if 'f'  ---- xyzzy
					// fmt.Println("s", columns[i], ": ", string(value.([]byte)))
					if columns[i] == "id" && j == 0 {
						id = fmt.Sprintf("%s", value)
					}
					oneRow[columns[i]] = string(value.([]byte))
				}

			case int64:
				// fmt.Println("i, %s", columns[i], ": ", value, godebug.LF())
				// oneRow[columns[i]] = fmt.Sprintf ( "%v", value )	// PJS-2014-03-06 - I suspect that this is a defect
				oneRow[columns[i]] = value

			case float64:
				// fmt.Println("f, %s", columns[i], ": ", value, godebug.LF())
				// oneRow[columns[i]] = fmt.Sprintf ( "%v", value )
				// fmt.Printf ( "yes it is a float\n" )
				oneRow[columns[i]] = value

			case bool:
				// fmt.Println("b, %s", columns[i], ": ", value, godebug.LF())
				// oneRow[columns[i]] = fmt.Sprintf ( "%v", value )		// PJS-2014-03-06
				// oneRow[columns[i]] = fmt.Sprintf ( "%t", value )		"true" or "false" as a value
				oneRow[columns[i]] = value

			case string:
				// fmt.Printf("string, %s\n", godebug.LF())
				if columns[i] == "id" && j == 0 {
					id = fmt.Sprintf("%s", value)
				}
				// fmt.Println("S", columns[i], ": ", value)
				oneRow[columns[i]] = fmt.Sprintf("%s", value)

			// Xyzzy - there is a timeNull structure in the driver - why is that not returned?  Maybee it is????
			// oneRow[columns[i]] = nil
			case time.Time:
				oneRow[columns[i]] = (value.(time.Time)).Format(ISO8601output)

			default:
				fmt.Printf("%s--- In default Case [%s] - %T %s\n", MiscLib.ColorRed, godebug.LF(), value, MiscLib.ColorReset)
				fmt.Fprintf(os.Stderr, "%s--- In default Case [%s] - %T %s\n", MiscLib.ColorRed, godebug.LF(), value, MiscLib.ColorReset)
				// fmt.Printf ( "default, yes it is a... , i=%d, %T\n", i, value, godebug.LF() )
				// fmt.Println("r", columns[i], ": ", value)
				if columns[i] == "id" && j == 0 {
					id = fmt.Sprintf("%v", value)
				}
				oneRow[columns[i]] = fmt.Sprintf("%v", value)
			}
			//fmt.Printf("\nType: %s\n", reflect.TypeOf(value))
		}
		// fmt.Println("-----------------------------------")
		finalResult = append(finalResult, oneRow)
		j++
	}
	return finalResult, id, j
}

// -------------------------------------------------------------------------------------------------
// test: t-run1q.go, .sql, .out
// -------------------------------------------------------------------------------------------------
func Run1(db *sql.DB, q string, arg ...interface{}) error {
	stmt, err := db.Prepare(q)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(arg...)
	if err != nil {
		return err
	}

	return nil
}

// ISO format for date
const ISO8601 = "2006-01-02T15:04:05.99999Z07:00"

// ISO format for date
const ISO8601output = "2006-01-02T15:04:05.99999-0700"
