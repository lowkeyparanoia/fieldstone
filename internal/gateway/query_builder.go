package gateway

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// QueryBuilder builds safe, parameterized SQL queries for the auto-REST API.
type QueryBuilder struct {
	table   string
	selects []string
	wheres  []string
	args    []interface{}
	orders  []string
	limit   int
	offset  int
}

// NewQueryBuilder creates a builder for a table.
func NewQueryBuilder(table string) *QueryBuilder {
	return &QueryBuilder{
		table:  sanitizeTable(table),
		limit:  1000,
		offset: 0,
	}
}

// Select sets the columns to return (comma-separated).
func (qb *QueryBuilder) Select(cols string) {
	if cols == "" {
		qb.selects = []string{"*"}
		return
	}
	parts := strings.Split(cols, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			qb.selects = append(qb.selects, sanitizeCol(p))
		}
	}
	if len(qb.selects) == 0 {
		qb.selects = []string{"*"}
	}
}

// Where adds a raw where clause (used for ID filters).
func (qb *QueryBuilder) Where(col, op, val string) {
	qb.addWhere(col, op, val)
}

// Filters parses query parameters for eq, neq, gt, gte, lt, lte, like.
func (qb *QueryBuilder) Filters(params url.Values) {
	for key, vals := range params {
		if len(vals) == 0 {
			continue
		}
		val := vals[0]
		switch {
		case strings.HasPrefix(key, "eq."):
			qb.addWhere(strings.TrimPrefix(key, "eq."), "=", val)
		case strings.HasPrefix(key, "neq."):
			qb.addWhere(strings.TrimPrefix(key, "neq."), "!=", val)
		case strings.HasPrefix(key, "gt."):
			qb.addWhere(strings.TrimPrefix(key, "gt."), ">", val)
		case strings.HasPrefix(key, "gte."):
			qb.addWhere(strings.TrimPrefix(key, "gte."), ">=", val)
		case strings.HasPrefix(key, "lt."):
			qb.addWhere(strings.TrimPrefix(key, "lt."), "<", val)
		case strings.HasPrefix(key, "lte."):
			qb.addWhere(strings.TrimPrefix(key, "lte."), "<=", val)
		case strings.HasPrefix(key, "like."):
			qb.addWhereLike(strings.TrimPrefix(key, "like."), val)
		}
	}
}

// Order parses ordering like "created_at.desc,id.asc".
func (qb *QueryBuilder) Order(order string) {
	if order == "" {
		return
	}
	parts := strings.Split(order, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		sub := strings.SplitN(p, ".", 2)
		col := sanitizeCol(sub[0])
		dir := "ASC"
		if len(sub) == 2 && strings.ToLower(sub[1]) == "desc" {
			dir = "DESC"
		}
		qb.orders = append(qb.orders, fmt.Sprintf("%s %s", col, dir))
	}
}

// Limit sets the limit.
func (qb *QueryBuilder) Limit(limit string) {
	if limit == "" {
		return
	}
	if n, err := strconv.Atoi(limit); err == nil && n > 0 {
		if n > 1000 {
			n = 1000
		}
		qb.limit = n
	}
}

// Offset sets the offset.
func (qb *QueryBuilder) Offset(offset string) {
	if offset == "" {
		return
	}
	if n, err := strconv.Atoi(offset); err == nil && n >= 0 {
		qb.offset = n
	}
}

// BuildSelect returns the SQL and arguments.
func (qb *QueryBuilder) BuildSelect() (string, []interface{}) {
	cols := strings.Join(qb.selects, ", ")
	sql := fmt.Sprintf("SELECT %s FROM %s", cols, pgx.Identifier{qb.table}.Sanitize())
	if len(qb.wheres) > 0 {
		sql += " WHERE " + strings.Join(qb.wheres, " AND ")
	}
	if len(qb.orders) > 0 {
		sql += " ORDER BY " + strings.Join(qb.orders, ", ")
	}
	sql += fmt.Sprintf(" LIMIT %d OFFSET %d", qb.limit, qb.offset)
	return sql, qb.args
}

// BuildInsert generates an INSERT ... RETURNING * query.
func (qb *QueryBuilder) BuildInsert(data map[string]interface{}) (string, []interface{}) {
	var cols []string
	var placeholders []string
	var args []interface{}
	i := 1
	for k, v := range data {
		if k == "id" && v == "" {
			continue
		}
		cols = append(cols, sanitizeCol(k))
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		args = append(args, v)
		i++
	}
	colStr := strings.Join(cols, ", ")
	phStr := strings.Join(placeholders, ", ")
	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING *", pgx.Identifier{qb.table}.Sanitize(), colStr, phStr)
	return sql, args
}

// BuildUpdate generates an UPDATE ... RETURNING * query.
func (qb *QueryBuilder) BuildUpdate(id string, data map[string]interface{}) (string, []interface{}) {
	var sets []string
	var args []interface{}
	i := 1
	for k, v := range data {
		if k == "id" {
			continue
		}
		sets = append(sets, fmt.Sprintf("%s = $%d", sanitizeCol(k), i))
		args = append(args, v)
		i++
	}
	args = append(args, id)
	setStr := strings.Join(sets, ", ")
	sql := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d RETURNING *", pgx.Identifier{qb.table}.Sanitize(), setStr, i)
	return sql, args
}

// BuildDelete generates a DELETE query.
func (qb *QueryBuilder) BuildDelete(id string) (string, []interface{}) {
	sql := fmt.Sprintf("DELETE FROM %s WHERE id = $1", pgx.Identifier{qb.table}.Sanitize())
	return sql, []interface{}{id}
}

func (qb *QueryBuilder) addWhere(col, op, val string) {
	col = sanitizeCol(col)
	qb.wheres = append(qb.wheres, fmt.Sprintf("%s %s $%d", col, op, len(qb.args)+1))
	qb.args = append(qb.args, val)
}

func (qb *QueryBuilder) addWhereLike(col, val string) {
	col = sanitizeCol(col)
	qb.wheres = append(qb.wheres, fmt.Sprintf("%s ILIKE $%d", col, len(qb.args)+1))
	qb.args = append(qb.args, "%"+val+"%")
}

func sanitizeTable(name string) string {
	// Remove dangerous characters; keep alphanumerics and underscore
	name = strings.ReplaceAll(name, "\x00", "")
	name = strings.ReplaceAll(name, "\"", "")
	name = strings.ReplaceAll(name, "'", "")
	name = strings.ReplaceAll(name, ";", "")
	return name
}

func sanitizeCol(name string) string {
	name = strings.ReplaceAll(name, "\x00", "")
	name = strings.ReplaceAll(name, "\"", "")
	name = strings.ReplaceAll(name, "'", "")
	name = strings.ReplaceAll(name, ";", "")
	name = strings.ReplaceAll(name, "(", "")
	name = strings.ReplaceAll(name, ")", "")
	return name
}
