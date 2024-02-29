package primitive

import (
	"time"

	"github.com/turbot/pipe-fittings/schema"
)

func flowpipeMetadataOutput(startedAt, finshedAt time.Time) map[string]interface{} {

	output := map[string]interface{}{
		schema.AttributeTypeStartedAt:  startedAt,
		schema.AttributeTypeFinishedAt: finshedAt,
	}
	return output

}
