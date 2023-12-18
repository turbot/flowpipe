package trigger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/zclconf/go-cty/cty"
)

func TestTriggerQuery(t *testing.T) {
	ctx := context.Background()

	assert := assert.New(t)

	pipeline := map[string]cty.Value{
		"name": cty.StringVal("test"),
	}

	trigger := &modconfig.Trigger{
		HclResourceImpl: modconfig.HclResourceImpl{
			FullName: "foo.bar.baz",
		},
		Pipeline: cty.ObjectVal(pipeline),
	}
	trigger.Config = &modconfig.TriggerQuery{
		ConnectionString: "postgres://steampipe@host.docker.internal:9193/steampipe",
		Sql:              "select * from hackernews.hackernews_new",
		PrimaryKey:       "id",
	}
	triggerRunner := NewTriggerRunner(ctx, nil, trigger)

	assert.NotNil(triggerRunner, "trigger runner should not be nil")

	triggerRunner.Run()
}
