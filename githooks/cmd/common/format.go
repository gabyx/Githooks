package ccm

import strs "rycus86/githooks/strings"

func FormatCodeBlock(s string, lang string) string {
	return strs.Fmt("```%s\n%s\n```", lang, s)
}
