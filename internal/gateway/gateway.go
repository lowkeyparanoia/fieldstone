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

// setRLSContext applies the request's JWT claims + tenant to the transaction via
// SET LOCAL. IMPORTANT: SET LOCAL only takes effect inside a transaction, so this
// must run on a pgx.Tx (see beginRLS) — otherwise Postgres discards it and RLS
// policies see no claims.
func (g *Gateway) setRLSContext(ctx context.Context, tx pgx.Tx, r *http.Request) error {
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
		if _, err := tx.Exec(ctx, "SELECT set_config('request.jwt.claims', $1, true)", string(claimsJSON)); err != nil {
			return fmt.Errorf("failed to set jwt claims: %w", err)
		}
		if _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant', $1, true)", authCtx.TenantID); err != nil {
			return fmt.Errorf("failed to set tenant: %w", err)
		}
	}
	return nil
}

// beginRLS acquires a pooled connection and opens a transaction with the request's
// RLS context applied. The caller MUST: defer conn.Release(); defer tx.Rollback(ctx);
// and tx.Commit(ctx) on the success path of any write. (SET LOCAL / set_config(…,true)
// require a transaction to take effect.)
func (g *Gateway) beginRLS(ctx context.Context, r *http.Request) (*pgxpool.Conn, pgx.Tx, error) {
	conn, err := g.pool.Acquire(ctx)
	if err != nil {
		return nil, nil, err
	}
	tx, err := conn.Begin(ctx)
	if err != nil {
		conn.Release()
		return nil, nil, err
	}
	if err := g.setRLSContext(ctx, tx, r); err != nil {
		log.Error().Err(err).Msg("RLS setup failed")
	}
	return conn, tx, nil
}

func (g *Gateway) handleQuery(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	table := chi.URLParam(r, "table")

	conn, tx, err := g.beginRLS(ctx, r)
	if err != nil {
		http.Error(w, `{"error":"database unavailable"}`, http.StatusServiceUnavailable)
		return
	}
	defer conn.Release()
	defer tx.Rollback(ctx)

	qb := NewQueryBuilder(table)
	qb.Select(r.URL.Query().Get("select"))
	qb.Filters(r.URL.Query())
	qb.Order(r.URL.Query().Get("order"))
	qb.Limit(r.URL.Query().Get("limit"))
	qb.Offset(r.URL.Query().Get("offset"))

	sql, args := qb.BuildSelect()
	rows, err := tx.Query(ctx, sql, args...)
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

	conn, tx, err := g.beginRLS(ctx, r)
	if err != nil {
		http.Error(w, `{"error":"database unavailable"}`, http.StatusServiceUnavailable)
		return
	}
	defer conn.Release()
	defer tx.Rollback(ctx)

	qb := NewQueryBuilder(table)
	qb.Select(r.URL.Query().Get("select"))
	qb.Filters(r.URL.Query())
	qb.Where("id", "=", id)

	sql, args := qb.BuildSelect()
	rows, err := tx.Query(ctx, sql, args...)
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

	conn, tx, err := g.beginRLS(ctx, r)
	if err != nil {
		http.Error(w, `{"error":"database unavailable"}`, http.StatusServiceUnavailable)
		return
	}
	defer conn.Release()
	defer tx.Rollback(ctx)

	qb := NewQueryBuilder(table)
	sql, args := qb.BuildInsert(body)
	rows, err := tx.Query(ctx, sql, args...)
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

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
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

	conn, tx, err := g.beginRLS(ctx, r)
	if err != nil {
		http.Error(w, `{"error":"database unavailable"}`, http.StatusServiceUnavailable)
		return
	}
	defer conn.Release()
	defer tx.Rollback(ctx)

	qb := NewQueryBuilder(table)
	sql, args := qb.BuildUpdate(id, body)
	rows, err := tx.Query(ctx, sql, args...)
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

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
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

	conn, tx, err := g.beginRLS(ctx, r)
	if err != nil {
		http.Error(w, `{"error":"database unavailable"}`, http.StatusServiceUnavailable)
		return
	}
	defer conn.Release()
	defer tx.Rollback(ctx)

	qb := NewQueryBuilder(table)
	sql, args := qb.BuildDelete(id)
	tag, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		log.Error().Err(err).Str("sql", sql).Msg("Delete failed")
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	if tag.RowsAffected() == 0 {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}
