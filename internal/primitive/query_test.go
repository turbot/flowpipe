package primitive

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/types"
	"github.com/turbot/flowpipe/pipeparser/schema"
)

func TestQueryListAll(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := Query{
		Setting: "go-sqlmock",
	}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeSql: "SELECT * from aws_ec2_instance order by instance_id",
	})

	// Initialize the DB connection
	_, err := hr.InitializeDB(ctx, input)
	if err != nil {
		return
	}
	mock := *hr.Mock

	// Add the rows to the table
	rows := sqlmock.NewRows([]string{"instance_id", "arn", "type", "state"}).
		AddRow("i-000a000b0000c00d1", "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d1", "t2.micro", "stopped").
		AddRow("i-000a000b0000c00d2", "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d2", "t2.micro", "stopped").
		AddRow("i-000a000b0000c00d3", "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d3", "t2.micro", "running").
		AddRow("i-000a000b0000c00d4", "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d4", "t2.micro", "running").
		AddRow("i-000a000b0000c00d5", "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d5", "m5.xlarge", "stopped")

	mock.ExpectQuery("^SELECT (.+) from aws_ec2_instance order by instance_id$").WillReturnRows(rows)

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal(5, len(output.Get(schema.AttributeTypeQuery).([]map[string]interface{})))

	// Expected output from the query
	expectedResult := []map[string]interface{}{
		{
			"arn":         "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d1",
			"instance_id": "i-000a000b0000c00d1",
			"state":       "stopped",
			"type":        "t2.micro",
		},
		{
			"arn":         "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d2",
			"instance_id": "i-000a000b0000c00d2",
			"state":       "stopped",
			"type":        "t2.micro",
		},
		{
			"arn":         "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d3",
			"instance_id": "i-000a000b0000c00d3",
			"state":       "running",
			"type":        "t2.micro",
		},
		{
			"arn":         "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d4",
			"instance_id": "i-000a000b0000c00d4",
			"state":       "running",
			"type":        "t2.micro",
		},
		{
			"arn":         "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d5",
			"instance_id": "i-000a000b0000c00d5",
			"state":       "stopped",
			"type":        "m5.xlarge",
		},
	}

	expectedRow := output.Get(schema.AttributeTypeQuery).([]map[string]interface{})
	assert.Equal(expectedResult, expectedRow)
}

func TestQueryWithArgs(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := Query{
		Setting: "go-sqlmock",
	}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeSql:  "SELECT * from aws_ec2_instance where instance_id = $1",
		schema.AttributeTypeArgs: []interface{}{"i-000a000b0000c00d1"},
	})

	// Initialize the DB connection
	_, err := hr.InitializeDB(ctx, input)
	if err != nil {
		return
	}
	mock := *hr.Mock

	// Add the rows to the table
	rows := sqlmock.NewRows([]string{"instance_id", "arn", "type", "state"}).
		AddRow("i-000a000b0000c00d1", "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d1", "t2.micro", "stopped")

	mock.ExpectQuery("^SELECT \\* from aws_ec2_instance where instance_id = \\$1$").WillReturnRows(rows).WithArgs("i-000a000b0000c00d1")

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal(1, len(output.Get(schema.AttributeTypeQuery).([]map[string]interface{})))

	// Expected output from the query
	expectedResult := []map[string]interface{}{
		{
			"arn":         "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d1",
			"instance_id": "i-000a000b0000c00d1",
			"state":       "stopped",
			"type":        "t2.micro",
		},
	}

	expectedRow := output.Get(schema.AttributeTypeQuery).([]map[string]interface{})
	assert.Equal(expectedResult, expectedRow)
}

func TestQueryWithArgsContainsRegexExpression(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := Query{
		Setting: "go-sqlmock",
	}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeSql:  "SELECT * from aws_ec2_instance where type like $1",
		schema.AttributeTypeArgs: []interface{}{"t2%"},
	})

	// Initialize the DB connection
	_, err := hr.InitializeDB(ctx, input)
	if err != nil {
		return
	}
	mock := *hr.Mock

	// Add the rows to the table
	rows := sqlmock.NewRows([]string{"instance_id", "arn", "type", "state"}).
		AddRow("i-000a000b0000c00d1", "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d1", "t2.micro", "stopped").
		AddRow("i-000a000b0000c00d2", "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d2", "t2.micro", "stopped").
		AddRow("i-000a000b0000c00d3", "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d3", "t2.micro", "running").
		AddRow("i-000a000b0000c00d4", "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d4", "t2.micro", "running")

	mock.ExpectQuery("^SELECT \\* from aws_ec2_instance where type like \\$1$").WillReturnRows(rows).WithArgs("t2%")

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal(4, len(output.Get(schema.AttributeTypeQuery).([]map[string]interface{})))

	// Expected output from the query
	expectedResult := []map[string]interface{}{
		{
			"arn":         "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d1",
			"instance_id": "i-000a000b0000c00d1",
			"state":       "stopped",
			"type":        "t2.micro",
		},
		{
			"arn":         "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d2",
			"instance_id": "i-000a000b0000c00d2",
			"state":       "stopped",
			"type":        "t2.micro",
		},
		{
			"arn":         "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d3",
			"instance_id": "i-000a000b0000c00d3",
			"state":       "running",
			"type":        "t2.micro",
		},
		{
			"arn":         "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d4",
			"instance_id": "i-000a000b0000c00d4",
			"state":       "running",
			"type":        "t2.micro",
		},
	}

	expectedRow := output.Get(schema.AttributeTypeQuery).([]map[string]interface{})
	assert.Equal(expectedResult, expectedRow)
}

func TestQueryTableNotFound(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := Query{
		Setting: "go-sqlmock",
	}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeSql: "SELECT * from instance",
	})

	// Initialize the DB connection
	_, err := hr.InitializeDB(ctx, input)
	if err != nil {
		return
	}
	mock := *hr.Mock

	mock.ExpectQuery("^SELECT (.+) from instance$").WillReturnError(sql.ErrNoRows)

	output, err := hr.Run(ctx, input)
	assert.NotNil(err)                                       // Expect an error since the table does not exist
	assert.Equal(nil, output.Get(schema.AttributeTypeQuery)) // Expect no rows to be returned
}

func TestQueryNoRows(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := Query{
		Setting: "go-sqlmock",
	}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeSql: "SELECT * from aws_ec2_instance",
	})

	// Initialize the DB connection
	_, err := hr.InitializeDB(ctx, input)
	if err != nil {
		return
	}
	mock := *hr.Mock

	rows := sqlmock.NewRows([]string{"instance_id", "arn", "type", "state"})

	mock.ExpectQuery("^SELECT (.+) from aws_ec2_instance$").WillReturnRows(rows)

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal(0, len(output.Get(schema.AttributeTypeQuery).([]map[string]interface{})))
}

func TestQueryBadQueryStatement(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := Query{
		Setting: "go-sqlmock",
	}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeSql: "SELECT * instance",
	})

	// Initialize the DB connection
	_, err := hr.InitializeDB(ctx, input)
	if err != nil {
		return
	}
	mock := *hr.Mock

	mock.ExpectQuery("^SELECT (.+) instance$").WillReturnError(fperr.BadRequestWithMessage("syntax error at or near \"instance\""))

	_, err = hr.Run(ctx, input)
	assert.NotNil(err)
	assert.Contains(err.Error(), "syntax error at or near \"instance\"")
}

func TestQueryWithMissingAttributeSql(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := Query{}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeConnectionString: "this is a connection string",
	})

	_, err := hr.Run(ctx, input)
	assert.NotNil(err)

	fpErr := err.(fperr.ErrorModel)
	assert.Equal("Query input must define sql", fpErr.Detail)
	assert.Equal(400, fpErr.Status)
}

func TestQueryWithInvalidAttribute(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := Query{
		Setting: "go-sqlmock",
	}

	input := types.Input(map[string]interface{}{
		"sql1": "^SELECT (.+) from aws_ec2_instance order by instance_id$",
	})

	// Initialize the DB connection
	_, err := hr.InitializeDB(ctx, input)
	if err != nil {
		return
	}
	mock := *hr.Mock

	// Add the rows to the table
	rows := sqlmock.NewRows([]string{"instance_id", "arn", "type", "state"}).
		AddRow("i-000a000b0000c00d1", "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d1", "t2.micro", "stopped").
		AddRow("i-000a000b0000c00d2", "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d2", "t2.micro", "stopped").
		AddRow("i-000a000b0000c00d3", "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d3", "t2.micro", "running").
		AddRow("i-000a000b0000c00d4", "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d4", "t2.micro", "running").
		AddRow("i-000a000b0000c00d5", "arn:aws:ec2:ap-south-1:0123456789:instance/i-000a000b0000c00d5", "m5.xlarge", "stopped")

	mock.ExpectQuery("^SELECT (.+) from aws_ec2_instance order by instance_id$").WillReturnRows(rows)

	_, err = hr.Run(ctx, input)
	assert.NotNil(err)

	fpErr := err.(fperr.ErrorModel)
	assert.Equal("Query input must define sql", fpErr.Detail)
	assert.Equal(400, fpErr.Status)
}

func TestQueryMissingArgs(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := Query{}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeConnectionString: "this is a connection string",
		schema.AttributeTypeSql:              "SELECT * from aws_ec2_instance where instance_id = $1",
	})

	_, err := hr.Run(ctx, input)
	assert.NotNil(err)

	fpErr := err.(fperr.ErrorModel)
	assert.Equal("Query input must define args if the sql has placeholders", fpErr.Detail)
	assert.Equal(400, fpErr.Status)
}

func TestQueryMissingConnectionString(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := Query{}

	input := types.Input(map[string]interface{}{
		schema.AttributeTypeSql:  "SELECT * from aws_ec2_instance where instance_id = $1",
		schema.AttributeTypeArgs: []interface{}{"i-000a000b0000c00d1"},
	})

	_, err := hr.Run(ctx, input)
	assert.NotNil(err)

	fpErr := err.(fperr.ErrorModel)
	assert.Equal("Query input must define connection_string", fpErr.Detail)
	assert.Equal(400, fpErr.Status)
}
