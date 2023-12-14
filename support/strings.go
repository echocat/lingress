package support

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func Join(sep string, what ...interface{}) string {
	buf := new(bytes.Buffer)
	for i, part := range what {
		if i > 0 {
			buf.WriteString(sep)
		}
		buf.WriteString(fmt.Sprint(part))
	}
	return buf.String()
}

func QuoteIfNeeded(what string) string {
	if strings.ContainsRune(what, '\t') ||
		strings.ContainsRune(what, '\n') ||
		strings.ContainsRune(what, ' ') ||
		strings.ContainsRune(what, '\xFF') ||
		strings.ContainsRune(what, '\u0100') ||
		strings.ContainsRune(what, '"') ||
		strings.ContainsRune(what, '\\') {
		return strconv.Quote(what)
	}
	return what
}

func QuoteAllIfNeeded(in ...string) []string {
	out := make([]string, len(in))
	for i, a := range in {
		out[i] = QuoteIfNeeded(a)
	}
	return out
}

func QuoteAndJoin(in ...string) string {
	out := make([]string, len(in))
	for i, a := range in {
		out[i] = QuoteIfNeeded(a)
	}
	return strings.Join(out, " ")
}
