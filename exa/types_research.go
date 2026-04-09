package exa

// CreateResearchTaskRequest represents a request to create a research task
type CreateResearchTaskRequest struct {
	// Instructions for what the research task should accomplish (required, max 4096 chars)
	Instructions string `json:"instructions"`

	// Model to use: exa-research (default) or exa-research-pro
	Model string `json:"model,omitempty"`

	// Output configuration including schema
	Output *ResearchOutput `json:"output,omitempty"`
}

// ResearchOutput represents output configuration for research tasks
type ResearchOutput struct {
	// A JsonSchema specification of the desired output
	Schema map[string]any `json:"schema,omitempty"`

	// When set to true and no output schema is provided, an LLM will generate an output schema
	InferSchema bool `json:"inferSchema,omitempty"`
}

// ResearchTask represents a research task response
type ResearchTask struct {
	// The unique identifier for the research task
	ID string `json:"id"`

	// The current status of the research task: running, completed, or failed
	Status string `json:"status"`

	// The instructions or query for the research task
	Instructions string `json:"instructions"`

	// The JSON schema specification for the expected output format
	Schema map[string]any `json:"schema,omitempty"`

	// The research results data conforming to the specified schema
	Data map[string]any `json:"data,omitempty"`

	// Citations grouped by the root field they were used in
	Citations map[string][]ResearchCitation `json:"citations,omitempty"`
}

// ResearchCitation represents a citation in research results
type ResearchCitation struct {
	// Citation ID
	ID string `json:"id"`

	// The URL of the source
	URL string `json:"url"`

	// The title of the source
	Title string `json:"title"`

	// Snippet from the source
	Snippet string `json:"snippet"`
}

// ListResearchTasksResponse represents the response from listing research tasks
type ListResearchTasksResponse struct {
	// Unique identifier for the request
	RequestID string `json:"requestId"`

	// The list of research tasks
	Data []ResearchTask `json:"data"`

	// Whether there are more results to paginate through
	HasMore bool `json:"hasMore"`

	// The cursor to paginate through the next set of results
	NextCursor *string `json:"nextCursor,omitempty"`
}
