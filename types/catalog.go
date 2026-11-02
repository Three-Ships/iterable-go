package types

type Catalogs struct {
	CatalogNames       []string `json:"catalogNames"`
	TotalCatalogsCount int      `json:"totalCatalogsCount"`
	PreviousPageUrl    string   `json:"previousPageUrl,omitempty"`
	NextPageUrl        string   `json:"nextPageUrl,omitempty"`
}

// CatalogItemReplaceRequest contains the complete replacement item.
type CatalogItemReplaceRequest struct {
	// Value contains the complete item and must encode as a non-null JSON object.
	// Use a struct or a map with string keys. The client marshals the value;
	// Iterable validates the catalog's field types and item limits.
	Value any `json:"value"`
}

type CatalogFieldMapping struct {
	DefinedMappings map[string]interface{} `json:"definedMappings"`
	UndefinedFields []string               `json:"undefinedFields"`
}
