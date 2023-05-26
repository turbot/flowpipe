package types

// APIVersionRequestURI defines the requested API version.
type APIVersionRequestURI struct {
	APIVersion string `uri:"api_version" binding:"required,flowpipe_api_version"`
}

type ListRequestQuery struct {
	NextToken string `json:"next_token" form:"next_token" binding:"omitempty"`
	Limit     *int   `json:"limit,omitempty" form:"limit" binding:"omitempty,steampipe_limit"`
}
