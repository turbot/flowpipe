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

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/marcboeker/go-duckdb"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"

	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/db_common"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

const (
	DriverPostgres = "postgres"
	DriverMySQL    = "mysql"
	DriverDuckDB   = "duckdb"
	DriverSQLite   = "sqlite3"
)

type Query struct {
	databaseType string

	queryReader QueryReader
}

func (e *Query) ValidateInput(ctx context.Context, i modconfig.Input) error {
	// A database connection string must be provided to set up the connection, unless we are using the mock database for the tests
	if i[schema.AttributeTypeConnectionString] == nil {
		return perr.BadRequestWithMessage("Query input must define connection_string")
	}

	if i[schema.AttributeTypeSql] == nil {
		return perr.BadRequestWithMessage("Query input must define sql")
	}

	// Validate the timeout attribute
	if i[schema.AttributeTypeTimeout] != nil {
		switch duration := i[schema.AttributeTypeTimeout].(type) {
		case string:
			_, err := time.ParseDuration(duration)
			if err != nil {
				return perr.BadRequestWithMessage("invalid sleep duration " + duration)
			}
		case int64:
			if duration < 0 {
				return perr.BadRequestWithMessage("The attribute '" + schema.AttributeTypeTimeout + "' must be a positive whole number")
			}
		case float64:
			if duration < 0 {
				return perr.BadRequestWithMessage("The attribute '" + schema.AttributeTypeTimeout + "' must be a positive whole number")
			}
		default:
			return perr.BadRequestWithMessage("The attribute '" + schema.AttributeTypeTimeout + "' must be a string or a whole number")
		}
	}

	return nil
}

func (e *Query) InitializeDB(ctx context.Context, i modconfig.Input) (*sql.DB, error) {
	var db *sql.DB
	var err error

	dbConnectionString := i[schema.AttributeTypeConnectionString].(string)

	if strings.HasPrefix(dbConnectionString, "postgres://") || strings.HasPrefix(dbConnectionString, "postgresql://") {
		db, err = sql.Open("postgres", dbConnectionString)

	} else if strings.HasPrefix(dbConnectionString, "mysql:") {
		queryReader := &MySQLQueryReader{
			connectionString: dbConnectionString,
		}
		e.databaseType = DriverMySQL

		return queryReader.Initialize()

	} else if strings.HasPrefix(dbConnectionString, "duckdb:") {
		duckDBConnectionString := dbConnectionString[7:]
		if duckDBConnectionString == "" {
			return nil, perr.BadRequestWithMessage("Invalid DuckDB database connection string")
		}
		duckDBConnectionString, err = formatSqlConnectionString(duckDBConnectionString)
		if err != nil {
			return nil, err
		}

		slog.Debug("Opening DuckDB database", "connectionString", duckDBConnectionString)
		db, err = sql.Open(DriverDuckDB, duckDBConnectionString)

	} else if strings.HasPrefix(dbConnectionString, "sqlite:") {
		sqliteConnectionString := dbConnectionString[7:]
		if sqliteConnectionString == "" {
			return nil, perr.BadRequestWithMessage("Invalid sqlite database connection string")
		}
		sqliteConnectionString, err = formatSqlConnectionString(sqliteConnectionString)
		if err != nil {
			return nil, err
		}

		slog.Debug("Opening sqlite database", "connection string", sqliteConnectionString)

		e.databaseType = DriverSQLite
		db, err = sql.Open("sqlite3", sqliteConnectionString)

	} else {
		return nil, perr.BadRequestWithMessage("Invalid database connection string")

	}

	if err != nil {
		return nil, err
	}

	return db, nil
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

func (e *Query) RunWithMetadata(ctx context.Context, input modconfig.Input) (*modconfig.Output, map[string]*sql.ColumnType, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, nil, err
	}

	db, err := e.InitializeDB(ctx, input)
	if err != nil {
		slog.Error("Error initializing the database", "error", err)
		return nil, nil, perr.InternalWithMessage("Error initializing the database: " + err.Error())
	}
	defer db.Close()

	// Get the inputs
	queryString := input[schema.AttributeTypeSql].(string)

	var args []interface{}
	if input[schema.AttributeTypeArgs] != nil {
		args = input[schema.AttributeTypeArgs].([]interface{})
	}

	var timeout time.Duration
	if input[schema.AttributeTypeTimeout] != nil {
		switch timeoutDuration := input[schema.AttributeTypeTimeout].(type) {
		case string:
			timeout, _ = time.ParseDuration(timeoutDuration)
		case int64:
			timeout = time.Duration(timeoutDuration) * time.Millisecond // in milliseconds
		case float64:
			timeout = time.Duration(timeoutDuration) * time.Millisecond // in milliseconds
		}
	}

	results := []map[string]interface{}{}
	var rows *sql.Rows

	start := time.Now().UTC()
	// For the query timeout we use a context with a timeout.
	// When we test the query test, it runs the primitive directly, so the context is clean.
	// But, when we run it inside watermill (e.g. in the integration tests), the context is already full of stuff which
	// causes a context cancellation error for some test which don't have timeout set.
	// So, for now we use 2 different methods to run the query, depending on whether the timeout is set or not.
	// If set, we use the context with timeout, otherwise we use the sql.Query method.
	if timeout > 0 {
		var cancel context.CancelFunc

		// We can't use the watermill context to set the query timeout, since it will be cancelled by watermill.
		// So, we create a new context with timeout and use it to run the query.
		// Set the timeout
		contextWithTimeout, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		rows, err = db.QueryContext(contextWithTimeout, queryString, args...)
	} else {
		//
		// potential for (?) but doesn't seem to make any difference: https://github.com/go-sql-driver/mysql/issues/407
		//
		// if e.databaseType == DriverMySQL {
		// 	stmt, err := db.Prepare(queryString)
		// 	if err != nil {
		// 		return nil, perr.InternalWithMessage("Error preparing query: " + err.Error())
		// 	}
		// 	defer stmt.Close()
		// 	rows, err = stmt.Query(args...)
		// 	if err != nil {
		// 		return nil, perr.InternalWithMessage("Error executing query: " + err.Error())
		// 	}
		// } else {
		rows, err = db.Query(queryString, args...)
		// }
	}

	if err != nil {
		return nil, nil, perr.InternalWithMessage("Error executing query: " + err.Error())
	}
	defer rows.Close()

	finish := time.Now().UTC()

	columnsTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, perr.InternalWithMessage("Error getting column types: " + err.Error())
	}
	columnTypeMap := map[string]*sql.ColumnType{}
	for _, columnType := range columnsTypes {
		columnTypeMap[columnType.Name()] = columnType
	}

	for rows.Next() {
		row := make(map[string]interface{})
		err = mapScan(rows, row)
		if err != nil {
			return nil, nil, perr.InternalWithMessage("Failed to scan row: " + err.Error())
		}
		// sqlx doesn't handle jsonb columns, so we need to do it manually
		// https://github.com/jmoiron/sqlx/issues/225

		// TODO: refactor this, add abstraction make it extensible to future database types
		for k, encoded := range row {
			switch ba := encoded.(type) {
			case []byte:
				if e.databaseType == DriverMySQL { // Check it it's a valid JSON object
					if isJSON, _ := db_common.IsJSON(ba); isJSON {
						var col interface{}
						err := json.Unmarshal(ba, &col)
						if err != nil {
							slog.Error("error unmarshalling jsonb", "column", k, "error", err)
							return nil, nil, perr.InternalWithMessage("Error unmarshalling jsonb column: " + err.Error())
						}
						row[k] = col
						continue
					}

					row[k], err = mysqlReadCell(ba, columnTypeMap[k])
					if err != nil {
						return nil, nil, perr.InternalWithMessage("Error reading cell: " + err.Error())
					}
				} else {
					// Check it it's a valid JSON object
					if isJSON, _ := db_common.IsJSON(ba); isJSON {
						var col interface{}
						err := json.Unmarshal(ba, &col)
						if err != nil {
							slog.Error("error unmarshalling jsonb", "column", k, "error", err)
							return nil, nil, perr.InternalWithMessage("Error unmarshalling jsonb column: " + err.Error())
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
				}
			}
		}
		results = append(results, row)
	}

	if err = rows.Err(); err != nil {
		// Check for context deadline exceeded error
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, nil, perr.TimeoutWithMessage("Query execution exceeded timeout")
		}
		return nil, nil, perr.InternalWithMessage("Error iterating over query results: " + err.Error())
	}

	output := &modconfig.Output{
		Data: map[string]interface{}{},
	}

	output.Data[schema.AttributeTypeRows] = results
	output.Data[schema.AttributeTypeStartedAt] = start
	output.Data[schema.AttributeTypeFinishedAt] = finish

	return output, columnTypeMap, nil
}

func (e *Query) Run(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {
	output, _, err := e.RunWithMetadata(ctx, input)
	return output, err
}

func mysqlReadCell(columnValue any, columnType *sql.ColumnType) (result any, err error) {
	if columnValue != nil {
		asStr := string(columnValue.([]byte))
		switch columnType.DatabaseTypeName() {
		case "INT", "TINYINT", "SMALLINT", "MEDIUMINT", "BIGINT", "YEAR":
			result, err = strconv.ParseInt(asStr, 10, 64)
		case "DECIMAL", "NUMERIC", "FLOAT", "DOUBLE":
			result, err = strconv.ParseFloat(asStr, 64)
		case "DATE":
			result, err = time.Parse(time.DateOnly, asStr)
		case "TIME":
			result, err = time.Parse(time.TimeOnly, asStr)
		case "DATETIME", "TIMESTAMP":
			result, err = time.Parse(time.DateTime, asStr)
		case "BIT", "BLOB", "BINARY", "VARBINARY":
			result = columnValue.([]byte)
		// case "CHAR", "VARCHAR", "TEXT", "ENUM", "SET":
		default:
			result = asStr
		}
	}
	return result, err
}

func mapScan(r *sql.Rows, dest map[string]interface{}) error {
	// ignore r.started, since we needn't use reflect for anything.
	columns, err := r.Columns()
	if err != nil {
		return err
	}

	values := make([]interface{}, len(columns))
	for i := range values {
		values[i] = new(interface{})
	}

	err = r.Scan(values...)
	if err != nil {
		return err
	}

	for i, column := range columns {
		dest[column] = *(values[i].(*interface{}))
	}

	return r.Err()
}
