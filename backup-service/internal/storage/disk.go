package storage

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const kbPerGB = 1 << 20

type DiskChecker struct{}

func NewDiskChecker() *DiskChecker { return &DiskChecker{} }

func (d *DiskChecker) CheckFreeSpace(ctx context.Context, path string, minFreeGB int) error {
	if minFreeGB <= 0 {
		return nil
	}
	cmd := exec.CommandContext(ctx, "df", "-Pk", path)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("run df on %q: %w", path, err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return fmt.Errorf("unexpected df output: %q", string(output))
	}
	fields := strings.Fields(lines[len(lines)-1])
	if len(fields) < 4 {
		return fmt.Errorf("unexpected df output format: %q", lines[len(lines)-1])
	}

	freeKB, err := strconv.ParseUint(fields[3], 10, 64)
	if err != nil {
		return fmt.Errorf("parse free space %q: %w", fields[3], err)
	}
	freeGB := freeKB / kbPerGB
	if freeGB < uint64(minFreeGB) {
		return fmt.Errorf("insufficient disk space at %q: %d GiB free, %d GiB required", path, freeGB, minFreeGB)
	}
	return nil
}
