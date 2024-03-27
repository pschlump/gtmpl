package sizlib

import (
	"database/sql"
)

// SQLQuery runs stmt and returns rows.
// func Run1(db *sql.DB, q string, arg ...interface{}) error {
func SQLQuery(db *sql.DB, stmt string, data ...interface{}) (resultSet *sql.Rows, err error) {
	// start := time.Now()
	//	stmt, data, _ = BindFixer(stmt, data)
	resultSet, err = db.Query(stmt, data...)
	// elapsed := time.Since(start)
	// logQueries(stmt, err, data, elapsed)
	return
}
