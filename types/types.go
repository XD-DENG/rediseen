package types

// ResponseType acts as the JSON template for API response (successful calls)
type ResponseType struct {
	ValueType string      `json:"type"`
	Value     interface{} `json:"value"`
}

// ErrorType acts as the JSON template for API response (failed calls)
type ErrorType struct {
	Error string `json:"error"`
}

// KeyInfoType acts as the JSON template for element in KeyListType
type KeyInfoType struct {
	Key  string `json:"key"`
	Type string `json:"type"`
}

// KeyListType acts as the JSON template for API response (successful calls)
type KeyListType struct {
	Count int           `json:"count"`
	Total int           `json:"total"`
	Keys  []KeyInfoType `json:"keys"`
}
