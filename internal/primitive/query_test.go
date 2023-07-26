package primitive

import (
	"context"
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

func TestQueryWithMissingAttributeSql(t *testing.T) {
	ctx := context.Background()
	ctx = fplog.ContextWithLogger(ctx)

	assert := assert.New(t)
	hr := Query{
		Setting: "go-sqlmock",
	}

	input := types.Input(map[string]interface{}{})

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
