package wal

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"fcstask-backend/backup-service/internal/config"
)

const (
	baseMetadataFile = "wal_base_metadata.json"
	basePrefix       = "base_"

	recoverySignal = "recovery.signal"
)

func connArgs(c config.ConnConfig) []string {
	return []string{"-h", c.Host, "-p", strconv.Itoa(c.Port), "-U", c.User}
}

func env(c config.ConnConfig) []string {
	e := os.Environ()
	if c.Password != "" {
		e = append(e, "PGPASSWORD="+c.Password)
	}
	if c.SSLMode != "" {
		e = append(e, "PGSSLMODE="+c.SSLMode)
	}
	if c.ConnectTimeoutSeconds > 0 {
		e = append(e, "PGCONNECT_TIMEOUT="+strconv.Itoa(c.ConnectTimeoutSeconds))
	}
	return e
}

func runCommand(ctx context.Context, conn config.ConnConfig, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = env(conn)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return stdout.String(), fmt.Errorf("%s cancelled: %w", name, ctx.Err())
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return stdout.String(), fmt.Errorf("%s failed: %s", name, msg)
	}
	return stdout.String(), nil
}

func toolAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func parseStartWALFile(labelPath string) (string, error) {
	f, err := os.Open(labelPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "START WAL LOCATION") {
			continue
		}
		open := strings.Index(line, "(file ")
		if open < 0 {
			continue
		}
		rest := line[open+len("(file "):]
		close := strings.IndexByte(rest, ')')
		if close < 0 {
			continue
		}
		return strings.TrimSpace(rest[:close]), nil
	}
	if err := sc.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("START WAL LOCATION not found in %s", labelPath)
}
