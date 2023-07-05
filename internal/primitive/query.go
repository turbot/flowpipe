package primitive

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/turbot/flowpipe/fperr"
	"github.com/turbot/flowpipe/internal/fplog"
	"github.com/turbot/flowpipe/internal/types"
)

type Query struct{}

func (e *Query) ValidateInput(ctx context.Context, i types.Input) error {
	if i["sql"] == nil {
		return fperr.BadRequestWithMessage("Query input must define sql")
	}
	return nil
}

func (e *Query) Run(ctx context.Context, input types.Input) (*types.StepOutput, error) {
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
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	finish := time.Now().UTC()
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
		"rows":        results,
		"started_at":  start,
		"finished_at": finish,
	}

	return output, nil
}
