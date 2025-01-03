package drivers

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	// import sqlite driver
	_ "modernc.org/sqlite"

	"github.com/jorgerojas26/lazysql/models"
)

type SQLite struct {
	Connection *sql.DB
	Provider   string
}

func (db *SQLite) TestConnection(urlstr string) (err error) {
	return db.Connect(urlstr)
}

func (db *SQLite) Connect(urlstr string) (err error) {
	db.SetProvider(DriverSqlite)

	db.Connection, err = sql.Open("sqlite", urlstr)
	if err != nil {
		return err
	}

	err = db.Connection.Ping()
	if err != nil {
		return err
	}

	return nil
}

func (db *SQLite) GetDatabases() ([]string, error) {
	var databases []string

	rows, err := db.Connection.Query("SELECT file FROM pragma_database_list WHERE name='main'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var database string
		err := rows.Scan(&database)
		if err != nil {
			return nil, err
		}

		split := strings.Split(database, "/")
		dbName := split[len(split)-1]

		databases = append(databases, dbName)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return databases, nil
}

func (db *SQLite) GetTables(database string) (map[string][]string, error) {
	if database == "" {
		return nil, errors.New("database name is required")
	}

	rows, err := db.Connection.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make(map[string][]string)

	for rows.Next() {
		var table string
		err = rows.Scan(&table)
		if err != nil {
			return nil, err
		}

		tables[database] = append(tables[database], table)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

func (db *SQLite) GetTableColumns(_, table string) (results [][]string, err error) {
	if table == "" {
		return nil, errors.New("table name is required")
	}

	rows, err := db.Connection.Query(fmt.Sprintf("PRAGMA table_info(%s)", db.formatTableName(table)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results = append(results, columns[1:])

	for rows.Next() {
		rowValues := make([]interface{}, len(columns))
		for i := range columns {
			rowValues[i] = new(sql.RawBytes)
		}

		err = rows.Scan(rowValues...)
		if err != nil {
			return nil, err
		}

		var row []string

		for _, col := range rowValues {
			if col == nil {
				row = append(row, "NULL")
			} else {
				row = append(row, string(*col.(*sql.RawBytes)))
			}
		}

		results = append(results, row[1:])
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return
}

func (db *SQLite) GetConstraints(_, table string) (results [][]string, err error) {
	if table == "" {
		return nil, errors.New("table name is required")
	}

	query := "SELECT sql FROM sqlite_master "
	query += "WHERE type='table' AND name = ?"

	rows, err := db.Connection.Query(query, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results = append(results, columns)

	for rows.Next() {
		rowValues := make([]interface{}, len(columns))
		for i := range columns {
			rowValues[i] = new(sql.RawBytes)
		}

		err = rows.Scan(rowValues...)
		if err != nil {
			return nil, err
		}

		var row []string
		for _, col := range rowValues {
			if col == nil {
				row = append(row, "NULL")
			} else {
				row = append(row, string(*col.(*sql.RawBytes)))
			}
		}

		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return
}

func (db *SQLite) GetForeignKeys(_, table string) (results [][]string, err error) {
	if table == "" {
		return nil, errors.New("table name is required")
	}

	rows, err := db.Connection.Query("PRAGMA foreign_key_list(" + table + ")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results = append(results, columns)

	for rows.Next() {
		rowValues := make([]interface{}, len(columns))
		for i := range columns {
			rowValues[i] = new(sql.RawBytes)
		}

		err = rows.Scan(rowValues...)
		if err != nil {
			return nil, err
		}

		var row []string
		for _, col := range rowValues {
			if col == nil {
				row = append(row, "NULL")
			} else {
				row = append(row, string(*col.(*sql.RawBytes)))
			}
		}

		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return
}

func (db *SQLite) GetIndexes(_, table string) (results [][]string, err error) {
	if table == "" {
		return nil, errors.New("table name is required")
	}

	rows, err := db.Connection.Query("PRAGMA index_list(" + table + ")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results = append(results, columns)

	for rows.Next() {
		rowValues := make([]interface{}, len(columns))
		for i := range columns {
			rowValues[i] = new(sql.RawBytes)
		}

		err = rows.Scan(rowValues...)
		if err != nil {
			return nil, err
		}

		var row []string
		for _, col := range rowValues {
			if col == nil {
				row = append(row, "NULL")
			} else {
				row = append(row, string(*col.(*sql.RawBytes)))
			}
		}

		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return
}

func (db *SQLite) GetRecords(_, table, where, sort string, offset, limit int) (paginatedResults [][]string, totalRecords int, err error) {
	if table == "" {
		return nil, 0, errors.New("table name is required")
	}

	if limit == 0 {
		limit = DefaultRowLimit
	}

	query := "SELECT * FROM "
	query += db.formatTableName(table)

	if where != "" {
		query += fmt.Sprintf(" %s", where)
	}

	if sort != "" {
		query += fmt.Sprintf(" ORDER BY %s", sort)
	}

	query += " LIMIT ?, ?"

	paginatedRows, err := db.Connection.Query(query, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	defer paginatedRows.Close()

	columns, err := paginatedRows.Columns()
	if err != nil {
		return nil, 0, err
	}

	paginatedResults = append(paginatedResults, columns)

	for paginatedRows.Next() {
		nullStringSlice := make([]sql.NullString, len(columns))

		rowValues := make([]interface{}, len(columns))

		for i := range nullStringSlice {
			rowValues[i] = &nullStringSlice[i]
		}

		err = paginatedRows.Scan(rowValues...)
		if err != nil {
			return nil, 0, err
		}

		var row []string
		for _, col := range nullStringSlice {
			if col.Valid {
				if col.String == "" {
					row = append(row, "EMPTY&")
				} else {
					row = append(row, col.String)
				}
			} else {
				row = append(row, "NULL&")
			}
		}

		paginatedResults = append(paginatedResults, row)
	}
	if err := paginatedRows.Err(); err != nil {
		return nil, 0, err
	}
	// close to release the connection
	if err := paginatedRows.Close(); err != nil {
		return nil, 0, err
	}

	countQuery := "SELECT COUNT(*) FROM "
	countQuery += db.formatTableName(table)
	row := db.Connection.QueryRow(countQuery)
	if err := row.Scan(&totalRecords); err != nil {
		return nil, 0, err
	}

	return
}

func (db *SQLite) ExecuteQuery(query string) (results [][]string, err error) {
	rows, err := db.Connection.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results = append(results, columns)

	for rows.Next() {
		rowValues := make([]interface{}, len(columns))
		for i := range columns {
			rowValues[i] = new(sql.RawBytes)
		}

		err = rows.Scan(rowValues...)
		if err != nil {
			return nil, err
		}

		var row []string
		for _, col := range rowValues {
			if col == nil {
				row = append(row, "NULL")
			} else {
				row = append(row, string(*col.(*sql.RawBytes)))
			}
		}

		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return
}

func (db *SQLite) UpdateRecord(_, table, column, value, primaryKeyColumnName, primaryKeyValue string) error {
	if table == "" {
		return errors.New("table name is required")
	}

	if column == "" {
		return errors.New("column name is required")
	}

	if value == "" {
		return errors.New("value is required")
	}

	if primaryKeyColumnName == "" {
		return errors.New("primary key column name is required")
	}

	if primaryKeyValue == "" {
		return errors.New("primary key value is required")
	}

	query := "UPDATE "
	query += db.formatTableName(table)
	query += fmt.Sprintf(" SET %s = ? WHERE %s = ?", column, primaryKeyColumnName)

	_, err := db.Connection.Exec(query, value, primaryKeyValue)

	return err
}

func (db *SQLite) DeleteRecord(_, table, primaryKeyColumnName, primaryKeyValue string) error {
	if table == "" {
		return errors.New("table name is required")
	}

	if primaryKeyColumnName == "" {
		return errors.New("primary key column name is required")
	}

	if primaryKeyValue == "" {
		return errors.New("primary key value is required")
	}

	query := "DELETE FROM "
	query += db.formatTableName(table)
	query += fmt.Sprintf(" WHERE %s = ?", primaryKeyColumnName)

	_, err := db.Connection.Exec(query)

	return err
}

func (db *SQLite) ExecuteDMLStatement(query string) (result string, err error) {
	res, err := db.Connection.Exec(query)
	if err != nil {
		return "", err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d rows affected", rowsAffected), nil
}

func (db *SQLite) ExecutePendingChanges(changes []models.DBDMLChange) (err error) {
	var queries []models.Query

	for _, change := range changes {
		columnNames := []string{}
		values := []interface{}{}
		valuesPlaceholder := []string{}

		switch change.Type {
		case models.DMLInsertType:
			for _, cell := range change.Values {
				switch cell.Type {
				case models.Null:
					columnNames = append(columnNames, cell.Column)
					valuesPlaceholder = append(valuesPlaceholder, "NULL")
				case models.Empty, models.String:
					columnNames = append(columnNames, cell.Column)
					valuesPlaceholder = append(valuesPlaceholder, "?")
				}
			}

			for _, cell := range change.Values {
				switch cell.Type {
				case models.Empty:
					values = append(values, "")
				case models.String:
					values = append(values, cell.Value)
				}
			}

			queryStr := "INSERT INTO "
			queryStr += db.formatTableName(change.Table)
			queryStr += fmt.Sprintf(" (%s) VALUES (%s)", strings.Join(columnNames, ", "), strings.Join(valuesPlaceholder, ", "))

			newQuery := models.Query{
				Query: queryStr,
				Args:  values,
			}

			queries = append(queries, newQuery)
		case models.DMLUpdateType:

			for _, cell := range change.Values {
				switch cell.Type {
				case models.Null:
					columnNames = append(columnNames, cell.Column)
					valuesPlaceholder = append(valuesPlaceholder, "NULL")
				case models.Empty, models.String:
					columnNames = append(columnNames, cell.Column)
					valuesPlaceholder = append(valuesPlaceholder, "?")
					/// Leaves "DEFAULT" type out because it's not supported by sqlite
				}
			}

			for _, cell := range change.Values {
				switch cell.Type {
				case models.Empty:
					values = append(values, "")
				case models.String:
					values = append(values, cell.Value)
				}
			}

			queryStr := "UPDATE "
			queryStr += db.formatTableName(change.Table)

			for i, column := range columnNames {
				if i == 0 {
					queryStr += fmt.Sprintf(" SET `%s` = %s", column, valuesPlaceholder[i])
				} else {
					queryStr += fmt.Sprintf(", `%s` = %s", column, valuesPlaceholder[i])
				}
			}

			args := make([]interface{}, len(values))

			copy(args, values)

			for i, pki := range change.PrimaryKeyInfo {
				if i == 0 {
					queryStr += fmt.Sprintf(" WHERE `%s` = ?", pki.Name)
				} else {
					queryStr += fmt.Sprintf(" AND `%s` = ?", pki.Name)
				}
				args = append(args, pki.Value)
			}

			newQuery := models.Query{
				Query: queryStr,
				Args:  args,
			}

			queries = append(queries, newQuery)
		case models.DMLDeleteType:
			queryStr := "DELETE FROM "
			queryStr += db.formatTableName(change.Table)

			args := make([]interface{}, len(change.PrimaryKeyInfo))

			for i, pki := range change.PrimaryKeyInfo {
				if i == 0 {
					queryStr += fmt.Sprintf(" WHERE `%s` = ?", pki.Name)
				} else {
					queryStr += fmt.Sprintf(" AND `%s` = ?", pki.Name)
				}
				args[i] = pki.Value
			}

			newQuery := models.Query{
				Query: queryStr,
				Args:  args,
			}

			queries = append(queries, newQuery)
		}
	}
	return queriesInTransaction(db.Connection, queries)
}

func (db *SQLite) GetPrimaryKeyColumnNames(database, table string) (primaryKeyColumnName []string, err error) {
	columns, err := db.GetTableColumns(database, table)
	if err != nil {
		return nil, err
	}

	indexOfPkColumn := -1

	for i, col := range columns[0] {
		if col == "pk" {
			indexOfPkColumn = i
		}
	}

	for i, col := range columns {
		if i > 0 && col[indexOfPkColumn] != "0" {
			primaryKeyColumnName = append(primaryKeyColumnName, col[0])
		}
	}

	return primaryKeyColumnName, nil
}

func (db *SQLite) SetProvider(provider string) {
	db.Provider = provider
}

func (db *SQLite) GetProvider() string {
	return db.Provider
}

func (db *SQLite) formatTableName(table string) string {
	return fmt.Sprintf("`%s`", table)
}
