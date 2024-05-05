package helper

import "strings"

func PadStart(s, pad string, length int) string {
	if len(s) >= length {
		return s
	}
	return strings.Repeat(pad, length-len(s)) + s
}
