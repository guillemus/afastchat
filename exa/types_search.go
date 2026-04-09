package exa

// SearchRequest represents a request to the /search endpoint
type SearchRequest struct {
	// The query string for the search (required)
	Query string `json:"query"`

	// Additional query variations for deep search. Only works with type="deep" or type="deep-reasoning"
	AdditionalQueries []string `json:"additionalQueries,omitempty"`

	// The type of search: neural, fast, auto (default), deep, deep-reasoning, or instant
	Type string `json:"type,omitempty"`

	// JSON schema for deep search structured output mode
	OutputSchema map[string]any `json:"outputSchema,omitempty"`

	// A data category to focus on: company, research paper, news, pdf, github, personal site, people, financial report
	Category string `json:"category,omitempty"`

	// The two-letter ISO country code of the user, e.g. US
	UserLocation string `json:"userLocation,omitempty"`

	// Number of results to return (up to thousands for custom plans, default: 10, max: 100)
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

	// List of strings that must be present in webpage text of results (currently only 1 string up to 5 words)
	IncludeText []string `json:"includeText,omitempty"`

	// List of strings that must not be present in webpage text of results (currently only 1 string up to 5 words)
	ExcludeText []string `json:"excludeText,omitempty"`

	// Deprecated: Use highlights or text instead
	Context any `json:"context,omitempty"`

	// Enable content moderation to filter unsafe content
	Moderation bool `json:"moderation,omitempty"`

	// Contents options for retrieving text, highlights, summaries, etc.
	Contents *ContentsOptions `json:"contents,omitempty"`
}

// ContentsOptions represents options for retrieving content from search results
type ContentsOptions struct {
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
	// Positive: use cache if less than this many hours old
	// 0: always livecrawl
	// -1: never livecrawl
	// Omit: livecrawl as fallback only
	MaxAgeHours int `json:"maxAgeHours,omitempty"`

	// The number of subpages to crawl (default: 0)
	Subpages int `json:"subpages,omitempty"`

	// Keyword(s) to find specific subpages (single string or array)
	SubpageTarget any `json:"subpageTarget,omitempty"`

	// Extra parameters like links and imageLinks
	Extras *ExtrasOptions `json:"extras,omitempty"`
}

// TextOptions represents options for text extraction
type TextOptions struct {
	// Maximum character limit for the full page text
	MaxCharacters int `json:"maxCharacters,omitempty"`

	// Include HTML tags in the response
	IncludeHTMLTags bool `json:"includeHtmlTags,omitempty"`

	// Verbosity level: compact, standard, or full
	Verbosity string `json:"verbosity,omitempty"`

	// Only include content from these semantic page sections
	IncludeSections []string `json:"includeSections,omitempty"`

	// Exclude content from these semantic page sections
	ExcludeSections []string `json:"excludeSections,omitempty"`
}

// HighlightsOptions represents options for highlight extraction
type HighlightsOptions struct {
	// Maximum number of characters to return for highlights
	MaxCharacters int `json:"maxCharacters,omitempty"`

	// Deprecated: use maxCharacters instead
	NumSentences int `json:"numSentences,omitempty"`

	// Deprecated: use maxCharacters instead
	HighlightsPerURL int `json:"highlightsPerUrl,omitempty"`

	// Custom query to direct the LLM's selection of highlights
	Query string `json:"query,omitempty"`
}

// SummaryOptions represents options for summary generation
type SummaryOptions struct {
	// Custom query for the LLM-generated summary
	Query string `json:"query,omitempty"`

	// JSON schema for structured output from summary
	Schema map[string]any `json:"schema,omitempty"`
}

// ExtrasOptions represents extra parameters for content retrieval
type ExtrasOptions struct {
	// Number of URLs to return from each webpage
	Links int `json:"links,omitempty"`

	// Number of images to return for each result
	ImageLinks int `json:"imageLinks,omitempty"`
}

// SearchResponse represents the response from the /search endpoint
type SearchResponse struct {
	// Unique identifier for the request
	RequestID string `json:"requestId"`

	// A list of search results containing title, URL, published date, and author
	Results []Result `json:"results"`

	// For auto searches, indicates which search type was selected: neural, deep, or deep-reasoning
	SearchType string `json:"searchType,omitempty"`

	// Deprecated: Combined context string from search results. Use highlights or text instead
	Context string `json:"context,omitempty"`

	// Deep-search synthesized output. Returned for deep search variants.
	Output *DeepSearchOutput `json:"output,omitempty"`

	// Cost information for the request
	CostDollars *CostDollars `json:"costDollars,omitempty"`
}

// Result represents a single search result with optional content
type Result struct {
	// The title of the search result
	Title string `json:"title"`

	// The URL of the search result
	URL string `json:"url"`

	// An estimate of the creation date, from parsing HTML content
	PublishedDate *string `json:"publishedDate,omitempty"`

	// If available, the author of the content
	Author *string `json:"author,omitempty"`

	// A number from 0 to 1 representing similarity between the query/url and the result
	Score *float64 `json:"score,omitempty"`

	// The temporary ID for the document
	ID string `json:"id"`

	// The URL of an image associated with the search result
	Image *string `json:"image,omitempty"`

	// The URL of the favicon for the search result's domain
	Favicon *string `json:"favicon,omitempty"`

	// The full content text of the search result (when contents requested)
	Text string `json:"text,omitempty"`

	// Array of highlights extracted from the search result content
	Highlights []string `json:"highlights,omitempty"`

	// Array of cosine similarity scores for each highlight
	HighlightScores []float64 `json:"highlightScores,omitempty"`

	// Summary of the webpage
	Summary string `json:"summary,omitempty"`

	// Array of subpages for the search result
	Subpages []Result `json:"subpages,omitempty"`

	// Results from extras (links, entities)
	Extras *ResultExtras `json:"extras,omitempty"`
}

// ResultExtras represents extra information in a result
type ResultExtras struct {
	// Array of links from the search result
	Links []string `json:"links,omitempty"`

	// Structured entity data for company or person search results
	Entities []Entity `json:"entities,omitempty"`
}

// Entity represents structured entity data
type Entity struct {
	// Exa entity ID
	ID string `json:"id"`

	// Entity type: company or person
	Type string `json:"type"`

	// Schema version number
	Version int `json:"version"`

	// Structured properties based on entity type
	Properties EntityProperties `json:"properties"`
}

// EntityProperties contains properties for company or person entities
type EntityProperties struct {
	// Company-specific properties
	Name        *string `json:"name,omitempty"`
	FoundedYear *int    `json:"foundedYear,omitempty"`
	Description *string `json:"description,omitempty"`

	// Workforce information
	Workforce *WorkforceInfo `json:"workforce,omitempty"`

	// Headquarters information
	Headquarters *HeadquartersInfo `json:"headquarters,omitempty"`

	// Financial information
	Financials *FinancialsInfo `json:"financials,omitempty"`

	// Web traffic information
	WebTraffic *WebTrafficInfo `json:"webTraffic,omitempty"`

	// Person-specific properties
	Location    *string            `json:"location,omitempty"`
	WorkHistory []WorkHistoryEntry `json:"workHistory,omitempty"`
}

// WorkforceInfo represents company workforce information
type WorkforceInfo struct {
	Total *int `json:"total,omitempty"`
}

// HeadquartersInfo represents company headquarters information
type HeadquartersInfo struct {
	Address    *string `json:"address,omitempty"`
	City       *string `json:"city,omitempty"`
	PostalCode *string `json:"postalCode,omitempty"`
	Country    *string `json:"country,omitempty"`
}

// FinancialsInfo represents company financial information
type FinancialsInfo struct {
	RevenueAnnual      *int          `json:"revenueAnnual,omitempty"`
	FundingTotal       *int          `json:"fundingTotal,omitempty"`
	FundingLatestRound *FundingRound `json:"fundingLatestRound,omitempty"`
}

// FundingRound represents a funding round
type FundingRound struct {
	Name   *string `json:"name,omitempty"`
	Date   *string `json:"date,omitempty"`
	Amount *int    `json:"amount,omitempty"`
}

// WebTrafficInfo represents company web traffic information
type WebTrafficInfo struct {
	VisitsMonthly *int `json:"visitsMonthly,omitempty"`
}

// WorkHistoryEntry represents a single work history entry for a person
type WorkHistoryEntry struct {
	Title    *string     `json:"title,omitempty"`
	Location *string     `json:"location,omitempty"`
	Dates    *DateRange  `json:"dates,omitempty"`
	Company  *CompanyRef `json:"company,omitempty"`
}

// DateRange represents a date range
type DateRange struct {
	From *string `json:"from,omitempty"`
	To   *string `json:"to,omitempty"`
}

// CompanyRef represents a reference to a company
type CompanyRef struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

// DeepSearchOutput represents deep search synthesized output
type DeepSearchOutput struct {
	// Deep-search synthesized content. String by default, or object when outputSchema is provided.
	Content any `json:"content"`

	// Field-level grounding for synthesized output
	Grounding []GroundingInfo `json:"grounding"`
}

// GroundingInfo represents field-level grounding information
type GroundingInfo struct {
	// Field path in output.content
	Field string `json:"field"`

	// Sources supporting this output field
	Citations []CitationRef `json:"citations"`

	// Confidence level: low, medium, or high
	Confidence string `json:"confidence"`
}

// CitationRef represents a citation reference
type CitationRef struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

// CostDollars represents cost information for a request
type CostDollars struct {
	// Total dollar cost for your request
	Total float64 `json:"total"`

	// Breakdown of costs by operation type
	Breakdown *CostBreakdown `json:"breakdown,omitempty"`
}

// CostBreakdown represents detailed cost breakdown
type CostBreakdown struct {
	NeuralSearch     float64 `json:"neuralSearch,omitempty"`
	DeepSearch       float64 `json:"deepSearch,omitempty"`
	ContentText      float64 `json:"contentText,omitempty"`
	ContentHighlight float64 `json:"contentHighlight,omitempty"`
	ContentSummary   float64 `json:"contentSummary,omitempty"`
}
