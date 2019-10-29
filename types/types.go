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
