package backup

import (
	"testing"
	"time"

	"fcstask-backend/backup-service/internal/config"
)

func meta(id, base string, typ BackupType, t time.Time) *Metadata {
	return &Metadata{ID: id, BaseFull: base, Type: typ, SnapshotTime: t}
}

func TestDecideBackup(t *testing.T) {
	base := time.Date(2026, 6, 1, 2, 0, 0, 0, time.UTC)

	b := &Backuper{cfg: config.BackupConfig{FullBackupEvery: 3}, pitr: config.PITRConfig{Enabled: true}}

	if d, _ := b.decideBackup(nil); d != DecisionFull {
		t.Errorf("empty history should be full, got %s", d)
	}

	chain := []*Metadata{
		meta("full_1", "full_1", FullBackup, base),
		meta("inc_1", "full_1", IncrementalBackup, base.Add(time.Hour)),
	}
	if d, parent := b.decideBackup(chain); d != DecisionIncremental || parent.ID != "inc_1" {
		t.Errorf("expected incremental off inc_1, got %s", d)
	}

	full3 := append(chain, meta("inc_2", "full_1", IncrementalBackup, base.Add(2*time.Hour)))
	if d, _ := b.decideBackup(full3); d != DecisionFull {
		t.Errorf("chain length == FullBackupEvery should force full, got %s", d)
	}

	b.forceFull = true
	if d, _ := b.decideBackup(chain); d != DecisionFull {
		t.Errorf("forceFull should yield full")
	}
	b.forceFull = false

	b.pitr.Enabled = false
	if d, _ := b.decideBackup(chain); d != DecisionFull {
		t.Errorf("pitr disabled should yield full")
	}
}
