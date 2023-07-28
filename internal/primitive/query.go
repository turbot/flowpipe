package primitive

import (
	"context"
	"database/sql"
	"encoding/json"
	"regexp"
	"time"

	_ "github.com/lib/pq"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser/schema"

	"github.com/DATA-DOG/go-sqlmock"
)

type Query struct {
	Setting string
	Mock    *sqlmock.Sqlmock
	DB      *sql.DB
}

func (e *Query) ValidateInput(ctx context.Context, i types.Input) error {
	// A database connection string must be provided to set up the connection, unless we are using the mock database for the tests
	if e.Setting != "go-sqlmock" && i[schema.AttributeTypeConnectionString] == nil {
		return fperr.BadRequestWithMessage("Query input must define connection_string")
	}

	if i[schema.AttributeTypeSql] == nil {
		return fperr.BadRequestWithMessage("Query input must define sql")
	}

	sql := i[schema.AttributeTypeSql].(string)
	if hasPlaceholder(sql) && i[schema.AttributeTypeArgs] == nil {
		return fperr.BadRequestWithMessage("Query input must define args if the sql has placeholders")
	}
	return nil
}

func (e *Query) InitializeDB(ctx context.Context, i types.Input) (*sql.DB, error) {
	var db *sql.DB
	var err error

	// The Run method opens a database connection by connecting to the provided database connection string.
	// But, while running the tests, we can't pass the connection string, hence we need to mock a database connection.
	if e.Setting == "go-sqlmock" {
		db, mock, err := sqlmock.New()
		if err != nil {
			return nil, fperr.BadRequestWithMessage("Failed to open stub database connection: " + err.Error())
		}
		e.Mock = &mock
		e.DB = db

		return db, nil
	}

	dbConnectionString := i[schema.AttributeTypeConnectionString].(string)
	db, err = sql.Open("postgres", dbConnectionString)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (e *Query) Run(ctx context.Context, input types.Input) (*types.StepOutput, error) {
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
	sql := input["sql"].(string)

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
				var col interface{}
				err := json.Unmarshal(ba, &col)
				if err != nil {
					fplog.Logger(ctx).Error("error unmarshalling jsonb column %s: %v", k, err)
					return nil, err
				}
				row[k] = col
			}
		}
		results = append(results, row)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	output := &types.StepOutput{
		OutputVariables: map[string]interface{}{},
	}

	output.OutputVariables[schema.AttributeTypeQuery] = results
	output.OutputVariables[schema.AttributeTypeStartedAt] = start
	output.OutputVariables[schema.AttributeTypeFinishedAt] = finish

	return output, nil
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

func hasPlaceholder(query string) bool {
	// $1, $2, $3, etc. - PostgreSQL and some other databases use the dollar-sign syntax for positional parameters.
	// ? - Commonly used as a positional parameter placeholder in many databases, including MySQL, SQLite, and SQL Server.
	// :name - Oracle and some other databases use colon : followed by a parameter name for named parameters.
	placeholderPattern := regexp.MustCompile(`(\$[0-9]+|\?|:\w+)`)
	return placeholderPattern.MatchString(query)
}
