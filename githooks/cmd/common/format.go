package ccm

import strs "gabyx/githooks/strings"

func FormatCodeBlock(s string, lang string) string {
	return strs.Fmt("```%s\n%s\n```", lang, s)
}
