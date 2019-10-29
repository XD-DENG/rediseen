package types

type ResponseType struct {
	ValueType string      `json:"type"`
	Value     interface{} `json:"value"`
}

type ErrorType struct {
	Error string `json:"error"`
}
