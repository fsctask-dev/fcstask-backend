package pg

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var majorVersionRe = regexp.MustCompile(`\d+`)

func parseToolMajor(versionOutput string) (int, error) {
	m := majorVersionRe.FindString(versionOutput)
	if m == "" {
		return 0, fmt.Errorf("could not find a version number in %q", strings.TrimSpace(versionOutput))
	}
	return strconv.Atoi(m)
}

func (c *Client) ToolMajorVersion(ctx context.Context, tool string) (int, error) {
	out, err := c.run(ctx, tool, []string{"--version"})
	if err != nil {
		return 0, err
	}
	return parseToolMajor(string(out))
}

func (c *Client) ServerMajorVersion(ctx context.Context) (int, error) {
	s, err := c.QueryScalar(ctx, "SHOW server_version_num", nil)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, fmt.Errorf("parse server_version_num %q: %w", s, err)
	}
	return n / 10000, nil
}

func (c *Client) CheckDumpCompatibility(ctx context.Context) error {
	server, err := c.ServerMajorVersion(ctx)
	if err != nil {
		return fmt.Errorf("read server version: %w", err)
	}
	for _, tool := range []string{"pg_dump", "pg_restore"} {
		tv, err := c.ToolMajorVersion(ctx, tool)
		if err != nil {
			return fmt.Errorf("read %s version: %w", tool, err)
		}
		if tv < server {
			return fmt.Errorf("%s major version %d is older than the server major version %d; "+
				"pg_dump/pg_restore must be at least the server version "+
				"(rebuild the image with --build-arg PG_MAJOR=%d)", tool, tv, server, server)
		}
	}
	return nil
}
