package exa

// FindSimilarRequest represents a request to the /findSimilar endpoint
type FindSimilarRequest struct {
	// The url for which you would like to find similar links (required)
	URL string `json:"url"`

	// If true, excludes links from the same domain as the provided URL from the results
	ExcludeSourceDomain bool `json:"excludeSourceDomain,omitempty"`

	// Number of results to return (default: 10, max: 100)
	NumResults int `json:"numResults,omitempty"`

	// List of domains to include in the search
	IncludeDomains []string `json:"includeDomains,omitempty"`

	// List of domains to exclude from search results
	ExcludeDomains []string `json:"excludeDomains,omitempty"`

	// Crawl date after which results will be included (ISO 8601)
	StartCrawlDate string `json:"startCrawlDate,omitempty"`

	// Crawl date before which results will be included (ISO 8601)
	EndCrawlDate string `json:"endCrawlDate,omitempty"`

	// Only links with a published date after this will be returned (ISO 8601)
	StartPublishedDate string `json:"startPublishedDate,omitempty"`

	// Only links with a published date before this will be returned (ISO 8601)
	EndPublishedDate string `json:"endPublishedDate,omitempty"`

	// List of strings that must be present in webpage text of results
	IncludeText []string `json:"includeText,omitempty"`

	// List of strings that must not be present in webpage text of results
	ExcludeText []string `json:"excludeText,omitempty"`

	// Deprecated: Use highlights or text instead
	Context any `json:"context,omitempty"`

	// Enable content moderation to filter unsafe content
	Moderation bool `json:"moderation,omitempty"`

	// Contents options for retrieving text, highlights, summaries, etc.
	Contents *ContentsOptions `json:"contents,omitempty"`
}

// FindSimilarResponse represents the response from the /findSimilar endpoint
type FindSimilarResponse struct {
	// Unique identifier for the request
	RequestID string `json:"requestId"`

	// Deprecated: Combined context string from search results
	Context string `json:"context,omitempty"`

	// A list of search results containing title, URL, published date, and author
	Results []Result `json:"results"`

	// Cost information for the request
	CostDollars *CostDollars `json:"costDollars,omitempty"`
}
