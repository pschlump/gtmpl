package sizlib

// (C) Copyright Philip Schlump, 2013-2018

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pschlump/HashStr"
	"github.com/pschlump/MiscLib"
	"github.com/pschlump/dbgo"
	"github.com/pschlump/uuid"
)

var ctx = context.Background()

// -------------------------------------------------------------------------------------------------
func ConnectToDb(auth string) *pgx.Conn {
	// conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	conn, err := pgx.Connect(ctx, auth)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	return conn
}

// -------------------------------------------------------------------------------------------------
func GetColumns(rows pgx.Rows) (columns []string, err error) {
	// var fd []pgx.FieldDescription
	fd := rows.FieldDescriptions()
	columns = make([]string, 0, len(fd))
	for _, vv := range fd {
		columns = append(columns, vv.Name)
	}
	return
}

// -------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------
func RowsToInterface(rows pgx.Rows) ([]map[string]interface{}, string, int) {
	// func RowsToInterface(rows *sql.Rows) ([]map[string]interface{}, string, int) {

	var finalResult []map[string]interface{}
	var oneRow map[string]interface{}
	var id string

	id = ""

	if rows == nil {
		return nil, "", 0
	}

	// Get column names
	// columns, err := rows.Columns()
	columns, err := GetColumns(rows)
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
				// fmt.Println("n, %s", columns[i], ": NULL", dbgo.LF())
				oneRow[columns[i]] = nil

			case []byte:
				// fmt.Printf("[]byte, len = %d, %s\n", len(value.([]byte)), dbgo.LF())
				// if len==16 && odbc - then - convert from UniversalIdentifier to string (UUID convert?)
				if len(value.([]byte)) == 16 {
					// var u *uuid.UUID
					//
					if uuid.IsUUID(fmt.Sprintf("%s", value.([]byte))) {
						u, err := uuid.Parse(value.([]byte))
						if err != nil {
							// fmt.Printf("Error: Invalid UUID parse, %s\n", dbgo.LF())
							oneRow[columns[i]] = string(value.([]byte))
							if columns[i] == "id" && j == 0 {
								id = fmt.Sprintf("%s", value)
							}
						} else {
							if columns[i] == "id" && j == 0 {
								id = u.String()
							}
							oneRow[columns[i]] = u.String()
							// fmt.Printf(">>>>>>>>>>>>>>>>>> %s, %s\n", value, dbgo.LF())
						}
					} else {
						if columns[i] == "id" && j == 0 {
							id = fmt.Sprintf("%s", value)
						}
						oneRow[columns[i]] = string(value.([]byte))
						// fmt.Printf(">>>>> 2 >>>>>>>>>>>>> %s, %s\n", value, dbgo.LF())
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
				// fmt.Println("i, %s", columns[i], ": ", value, dbgo.LF())
				// oneRow[columns[i]] = fmt.Sprintf ( "%v", value )	// PJS-2014-03-06 - I suspect that this is a defect
				oneRow[columns[i]] = value

			case int32:
				// fmt.Println("i, %s", columns[i], ": ", value, dbgo.LF())
				// oneRow[columns[i]] = fmt.Sprintf ( "%v", value )	// PJS-2014-03-06 - I suspect that this is a defect
				oneRow[columns[i]] = int(value.(int32))

			case float64:
				// fmt.Println("f, %s", columns[i], ": ", value, dbgo.LF())
				// oneRow[columns[i]] = fmt.Sprintf ( "%v", value )
				// fmt.Printf ( "yes it is a float\n" )
				oneRow[columns[i]] = value

			case bool:
				// fmt.Println("b, %s", columns[i], ": ", value, dbgo.LF())
				// oneRow[columns[i]] = fmt.Sprintf ( "%v", value )		// PJS-2014-03-06
				// oneRow[columns[i]] = fmt.Sprintf ( "%t", value )		"true" or "false" as a value
				oneRow[columns[i]] = value

			case string:
				// fmt.Printf("string, %s\n", dbgo.LF())
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
				fmt.Printf("%s--- In default Case [%s] - %T %s\n", MiscLib.ColorRed, dbgo.LF(), value, MiscLib.ColorReset)
				fmt.Fprintf(os.Stderr, "%s--- In default Case [%s] - %T %s\n", MiscLib.ColorRed, dbgo.LF(), value, MiscLib.ColorReset)
				// fmt.Printf ( "default, yes it is a... , i=%d, %T\n", i, value, dbgo.LF() )
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
// -------------------------------------------------------------------------------------------------
func SelQ(db *pgx.Conn, q string, data ...interface{}) (Rows pgx.Rows, err error) {
	//func SelQ(db *sql.DB, q string, data ...interface{}) (Rows *sql.Rows, err error) {
	if len(data) == 0 {
		Rows, err = db.Query(ctx, q)
	} else {
		Rows, err = db.Query(ctx, q, data...)
	}
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("Database error (%v) at %s:%d, query=%s\n", err, file, line, q)
	}
	return
}

// -------------------------------------------------------------------------------------------------
// -------------------------------------------------------------------------------------------------
func SelData2(db *pgx.Conn, q string, data ...interface{}) ([]map[string]interface{}, error) {
	// func SelData2(db *sql.DB, q string, data ...interface{}) ([]map[string]interface{}, error) {
	// 1 use "sel" to do the query
	// func sel ( res http.ResponseWriter, req *http.Request, db *pgx.Conn, q string, data ...interface{} ) ( Rows *sql.Rows, err error ) {
	Rows, err := SelQ(db, q, data...)

	if err != nil {
		fmt.Printf("Params: %s\n", SVar(data))
		// dbgo.IAmAt2( fmt.Sprintf ( "Error (%s)", err ) )
		return make([]map[string]interface{}, 0, 1), err
	}

	defer Rows.Close()

	rv, _, n := RowsToInterface(Rows)

	_ = n
	return rv, err
}

// SelData seelct data from the database and return it.
func SelData(db *pgx.Conn, q string, data ...interface{}) []map[string]interface{} {
	// func SelData(db *sql.DB, q string, data ...interface{}) []map[string]interface{} {
	// 1 use "sel" to do the query
	// func sel ( res http.ResponseWriter, req *http.Request, db *pgx.Conn, q string, data ...interface{} ) ( Rows *sql.Rows, err error ) {
	// fmt.Printf("in SelData, %s\n", dbgo.LF())

	Rows, err := SelQ(db, q, data...)

	if err != nil {
		fmt.Printf("Params: %s\n", SVar(data))
		return make([]map[string]interface{}, 0, 1)
	}

	defer Rows.Close()

	rv, _, n := RowsToInterface(Rows)
	_ = n

	return rv
}

// -------------------------------------------------------------------------------------------------
// test: t-run1q.go, .sql, .out
// -------------------------------------------------------------------------------------------------
func Run1(db *pgx.Conn, q string, arg ...interface{}) error {
	h := HashStr.HashStrToName(q) + q
	ps, err := db.Prepare(ctx, h, q)
	if err != nil {
		return err
	}
	_ = ps

	_, err = db.Exec(ctx, h, arg...)
	if err != nil {
		return err
	}
	return nil
}

func Run2(db *pgx.Conn, q string, arg ...interface{}) (nr int64, err error) {
	nr = 0
	err = nil

	h := HashStr.HashStrToName(q) + q

	ps, err := db.Prepare(ctx, h, q)
	if err != nil {
		return
	}
	_ = ps

	R, err := db.Exec(ctx, h, arg...)
	if err != nil {
		return
	}

	nr = R.RowsAffected()
	return
}

// -------------------------------------------------------------------------------------------------
func InsUpd(db *pgx.Conn, ins string, upd string, mdata map[string]string) {
	// func InsUpd(db *sql.DB, ins string, upd string, mdata map[string]string) {
	ins_q := Qt(ins, mdata)
	// fmt.Printf("     insUpd(ins) %s\n", ins_q)
	err := Run1(db, ins_q)
	if err != nil {
		// fmt.Printf("Error (1) in insUpd = %s\n", err)
		upd_q := Qt(upd, mdata)
		// fmt.Printf("     insUpd(upd) %s\n", upd_q)
		err = Run1(db, upd_q)
		if err != nil {
			fmt.Printf("Error (2) in insUpd = %s\n", err)
		}
	}
}

// -------------------------------------------------------------------------------------------------
// xyzzy-Rewrite
//
//	mdata["group_id"] = insSel ( "select \"id\" from \"img_group\" where \"group_name\" = '%{user_id%}'",
//
// -------------------------------------------------------------------------------------------------
func InsSel(db *pgx.Conn, sel string, ins string, mdata map[string]string) (id string) {
	// func InsSel(db *sql.DB, sel string, ins string, mdata map[string]string) (id string) {

	id = ""
	q := Qt(sel, mdata)

	Rows, err := db.Query(ctx, q)
	if err != nil {
		fmt.Printf("Error (237) on talking to database, %s\n", err)
		return
	}
	defer Rows.Close()

	var x_id string
	n_row := 0
	for Rows.Next() {
		//  fmt.Printf ("Inside Rows Next\n" );
		n_row++
		err = Rows.Scan(&x_id)
		if err != nil {
			fmt.Printf("Error (249) on retreiving row from database, %s\n", err)
			return
		}
	}
	if n_row > 1 {
		fmt.Printf("Error (260) too many rows returned, n_rows=%d\n", n_row)
		return
	}
	if n_row == 1 {
		id = x_id
		return
	}

	y_id, _ := uuid.NewV4()
	id = y_id.String()
	mdata["id"] = id

	q = Qt(ins, mdata)

	Run1(db, q)
	return
}

// -------------------------------------------------------------------------------------------------
// Rows to JSON -- Go from a set of "rows" returned by db.Query to a JSON string.
// -------------------------------------------------------------------------------------------------
func RowsToJson(rows pgx.Rows) (string, string) {

	var finalResult []map[string]interface{}
	var oneRow map[string]interface{}
	var id string

	id = ""

	// Get column names
	// columns, err := rows.Columns()
	columns, err := GetColumns(rows)
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
				// fmt.Println("n", columns[i], ": NULL")
				oneRow[columns[i]] = nil

			case []byte:
				// Floats seem to end up at this point - xyzzy - instead of float64 -- so....  Need to check our column type info and see if 'f'  ---- xyzzy
				// fmt.Println("s", columns[i], ": ", string(value.([]byte)))
				if columns[i] == "id" && j == 0 {
					id = fmt.Sprintf("%s", value)
				}
				oneRow[columns[i]] = string(value.([]byte))

			case int64:
				// fmt.Println("i", columns[i], ": ", value)
				// oneRow[columns[i]] = fmt.Sprintf ( "%v", value )	// PJS-2014-03-06 - I suspect that this is a defect
				oneRow[columns[i]] = value

			case float64:
				//fmt.Println("f", columns[i], ": ", value)
				// oneRow[columns[i]] = fmt.Sprintf ( "%v", value )
				// fmt.Printf ( "yes it is a float\n" )
				oneRow[columns[i]] = value

			case bool:
				//fmt.Println("b", columns[i], ": ", value)
				// oneRow[columns[i]] = fmt.Sprintf ( "%v", value )		// PJS-2014-03-06
				// oneRow[columns[i]] = fmt.Sprintf ( "%t", value )		"true" or "false" as a value
				oneRow[columns[i]] = value

			case string:
				if columns[i] == "id" && j == 0 {
					id = fmt.Sprintf("%s", value)
				}
				// fmt.Println("S", columns[i], ": ", value)
				oneRow[columns[i]] = fmt.Sprintf("%s", value)

			// Xyzzy - there is a timeNull structure in the driver - why is that not returned?  Maybee it is????
			case time.Time:
				//fmt.Printf("time.Time - %s, %s\n", columns[i], dbgo.LF())
				//oneRow[columns[i]] = value
				oneRow[columns[i]] = (value.(time.Time)).Format(ISO8601output)

			default:
				// fmt.Printf ( "default, yes it is a... , i=%d, %T\n", i, value )
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
	if j > 0 {
		s, err := json.MarshalIndent(finalResult, "", "\t")
		if err != nil {
			fmt.Printf("Unable to convert to JSON data, %v\n", err)
		}
		return string(s), id
	} else {
		return "[]", ""
	}
}

// -------------------------------------------------------------------------------------------------
// Rows to JSON -- Go from a set of "rows" returned by db.Query to a JSON string.
// -------------------------------------------------------------------------------------------------
func RowsToJsonFirstRow(rows pgx.Rows) (string, string) {
	// func RowsToJsonFirstRow(rows *sql.Rows) (string, string) {

	// var finalResult   []map[string]interface{}
	var oneRow map[string]interface{}
	var id string

	id = ""

	// Get column names
	// columns, err := rows.Columns()
	columns, err := GetColumns(rows)
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

		if j == 0 {

			// Print data
			for i, value := range values {
				// fmt.Printf ( "at top i=%d %T\n", i, value )
				switch value.(type) {
				case nil:
					// fmt.Println("n", columns[i], ": NULL")
					oneRow[columns[i]] = nil

				case []byte:
					// Floats seem to end up at this point - xyzzy - instead of float64 -- so....  Need to check our column type info and see if 'f'  ---- xyzzy
					// fmt.Println("s", columns[i], ": ", string(value.([]byte)))
					if columns[i] == "id" && j == 0 {
						id = fmt.Sprintf("%s", value)
					}
					oneRow[columns[i]] = string(value.([]byte))

				case int64:
					// fmt.Println("i", columns[i], ": ", value)
					// oneRow[columns[i]] = fmt.Sprintf ( "%v", value )	// PJS-2014-03-06 - I suspect that this is a defect
					oneRow[columns[i]] = value

				case float64:
					//fmt.Println("f", columns[i], ": ", value)
					// oneRow[columns[i]] = fmt.Sprintf ( "%v", value )
					// fmt.Printf ( "yes it is a float\n" )
					oneRow[columns[i]] = value

				case bool:
					//fmt.Println("b", columns[i], ": ", value)
					// oneRow[columns[i]] = fmt.Sprintf ( "%v", value )		// PJS-2014-03-06
					// oneRow[columns[i]] = fmt.Sprintf ( "%t", value )		"true" or "false" as a value
					oneRow[columns[i]] = value

				case string:
					if columns[i] == "id" && j == 0 {
						id = fmt.Sprintf("%s", value)
					}
					// fmt.Println("S", columns[i], ": ", value)
					oneRow[columns[i]] = fmt.Sprintf("%s", value)

				// Xyzzy - there is a timeNull structure in the driver - why is that not returned?  Maybee it is????
				case time.Time:
					oneRow[columns[i]] = (value.(time.Time)).Format(ISO8601output)

				default:
					// fmt.Printf ( "default, yes it is a... , i=%d, %T\n", i, value )
					// fmt.Println("r", columns[i], ": ", value)
					if columns[i] == "id" && j == 0 {
						id = fmt.Sprintf("%v", value)
					}
					oneRow[columns[i]] = fmt.Sprintf("%v", value)
				}
				//fmt.Printf("\nType: %s\n", reflect.TypeOf(value))
			}
		}
		// fmt.Println("-----------------------------------")
		// finalResult = append ( finalResult, oneRow )
		j++
	}
	if j > 0 {
		s, err := json.MarshalIndent(oneRow, "", "\t")
		if err != nil {
			fmt.Printf("Unable to convert to JSON data, %v\n", err)
		}
		return string(s), id
	} else {
		return "{}", ""
	}
}
