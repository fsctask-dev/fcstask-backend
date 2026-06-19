package backup

import (
	"fmt"
	"io"
	"os"

	"fcstask-backend/backup-service/internal/logging"
)

const PartSuffix = ".part"

type FileSplitter struct {
	maxBytes int64
	logger   *logging.Logger
}

func NewFileSplitter(maxSizeMB int, logger *logging.Logger) *FileSplitter {
	return &FileSplitter{maxBytes: int64(maxSizeMB) * 1024 * 1024, logger: logger}
}

func (s *FileSplitter) SplitFile(filePath string) (int, error) {
	in, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("open %q: %w", filePath, err)
	}
	defer in.Close()

	partNum := 1
	parts := 0
	for {
		partName := fmt.Sprintf("%s%s%03d", filePath, PartSuffix, partNum)
		part, err := os.Create(partName)
		if err != nil {
			return parts, fmt.Errorf("create part %q: %w", partName, err)
		}

		written, err := io.CopyN(part, in, s.maxBytes)
		closeErr := part.Close()
		if err != nil && err != io.EOF {
			_ = os.Remove(partName)
			return parts, fmt.Errorf("write part %q: %w", partName, err)
		}
		if closeErr != nil {
			return parts, fmt.Errorf("close part %q: %w", partName, closeErr)
		}
		if written == 0 {
			_ = os.Remove(partName)
			break
		}
		parts++
		partNum++
		if err == io.EOF || written < s.maxBytes {
			break
		}
	}

	if err := os.Remove(filePath); err != nil {
		return parts, fmt.Errorf("remove original %q after split: %w", filePath, err)
	}
	return parts, nil
}
