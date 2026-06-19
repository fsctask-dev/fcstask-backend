package sqlutil

import "testing"

func TestQuoteIdent(t *testing.T) {
	cases := map[string]string{
		"users":     `"users"`,
		`a"b`:       `"a""b"`,
		`x; DROP--`: `"x; DROP--"`,
	}
	for in, want := range cases {
		if got := QuoteIdent(in); got != want {
			t.Errorf("QuoteIdent(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestQuoteQualified(t *testing.T) {
	if got := QuoteQualified("public", "users"); got != `"public"."users"` {
		t.Errorf("got %q", got)
	}
}

func TestValidateIdentifier(t *testing.T) {
	good := []string{"created_at", "Updated_At", "_x", "t1"}
	for _, s := range good {
		if err := ValidateIdentifier("col", s); err != nil {
			t.Errorf("expected %q valid: %v", s, err)
		}
	}
	bad := []string{"", "a b", "x'y", `a"b`, "drop;table", "a-b"}
	for _, s := range bad {
		if err := ValidateIdentifier("col", s); err == nil {
			t.Errorf("expected %q invalid", s)
		}
	}
}
