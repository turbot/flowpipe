package primitive

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/marcboeker/go-duckdb"

	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"

	"github.com/DATA-DOG/go-sqlmock"
)

const (
	DriverPostgres = "postgres"
	DriverMySQL    = "mysql"
	DriverDuckDB   = "duckdb"
)

type Query struct {
	Setting string
	Mock    *sqlmock.Sqlmock
	DB      *sql.DB
}

func (e *Query) ValidateInput(ctx context.Context, i modconfig.Input) error {
	// A database connection string must be provided to set up the connection, unless we are using the mock database for the tests
	if e.Setting != "go-sqlmock" && i[schema.AttributeTypeConnectionString] == nil {
		return perr.BadRequestWithMessage("Query input must define connection_string")
	}

	if i[schema.AttributeTypeSql] == nil {
		return perr.BadRequestWithMessage("Query input must define sql")
	}
	return nil
}

func (e *Query) InitializeDB(ctx context.Context, i modconfig.Input) (*sql.DB, error) {
	var db *sql.DB
	var err error

	// The Run method opens a database connection by connecting to the provided database connection string.
	// But, while running the tests, we can't pass the connection string, hence we need to mock a database connection.
	if e.Setting == "go-sqlmock" {
		db, mock, err := sqlmock.New()
		if err != nil {
			return nil, perr.BadRequestWithMessage("Failed to open stub database connection: " + err.Error())
		}
		e.Mock = &mock
		e.DB = db

		return db, nil
	}

	dbConnectionString := i[schema.AttributeTypeConnectionString].(string)
	switch {
	case strings.Contains(dbConnectionString, "postgres://"):
		db, err = sql.Open(DriverPostgres, dbConnectionString)
	case strings.Contains(dbConnectionString, "mysql://"):
		trimmedDBConnectionString := strings.TrimPrefix(dbConnectionString, "mysql://")
		db, err = sql.Open(DriverMySQL, trimmedDBConnectionString)
	case strings.HasPrefix(dbConnectionString, "duckdb:"):
		db, err = sql.Open(DriverDuckDB, dbConnectionString)
	default:
		return nil, perr.BadRequestWithMessage("Unsupported database type")
	}

	if err != nil {
		return nil, err
	}

	return db, nil
}

func (e *Query) Run(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	var db *sql.DB
	var err error

	if e.DB == nil {
		db, err = e.InitializeDB(ctx, input)
		if err != nil {
			return nil, err
		}
	} else {
		db = e.DB
	}
	defer db.Close()

	// Get the inputs
	sql := input[schema.AttributeTypeSql].(string)

	var args []interface{}
	if input[schema.AttributeTypeArgs] != nil {
		args = input[schema.AttributeTypeArgs].([]interface{})
	}

	results := []map[string]interface{}{}

	start := time.Now().UTC()
	rows, err := db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	finish := time.Now().UTC()
	for rows.Next() {
		row := make(map[string]interface{})
		err = mapScan(rows, row)
		if err != nil {
			return nil, err
		}
		// sqlx doesn't handle jsonb columns, so we need to do it manually
		// https://github.com/jmoiron/sqlx/issues/225
		for k, encoded := range row {
			switch ba := encoded.(type) {
			case []byte:
				// Check it it's a valid JSON object
				if isJSON, _ := isJSON(ba); isJSON {
					var col interface{}
					err := json.Unmarshal(ba, &col)
					if err != nil {
						slog.Error("error unmarshalling jsonb", "column", k, "error", err)
						return nil, err
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
		return nil, err
	}

	output := &modconfig.Output{
		Data: map[string]interface{}{},
	}

	output.Data[schema.AttributeTypeRows] = results
	output.Data[schema.AttributeTypeStartedAt] = start
	output.Data[schema.AttributeTypeFinishedAt] = finish

	return output, nil
}

func isJSON(b []byte) (bool, error) {
	var col interface{}
	err := json.Unmarshal(b, &col)
	if err != nil {
		return false, err
	}

	// Check if it's a JSON object (map) or array (slice)
	_, isObject := col.(map[string]interface{})
	_, isArray := col.([]interface{})

	return isObject || isArray, nil
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
