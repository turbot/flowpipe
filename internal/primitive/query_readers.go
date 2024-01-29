package primitive

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/db_common"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func NewQueryReader(dbConnectionString string) (QueryReader, error) {
	parts := strings.SplitN(dbConnectionString, ":", 2)

	if len(parts) != 2 {
		return nil, perr.BadRequestWithMessage("Invalid database connection string")
	}

	var queryReader QueryReader
	switch parts[0] {
	case DriverPostgres, DriverPostgresql:
		queryReader = &PostgresQueryReader{
			QueryReaderImpl{
				connectionString: dbConnectionString,
				rowReader:        &RowReaderImpl{},
			},
		}
	case DriverDuckDB:
		queryReader = &FileBasedQueryReader{
			QueryReaderImpl{
				connectionString: dbConnectionString,
				rowReader:        &RowReaderImpl{},
			},
		}
	case DriverSQLite, DriverSQLite3:
		queryReader = &SQLiteQueryReader{
			QueryReaderImpl{
				connectionString: dbConnectionString,
				rowReader:        &RowReaderImpl{},
			},
		}
	case DriverMySQL:
		queryReader = &MySQLQueryReader{
			QueryReaderImpl{
				connectionString: dbConnectionString,
				rowReader:        &MySQLRowReader{},
			},
		}
	default:
		return nil, perr.BadRequestWithMessage("Unsupported database type " + parts[0] + ". Supported types are: " + DriverPostgres + ", " + DriverPostgresql + ", " + DriverMySQL + ", " + DriverDuckDB + ", " + DriverSQLite3 + ".")
	}

	err := queryReader.Initialize()
	return queryReader, err
}

type QueryReader interface {
	GetConnectionString() string
	Initialize() error
	Query(context.Context, string, ...interface{}) ([]map[string]interface{}, map[string]*sql.ColumnType, error)
	RowsToCty(rows []map[string]interface{}, columnTypes map[string]*sql.ColumnType) ([]cty.Value, error)
	Close()
}

type QueryReaderImpl struct {
	connectionString string
	db               *sql.DB
	rowReader        RowReader
}

func (q *QueryReaderImpl) Initialize() error {
	parts := strings.SplitN(q.connectionString, ":", 2)

	if len(parts) != 2 {
		return perr.BadRequestWithMessage("Invalid database connection string")
	}

	driver := parts[0]
	trimmedConnectionString := parts[1]
	trimmedConnectionString = strings.TrimPrefix(trimmedConnectionString, "//")

	db, err := sql.Open(driver, trimmedConnectionString)
	q.db = db
	return err
}

func (q *QueryReaderImpl) Close() {
	if q.db != nil {
		q.db.Close()
	}
}

func (q *QueryReaderImpl) queryRows(rows *sql.Rows) ([]map[string]interface{}, map[string]*sql.ColumnType, error) {
	columnsTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, perr.InternalWithMessage("Error getting column types: " + err.Error())
	}
	columnTypeMap := map[string]*sql.ColumnType{}
	for _, columnType := range columnsTypes {
		columnTypeMap[columnType.Name()] = columnType
	}

	results, err := q.rowReader.Read(rows, columnTypeMap)
	if err != nil {
		return nil, nil, err
	}

	return results, columnTypeMap, nil
}

func (q *QueryReaderImpl) Query(ctx context.Context, queryString string, args ...interface{}) ([]map[string]interface{}, map[string]*sql.ColumnType, error) {
	rows, err := q.db.QueryContext(ctx, queryString, args...)

	if err != nil {
		return nil, nil, perr.InternalWithMessage("Error executing query: " + err.Error())
	}

	defer rows.Close()

	return q.queryRows(rows)
}

func (q *QueryReaderImpl) RowsToCty(rows []map[string]interface{}, columnTypes map[string]*sql.ColumnType) ([]cty.Value, error) {
	var rowsCty []cty.Value
	for _, r := range rows {
		rowCty, err := q.rowReader.RowToCty(r, columnTypes)
		if err != nil {
			return nil, err
		}
		rowsCty = append(rowsCty, rowCty)
	}
	return rowsCty, nil
}

func (q *QueryReaderImpl) GetConnectionString() string {
	return q.connectionString
}

type PostgresQueryReader struct {
	QueryReaderImpl
}

func (p *PostgresQueryReader) Initialize() error {
	db, err := sql.Open("postgres", p.connectionString)
	p.db = db

	return err
}

type MySQLQueryReader struct {
	QueryReaderImpl
}

func (m *MySQLQueryReader) Query(ctx context.Context, queryString string, args ...interface{}) ([]map[string]interface{}, map[string]*sql.ColumnType, error) {
	// potential for (?) but doesn't seem to make any difference: https://github.com/go-sql-driver/mysql/issues/407
	stmt, err := m.db.PrepareContext(ctx, queryString)
	if err != nil {
		return nil, nil, perr.InternalWithMessage("Error preparing query: " + err.Error())
	}

	defer stmt.Close()
	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, nil, perr.InternalWithMessage("Error executing query: " + err.Error())
	}
	defer rows.Close()

	return m.QueryReaderImpl.queryRows(rows)
}

func (m *MySQLQueryReader) RowsToCty(rows []map[string]interface{}, columnTypes map[string]*sql.ColumnType) ([]cty.Value, error) {
	var rowsCty []cty.Value
	for _, r := range rows {
		rowCty, err := m.rowReader.RowToCty(r, columnTypes)
		if err != nil {
			return nil, err
		}
		rowsCty = append(rowsCty, rowCty)
	}
	return rowsCty, nil
}

type SQLiteQueryReader struct {
	QueryReaderImpl
}

func (s *SQLiteQueryReader) Initialize() error {
	parts := strings.SplitN(s.QueryReaderImpl.connectionString, ":", 2)

	if len(parts) != 2 {
		return perr.BadRequestWithMessage("Invalid database connection string")
	}

	driver := parts[0]
	if driver != `sqlite3` && driver != `sqlite` {
		return perr.BadRequestWithMessage("Invalid database driver. Only sqlite3 and sqlite are supported")
	}

	driver = "sqlite3"
	trimmedConnectionString := parts[1]
	trimmedConnectionString = strings.TrimPrefix(trimmedConnectionString, "//")

	var err error
	trimmedConnectionString, err = formatSqlConnectionString(trimmedConnectionString)
	if err != nil {
		return err
	}

	db, err := sql.Open(driver, trimmedConnectionString)
	s.db = db
	return err
}

type FileBasedQueryReader struct {
	QueryReaderImpl
}

func (s *FileBasedQueryReader) Initialize() error {
	parts := strings.SplitN(s.QueryReaderImpl.connectionString, ":", 2)

	if len(parts) != 2 {
		return perr.BadRequestWithMessage("Invalid database connection string")
	}

	driver := parts[0]
	trimmedConnectionString := parts[1]
	trimmedConnectionString = strings.TrimPrefix(trimmedConnectionString, "//")

	var err error
	trimmedConnectionString, err = formatSqlConnectionString(trimmedConnectionString)
	if err != nil {
		return err
	}

	db, err := sql.Open(driver, trimmedConnectionString)
	s.db = db
	return err
}

type RowReader interface {
	Read(*sql.Rows, map[string]*sql.ColumnType) ([]map[string]interface{}, error)
	RowToCty(row map[string]interface{}, columnTypes map[string]*sql.ColumnType) (cty.Value, error)
}

type MySQLRowReader struct {
	RowReaderImpl
}

func (m *MySQLRowReader) Read(rows *sql.Rows, columnTypeMap map[string]*sql.ColumnType) ([]map[string]interface{}, error) {

	results := []map[string]interface{}{}

	columns, err := rows.Columns()
	if err != nil {
		return nil, perr.InternalWithMessage("Error getting columns: " + err.Error())
	}

	for rows.Next() {
		row := make(map[string]interface{})
		err = mapScan(rows, columns, row)
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
					return nil, perr.InternalWithMessage("Error reading cell: " + err.Error())
				}
			}
		}
		results = append(results, row)
	}

	if err = rows.Err(); err != nil {
		// Check for context deadline exceeded error
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, perr.TimeoutWithMessage("Query execution exceeded timeout")
		}
		return nil, perr.InternalWithMessage("Error iterating over query results: " + err.Error())
	}

	return results, err
}

func mysqlReadCell(columnValue any, columnType *sql.ColumnType) (result any, err error) {
	// https: //github.com/go-sql-driver/mysql/blob/master/fields.go

	if columnValue != nil {
		asStr := string(columnValue.([]byte))
		switch columnType.DatabaseTypeName() {
		case "INTEGER", "INT", "INT8", "TINYINT", "SMALLINT", "MEDIUMINT", "BIGINT", "YEAR", "UNSIGNED MEDIUMINT", "UNSIGNED INT", "UNSIGNED SMALLINT", "UNSIGNED TINYINT":
			result, err = strconv.ParseInt(asStr, 10, 64)
		case "DECIMAL", "NUMERIC", "FLOAT", "DOUBLE":
			result, err = strconv.ParseFloat(asStr, 64)
		case "DATE":
			result, err = time.Parse(time.DateOnly, asStr)
		case "TIME":
			result, err = time.Parse(time.TimeOnly, asStr)
		case "DATETIME", "TIMESTAMP":
			result, err = time.Parse(time.DateTime, asStr)
		// case "BIT", "BLOB", "BINARY", "VARBINARY", "MEDIUMBLOB", "TINYBLOB", "JSON", "LONGBLOB":
		// 	result = columnValue.([]byte)
		// case "CHAR", "VARCHAR", "TEXT", "ENUM", "SET", "GEOMETRY", "NULL"
		default:
			result = asStr
		}
	}
	return result, err
}

func (m *MySQLRowReader) RowToCty(row map[string]interface{}, columnTypes map[string]*sql.ColumnType) (cty.Value, error) {
	return m.RowReaderImpl.RowToCty(row, columnTypes)
}

type RowReaderImpl struct {
}

func (r *RowReaderImpl) Read(rows *sql.Rows, columnTypeMap map[string]*sql.ColumnType) ([]map[string]interface{}, error) {
	var err error
	results := []map[string]interface{}{}

	columns, err := rows.Columns()
	if err != nil {
		return nil, perr.InternalWithMessage("Error getting columns: " + err.Error())
	}

	for rows.Next() {
		row := make(map[string]interface{})
		err = mapScan(rows, columns, row)
		if err != nil {
			return nil, perr.InternalWithMessage("Failed to scan row: " + err.Error())
		}

		// sqlx doesn't handle jsonb columns, so we need to do it manually
		// https://github.com/jmoiron/sqlx/issues/225

		// TODO: refactor this, add abstraction make it extensible to future database types
		for k, encoded := range row {
			switch ba := encoded.(type) {
			case []byte:
				// Check it it's a valid JSON object
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

				// Check if it's base64 encoded
				if decodedData, err := base64.StdEncoding.DecodeString(string(ba)); err == nil {
					// It's valid base64
					row[k] = string(decodedData)
					continue
				}

				row[k] = string(ba)
			}
		}
		results = append(results, row)
	}

	if err = rows.Err(); err != nil {
		// Check for context deadline exceeded error
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, perr.TimeoutWithMessage("Query execution exceeded timeout")
		}
		return nil, perr.InternalWithMessage("Error iterating over query results: " + err.Error())
	}

	return results, err
}

// Attempt to have a generic function to convert a row to cty. It may not work for all the database that Flowpipe will support,
// the types are structured so it can be extended to various different DB Row Readers
func (r *RowReaderImpl) RowToCty(row map[string]interface{}, columnTypes map[string]*sql.ColumnType) (cty.Value, error) {

	// mysql: https: //github.com/go-sql-driver/mysql/blob/master/fields.go
	// sqlite:
	rowCty := map[string]cty.Value{}
	for k, v := range row {
		columnType := columnTypes[k]
		if columnType == nil {
			return cty.NilVal, perr.BadRequestWithMessage("Column type not found for column " + k)
		}

		switch strings.ToUpper(columnType.DatabaseTypeName()) {
		case "INTEGER", "INT", "INT8", "TINYINT", "SMALLINT", "MEDIUMINT", "BIGINT", "YEAR", "UNSIGNED MEDIUMINT", "UNSIGNED INT", "UNSIGNED SMALLINT", "UNSIGNED TINYINT", "DECIMAL", "NUMERIC", "FLOAT", "DOUBLE":
			if helpers.IsNil(v) {
				rowCty[k] = cty.NullVal(cty.Number)
				continue
			}

			val, err := gocty.ToCtyValue(v, cty.Number)
			if err != nil {
				return cty.NilVal, err
			}
			rowCty[k] = val

		case "JSON", "JSONB":
			if helpers.IsNil(v) {
				rowCty[k] = cty.NullVal(cty.EmptyObject)
				continue
			}

			val, err := hclhelpers.ConvertInterfaceToCtyValue(v)
			if err != nil {
				return cty.NilVal, err
			}
			rowCty[k] = val

		// All binary type will be converted to string
		case "BIT", "BLOB", "BINARY", "VARBINARY", "MEDIUMBLOB", "TINYBLOB", "LONGBLOB":
			if helpers.IsNil(v) {
				rowCty[k] = cty.NullVal(cty.String)
				continue
			}

			vals := fmt.Sprintf("%v", v)
			val := cty.StringVal(vals)

			rowCty[k] = val

		case "BOOLEAN":
			if helpers.IsNil(v) {
				rowCty[k] = cty.NullVal(cty.Bool)
				continue
			}

			val, err := gocty.ToCtyValue(v, cty.Bool)
			if err != nil {
				return cty.NilVal, err
			}
			rowCty[k] = val

		case "DATE", "TIME", "DATETIME", "TIMESTAMP":
			if helpers.IsNil(v) {
				rowCty[k] = cty.NullVal(cty.String)
				continue
			}

			if t, ok := v.(time.Time); ok {
				rfc3339Time := t.Format(time.RFC3339)
				rowCty[k] = cty.StringVal(rfc3339Time)
				continue
			}

			stringVal := fmt.Sprintf("%v", v)
			val, err := gocty.ToCtyValue(stringVal, cty.String)
			if err != nil {
				return cty.NilVal, err
			}
			rowCty[k] = val

		// case "CHAR", "VARCHAR", "TEXT", "ENUM", "SET", "GEOMETRY", "NULL",
		default:
			if helpers.IsNil(v) {
				rowCty[k] = cty.NullVal(cty.String)
				continue
			}
			val, err := gocty.ToCtyValue(v, cty.String)
			if err != nil {
				return cty.NilVal, err
			}
			rowCty[k] = val
		}
	}
	return cty.ObjectVal(rowCty), nil
}

// Function to append basePath to the file part of the  connection string
func formatSqlConnectionString(connStr string) (string, error) {
	parts := strings.SplitN(connStr, "?", 2)
	if len(parts) == 0 {
		return "", perr.BadRequestWithMessage(fmt.Sprintf("Invalid connection string: %s", connStr))
	}

	// Append the base path to the file part
	formatted := filepath.Join(viper.GetString(constants.ArgModLocation), parts[0])

	// If there are additional parameters, append them back
	if len(parts) > 1 {
		formatted += "?" + parts[1]
	}

	return formatted, nil
}
