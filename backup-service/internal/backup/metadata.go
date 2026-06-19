package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

func fileSHA256(path string) (string, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

var (
	partRe   = regexp.MustCompile(`\.part\d{3,}$`)
	unsafeRe = regexp.MustCompile(`[^A-Za-z0-9_.-]`)
)

func IsPart(name string) bool { return partRe.MatchString(filepath.Base(name)) }

func safeComponent(s string) string {
	s = unsafeRe.ReplaceAllString(s, "_")
	s = strings.TrimLeft(s, ".")
	if s == "" {
		s = "x"
	}
	return s
}

type BackupType string

const (
	FullBackup        BackupType = "full"
	IncrementalBackup BackupType = "incremental"
)

const (
	metadataFile    = "backup_metadata.json"
	dataSubdir      = "data"
	cdcSubdir       = "cdc"
	checksumsFile   = "SHA256SUMS"
	metadataVersion = 1
)

var ErrNoChecksums = fmt.Errorf("no %s present", checksumsFile)

func WriteChecksums(dir string) error {
	var lines []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		name := info.Name()
		if name == checksumsFile || name == metadataFile || name == metadataFile+".tmp" {
			return nil
		}
		sum, _, herr := fileSHA256(path)
		if herr != nil {
			return herr
		}
		rel, rerr := filepath.Rel(dir, path)
		if rerr != nil {
			return rerr
		}
		lines = append(lines, sum+"  "+filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return err
	}
	sort.Strings(lines)
	content := strings.Join(lines, "\n")
	if content != "" {
		content += "\n"
	}
	tmp := filepath.Join(dir, checksumsFile+".tmp")
	if err := os.WriteFile(tmp, []byte(content), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dir, checksumsFile))
}

func VerifyChecksums(dir string) error {
	data, err := os.ReadFile(filepath.Join(dir, checksumsFile))
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNoChecksums
		}
		return err
	}
	var problems []string
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		if line == "" {
			continue
		}
		idx := strings.Index(line, "  ")
		if idx < 0 {
			problems = append(problems, "malformed line: "+line)
			continue
		}
		want := line[:idx]
		rel := filepath.FromSlash(line[idx+2:])
		got, _, herr := fileSHA256(filepath.Join(dir, rel))
		if herr != nil {
			problems = append(problems, rel+": "+herr.Error())
			continue
		}
		if got != want {
			problems = append(problems, rel+": checksum mismatch")
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("integrity check failed for %q: %s", filepath.Base(dir), strings.Join(problems, "; "))
	}
	return nil
}

type TableInfo struct {
	Schema     string   `json:"schema"`
	Name       string   `json:"name"`
	CreatedAt  string   `json:"created_at_col,omitempty"`
	UpdatedAt  string   `json:"updated_at_col,omitempty"`
	DeletedAt  string   `json:"deleted_at_col,omitempty"`
	ChangeExpr string   `json:"change_expr"`
	Columns    []string `json:"columns,omitempty"`
	DataFile   string   `json:"data_file"`
	RowCount   int64    `json:"row_count"`
}

type Metadata struct {
	Version      int         `json:"version"`
	Type         BackupType  `json:"type"`
	ID           string      `json:"id"`
	Database     string      `json:"database"`
	StartedAt    time.Time   `json:"started_at"`
	FinishedAt   time.Time   `json:"finished_at"`
	SnapshotTime time.Time   `json:"snapshot_time"`
	Since        time.Time   `json:"since,omitempty"`
	Parent       string      `json:"parent,omitempty"`
	BaseFull     string      `json:"base_full"`
	Tables       []TableInfo `json:"tables,omitempty"`
	Splits       []SplitInfo `json:"splits,omitempty"`
	SizeBytes    int64       `json:"size_bytes"`
	FileCount    int         `json:"file_count"`
}

type SplitInfo struct {
	File   string `json:"file"`
	Parts  int    `json:"parts"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

func writeMetadata(dir string, m *Metadata) error {
	m.Version = metadataVersion
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	tmp := filepath.Join(dir, metadataFile+".tmp")
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}

	return os.Rename(tmp, filepath.Join(dir, metadataFile))
}

func ReadMetadata(dir string) (*Metadata, error) {
	data, err := os.ReadFile(filepath.Join(dir, metadataFile))
	if err != nil {
		return nil, err
	}
	var m Metadata
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse metadata in %q: %w", dir, err)
	}
	return &m, nil
}

func ListBackups(root string) ([]*Metadata, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read backup root %q: %w", root, err)
	}
	var out []*Metadata
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, string(FullBackup)+"_") &&
			!strings.HasPrefix(name, string(IncrementalBackup)+"_") {
			continue
		}
		m, err := ReadMetadata(filepath.Join(root, name))
		if err != nil {
			continue
		}
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SnapshotTime.Before(out[j].SnapshotTime) })
	return out, nil
}

func dirSize(dir string) (int64, int, error) {
	var size int64
	var count int
	err := filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
			count++
		}
		return nil
	})
	return size, count, err
}
