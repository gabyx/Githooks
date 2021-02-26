package result

type JSONResult struct {
	Version int
	Timeout bool
	Result  interface{} `yaml:",inline"`
}

// NewJSONResult returns a new JSON result.
func NewJSONResult(res interface{}) JSONResult {
	return JSONResult{Version: 1, Result: res}
}
