package backup

import "testing"

func TestIsPart(t *testing.T) {
	yes := []string{"data.part001", "x.part123", "a/b/data.csv.part010"}
	for _, n := range yes {
		if !IsPart(n) {
			t.Errorf("expected %q to be a part", n)
		}
	}
	no := []string{"data.csv", "table.partition.csv", "x.part", "report.partial"}
	for _, n := range no {
		if IsPart(n) {
			t.Errorf("expected %q not to be a part", n)
		}
	}
}

func TestSafeComponent(t *testing.T) {
	cases := map[string]string{
		"users":       "users",
		"../../etc":   "_.._etc",
		"a/b":         "a_b",
		"..":          "x",
		"public":      "public",
		"weird name;": "weird_name_",
	}
	for in, want := range cases {
		if got := safeComponent(in); got != want {
			t.Errorf("safeComponent(%q) = %q, want %q", in, got, want)
		}
	}
}
