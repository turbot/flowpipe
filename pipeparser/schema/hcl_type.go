package schema

import "github.com/turbot/go-kit/helpers"

// NOTE: when adding a block type, be sure to update  QueryProviderBlocks/ReferenceBlocks/AllBlockTypes as needed
const (
	// require blocks
	BlockTypeSteampipe = "steampipe"
	BlockTypeMod       = "mod"
	BlockTypePlugin    = "plugin"
	// resource blocks
	BlockTypeQuery          = "query"
	BlockTypeControl        = "control"
	BlockTypeBenchmark      = "benchmark"
	BlockTypeDashboard      = "dashboard"
	BlockTypeContainer      = "container"
	BlockTypeChart          = "chart"
	BlockTypeCard           = "card"
	BlockTypeFlow           = "flow"
	BlockTypeGraph          = "graph"
	BlockTypeHierarchy      = "hierarchy"
	BlockTypeImage          = "image"
	BlockTypeInput          = "input"
	BlockTypeTable          = "table"
	BlockTypeText           = "text"
	BlockTypeLocals         = "locals"
	BlockTypeVariable       = "variable"
	BlockTypeParam          = "param"
	BlockTypeRequire        = "require"
	BlockTypeNode           = "node"
	BlockTypeEdge           = "edge"
	BlockTypeLegacyRequires = "requires"
	BlockTypeCategory       = "category"
	BlockTypeWith           = "with"
	BlockTypeError          = "error"

	// config blocks
	BlockTypeConnection       = "connection"
	BlockTypeOptions          = "options"
	BlockTypeWorkspaceProfile = "workspace"
	BlockTypePipeline         = "pipeline"
	BlockTypePipelineStep     = "step"
	BlockTypePipelineOutput   = "output"

	AttributeTypeValue     = "value"
	AttributeTypeSensitive = "sensitive"

	// Pipeline blocks
	BlockTypePipelineStepHttp     = "http"
	BlockTypePipelineStepSleep    = "sleep"
	BlockTypePipelineStepEmail    = "email"
	BlockTypePipelineStepEcho     = "echo"
	BlockTypePipelineStepQuery    = "query"
	BlockTypePipelineStepExec     = "exec"
	BlockTypePipelineStepPipeline = "pipeline"

	// error block
	AttributeTypeIgnore  = "ignore"
	AttributeTypeRetries = "retries"

	// Common step attributes
	AttributeTypeTitle       = "title"
	AttributeTypeDependsOn   = "depends_on"
	AttributeTypeForEach     = "for_each"
	AttributeTypeDescription = "description"
	AttributeTypeIf          = "if"

	AttributeTypeStartedAt  = "started_at"
	AttributeTypeFinishedAt = "finished_at"

	// Used by query step
	AttributeTypeSql              = "sql"
	AttributeTypeArgs             = "args"
	AttributeTypeQuery            = "query"
	AttributeTypeRows             = "rows"
	AttributeTypeConnectionString = "connection_string"

	// Used by email step
	AttributeTypeBcc              = "bcc"
	AttributeTypeBody             = "body"
	AttributeTypeCc               = "cc"
	AttributeTypeContentType      = "content_type"
	AttributeTypeFrom             = "from"
	AttributeTypeHost             = "host"
	AttributeTypePort             = "port"
	AttributeTypeSenderCredential = "sender_credential" //nolint:gosec // Getting Potential hardcoded credentials warning
	AttributeTypeSenderName       = "sender_name"
	AttributeTypeSubject          = "subject"
	AttributeTypeTo               = "to"

	// Used by sleep step
	AttributeTypeDuration = "duration"

	// Used by http step
	AttributeTypeUrl              = "url"
	AttributeTypeMethod           = "method"
	AttributeTypeRequestBody      = "request_body"
	AttributeTypeRequestHeaders   = "request_headers"
	AttributeTypeRequestTimeoutMs = "request_timeout_ms"
	AttributeTypeCaCertPem        = "ca_cert_pem"
	AttributeTypeInsecure         = "insecure"
	AttributeTypeResponseHeaders  = "response_headers"
	AttributeTypeResponseBody     = "response_body"
	AttributeTypeStatusCode       = "status_code"
	AttributeTypeStatus           = "status"

	// Used by echo step
	AttributeTypeText = "text"
	AttributeTypeJson = "json"

	// Used byy Pipeline step
	AttributeTypePipeline = "pipeline"

	AttributeTypeMessage = "message"

	LabelName = "name"
	LabelType = "type"
)

// QueryProviderBlocks is a list of block types which implement QueryProvider
var QueryProviderBlocks = []string{
	BlockTypeCard,
	BlockTypeChart,
	BlockTypeControl,
	BlockTypeEdge,
	BlockTypeFlow,
	BlockTypeGraph,
	BlockTypeHierarchy,
	BlockTypeImage,
	BlockTypeInput,
	BlockTypeQuery,
	BlockTypeNode,
	BlockTypeTable,
	BlockTypeWith,
}

// NodeAndEdgeProviderBlocks is a list of block types which implementnodeAndEdgeProvider
var NodeAndEdgeProviderBlocks = []string{
	BlockTypeHierarchy,
	BlockTypeFlow,
	BlockTypeGraph,
}

// ReferenceBlocks is a list of block types we store references for
var ReferenceBlocks = []string{
	BlockTypeMod,
	BlockTypeQuery,
	BlockTypeControl,
	BlockTypeBenchmark,
	BlockTypeDashboard,
	BlockTypeContainer,
	BlockTypeCard,
	BlockTypeChart,
	BlockTypeFlow,
	BlockTypeGraph,
	BlockTypeHierarchy,
	BlockTypeImage,
	BlockTypeInput,
	BlockTypeTable,
	BlockTypeText,
	BlockTypeParam,
	BlockTypeCategory,
	BlockTypeWith,
}

var ValidResourceItemTypes = []string{
	BlockTypeMod,
	BlockTypeQuery,
	BlockTypeControl,
	BlockTypeBenchmark,
	BlockTypeDashboard,
	BlockTypeContainer,
	BlockTypeChart,
	BlockTypeCard,
	BlockTypeFlow,
	BlockTypeGraph,
	BlockTypeHierarchy,
	BlockTypeImage,
	BlockTypeInput,
	BlockTypeTable,
	BlockTypeText,
	BlockTypeLocals,
	BlockTypeVariable,
	BlockTypeParam,
	BlockTypeRequire,
	BlockTypeNode,
	BlockTypeEdge,
	BlockTypeLegacyRequires,
	BlockTypeCategory,
	BlockTypeConnection,
	BlockTypeOptions,
	BlockTypeWorkspaceProfile,
	BlockTypePipeline,
	BlockTypeWith,
	// local is not an actual block name but is a resource type
	"local",
	// references
	"ref",
	// variables
	"var",
}

func IsValidResourceItemType(blockType string) bool {
	return helpers.StringSliceContains(ValidResourceItemTypes, blockType)
}
