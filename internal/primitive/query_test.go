package primitive

import (
	"context"
	"testing"
	"time"

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
		schema.AttributeTypeDatabase: "sqlite:./database_files/employee.db",
		schema.AttributeTypeSql:      "select * from employee order by id;",
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
		schema.AttributeTypeDatabase: "sqlite:database_files/employee.db",
		schema.AttributeTypeSql:      "select * from employee where id = $1;",
		schema.AttributeTypeArgs:     []interface{}{10},
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
		schema.AttributeTypeDatabase: "sqlite:database_files/employee.db",
		schema.AttributeTypeSql:      "SELECT * from employee where name like $1;",
		schema.AttributeTypeArgs:     []interface{}{"Jo%"},
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
		schema.AttributeTypeDatabase: "sqlite:database_files/employee.db",
		schema.AttributeTypeSql:      "select * from user;",
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
		schema.AttributeTypeDatabase: "sqlite:database_files/employee.db",
		schema.AttributeTypeSql:      "select * from department;",
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
		schema.AttributeTypeDatabase: "sqlite:database_files/employee.db",
		schema.AttributeTypeSql:      "SELECT * employee;",
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
		schema.AttributeTypeDatabase: "this is a connection string",
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
		schema.AttributeTypeDatabase: "this is a connection string",
		"sql1":                       "select * from employee;",
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
		schema.AttributeTypeDatabase: "this is a connection string",

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
	assert.Equal("Query input must define database", fpErr.Detail)
	assert.Equal(400, fpErr.Status)
}

func TestQueryDuckDB(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	hr := Query{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeDatabase: "duckdb:./database_files/new_database.duckdb",
		schema.AttributeTypeSql:      "select * from employee order by id;",
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
		schema.AttributeTypeDatabase: "sqlite:database_files/employee.db",
		schema.AttributeTypeSql:      "select * from employee order by id;",
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
		schema.AttributeTypeDatabase: "abcd",
		schema.AttributeTypeSql:      "select * from employee order by id;",
	})

	_, err := hr.Run(ctx, input)
	assert.NotNil(err)
	assert.Contains(err.Error(), "Bad Request: Invalid database connection string")
}

func XTestQueryMariaDB(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)
	hr := Query{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeDatabase: "mysql://root:flowpipe@tcp(localhost:3306)/flowpipe_test",
		schema.AttributeTypeSql:      "select * from DataTypeDemo;",
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.NotNil(output)
	/**
	mariadb -u root -pflowpipe

	create database flowpipe_test;

	use flowpipe_test;

	CREATE TABLE DataTypeDemo (
	    id INT AUTO_INCREMENT,
	    sample_int INT,
	    sample_varchar VARCHAR(50),
	    sample_text TEXT,
	    sample_date DATE,
	    sample_datetime DATETIME,
	    sample_float FLOAT,
	    sample_double DOUBLE,
	    sample_decimal DECIMAL(10,2),
	    sample_bool BOOLEAN, sample_json JSON, sample_blob BLOB, PRIMARY KEY (id) )

	-- Insert statement 1
	INSERT INTO DataTypeDemo
	(sample_int, sample_varchar, sample_text, sample_date, sample_datetime, sample_float, sample_double, sample_decimal, sample_bool, sample_json, sample_blob)
	VALUES
	(4, 'Example 1', 'Text for example 1.', '2024-03-01', '2024-03-01 12:00:00', 1.23, 456.789, 100.10, FALSE, '{"name": "John", "age": 30, "city": "New York"}', CAST('Binary data example 1' AS BINARY));

	-- Insert statement 2
	INSERT INTO DataTypeDemo
	(sample_int, sample_varchar, sample_text, sample_date, sample_datetime, sample_float, sample_double, sample_decimal, sample_bool, sample_json, sample_blob)
	VALUES
	(5, 'Example 2', 'Text for example 2.', '2024-04-10', '2024-04-10 15:30:00', 9.87, 321.654, 200.20, TRUE, '{"product": "Table", "price": 150.75}', CAST('Binary data example 2' AS BINARY));

	-- Insert statement 3
	INSERT INTO DataTypeDemo
	(sample_int, sample_varchar, sample_text, sample_date, sample_datetime, sample_float, sample_double, sample_decimal, sample_bool, sample_json, sample_blob)
	VALUES
	(6, 'Example 3', 'Text for example 3.', '2024-05-20', '2024-05-20 18:45:00', 6.54, 987.321, 300.30, FALSE, '{"animal": "Dog", "breed": "Labrador"}', CAST('Binary data example 3' AS BINARY));

	INSERT INTO DataTypeDemo
	(sample_int, sample_varchar, sample_text, sample_date, sample_datetime, sample_float, sample_double, sample_decimal, sample_bool, sample_json, sample_blob)
	VALUES
	(7, 'Example 4', 'Text for example 4.', '2024-05-20', '2024-05-20 18:45:00', 6.54, 987.321, 300.30, FALSE, '{"animal": "Dog", "breed": "Labrador"}', CAST('Binary data example 4' AS BINARY));


	*/
}

func TestMariaDBQueryListAll(t *testing.T) {
	ctx := context.Background()

	connectionString := "flowpipe:password@tcp(localhost:3306)/flowpipe-test"
	err := ApplyDatabaseScript(DriverMySQL, connectionString, "./database_files/mariadb_populate_data.sql")
	if err != nil {
		t.Fatalf("Error setting up the database: " + err.Error())
	}

	assert := assert.New(t)
	hr := Query{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeConnectionString: "mysql://" + connectionString,
		schema.AttributeTypeSql:              "select * from employee order by id;",
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal(3, len(output.Get(schema.AttributeTypeRows).([]map[string]interface{})))

	expectedResult := []map[string]interface{}{
		{
			"id":                      int64(1),
			"name":                    "John",
			"email":                   "john@example.com",
			"preferences":             map[string]interface{}{"theme": "dark", "notifications": true},
			"salary":                  float64(50000),
			"birth_date":              time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC),
			"hire_datetime":           time.Date(2020, time.January, 1, 8, 30, 0, 0, time.UTC),
			"part_time":               int64(1),
			"biography":               "John has been a part of our company for over a decade...",
			"profile_picture":         nil,
			"last_login":              time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC),
			"vacation_days":           int64(10),
			"contract_length":         int64(12),
			"employee_number":         int64(100001),
			"working_hours":           time.Date(0, time.January, 1, 9, 0, 0, 0, time.UTC), // Zero year for time.Time denotes a Time-of-Day value
			"yearly_bonus":            float64(3000),
			"employee_code":           "EMP00001",
			"health_status":           "good",
			"security_level":          int64(3),
			"resume":                  nil,
			"linkedin_url":            "https://linkedin.com/in/john",
			"personal_website":        "https://johnsblog.com",
			"notes":                   "John has consistently performed well.",
			"office_location":         "\x00\x00\x00\x00\x01\x01\x00\x00\x00\xaa\xf1\xd2Mb\x80R\xc0^K\xc8\a=[D@",
			"department_id":           int64(1),
			"fingerprint":             nil,
			"schedule":                nil, // NOTE: The schedule appears to be nil in the provided structure
			"last_performance_review": int64(2023),
			"nationality":             "USA",
			"languages":               map[string]interface{}{"English": "fluent", "Spanish": "intermediate"},
			"hire_date_year":          int64(2020),
		},
		{
			"id":                      int64(2),
			"name":                    "Adam",
			"email":                   "adam@example.com",
			"preferences":             map[string]interface{}{"theme": "dark", "notifications": true},
			"salary":                  float64(52000),
			"birth_date":              time.Date(1982, time.May, 12, 0, 0, 0, 0, time.UTC),
			"hire_datetime":           time.Date(2020, time.March, 15, 9, 0, 0, 0, time.UTC),
			"part_time":               int64(0),
			"biography":               "Adam is a recent addition to the team...",
			"profile_picture":         nil,
			"last_login":              time.Date(2023, time.February, 2, 14, 30, 0, 0, time.UTC),
			"vacation_days":           int64(15),
			"contract_length":         int64(24),
			"employee_number":         int64(100002),
			"working_hours":           time.Date(0, time.January, 1, 10, 0, 0, 0, time.UTC), // Zero year for time.Time denotes a Time-of-Day value
			"yearly_bonus":            float64(2500),
			"employee_code":           "EMP00002",
			"health_status":           "excellent",
			"security_level":          int64(4),
			"resume":                  nil,
			"linkedin_url":            "https://linkedin.com/in/adam",
			"personal_website":        "https://adamportfolio.com",
			"notes":                   "Adam brings fresh perspectives.",
			"office_location":         "\x00\x00\x00\x00\x01\x01\x00\x00\x00\xa0\x1a/\xdd$\x06\x10\xc0w-!\x1f\xf4l)@",
			"department_id":           int64(2),
			"fingerprint":             nil,
			"schedule":                nil, // NOTE: The schedule appears to be nil in the provided structure
			"last_performance_review": int64(2023),
			"nationality":             "CAN",
			"languages":               map[string]interface{}{"French": "fluent", "English": "fluent"},
			"hire_date_year":          int64(2020),
		},
		{
			"id":                      int64(16),
			"name":                    "Diana",
			"email":                   "diana@example.com",
			"preferences":             map[string]interface{}{"theme": "dark", "notifications": true},
			"salary":                  float64(55000),
			"birth_date":              time.Date(1990, time.April, 5, 0, 0, 0, 0, time.UTC),
			"hire_datetime":           time.Date(2021, time.April, 15, 9, 30, 0, 0, time.UTC),
			"part_time":               int64(0),
			"biography":               "Diana is known for her attention to detail...",
			"profile_picture":         nil,
			"last_login":              time.Date(2024, time.January, 20, 10, 0, 0, 0, time.UTC),
			"vacation_days":           int64(20),
			"contract_length":         int64(36),
			"employee_number":         int64(100016),
			"working_hours":           time.Date(0, time.January, 1, 8, 0, 0, 0, time.UTC), // Zero year for time.Time denotes a Time-of-Day value
			"yearly_bonus":            float64(4500),
			"employee_code":           "EMP00016",
			"health_status":           "excellent",
			"security_level":          int64(5),
			"resume":                  nil,
			"linkedin_url":            "https://linkedin.com/in/diana",
			"personal_website":        "https://dianasportfolio.com",
			"notes":                   "Diana has led several successful projects.",
			"office_location":         "\x00\x00\x00\x00\x01\x01\x00\x00\x00P\x8d\x97n\x12\x03,@\xaf%䃞-T\xc0",
			"department_id":           int64(3),
			"fingerprint":             nil,
			"schedule":                nil, // Assuming 'schedule' is not set for Diana, thus nil
			"last_performance_review": int64(2024),
			"nationality":             "GBR",
			"languages":               map[string]interface{}{"English": "fluent", "French": "basic"},
			"hire_date_year":          int64(2021),
		},
	}

	expectedRow := output.Get(schema.AttributeTypeRows).([]map[string]interface{})
	assert.Equal(expectedResult, expectedRow)
}

func TestPostgresSQLQueryListAll(t *testing.T) {
	ctx := context.Background()

	connectionString := "postgres://flowpipe:password@localhost:5432/flowpipe-test?sslmode=disable"

	err := ApplyDatabaseScript(DriverPostgres, connectionString, "./database_files/postgres_populate_data.sql")
	if err != nil {
		t.Fatalf("Error setting up the database: " + err.Error())
	}

	assert := assert.New(t)
	hr := Query{}

	input := modconfig.Input(map[string]interface{}{
		schema.AttributeTypeConnectionString: connectionString,
		schema.AttributeTypeSql:              "select * from employee order by id;",
	})

	output, err := hr.Run(ctx, input)
	assert.Nil(err)
	assert.Equal(3, len(output.Get(schema.AttributeTypeRows).([]map[string]interface{})))

	location, err := time.LoadLocation("Etc/UTC")
	if err != nil {
		t.Fatal(err)
	}

	expectedResult := []map[string]interface{}{
		{
			"id":                      int64(1),
			"name":                    "John",
			"email":                   "john@example.com",
			"preferences":             map[string]interface{}{"theme": "dark", "notifications": true},
			"salary":                  "50000.00",
			"birth_date":              time.Date(1980, time.January, 1, 0, 0, 0, 0, time.FixedZone("", 0)),
			"hire_datetime":           time.Date(2020, time.January, 1, 6, 30, 0, 0, location),
			"part_time":               true,
			"biography":               "John has been a part of our company for over a decade...",
			"profile_picture":         nil,
			"last_login":              time.Date(2023, time.January, 1, 12, 0, 0, 0, time.FixedZone("", 0)),
			"vacation_days":           int64(10),
			"contract_length":         int64(12),
			"employee_number":         int64(100001),
			"office_location":         nil,
			"working_hours":           "08:00:00",
			"yearly_bonus":            float64(3000),
			"employee_code":           "EMP00001  ",
			"health_status":           "good",
			"security_level":          int64(3),
			"resume":                  nil,
			"linkedin_url":            "https://linkedin.com/in/john",
			"personal_website":        "https://johnsblog.com",
			"notes":                   "John has consistently performed well.",
			"department_id":           int64(1),
			"fingerprint":             nil,
			"schedule":                `{morning,afternoon}`,
			"last_performance_review": time.Date(2023, time.January, 1, 0, 0, 0, 0, time.FixedZone("", 0)),
			"nationality":             "USA",
			"languages":               map[string]interface{}{"English": "fluent", "Spanish": "intermediate"},
			"hire_date_year":          int64(2020),
		},
		{
			"id":                      int64(2),
			"name":                    "Adam",
			"email":                   "adam@example.com",
			"preferences":             map[string]interface{}{"theme": "dark", "notifications": true},
			"salary":                  "52000.00",
			"birth_date":              time.Date(1982, time.May, 12, 0, 0, 0, 0, time.FixedZone("", 0)), // NOTE: birth_date value is suspect
			"hire_datetime":           time.Date(2020, time.March, 15, 9, 0, 0, 0, location),
			"part_time":               false,
			"biography":               "Adam is a recent addition to the team...",
			"profile_picture":         nil,
			"last_login":              time.Date(2023, time.February, 2, 14, 30, 0, 0, time.FixedZone("", 0)), // NOTE: last_login value is suspect
			"vacation_days":           int64(15),
			"contract_length":         int64(24),
			"employee_number":         int64(100002),
			"office_location":         nil,
			"working_hours":           "09:00:00",
			"yearly_bonus":            float64(2500),
			"employee_code":           "EMP00002  ",
			"health_status":           "excellent",
			"security_level":          int64(4),
			"resume":                  nil,
			"linkedin_url":            "https://linkedin.com/in/adam",
			"personal_website":        "https://adamportfolio.com",
			"notes":                   "Adam brings fresh perspectives.",
			"department_id":           int64(2),
			"fingerprint":             nil,               // Assuming bytea returns nil when not set
			"schedule":                `{morning,night}`, // Assuming your data layer converts PG arrays to Go slices
			"last_performance_review": time.Date(2023, time.February, 2, 0, 0, 0, 0, time.FixedZone("", 0)),
			"nationality":             "CAN",
			"languages":               map[string]interface{}{"French": "fluent", "English": "fluent"},
			"hire_date_year":          int64(2020),
		},
		{
			"id":                      int64(3),
			"name":                    "Diana",
			"email":                   "diana@example.com",
			"preferences":             map[string]interface{}{"theme": "dark", "notifications": true},
			"salary":                  "55000.00",
			"birth_date":              time.Date(1990, time.April, 5, 0, 0, 0, 0, time.FixedZone("", 0)), // NOTE: birth_date value is suspect
			"hire_datetime":           time.Date(2021, time.April, 15, 9, 30, 0, 0, location),
			"part_time":               false,
			"biography":               "Diana is known for her attention to detail...",
			"profile_picture":         nil,
			"last_login":              time.Date(2024, time.January, 20, 10, 0, 0, 0, time.FixedZone("", 0)), // NOTE: last_login value is suspect
			"vacation_days":           int64(20),
			"contract_length":         int64(36),
			"employee_number":         int64(100016),
			"office_location":         nil,
			"working_hours":           "08:00:00",
			"yearly_bonus":            float64(4500),
			"employee_code":           "EMP00016  ",
			"health_status":           "excellent",
			"security_level":          int64(5),
			"resume":                  nil,
			"linkedin_url":            "https://linkedin.com/in/diana",
			"personal_website":        "https://dianasportfolio.com",
			"notes":                   "Diana has led several successful projects.",
			"department_id":           int64(3),
			"fingerprint":             nil,                 // Assuming bytea returns nil when not set
			"schedule":                `{afternoon,night}`, // Assuming your data layer converts PG arrays to Go slices
			"last_performance_review": time.Date(2024, time.January, 20, 0, 0, 0, 0, time.FixedZone("", 0)),
			"nationality":             "GBR",
			"languages":               map[string]interface{}{"English": "fluent", "French": "basic"},
			"hire_date_year":          int64(2021),
		},
	}

	expectedRow := output.Get(schema.AttributeTypeRows).([]map[string]interface{})
	assert.Equal(expectedResult, expectedRow)
}
