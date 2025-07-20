package models

// ParseOptions contains optional parameters for parsing statements
// This is the primary interface for parser configuration
type ParseOptions struct {
	SkipRows int             `json:"skip_rows"`
	Mappings []ColumnMapping `json:"mappings"`
}

// CreateStatementMetadata is an alias for ParseOptions to maintain API compatibility
type CreateStatementMetadata = ParseOptions

// NewParseOptions creates a new ParseOptions instance with default values
func NewParseOptions() ParseOptions {
	return ParseOptions{
		SkipRows: 0,
		Mappings: []ColumnMapping{},
	}
}

// NewCreateStatementMetadata creates a new CreateStatementMetadata instance with default values
func NewCreateStatementMetadata() CreateStatementMetadata {
	return CreateStatementMetadata{
		SkipRows: 0,
		Mappings: []ColumnMapping{},
	}
}

// HasCustomMappings checks if custom column mappings are provided
func (p ParseOptions) HasCustomMappings() bool {
	return len(p.Mappings) > 0
}

// HasRowSkipping checks if row skipping is enabled
func (p ParseOptions) HasRowSkipping() bool {
	return p.SkipRows > 0
}

// IsEmpty checks if metadata has any non-default values
func (p ParseOptions) IsEmpty() bool {
	return p.SkipRows == 0 && len(p.Mappings) == 0
}

// Note: Since CreateStatementMetadata is an alias for ParseOptions,
// all ParseOptions methods are automatically available for CreateStatementMetadata