package exa

// ContentsRequest represents a request to the /contents endpoint
type ContentsRequest struct {
	// Array of URLs to crawl (required)
	URLs []string `json:"urls"`

	// Deprecated: use 'urls' instead
	IDs []string `json:"ids,omitempty"`

	// Text retrieval options
	Text *TextOptions `json:"text,omitempty"`

	// Highlights extraction options
	Highlights *HighlightsOptions `json:"highlights,omitempty"`

	// Summary generation options
	Summary *SummaryOptions `json:"summary,omitempty"`

	// Deprecated: Use maxAgeHours instead
	Livecrawl string `json:"livecrawl,omitempty"`

	// The timeout for livecrawling in milliseconds (default: 10000)
	LivecrawlTimeout int `json:"livecrawlTimeout,omitempty"`

	// Maximum age of cached content in hours
	MaxAgeHours int `json:"maxAgeHours,omitempty"`

	// The number of subpages to crawl (default: 0)
	Subpages int `json:"subpages,omitempty"`

	// Keyword(s) to find specific subpages
	SubpageTarget any `json:"subpageTarget,omitempty"`

	// Extra parameters like links and imageLinks
	Extras *ExtrasOptions `json:"extras,omitempty"`

	// Deprecated: Use highlights or text instead
	Context any `json:"context,omitempty"`
}

// ContentsResponse represents the response from the /contents endpoint
type ContentsResponse struct {
	// Unique identifier for the request
	RequestID string `json:"requestId"`

	// Array of results with content
	Results []Result `json:"results"`

	// Deprecated: Combined context string from search results
	Context string `json:"context,omitempty"`

	// Status information for each requested URL
	Statuses []ContentStatus `json:"statuses"`

	// Cost information for the request
	CostDollars *CostDollars `json:"costDollars,omitempty"`
}

// ContentStatus represents the status of content retrieval for a URL
type ContentStatus struct {
	// The URL that was requested
	ID string `json:"id"`

	// Status of the content fetch operation: success or error
	Status string `json:"status"`

	// Error details, only present when status is "error"
	Error *ContentError `json:"error,omitempty"`
}

// ContentError represents error details for content retrieval
type ContentError struct {
	// Specific error type
	Tag string `json:"tag"`

	// The corresponding HTTP status code
	HTTPStatusCode *int `json:"httpStatusCode,omitempty"`
}
