package ccm

import strs "gabyx/githooks/strings"

// FormatCodeBlock formats a code block in markdown style.
func FormatCodeBlock(s string, lang string) string {
	return strs.Fmt("```%s\n%s\n```", lang, s)
}
