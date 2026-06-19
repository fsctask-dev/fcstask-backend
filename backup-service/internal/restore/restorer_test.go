package restore

import (
	"testing"
	"time"
)

func TestBoundValue(t *testing.T) {
	if got := boundValue(nil); got != "infinity" {
		t.Errorf("nil bound = %q, want infinity", got)
	}
	tm := time.Date(2026, 6, 7, 0, 0, 0, 0, time.UTC)
	if got := boundValue(&tm); got != "2026-06-07 00:00:00+00" {
		t.Errorf("bound = %q", got)
	}
}

func TestBuildUpsert(t *testing.T) {
	got := buildUpsert([]string{"id"}, []string{"id", "name"})
	want := `ON CONFLICT ("id") DO UPDATE SET "name" = EXCLUDED."name"`
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}

	gotPKOnly := buildUpsert([]string{"id"}, []string{"id"})
	wantPKOnly := `ON CONFLICT ("id") DO NOTHING`
	if gotPKOnly != wantPKOnly {
		t.Errorf("got %q want %q", gotPKOnly, wantPKOnly)
	}
}
