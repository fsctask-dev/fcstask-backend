package recovery

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type FileAssembler struct {
	logger *RestoreLogger
}

func NewFileAssembler(logger *RestoreLogger) *FileAssembler {
	return &FileAssembler{logger: logger}
}

func (a *FileAssembler) AssembleFiles(rootDir string) error {
	a.logger.Info("Assembling split files in %s", rootDir)

	partsMap := make(map[string][]string)

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		base := filepath.Base(path)
		if strings.Contains(base, ".part") {
			idx := strings.LastIndex(base, ".part")
			if idx == -1 {
				return nil
			}
			origName := base[:idx]
			origPath := filepath.Join(filepath.Dir(path), origName)
			partsMap[origPath] = append(partsMap[origPath], path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk error: %w", err)
	}

	if len(partsMap) == 0 {
		a.logger.Debug("No split files found")
		return nil
	}

	for origPath, parts := range partsMap {
		if err := a.assembleOne(origPath, parts); err != nil {
			return fmt.Errorf("failed to assemble %s: %w", origPath, err)
		}
	}
	return nil
}

func (a *FileAssembler) assembleOne(origPath string, parts []string) error {
	a.logger.Info("Assembling %s from %d parts", origPath, len(parts))

	sort.Slice(parts, func(i, j int) bool {
		numI := extractPartNumber(parts[i])
		numJ := extractPartNumber(parts[j])
		return numI < numJ
	})

	out, err := os.Create(origPath)
	if err != nil {
		return err
	}
	defer out.Close()

	for _, part := range parts {
		in, err := os.Open(part)
		if err != nil {
			return err
		}
		_, err = io.Copy(out, in)
		in.Close()
		if err != nil {
			return err
		}
	}
	out.Close()
	for _, part := range parts {
		if err := os.Remove(part); err != nil {
			a.logger.Warn("Failed to remove part %s: %v", part, err)
		}
	}
	a.logger.Debug("Successfully assembled and cleaned up %s", origPath)
	return nil
}

func extractPartNumber(partPath string) int {
	base := filepath.Base(partPath)
	idx := strings.LastIndex(base, ".part")
	if idx == -1 {
		return 0
	}
	numStr := base[idx+5:]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0
	}
	return num
}
