package gui

import (
	"encoding/json"
	"text/template"
)

// Template functions for `osascript` templates.
var templateFuncs = template.FuncMap{"json": func(v any) (string, error) {
	b, err := json.Marshal(v)

	return string(b), err
}}
