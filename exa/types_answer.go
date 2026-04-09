package exa

// AnswerRequest represents a request to the /answer endpoint
type AnswerRequest struct {
	// The question or query to answer (required)
	Query string `json:"query"`

	// If true, the response is returned as a server-sent events (SSE) stream
	Stream bool `json:"stream,omitempty"`

	// If true, the response includes full text content in the search results
	Text bool `json:"text,omitempty"`

	// JSON Schema Draft 7 specification for the desired answer structure
	// When provided, the answer will be returned as a structured object
	OutputSchema *OutputSchema `json:"outputSchema,omitempty"`
}

// OutputSchema represents a JSON schema for structured answer output
type OutputSchema struct {
	// The root schema type (typically "object")
	Type string `json:"type"`

	// An object where each key is a property name and each value is a JSON Schema
	Properties map[string]SchemaProperty `json:"properties"`

	// List of required property names
	Required []string `json:"required,omitempty"`

	// A description of the schema
	Description string `json:"description,omitempty"`

	// Whether to allow properties not listed in properties
	AdditionalProperties bool `json:"additionalProperties,omitempty"`
}

// SchemaProperty represents a property definition in a JSON schema
type SchemaProperty struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

// AnswerResponse represents the response from the /answer endpoint
type AnswerResponse struct {
	// The generated answer based on search results
	// Returns a string by default, or a structured object when outputSchema is provided
	Answer any `json:"answer"`

	// Search results used to generate the answer
	Citations []AnswerCitation `json:"citations"`

	// Cost information for the request
	CostDollars *CostDollars `json:"costDollars,omitempty"`

	// Partial answer chunk when streaming is enabled
	StreamingAnswer string `json:"answer,omitempty"`
}

// AnswerCitation represents a citation in an answer response
type AnswerCitation struct {
	// The temporary ID for the document
	ID string `json:"id"`

	// The URL of the search result
	URL string `json:"url"`

	// The title of the search result
	Title string `json:"title"`

	// If available, the author of the content
	Author *string `json:"author,omitempty"`

	// An estimate of the creation date
	PublishedDate *string `json:"publishedDate,omitempty"`

	// The full text content of each source (only when text is enabled)
	Text string `json:"text,omitempty"`

	// The URL of the image associated with the search result
	Image *string `json:"image,omitempty"`

	// The URL of the favicon for the search result's domain
	Favicon *string `json:"favicon,omitempty"`
}
