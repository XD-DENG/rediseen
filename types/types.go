package types

type ResponseType struct {
	ValueType string      `json:"type"`
	Value     interface{} `json:"value"`
}

type ErrorType struct {
	Error string `json:"error"`
}

type ResultType struct {
	Action string `json:"action"`
	Result string `json:"result"`
}
