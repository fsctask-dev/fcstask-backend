package backup

import (
	"fcstask-backend/backup-tool/pkg/logging"
	"fmt"
	"io"
	"os"
)

type FileSplitter struct {
	maxSizeMB int
	logger    *logging.Logger
}

func NewFileSplitter(maxSizeMB int, logger *logging.Logger) *FileSplitter {
	return &FileSplitter{
		maxSizeMB: maxSizeMB,
		logger:    logger,
	}
}

func (s *FileSplitter) SplitFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	maxBytes := int64(s.maxSizeMB) * 1024 * 1024
	partNum := 1
	buffer := make([]byte, maxBytes)

	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}

		if n == 0 {
			break
		}

		partFileName := fmt.Sprintf("%s.part%03d", filePath, partNum)
		partFile, err := os.Create(partFileName)
		if err != nil {
			return err
		}

		if _, err := partFile.Write(buffer[:n]); err != nil {
			partFile.Close()
			return err
		}
		partFile.Close()

		partNum++

		if err == io.EOF {
			break
		}
	}

	if err := os.Remove(filePath); err != nil {
		s.logger.Warn("Failed to remove original file: %v", err)
	}

	return nil
}
