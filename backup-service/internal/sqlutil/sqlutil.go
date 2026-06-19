package sqlutil

import (
	"fmt"
	"regexp"
	"strings"
)

var identRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_$]*$`)

func ValidateIdentifier(kind, s string) error {
	if s == "" {
		return fmt.Errorf("empty %s identifier", kind)
	}
	if len(s) > 63 {
		return fmt.Errorf("%s identifier %q exceeds 63 bytes", kind, s)
	}
	if !identRe.MatchString(s) {
		return fmt.Errorf("%s identifier %q contains unsupported characters", kind, s)
	}
	return nil
}

func QuoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func QuoteQualified(schema, table string) string {
	return QuoteIdent(schema) + "." + QuoteIdent(table)
}

func JoinIdents(cols []string) string {
	out := make([]string, len(cols))
	for i, c := range cols {
		out[i] = QuoteIdent(c)
	}
	return strings.Join(out, ", ")
}

func EscapeLiteral(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
