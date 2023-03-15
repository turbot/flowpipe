package primitive

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/turbot/steampipe-pipelines/pipeline"
)

type Query struct{}

func (e *Query) ValidateInput(ctx context.Context, i pipeline.StepInput) error {
	if i["sql"] == nil {
		return errors.New("Query input must define sql")
	}
	return nil
}

func (e *Query) Run(ctx context.Context, input pipeline.StepInput) (*pipeline.Output, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	db, err := sqlx.Connect("postgres", "postgres://steampipe@localhost:9193/steampipe")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	sql := input["sql"].(string)

	results := []map[string]interface{}{}

	start := time.Now().UTC()
	rows, err := db.Queryx(sql)
	finish := time.Now().UTC()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		row := make(map[string]interface{})
		err = rows.MapScan(row)
		if err != nil {
			return nil, err
		}
		// sqlx doesn't handle jsonb columns, so we need to do it manually
		// https://github.com/jmoiron/sqlx/issues/225
		for k, encoded := range row {
			switch ba := encoded.(type) {
			case []byte:
				var col interface{}
				json.Unmarshal(ba, &col)
				row[k] = col
			}
		}
		results = append(results, row)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	output := &pipeline.Output{
		"rows":        results,
		"started_at":  start,
		"finished_at": finish,
	}

	return output, nil
}
