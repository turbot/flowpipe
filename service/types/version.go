package types

// APIVersionRequestURI defines the requested API version.
type APIVersionRequestURI struct {
	APIVersion string `uri:"api_version" binding:"required,flowpipe_api_version"`
}
