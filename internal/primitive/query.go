package primitive

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/marcboeker/go-duckdb"
	_ "github.com/mattn/go-sqlite3"

	"github.com/turbot/flowpipe/internal/resources"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

const (
	DriverPostgres   = "postgres"
	DriverPostgresql = "postgresql"
	DriverMySQL      = "mysql"
	DriverDuckDB     = "duckdb"
	DriverSQLite3    = "sqlite3"
	DriverSQLite     = "sqlite"
)

type Query struct {
	QueryReader QueryReader
}

func (e *Query) ValidateInput(ctx context.Context, i resources.Input) error {
	// A database connection string must be provided to set up the connection, unless we are using the mock database for the tests
	if i[schema.AttributeTypeDatabase] == nil {
		return perr.BadRequestWithMessage("Query input must define database")
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

func (e *Query) RunWithMetadata(ctx context.Context, input resources.Input) (*resources.Output, map[string]*sql.ColumnType, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, nil, err
	}

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

	queryReader, err := NewQueryReader(input[schema.AttributeTypeDatabase].(string))
	if err != nil {
		slog.Error("Error initializing the database", "error", err)
		return nil, nil, perr.InternalWithMessage("Error initializing the database: " + err.Error())
	}
	e.QueryReader = queryReader
	defer queryReader.Close()

	var results []map[string]interface{}
	var md map[string]*sql.ColumnType

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
		results, md, err = queryReader.Query(contextWithTimeout, queryString, args...)
	} else {
		results, md, err = queryReader.Query(context.Background(), queryString, args...)
	}

	if err != nil {
		return nil, nil, perr.InternalWithMessage("Error executing query: " + err.Error())
	}

	finish := time.Now().UTC()

	output := &resources.Output{
		Data: map[string]interface{}{},
	}

	output.Data[schema.AttributeTypeRows] = results
	output.Flowpipe = FlowpipeMetadataOutput(start, finish)

	return output, md, nil
}

func (e *Query) Run(ctx context.Context, input resources.Input) (*resources.Output, error) {
	output, _, err := e.RunWithMetadata(ctx, input)
	return output, err
}

func mapScan(r *sql.Rows, columns []string, dest map[string]interface{}) error {
	// ignore r.started, since we needn't use reflect for anything.
	values := make([]interface{}, len(columns))
	for i := range values {
		values[i] = new(interface{})
	}

	err := r.Scan(values...)
	if err != nil {
		return err
	}

	for i, column := range columns {
		dest[column] = *(values[i].(*interface{}))
	}

	return r.Err()
}
