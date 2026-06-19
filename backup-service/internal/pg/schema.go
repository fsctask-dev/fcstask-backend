package pg

import (
	"context"
	"fmt"
	"strings"

	"fcstask-backend/backup-service/internal/sqlutil"
)

type TableRef struct {
	Schema    string
	Name      string
	CreatedAt string
	UpdatedAt string
	DeletedAt string
}

func (t TableRef) Qualified() string {
	return sqlutil.QuoteQualified(t.Schema, t.Name)
}

func (t TableRef) ChangeExpr() string {
	var cols []string
	for _, c := range []string{t.UpdatedAt, t.CreatedAt, t.DeletedAt} {
		if c != "" {
			cols = append(cols, sqlutil.QuoteIdent(c))
		}
	}
	if len(cols) == 1 {
		return cols[0]
	}
	return "GREATEST(" + strings.Join(cols, ", ") + ")"
}

func arrayParam(prefix string, vals []string) (string, Params) {
	if len(vals) == 0 {
		return "ARRAY[]::text[]", nil
	}
	parts := make([]string, len(vals))
	p := Params{}
	for i, v := range vals {
		key := fmt.Sprintf("%s%d", prefix, i)
		parts[i] = ":'" + key + "'"
		p[key] = v
	}
	return "ARRAY[" + strings.Join(parts, ", ") + "]", p
}

func mergeParams(maps ...Params) Params {
	out := Params{}
	for _, m := range maps {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

func (c *Client) DiscoverCDCTables(ctx context.Context, schemas, createdCols, updatedCols, deletedCols []string) ([]TableRef, []string, error) {
	tracked := uniqueLower(append(append(append([]string{}, createdCols...), updatedCols...), deletedCols...))
	if len(schemas) == 0 || len(tracked) == 0 {
		return nil, nil, nil
	}

	schemasArr, sp := arrayParam("s", schemas)
	colsArr, cp := arrayParam("c", tracked)
	params := mergeParams(sp, cp)

	sql := fmt.Sprintf(`
SELECT c.table_schema, c.table_name, string_agg(lower(c.column_name), ',')
FROM information_schema.columns c
JOIN information_schema.tables t
  ON t.table_schema = c.table_schema AND t.table_name = c.table_name
WHERE t.table_type = 'BASE TABLE'
  AND c.table_schema = ANY (%s)
  AND lower(c.column_name) = ANY (%s)
GROUP BY c.table_schema, c.table_name
ORDER BY 1, 2`, schemasArr, colsArr)

	rows, err := c.Query(ctx, sql, params)
	if err != nil {
		return nil, nil, fmt.Errorf("discover tables: %w", err)
	}

	createdSet := toSet(createdCols)
	updatedSet := toSet(updatedCols)
	deletedSet := toSet(deletedCols)

	var tables []TableRef
	var skipped []string
	for _, r := range rows {
		if len(r) < 3 {
			continue
		}
		ref := TableRef{Schema: r[0], Name: r[1]}
		for _, col := range strings.Split(r[2], ",") {
			switch {
			case createdSet[col] && ref.CreatedAt == "":
				ref.CreatedAt = col
			case updatedSet[col] && ref.UpdatedAt == "":
				ref.UpdatedAt = col
			case deletedSet[col] && ref.DeletedAt == "":
				ref.DeletedAt = col
			}
		}

		if ref.CreatedAt == "" && ref.UpdatedAt == "" {
			continue
		}
		pk, err := c.PrimaryKey(ctx, ref.Schema, ref.Name)
		if err != nil {
			return nil, nil, err
		}
		if len(pk) == 0 {
			skipped = append(skipped, ref.Schema+"."+ref.Name+" (no primary key)")
			continue
		}
		tables = append(tables, ref)
	}
	return tables, skipped, nil
}

func (c *Client) PrimaryKey(ctx context.Context, schema, table string) ([]string, error) {
	sql := `
SELECT a.attname
FROM pg_index i
JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY (i.indkey)
WHERE i.indrelid = format('%I.%I', :'schema', :'table')::regclass
  AND i.indisprimary
ORDER BY array_position(i.indkey, a.attnum)`
	rows, err := c.Query(ctx, sql, Params{"schema": schema, "table": table})
	if err != nil {
		return nil, fmt.Errorf("primary key of %s.%s: %w", schema, table, err)
	}
	var cols []string
	for _, r := range rows {
		if len(r) > 0 {
			cols = append(cols, r[0])
		}
	}
	return cols, nil
}

type ColumnInfo struct {
	Name           string
	IdentityAlways bool
}

func (c *Client) ColumnsInfo(ctx context.Context, schema, table string) ([]ColumnInfo, error) {
	sql := `
SELECT column_name,
       (is_identity = 'YES' AND identity_generation = 'ALWAYS')
FROM information_schema.columns
WHERE table_schema = :'schema' AND table_name = :'table'
  AND is_generated <> 'ALWAYS'
ORDER BY ordinal_position`
	rows, err := c.Query(ctx, sql, Params{"schema": schema, "table": table})
	if err != nil {
		return nil, fmt.Errorf("columns of %s.%s: %w", schema, table, err)
	}
	var cols []ColumnInfo
	for _, r := range rows {
		if len(r) < 2 {
			continue
		}
		cols = append(cols, ColumnInfo{Name: r[0], IdentityAlways: r[1] == "t"})
	}
	return cols, nil
}

func (c *Client) Columns(ctx context.Context, schema, table string) ([]string, error) {
	info, err := c.ColumnsInfo(ctx, schema, table)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(info))
	for i, ci := range info {
		names[i] = ci.Name
	}
	return names, nil
}

func (c *Client) HasIdentityAlways(ctx context.Context, schema, table string) (bool, error) {
	sql := `
SELECT count(*)
FROM information_schema.columns
WHERE table_schema = :'schema' AND table_name = :'table'
  AND identity_generation = 'ALWAYS'`
	s, err := c.QueryScalar(ctx, sql, Params{"schema": schema, "table": table})
	if err != nil {
		return false, fmt.Errorf("identity check of %s.%s: %w", schema, table, err)
	}
	return strings.TrimSpace(s) != "" && strings.TrimSpace(s) != "0", nil
}

func toSet(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, it := range items {
		m[strings.ToLower(it)] = true
	}
	return m
}

func uniqueLower(items []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, it := range items {
		l := strings.ToLower(it)
		if l == "" || seen[l] {
			continue
		}
		seen[l] = true
		out = append(out, l)
	}
	return out
}
