package restore

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"fcstask-backend/backup-service/internal/backup"
	"fcstask-backend/backup-service/internal/logging"
)

type FileAssembler struct {
	logger *logging.Logger
}

func NewFileAssembler(logger *logging.Logger) *FileAssembler {
	return &FileAssembler{logger: logger}
}

func (a *FileAssembler) AssembleFiles(rootDir string, expected []backup.SplitInfo) error {
	manifest := map[string]backup.SplitInfo{}
	for _, s := range expected {
		manifest[filepath.FromSlash(s.File)] = s
	}

	parts := map[string][]string{}
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !backup.IsPart(path) {
			return nil
		}
		base := filepath.Base(path)
		idx := strings.LastIndex(base, backup.PartSuffix)
		if idx == -1 {
			return nil
		}
		origPath := filepath.Join(filepath.Dir(path), base[:idx])
		parts[origPath] = append(parts[origPath], path)
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk %q: %w", rootDir, err)
	}

	for orig, ps := range parts {
		rel, rerr := filepath.Rel(rootDir, orig)
		if rerr != nil {
			return rerr
		}
		exp, hasExp := manifest[rel]
		if hasExp && len(ps) != exp.Parts {
			return fmt.Errorf("file %q: found %d parts, manifest expects %d", rel, len(ps), exp.Parts)
		}
		if err := a.assembleOne(orig, ps); err != nil {
			return fmt.Errorf("assemble %q: %w", orig, err)
		}
		if hasExp {
			sum, size, verr := fileSHA256(orig)
			if verr != nil {
				return fmt.Errorf("verify %q: %w", rel, verr)
			}
			if size != exp.Size || sum != exp.SHA256 {
				return fmt.Errorf("integrity check failed for %q: size/sha256 mismatch", rel)
			}
			a.logger.Info("Verified reassembled %s (%d bytes)", rel, size)
		}
	}

	for rel, exp := range manifest {
		if _, done := parts[filepath.Join(rootDir, rel)]; done {
			continue
		}
		orig := filepath.Join(rootDir, rel)
		sum, size, verr := fileSHA256(orig)
		if verr != nil {
			return fmt.Errorf("missing reassembled file %q: %w", rel, verr)
		}
		if size != exp.Size || sum != exp.SHA256 {
			return fmt.Errorf("integrity check failed for %q: size/sha256 mismatch", rel)
		}
	}
	return nil
}

func (a *FileAssembler) assembleOne(origPath string, parts []string) error {
	sort.Slice(parts, func(i, j int) bool {
		return partNumber(parts[i]) < partNumber(parts[j])
	})
	a.logger.Info("Reassembling %s from %d parts", origPath, len(parts))

	out, err := os.OpenFile(origPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	for _, p := range parts {
		in, err := os.Open(p)
		if err != nil {
			out.Close()
			return err
		}
		_, cerr := io.Copy(out, in)
		in.Close()
		if cerr != nil {
			out.Close()
			return cerr
		}
	}
	if err := out.Close(); err != nil {
		return err
	}
	for _, p := range parts {
		if err := os.Remove(p); err != nil {
			a.logger.Warn("Failed to remove part %s: %v", p, err)
		}
	}
	return nil
}

func partNumber(path string) int {
	base := filepath.Base(path)
	idx := strings.LastIndex(base, backup.PartSuffix)
	if idx == -1 {
		return 0
	}
	n, err := strconv.Atoi(base[idx+len(backup.PartSuffix):])
	if err != nil {
		return 0
	}
	return n
}

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
