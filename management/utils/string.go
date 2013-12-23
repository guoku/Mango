package utils

import (
    "regexp"
)

func StripPuncsAndSymbols(str string) string {
	re := regexp.MustCompile("[\\pP\\pS\\pZ\\pM]")
	re2 := regexp.MustCompile("[@\\-`'!_ +&＆]")
	var rp = func(repl string) string {
		if !re2.MatchString(repl) {
			return " "
		}
		return repl
	}
	return re.ReplaceAllStringFunc(str, rp)
}
