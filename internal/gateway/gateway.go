// Package gateway provides a PostgREST-equivalent auto-REST API over real Postgres tables.
package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fieldstone/fieldstone/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// Gateway is a PostgREST-like query gateway over Postgres tables.
type Gateway struct {
	pool *pgxpool.Pool
}

// NewGateway creates a new query gateway.
func NewGateway(pool *pgxpool.Pool) *Gateway {
	return &Gateway{pool: pool}
}

// HasPool returns true if the gateway has a valid pool.
func (g *Gateway) HasPool() bool {
	return g.pool != nil
}

// RegisterRoutes mounts the gateway routes on the provided chi router.
func (g *Gateway) RegisterRoutes(r chi.Router) {
	r.Route("/v1", func(r chi.Router) {
		r.Get("/{table}", g.handleQuery)
		r.Get("/{table}/{id}", g.handleGetByID)
		r.Post("/{table}", g.handleInsert)
		r.Patch("/{table}", g.handleUpdate)
		r.Delete("/{table}", g.handleDelete)
	})
}

// setRLSContext sets JWT claims and tenant on a dedicated connection.
func (g *Gateway) setRLSContext(ctx context.Context, conn *pgxpool.Conn, r *http.Request) error {
	authCtx, ok := auth.FromContext(r.Context())
	if ok {
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
			return fmt.Errorf("failed to set jwt claims: %w", err)
		}
		if _, err := conn.Exec(ctx, "SET LOCAL app.current_tenant = $1", authCtx.TenantID); err != nil {
			return fmt.Errorf("failed to set tenant: %w", err)
		}
	}
	return nil
}

func (g *Gateway) handleQuery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	table := chi.URLParam(r, "table")

	conn, err := g.pool.Acquire(ctx)
	if err != nil {
		http.Error(w, `{"error":"database unavailable"}`, http.StatusServiceUnavailable)
		return
	}
	defer conn.Release()

	if err := g.setRLSContext(ctx, conn, r); err != nil {
		log.Error().Err(err).Msg("RLS setup failed")
	}

	qb := NewQueryBuilder(table)
	qb.Select(r.URL.Query().Get("select"))
	qb.Filters(r.URL.Query())
	qb.Order(r.URL.Query().Get("order"))
	qb.Limit(r.URL.Query().Get("limit"))
	qb.Offset(r.URL.Query().Get("offset"))

	sql, args := qb.BuildSelect()
	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		log.Error().Err(err).Str("sql", sql).Msg("Query failed")
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	defer rows.Close()

	results, err := pgx.CollectRows(rows, pgx.RowToMap)
	if err != nil {
		log.Error().Err(err).Msg("Collect rows failed")
		http.Error(w, `{"error":"failed to read results"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": results,
	})
}

func (g *Gateway) handleGetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	table := chi.URLParam(r, "table")
	id := chi.URLParam(r, "id")

	conn, err := g.pool.Acquire(ctx)
	if err != nil {
		http.Error(w, `{"error":"database unavailable"}`, http.StatusServiceUnavailable)
		return
	}
	defer conn.Release()

	if err := g.setRLSContext(ctx, conn, r); err != nil {
		log.Error().Err(err).Msg("RLS setup failed")
	}

	qb := NewQueryBuilder(table)
	qb.Select(r.URL.Query().Get("select"))
	qb.Filters(r.URL.Query())
	qb.Where("id", "=", id)

	sql, args := qb.BuildSelect()
	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		log.Error().Err(err).Str("sql", sql).Msg("Query failed")
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	defer rows.Close()

	result, err := pgx.CollectOneRow(rows, pgx.RowToMap)
	if err == pgx.ErrNoRows {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": result,
	})
}

func (g *Gateway) handleInsert(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	table := chi.URLParam(r, "table")

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	conn, err := g.pool.Acquire(ctx)
	if err != nil {
		http.Error(w, `{"error":"database unavailable"}`, http.StatusServiceUnavailable)
		return
	}
	defer conn.Release()

	if err := g.setRLSContext(ctx, conn, r); err != nil {
		log.Error().Err(err).Msg("RLS setup failed")
	}

	qb := NewQueryBuilder(table)
	sql, args := qb.BuildInsert(body)
	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		log.Error().Err(err).Str("sql", sql).Msg("Insert failed")
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	defer rows.Close()

	returning, err := pgx.CollectOneRow(rows, pgx.RowToMap)
	if err != nil {
		log.Error().Err(err).Str("sql", sql).Msg("Insert collect failed")
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": returning,
	})
}

func (g *Gateway) handleUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	table := chi.URLParam(r, "table")

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}

	id, ok := body["id"].(string)
	if !ok {
		// Allow id via query param
		id = r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
			return
		}
	}

	conn, err := g.pool.Acquire(ctx)
	if err != nil {
		http.Error(w, `{"error":"database unavailable"}`, http.StatusServiceUnavailable)
		return
	}
	defer conn.Release()

	if err := g.setRLSContext(ctx, conn, r); err != nil {
		log.Error().Err(err).Msg("RLS setup failed")
	}

	qb := NewQueryBuilder(table)
	sql, args := qb.BuildUpdate(id, body)
	rows, err := conn.Query(ctx, sql, args...)
	if err != nil {
		log.Error().Err(err).Str("sql", sql).Msg("Update failed")
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	defer rows.Close()

	returning, err := pgx.CollectOneRow(rows, pgx.RowToMap)
	if err == pgx.ErrNoRows {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		log.Error().Err(err).Str("sql", sql).Msg("Update collect failed")
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": returning,
	})
}

func (g *Gateway) handleDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	table := chi.URLParam(r, "table")
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, `{"error":"id required"}`, http.StatusBadRequest)
		return
	}

	conn, err := g.pool.Acquire(ctx)
	if err != nil {
		http.Error(w, `{"error":"database unavailable"}`, http.StatusServiceUnavailable)
		return
	}
	defer conn.Release()

	if err := g.setRLSContext(ctx, conn, r); err != nil {
		log.Error().Err(err).Msg("RLS setup failed")
	}

	qb := NewQueryBuilder(table)
	sql, args := qb.BuildDelete(id)
	tag, err := conn.Exec(ctx, sql, args...)
	if err != nil {
		log.Error().Err(err).Str("sql", sql).Msg("Delete failed")
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}
