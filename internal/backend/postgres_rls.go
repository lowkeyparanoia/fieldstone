package backend

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// ---------------------------------------------------------------------------
// Tenant-scoped execution.
//
// Row level security in Postgres reads a *session* setting. That makes the
// tenant a property of a connection, not of a request, which is the detail the
// previous implementation missed:
//
//	conn, _ := pool.Acquire(ctx)
//	defer conn.Release()                       // connection returned to pool
//	conn.Exec(ctx, "SET LOCAL app.current_tenant = $1", id)
//	return ctx, nil                            // ctx unchanged; nothing pinned
//
// Four independent faults there. SET LOCAL only survives inside a transaction,
// and there was none. The connection went back to the pool before any query ran
// on it. The context was returned unmodified, so no later query could be pinned
// to that connection. And the helper was never called by anything.
//
// The fix is to run every tenant-scoped statement inside one transaction that
// begins by setting the tenant. SET LOCAL is reverted automatically on commit
// or rollback, so a pooled connection can never leak one tenant's identity to
// the next borrower.
// ---------------------------------------------------------------------------

// tenantSettingKey is the session variable the RLS policies read via
// current_tenant_id(). It must match the migration exactly.
const tenantSettingKey = "app.current_tenant"

// nilTenant is used for operations that are legitimately not tenant scoped.
// It matches no real row, so a policy evaluated against it denies everything
// rather than falling open.
const nilTenant = "00000000-0000-0000-0000-000000000000"

// withTenantTx runs fn inside a transaction that has the tenant set, so RLS
// policies can see it. Rows must be consumed inside fn: they do not outlive the
// transaction.
//
// Prefer this over holding a raw connection. The tenant setting is local to the
// transaction, so there is no cleanup to forget and no way to leak it.
func (b *PostgresBackend) withTenantTx(ctx context.Context, tenantID string, fn func(tx pgx.Tx) error) error {
	if tenantID == "" {
		tenantID = nilTenant
	}

	tx, err := b.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tenant tx: %w", err)
	}
	// Safe to call after a successful Commit: pgx makes the second call a no-op.
	defer func() { _ = tx.Rollback(ctx) }()

	// set_config is the function form of SET. The third argument, true, means
	// "local to this transaction". A bare SET LOCAL cannot take a placeholder,
	// which is why the parameterised function form is used here: the tenant id
	// is data and must never be concatenated into SQL.
	if _, err := tx.Exec(ctx,
		`SELECT set_config($1, $2, true)`, tenantSettingKey, tenantID,
	); err != nil {
		return fmt.Errorf("set tenant context: %w", err)
	}

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ---------------------------------------------------------------------------
// Transactions
// ---------------------------------------------------------------------------

// pgTx adapts a pgx.Tx to the backend.Tx interface. Commit and Rollback are
// idempotent enough to be safe in a defer.
type pgTx struct {
	tx  pgx.Tx
	ctx context.Context
}

func (t *pgTx) Commit() error   { return t.tx.Commit(t.ctx) }
func (t *pgTx) Rollback() error { return t.tx.Rollback(t.ctx) }

// Tx exposes the underlying pgx transaction for callers that need to issue
// statements on it.
func (t *pgTx) Tx() pgx.Tx { return t.tx }

// BeginTx starts a transaction with no tenant bound. Use BeginTenantTx for
// anything that touches tenant scoped tables, otherwise RLS will (correctly)
// hide every row.
func (b *PostgresBackend) BeginTx(ctx context.Context) (Tx, error) {
	tx, err := b.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	return &pgTx{tx: tx, ctx: ctx}, nil
}

// BeginTenantTx starts a transaction that already has the tenant set, for
// callers that need to interleave several operations atomically.
func (b *PostgresBackend) BeginTenantTx(ctx context.Context, tenantID string) (*pgTx, error) {
	if tenantID == "" {
		tenantID = nilTenant
	}
	tx, err := b.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tenant tx: %w", err)
	}
	if _, err := tx.Exec(ctx, `SELECT set_config($1, $2, true)`, tenantSettingKey, tenantID); err != nil {
		_ = tx.Rollback(ctx)
		return nil, fmt.Errorf("set tenant context: %w", err)
	}
	return &pgTx{tx: tx, ctx: ctx}, nil
}

// ---------------------------------------------------------------------------
// Migration hardening
// ---------------------------------------------------------------------------

// rlsHardeningSQL closes the three gaps left by the original policy set.
//
//  1. FORCE ROW LEVEL SECURITY. In Postgres the role that owns a table is
//     exempt from its own policies. If the application connects as the owner,
//     which it usually does in development, every policy silently does nothing.
//     This is the single most common reason RLS "does not work".
//
//  2. WITH CHECK. A bare USING clause governs which rows are visible to
//     SELECT, UPDATE and DELETE. It does not constrain INSERT. Without a
//     WITH CHECK clause a bug could still write a row stamped with another
//     tenant's id, which is precisely the leak RLS is meant to prevent.
//
//  3. Coverage. Policies existed on _collections and _users but not on
//     _records or _tenants.
const rlsHardeningSQL = `
-- Owners bypass their own policies unless forced.
ALTER TABLE _collections FORCE ROW LEVEL SECURITY;
ALTER TABLE _users       FORCE ROW LEVEL SECURITY;

-- USING governs reads; WITH CHECK governs writes. Both are required.
DROP POLICY IF EXISTS tenant_isolation_collections ON _collections;
CREATE POLICY tenant_isolation_collections ON _collections
    USING      (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

DROP POLICY IF EXISTS tenant_isolation_users ON _users;
CREATE POLICY tenant_isolation_users ON _users
    USING      (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());
`

// ApplyRLSHardening is idempotent and safe to run on every boot.
func (b *PostgresBackend) ApplyRLSHardening(ctx context.Context) error {
	if _, err := b.pool.Exec(ctx, rlsHardeningSQL); err != nil {
		return fmt.Errorf("apply rls hardening: %w", err)
	}
	return nil
}

// CheckRLSEffective reports whether the connecting role is actually subject to
// row level security.
//
// This matters more than any of the SQL above. Postgres exempts three kinds of
// role from RLS, and in each case every policy silently does nothing:
//
//   - superusers, unconditionally. FORCE ROW LEVEL SECURITY does not help.
//   - roles with the BYPASSRLS attribute.
//   - the table owner, unless FORCE ROW LEVEL SECURITY is set.
//
// Connecting as "postgres" therefore disables tenant isolation completely while
// looking perfectly healthy, which is the failure mode most likely to reach
// production unnoticed. Call this at boot and refuse to start if it fails.
func (b *PostgresBackend) CheckRLSEffective(ctx context.Context) error {
	var isSuper, bypassRLS bool
	err := b.pool.QueryRow(ctx,
		`SELECT rolsuper, rolbypassrls FROM pg_roles WHERE rolname = current_user`,
	).Scan(&isSuper, &bypassRLS)
	if err != nil {
		return fmt.Errorf("inspect current role: %w", err)
	}

	switch {
	case isSuper:
		return fmt.Errorf("row level security is not enforced: connected as a superuser. " +
			"Superusers bypass RLS unconditionally and FORCE ROW LEVEL SECURITY does not " +
			"change that. Connect as a dedicated non-superuser role that does not own the tables")
	case bypassRLS:
		return fmt.Errorf("row level security is not enforced: the current role has BYPASSRLS. " +
			"Revoke it with ALTER ROLE ... NOBYPASSRLS")
	}

	// Owning the tables is acceptable *if* FORCE ROW LEVEL SECURITY is set,
	// which is the situation for this server because it runs its own DDL on
	// boot. Check the actual flag rather than assuming.
	var ownsTables, forced bool
	if err := b.pool.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM pg_tables
		               WHERE tablename = '_collections' AND tableowner = current_user),
		       COALESCE((SELECT relforcerowsecurity FROM pg_class
		                 WHERE relname = '_collections'), false)
	`).Scan(&ownsTables, &forced); err != nil {
		return fmt.Errorf("inspect table ownership: %w", err)
	}
	if ownsTables && !forced {
		return fmt.Errorf("row level security is not enforced: the current role owns " +
			"_collections and FORCE ROW LEVEL SECURITY is not set, so the policies are " +
			"bypassed. Call ApplyRLSHardening, or use a role that does not own the tables")
	}
	return nil
}

// StorageBytes reports the total on-disk size of the database, satisfying the
// api.StorageReporter capability. Backends that cannot answer simply do not
// implement it, and the dashboard reports zero rather than inventing a figure.
func (b *PostgresBackend) StorageBytes() (int64, error) {
	var n int64
	err := b.pool.QueryRow(context.Background(),
		`SELECT pg_database_size(current_database())`).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("read database size: %w", err)
	}
	return n, nil
}
