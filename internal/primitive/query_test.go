package primitive

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

func TestQueryListAll(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	hr := Query{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeConnectionString: "sqlite:///workspaces/flowpipe/internal/primitive/database_files/employee.db",
		schema.AttributeTypeSql:              "select * from employee order by id;",
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal(15, len(output.Get(schema.AttributeTypeRows).([]map[string]interface{})))

	// Expected output from the query
	expectedResult := []map[string]interface{}{
		{
			"email":       "john@example.com",
			"id":          int64(1),
			"name":        "John",
			"preferences": "{\"theme\": \"dark\", \"notifications\": true}",
		},
		{
			"email":       "adam@example.com",
			"id":          int64(2),
			"name":        "Adam",
			"preferences": "{\"theme\": \"dark\", \"notifications\": true}",
		},
		{
			"email":       "alice@example.com",
			"id":          int64(3),
			"name":        "Alice",
			"preferences": "{\"theme\": \"dark\", \"notifications\": false}",
		},
		{
			"email":       "bob@example.com",
			"id":          int64(4),
			"name":        "Bob",
			"preferences": "{\"theme\": \"light\", \"notifications\": false}",
		},
		{
			"email":       "alex@example.com",
			"id":          int64(5),
			"name":        "Alex",
			"preferences": "{\"theme\": \"dark\", \"notifications\": true}",
		},
		{
			"email":       "carey@example.com",
			"id":          int64(6),
			"name":        "Carey",
			"preferences": "{\"theme\": \"light\", \"notifications\": false}",
		},
		{
			"email":       "cody@example.com",
			"id":          int64(7),
			"name":        "Cody",
			"preferences": "{\"theme\": \"light\", \"notifications\": false}",
		},
		{
			"email":       "andrew@example.com",
			"id":          int64(8),
			"name":        "Andrew",
			"preferences": "{\"theme\": \"dark\", \"notifications\": true}",
		},
		{
			"email":       "alexandra@example.com",
			"id":          int64(9),
			"name":        "Alexandra",
			"preferences": "{\"theme\": \"light\", \"notifications\": true}",
		},
		{
			"email":       "jon@example.com",
			"id":          int64(10),
			"name":        "Jon",
			"preferences": "{\"theme\": \"dark\", \"notifications\": true}",
		},
		{
			"email":       "jennifer@example.com",
			"id":          int64(11),
			"name":        "Jennifer",
			"preferences": "{\"theme\": \"light\", \"notifications\": false}",
		},
		{
			"email":       "alan@example.com",
			"id":          int64(12),
			"name":        "Alan",
			"preferences": "{\"theme\": \"dark\", \"notifications\": true}",
		},
		{
			"email":       "mia@example.com",
			"id":          int64(13),
			"name":        "Mia",
			"preferences": "{\"theme\": \"light\", \"notifications\": true}",
		},
		{
			"email":       "aaron@example.com",
			"id":          int64(14),
			"name":        "Aaron",
			"preferences": "{\"theme\": \"light\", \"notifications\": true}",
		},
		{
			"email":       "adrian@example.com",
			"id":          int64(15),
			"name":        "Adrian",
			"preferences": "{\"theme\": \"dark\", \"notifications\": true}",
		},
	}

	expectedRow := output.Get(schema.AttributeTypeRows).([]map[string]interface{})
	assert.Equal(expectedResult, expectedRow)
}

func TestQueryWithArgs(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	hr := Query{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeConnectionString: "sqlite:database_files/employee.db",
		schema.AttributeTypeSql:              "select * from employee where id = $1;",
		schema.AttributeTypeArgs:             []interface{}{10},
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal(1, len(output.Get(schema.AttributeTypeRows).([]map[string]interface{})))

	// Expected output from the query
	expectedResult := []map[string]interface{}{
		{
			"email":       "jon@example.com",
			"id":          int64(10),
			"name":        "Jon",
			"preferences": "{\"theme\": \"dark\", \"notifications\": true}",
		},
	}

	expectedRow := output.Get(schema.AttributeTypeRows).([]map[string]interface{})
	assert.Equal(expectedResult, expectedRow)
}

func TestQueryWithArgsContainsRegexExpression(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	hr := Query{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeConnectionString: "sqlite:database_files/employee.db",
		schema.AttributeTypeSql:              "SELECT * from employee where name like $1;",
		schema.AttributeTypeArgs:             []interface{}{"Jo%"},
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal(2, len(output.Get(schema.AttributeTypeRows).([]map[string]interface{})))

	// Expected output from the query
	expectedResult := []map[string]interface{}{
		{
			"email":       "john@example.com",
			"id":          int64(1),
			"name":        "John",
			"preferences": "{\"theme\": \"dark\", \"notifications\": true}",
		},
		{
			"email":       "jon@example.com",
			"id":          int64(10),
			"name":        "Jon",
			"preferences": "{\"theme\": \"dark\", \"notifications\": true}",
		},
	}

	expectedRow := output.Get(schema.AttributeTypeRows).([]map[string]interface{})
	assert.Equal(expectedResult, expectedRow)
}

func TestQueryTableNotFound(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	hr := Query{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeConnectionString: "sqlite:database_files/employee.db",
		schema.AttributeTypeSql:              "select * from user;",
	})

	output, err := hr.Run(ctx, input)
	assert.NotNil(err)                                      // Expect an error since the table does not exist
	assert.Equal(nil, output.Get(schema.AttributeTypeRows)) // Expect no rows to be returned
}

func TestQueryNoRows(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	hr := Query{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeConnectionString: "sqlite:database_files/employee.db",
		schema.AttributeTypeSql:              "select * from department;",
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal(0, len(output.Get(schema.AttributeTypeRows).([]map[string]interface{})))
}

func TestQueryBadQueryStatement(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	hr := Query{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeConnectionString: "sqlite:database_files/employee.db",
		schema.AttributeTypeSql:              "SELECT * employee;",
	})

	_, err := hr.Run(ctx, input)
	assert.NotNil(err)
	assert.Contains(err.Error(), "syntax error")
}

func TestQueryWithMissingAttributeSql(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	hr := Query{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeConnectionString: "this is a connection string",
	})

	_, err := hr.Run(ctx, input)
	assert.NotNil(err)

	fpErr := err.(perr.ErrorModel)
	assert.Equal("Query input must define sql", fpErr.Detail)
	assert.Equal(400, fpErr.Status)
}

func TestQueryWithInvalidAttribute(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	hr := Query{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeConnectionString: "this is a connection string",
		"sql1":                               "select * from employee;",
	})

	_, err := hr.Run(ctx, input)
	assert.NotNil(err)

	fpErr := err.(perr.ErrorModel)
	assert.Equal("Query input must define sql", fpErr.Detail)
	assert.Equal(400, fpErr.Status)
}

func TestQueryWithTimestamp(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	queryPrimitive := Query{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeConnectionString: "this is a connection string",

		// This query string used to cause issue because we were trying to detect args in the query string, i.e. $1, ? or :name (Oracle)
		// In retrospect, we believe it's difficult to cover all possible cases especially with complex SQL. So, we decided to remove the
		// detection of args in the query string and let the query fails if user does not supply args with the query.
		schema.AttributeTypeSql: "select * from hackernews.hackernews_new where time::timestamp < now()",
	})

	err := queryPrimitive.ValidateInput(ctx, input)
	assert.Nil(err)
}

func TestQueryMissingConnectionString(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	hr := Query{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeSql:  "SELECT * from aws_ec2_instance where instance_id = $1",
		schema.AttributeTypeArgs: []interface{}{"i-000a000b0000c00d1"},
	})

	_, err := hr.Run(ctx, input)
	assert.NotNil(err)

	fpErr := err.(perr.ErrorModel)
	assert.Equal("Query input must define connection_string", fpErr.Detail)
	assert.Equal(400, fpErr.Status)
}

func TestQueryDuckDB(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	hr := Query{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeConnectionString: "duckdb:./database_files/new_database.duckdb",
		schema.AttributeTypeSql:              "select * from employee order by id;",
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal(3, len(output.Get(schema.AttributeTypeRows).([]map[string]interface{})))

	rows := output.Get(schema.AttributeTypeRows).([]map[string]interface{})

	// Row 1
	assert.Equal(int32(1), rows[0]["id"])
	assert.Equal("john@example.com", rows[0]["email"])
	assert.Equal("{\"theme\": \"dark\", \"notifications\": true}", rows[0]["preferences"])

	// Row 2
	assert.Equal(int32(2), rows[1]["id"])
	assert.Equal("adam@example.com", rows[1]["email"])
	assert.Equal("{\"theme\": \"light\", \"notifications\": true}", rows[1]["preferences"])

	// Row 3
	assert.Equal(int32(3), rows[2]["id"])
	assert.Equal("alice@example.com", rows[2]["email"])
	assert.Equal("{\"theme\": \"dark\", \"notifications\": true}", rows[2]["preferences"])
}

func TestQuerySQLiteDB(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	hr := Query{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeConnectionString: "sqlite:database_files/employee.db",
		schema.AttributeTypeSql:              "select * from employee order by id;",
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal(15, len(output.Get(schema.AttributeTypeRows).([]map[string]interface{})))

	rows := output.Get(schema.AttributeTypeRows).([]map[string]interface{})

	// Row 1
	assert.Equal(int64(1), rows[0]["id"])
	assert.Equal("john@example.com", rows[0]["email"])
	assert.Equal("{\"theme\": \"dark\", \"notifications\": true}", rows[0]["preferences"])

	// Row 2
	assert.Equal(int64(2), rows[1]["id"])
	assert.Equal("adam@example.com", rows[1]["email"])
	assert.Equal("{\"theme\": \"dark\", \"notifications\": true}", rows[1]["preferences"])

	// Row 3
	assert.Equal(int64(3), rows[2]["id"])
	assert.Equal("alice@example.com", rows[2]["email"])
	assert.Equal("{\"theme\": \"dark\", \"notifications\": false}", rows[2]["preferences"])
}

func TestQueryInvalidDatabase(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	hr := Query{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeConnectionString: "abcd",
		schema.AttributeTypeSql:              "select * from employee order by id;",
	})

	_, err := hr.Run(ctx, input)
	assert.NotNil(err)
	assert.Contains(err.Error(), "Bad Request: Invalid database connection string")
}
