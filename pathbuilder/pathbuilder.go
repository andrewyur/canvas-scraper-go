package pathbuilder

import (
	"path/filepath"
	"regexp"
	"strings"
)

var sanitizeRegexp = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

func Sanitize(s string) string {
	safe := sanitizeRegexp.ReplaceAllString(s, "_")
	safe = strings.Trim(safe, ".")
	return safe
}

type PathBuilder struct {
	components []string
}

func NewPathBuilder(base string) PathBuilder {
	return PathBuilder{
		components: []string{Sanitize(base)},
	}
}

func (c PathBuilder) Fork(part ...string) PathBuilder {
	for i, p := range part {
		part[i] = Sanitize(p)
	}

	return PathBuilder{
		components: append(c.components, part...),
	}
}

func (c PathBuilder) Build() string {
	return filepath.Join(c.components...)
}
