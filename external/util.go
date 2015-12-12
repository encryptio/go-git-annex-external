package external

import (
	"strings"
)

func filterNewlines(s string) string {
	return strings.Replace(s, "\n", " ", -1)
}
