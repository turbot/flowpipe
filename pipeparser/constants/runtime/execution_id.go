package runtime

import (
	"fmt"
	"time"

	"github.com/turbot/go-kit/helpers"
)

var (
	ExecutionID = helpers.GetMD5Hash(fmt.Sprintf("%d", time.Now().Nanosecond()))[:4]
	// PgClientAppName                    = fmt.Sprintf("%s_%s", constants.AppName, ExecutionID)
	PgClientAppNamePluginManagerPrefix = "pm_"
)
