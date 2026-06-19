package pg

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"fcstask-backend/backup-service/internal/config"
	"fcstask-backend/backup-service/internal/sqlutil"
)

type Client struct {
	conn config.ConnConfig
}

func New(conn config.ConnConfig) *Client { return &Client{conn: conn} }

const fieldSep = "\x1f"

type Params map[string]string

func (c *Client) env() []string {
	env := os.Environ()
	if c.conn.Password != "" {
		env = append(env, "PGPASSWORD="+c.conn.Password)
	}
	env = append(env, "PGTZ=UTC")
	if c.conn.SSLMode != "" {
		env = append(env, "PGSSLMODE="+c.conn.SSLMode)
	}
	if c.conn.ConnectTimeoutSeconds > 0 {
		env = append(env, "PGCONNECT_TIMEOUT="+strconv.Itoa(c.conn.ConnectTimeoutSeconds))
	}
	return env
}

func (c *Client) baseArgs() []string {
	return []string{"-h", c.conn.Host, "-p", strconv.Itoa(c.conn.Port), "-U", c.conn.User}
}

func (c *Client) Database() string { return c.conn.Database }

func paramArgs(params Params) []string {
	if len(params) == 0 {
		return nil
	}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	args := make([]string, 0, len(keys)*2)
	for _, k := range keys {
		args = append(args, "-v", k+"="+params[k])
	}
	return args
}

func (c *Client) Ping(ctx context.Context) error {
	_, err := c.QueryScalar(ctx, "SELECT 1", nil)
	return err
}

func (c *Client) Query(ctx context.Context, sql string, params Params) ([][]string, error) {
	args := append(c.baseArgs(), "-d", c.conn.Database, "-tA", "-F", fieldSep)
	args = append(args, paramArgs(params)...)
	args = append(args, "-c", sql)
	out, err := c.run(ctx, "psql", args)
	if err != nil {
		return nil, err
	}
	var rows [][]string
	for _, line := range strings.Split(strings.TrimRight(string(out), "\n"), "\n") {
		if line == "" {
			continue
		}
		rows = append(rows, strings.Split(line, fieldSep))
	}
	return rows, nil
}

func (c *Client) QueryScalar(ctx context.Context, sql string, params Params) (string, error) {
	rows, err := c.Query(ctx, sql, params)
	if err != nil {
		return "", err
	}
	if len(rows) == 0 || len(rows[0]) == 0 {
		return "", nil
	}
	return rows[0][0], nil
}

func (c *Client) Exec(ctx context.Context, sql string, params Params) error {
	args := append(c.baseArgs(), "-d", c.conn.Database, "-v", "ON_ERROR_STOP=1")
	args = append(args, paramArgs(params)...)
	args = append(args, "-c", sql)
	_, err := c.run(ctx, "psql", args)
	return err
}

func (c *Client) ExecOnDatabase(ctx context.Context, database, sql string, params Params) error {
	args := append(c.baseArgs(), "-d", database, "-v", "ON_ERROR_STOP=1")
	args = append(args, paramArgs(params)...)
	args = append(args, "-c", sql)
	_, err := c.run(ctx, "psql", args)
	return err
}

func (c *Client) RunScriptFile(ctx context.Context, path string, params Params) error {
	args := append(c.baseArgs(), "-d", c.conn.Database, "-v", "ON_ERROR_STOP=1")
	args = append(args, paramArgs(params)...)
	args = append(args, "-f", path)
	_, err := c.run(ctx, "psql", args)
	return err
}

type Copy struct {
	SelectSQL string
	DestPath  string
}

func (c *Client) RunCopyTransaction(ctx context.Context, copies []Copy) ([]int64, error) {
	if len(copies) == 0 {
		return nil, nil
	}
	var sb strings.Builder
	sb.WriteString("\\set ON_ERROR_STOP on\n")
	sb.WriteString("BEGIN TRANSACTION ISOLATION LEVEL REPEATABLE READ;\n")
	for _, cp := range copies {
		fmt.Fprintf(&sb, "\\copy (%s) TO '%s' WITH (FORMAT csv, HEADER true)\n",
			cp.SelectSQL, sqlutil.EscapeLiteral(cp.DestPath))
	}
	sb.WriteString("COMMIT;\n")

	f, err := os.CreateTemp("", "cdc-capture-*.sql")
	if err != nil {
		return nil, err
	}
	path := f.Name()
	defer os.Remove(path)
	if _, err := f.WriteString(sb.String()); err != nil {
		f.Close()
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}

	args := append(c.baseArgs(), "-d", c.conn.Database, "-v", "ON_ERROR_STOP=1", "-f", path)
	out, err := c.run(ctx, "psql", args)
	if err != nil {
		return nil, err
	}
	counts := parseCopyCounts(string(out))
	if len(counts) != len(copies) {
		return nil, fmt.Errorf("expected %d COPY results, parsed %d", len(copies), len(counts))
	}
	return counts, nil
}

func (c *Client) PgDumpDirectory(ctx context.Context, destDir string, jobs int) error {
	args := append(c.baseArgs(),
		"-d", c.conn.Database,
		"-F", "d",
		"-f", destDir,
		"-j", strconv.Itoa(jobs),
		"--no-password",
	)
	_, err := c.run(ctx, "pg_dump", args)
	return err
}

func (c *Client) PgRestoreDirectory(ctx context.Context, srcDir string, jobs int, clean bool) error {
	args := append(c.baseArgs(),
		"-d", c.conn.Database,
		"-F", "d",
		"-j", strconv.Itoa(jobs),
		"--no-password",
		"--exit-on-error",
	)
	if clean {
		args = append(args, "--clean", "--if-exists")
	}
	args = append(args, srcDir)
	_, err := c.run(ctx, "pg_restore", args)
	return err
}

func (c *Client) run(ctx context.Context, name string, args []string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = c.env()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("%s cancelled: %w", name, ctx.Err())
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("%s failed: %s", name, msg)
	}
	return stdout.Bytes(), nil
}

func parseCopyCounts(out string) []int64 {
	var counts []int64
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "COPY ") {
			if n, err := strconv.ParseInt(strings.TrimSpace(line[len("COPY "):]), 10, 64); err == nil {
				counts = append(counts, n)
			}
		}
	}
	return counts
}
