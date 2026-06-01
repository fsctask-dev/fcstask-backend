package storage

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const (
	KB_to_GB = 1 << 20
)

type DiskChecker struct{}

func NewDiskChecker() *DiskChecker {
	return &DiskChecker{}
}

func (d *DiskChecker) CheckFreeSpace(path string, minFreeGB int) error {
	cmd := exec.Command("df", "-k", path)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to run df: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return fmt.Errorf("unexpected df output")
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return fmt.Errorf("unexpected df output format")
	}

	freeKB, err := strconv.ParseUint(fields[3], 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse free space: %v", err)
	}

	freeGB := freeKB / KB_to_GB

	if freeGB < uint64(minFreeGB) {
		return fmt.Errorf("insufficient disk space: %d GB available, %d GB required", freeGB, minFreeGB)
	}

	return nil
}
