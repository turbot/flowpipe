package primitive

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/marcboeker/go-duckdb"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"

	"github.com/turbot/pipe-fittings/constants"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

const (
	DriverPostgres = "postgres"
	DriverMySQL    = "mysql"
	DriverDuckDB   = "duckdb"
)

type Query struct{}

func (e *Query) ValidateInput(ctx context.Context, i modconfig.Input) error {
	// A database connection string must be provided to set up the connection, unless we are using the mock database for the tests
	if i[schema.AttributeTypeConnectionString] == nil {
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

	dbConnectionString := i[schema.AttributeTypeConnectionString].(string)

	if strings.HasPrefix(dbConnectionString, "postgres://") || strings.HasPrefix(dbConnectionString, "postgresql://") {
		db, err = sql.Open("postgres", dbConnectionString)

	} else if strings.HasPrefix(dbConnectionString, "mysql://") {
		trimmedDBConnectionString := strings.TrimPrefix(dbConnectionString, "mysql://")
		db, err = sql.Open(DriverMySQL, trimmedDBConnectionString)

	} else if strings.HasPrefix(dbConnectionString, "duckdb:") {
		duckDBFile := dbConnectionString[7:]
		if duckDBFile == "" {
			return nil, perr.BadRequestWithMessage("Invalid duckDB database connection string")
		}
		dbFile := filepath.Join(viper.GetString(constants.ArgModLocation), duckDBFile)

		slog.Debug("Opening duckDB database", "file", dbFile)
		db, err = sql.Open(DriverDuckDB, dbFile)

	} else if strings.HasPrefix(dbConnectionString, "sqlite:") {
		sqlLiteFile := dbConnectionString[7:]
		if sqlLiteFile == "" {
			return nil, perr.BadRequestWithMessage("Invalid sqlite database connection string")
		}
		dbFile := filepath.Join(viper.GetString(constants.ArgModLocation), sqlLiteFile)

		slog.Debug("Opening sqlite database", "file", dbFile)
		db, err = sql.Open("sqlite3", dbFile)

	} else {
		return nil, perr.BadRequestWithMessage("Invalid database connection string")

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

	db, err := e.InitializeDB(ctx, input)
	if err != nil {
		slog.Error("Error initializing the database", "error", err)
		return nil, perr.InternalWithMessage("Error initializing the database: " + err.Error())
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
