package pg

import "testing"

func TestParseToolMajor(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"pg_dump (PostgreSQL) 17.2\n", 17},
		{"pg_restore (PostgreSQL) 16.4 (Debian 16.4-1.pgdg120+1)\n", 16},
		{"pg_basebackup (PostgreSQL) 15.7", 15},
		{"pg_receivewal (PostgreSQL) 18beta1", 18},
	}
	for _, c := range cases {
		got, err := parseToolMajor(c.in)
		if err != nil {
			t.Fatalf("parseToolMajor(%q) returned error: %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("parseToolMajor(%q) = %d, want %d", c.in, got, c.want)
		}
	}

	if _, err := parseToolMajor("no version here"); err == nil {
		t.Error("expected error for output without a version number")
	}
}
