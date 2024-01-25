package primitive

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/turbot/pipe-fittings/db_common"
	"github.com/turbot/pipe-fittings/perr"
)

type QueryReader interface {
	GetConnectionString() string
	Initialize() (*sql.DB, error)
	Query(string, ...interface{}) (*sql.Rows, []sql.ColumnType, error)
}

type QueryReaderImpl struct {
	connectionString string
	db               *sql.DB
}
type MySQLQueryReader struct {
	QueryReaderImpl
}

func (m *MySQLQueryReader) GetConnectionString() string {
	return m.connectionString
}

func (m *MySQLQueryReader) Initialize() (*sql.DB, error) {
	trimmedDBConnectionString := strings.TrimPrefix(m.connectionString, "mysql://")

	db, err := sql.Open(DriverMySQL, trimmedDBConnectionString)
	m.db = db

	return db, err
}

func (m *MySQLQueryReader) Query(queryString string, args ...interface{}) (*sql.Rows, map[string]*sql.ColumnType, error) {
	// potential for (?) but doesn't seem to make any difference: https://github.com/go-sql-driver/mysql/issues/407
	stmt, err := m.db.Prepare(queryString)
	if err != nil {
		return nil, nil, perr.InternalWithMessage("Error preparing query: " + err.Error())
	}

	defer stmt.Close()
	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, nil, perr.InternalWithMessage("Error executing query: " + err.Error())
	}

	columnsTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, perr.InternalWithMessage("Error getting column types: " + err.Error())
	}
	columnTypeMap := map[string]*sql.ColumnType{}
	for _, columnType := range columnsTypes {
		columnTypeMap[columnType.Name()] = columnType
	}
}

type MySQLRowReader struct {
}

func (m *MySQLRowReader) Read(rows *sql.Rows, columnTypeMap map[string]*sql.ColumnType) ([]map[string]interface{}, error) {
	var err error
	for rows.Next() {
		row := make(map[string]interface{})
		err = mapScan(rows, row)
		if err != nil {
			return nil, perr.InternalWithMessage("Failed to scan row: " + err.Error())
		}
		// sqlx doesn't handle jsonb columns, so we need to do it manually
		// https://github.com/jmoiron/sqlx/issues/225

		// TODO: refactor this, add abstraction make it extensible to future database types
		for k, encoded := range row {
			switch ba := encoded.(type) {
			case []byte:
				if isJSON, _ := db_common.IsJSON(ba); isJSON {
					var col interface{}
					err := json.Unmarshal(ba, &col)
					if err != nil {
						slog.Error("error unmarshalling jsonb", "column", k, "error", err)
						return nil, perr.InternalWithMessage("Error unmarshalling jsonb column: " + err.Error())
					}
					row[k] = col
					continue
				}

				row[k], err = mysqlReadCell(ba, columnTypeMap[k])
				if err != nil {
					return nil, nil, perr.InternalWithMessage("Error reading cell: " + err.Error())
				}
			}
		}
		results = append(results, row)
	}
}
