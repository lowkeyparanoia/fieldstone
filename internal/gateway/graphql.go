package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/fieldstone/fieldstone/internal/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// GraphQLGateway provides a GraphQL interface over Postgres
type GraphQLGateway struct {
	pool *pgxpool.Pool
}

// NewGraphQLGateway creates a GraphQL gateway
func NewGraphQLGateway(pool *pgxpool.Pool) *GraphQLGateway {
	return &GraphQLGateway{pool: pool}
}

// HasPool returns true if pool is available
func (g *GraphQLGateway) HasPool() bool {
	return g.pool != nil
}

// GraphQLRequest represents an incoming GraphQL query
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   json.RawMessage `json:"data,omitempty"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message string `json:"message"`
}

// HandleGraphQL processes GraphQL queries using pg_graphql
func (g *GraphQLGateway) HandleGraphQL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req GraphQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendGraphQLError(w, "invalid JSON body")
		return
	}

	if req.Query == "" {
		sendGraphQLError(w, "query is required")
		return
	}

	conn, err := g.pool.Acquire(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to acquire connection")
		sendGraphQLError(w, "database unavailable")
		return
	}
	defer conn.Release()

	// Set tenant context for RLS
	if authCtx, ok := auth.FromContext(r.Context()); ok {
		claims := map[string]interface{}{
			"sub":       authCtx.UserID,
			"tenant_id": authCtx.TenantID,
			"email":     authCtx.Email,
			"role":      authCtx.Role,
		}
		if authCtx.RawClaims != nil {
			for k, v := range authCtx.RawClaims {
				if _, exists := claims[k]; !exists {
					claims[k] = v
				}
			}
		}
		claimsJSON, _ := json.Marshal(claims)
		if _, err := conn.Exec(ctx, "SET LOCAL request.jwt.claims = $1", claimsJSON); err != nil {
			log.Error().Err(err).Msg("failed to set jwt claims")
		}
		if _, err := conn.Exec(ctx, "SET LOCAL app.current_tenant = $1", authCtx.TenantID); err != nil {
			log.Error().Err(err).Msg("failed to set tenant")
		}
	}

	// Check if pg_graphql extension is available
	var hasGraphQL bool
	err = conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'pg_graphql')").Scan(&hasGraphQL)
	if err != nil || !hasGraphQL {
		// Fallback: basic reflection-based GraphQL
		g.handleFallbackGraphQL(w, r, req)
		return
	}

	// Use pg_graphql
	var resultJSON string
	err = conn.QueryRow(ctx,
		"SELECT graphql.resolve($1, $2::jsonb, $3)",
		req.Query,
		req.Variables,
		req.OperationName,
	).Scan(&resultJSON)

	if err != nil {
		log.Error().Err(err).Str("query", req.Query).Msg("GraphQL query failed")
		sendGraphQLError(w, fmt.Sprintf("query failed: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(resultJSON))
}

// handleFallbackGraphQL provides a basic GraphQL implementation when pg_graphql is not available
func (g *GraphQLGateway) handleFallbackGraphQL(w http.ResponseWriter, r *http.Request, req GraphQLRequest) {
	ctx := r.Context()

	// Simple introspection for tables
	query := strings.TrimSpace(req.Query)

	// Handle introspection queries
	if strings.Contains(query, "__schema") || strings.Contains(query, "__type") {
		g.handleIntrospection(w, ctx)
		return
	}

	// Parse basic queries like: query { users { id email } }
	tableName, fields, err := parseSimpleQuery(query)
	if err != nil {
		sendGraphQLError(w, fmt.Sprintf("parse error: %v", err))
		return
	}

	conn, err := g.pool.Acquire(ctx)
	if err != nil {
		sendGraphQLError(w, "database unavailable")
		return
	}
	defer conn.Release()

	// Build simple SELECT
	cols := "*"
	if len(fields) > 0 {
		cols = strings.Join(fields, ", ")
	}

	// Sanitize table name
	tableName = sanitizeTableName(tableName)

	rows, err := conn.Query(ctx, fmt.Sprintf("SELECT %s FROM %s LIMIT 100", cols, tableName))
	if err != nil {
		sendGraphQLError(w, fmt.Sprintf("query failed: %v", err))
		return
	}
	defer rows.Close()

	results, err := pgx.CollectRows(rows, pgx.RowToMap)
	if err != nil {
		sendGraphQLError(w, fmt.Sprintf("collect failed: %v", err))
		return
	}

	resp := GraphQLResponse{
		Data: mustJSON(map[string]interface{}{
			capitalize(tableName): results,
		}),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleIntrospection returns basic schema info
func (g *GraphQLGateway) handleIntrospection(w http.ResponseWriter, ctx context.Context) {
	conn, err := g.pool.Acquire(ctx)
	if err != nil {
		sendGraphQLError(w, "database unavailable")
		return
	}
	defer conn.Release()

	// Get list of tables
	rows, err := conn.Query(ctx, `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE'
		AND table_name NOT LIKE 'pg_%'
		AND table_name NOT LIKE '_%'
	`)
	if err != nil {
		sendGraphQLError(w, fmt.Sprintf("introspection failed: %v", err))
		return
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			tables = append(tables, name)
		}
	}

	types := make([]map[string]interface{}, 0, len(tables))
	for _, t := range tables {
		types = append(types, map[string]interface{}{
			"kind": "OBJECT",
			"name": t,
			"fields": []map[string]interface{}{
				{"name": "id", "type": map[string]string{"name": "ID"}},
				{"name": "created_at", "type": map[string]string{"name": "String"}},
			},
		})
	}

	resp := GraphQLResponse{
		Data: mustJSON(map[string]interface{}{
			"__schema": map[string]interface{}{
				"types": types,
			},
		}),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func parseSimpleQuery(query string) (tableName string, fields []string, err error) {
	// Very basic parser for queries like: { users { id email } }
	query = strings.TrimSpace(query)

	// Remove 'query { ... }' wrapper
	if strings.HasPrefix(query, "query") {
		idx := strings.Index(query, "{")
		if idx > 0 {
			query = query[idx:]
		}
	}

	// Find first table name between { and {
	start := strings.Index(query, "{")
	if start < 0 {
		return "", nil, fmt.Errorf("no opening brace")
	}

	end := strings.Index(query[start+1:], "{")
	if end < 0 {
		// No fields specified
		tableName = strings.TrimSpace(strings.Trim(query[start+1:], "{} \n\t"))
		return tableName, nil, nil
	}

	tableName = strings.TrimSpace(query[start+1 : start+1+end])

	// Extract fields
	fieldsStart := start + 1 + end + 1
	fieldsEnd := strings.Index(query[fieldsStart:], "}")
	if fieldsEnd < 0 {
		return tableName, nil, nil
	}

	fieldsStr := query[fieldsStart : fieldsStart+fieldsEnd]
	for _, f := range strings.Fields(fieldsStr) {
		f = strings.TrimSpace(f)
		if f != "" {
			fields = append(fields, f)
		}
	}

	return tableName, fields, nil
}

func sendGraphQLError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // GraphQL returns 200 even for errors
	json.NewEncoder(w).Encode(GraphQLResponse{
		Errors: []GraphQLError{{Message: message}},
	})
}

func mustJSON(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return json.RawMessage(b)
}

func sanitizeTableName(name string) string {
	// Remove dangerous characters
	name = strings.ReplaceAll(name, "\x00", "")
	name = strings.ReplaceAll(name, "\"", "")
	name = strings.ReplaceAll(name, "'", "")
	name = strings.ReplaceAll(name, ";", "")
	return name
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
