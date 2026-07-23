package backend

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// These tests need a real Postgres, because row level security is a Postgres
// feature and there is nothing to assert without one:
//
//	docker run --rm -d -p 5433:5432 -e POSTGRES_PASSWORD=pw --name fs-test postgres:16
//	FIELDSTONE_TEST_DSN='postgres://postgres:pw@localhost:5433/postgres' go test ./internal/backend/ -run RLS -v
//
// Without the env var they skip, so the normal test run is unaffected.
func testDSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("FIELDSTONE_TEST_DSN")
	if dsn == "" {
		t.Skip("set FIELDSTONE_TEST_DSN to run RLS integration tests")
	}
	return dsn
}

// TestRLSBlocksCrossTenantReads is the test that actually proves the fix.
//
// It deliberately issues a query with NO tenant_id predicate. Under the old
// code that returned every tenant's rows, because the policies were never
// activated. If RLS is working, the database itself filters the result down to
// the tenant set on the transaction.
func TestRLSBlocksCrossTenantReads(t *testing.T) {
	dsn := testDSN(t)
	ctx := context.Background()

	b, err := NewPostgresBackend(dsn, 4, 1)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer b.Close()

	if err := b.ApplyRLSHardening(ctx); err != nil {
		t.Fatalf("harden: %v", err)
	}

	tenantA, tenantB := uuid.NewString(), uuid.NewString()
	seed(t, b, tenantA, "alpha")
	seed(t, b, tenantB, "beta")

	// Query as tenant A with no tenant filter of any kind.
	var names []string
	err = b.withTenantTx(ctx, tenantA, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `SELECT name FROM _collections`)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var n string
			if err := rows.Scan(&n); err != nil {
				return err
			}
			names = append(names, n)
		}
		return rows.Err()
	})
	if err != nil {
		t.Fatalf("query: %v", err)
	}

	if len(names) != 1 || names[0] != "alpha" {
		t.Fatalf("RLS is not enforcing isolation: got %v, want [alpha]. "+
			"If both tenants appear, check that the connecting role is not the "+
			"table owner, or that FORCE ROW LEVEL SECURITY was applied.", names)
	}
}

// TestRLSBlocksCrossTenantWrites proves the WITH CHECK clause. Without it a
// bug could insert a row stamped with somebody else's tenant id.
func TestRLSBlocksCrossTenantWrites(t *testing.T) {
	dsn := testDSN(t)
	ctx := context.Background()

	b, err := NewPostgresBackend(dsn, 4, 1)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer b.Close()
	if err := b.ApplyRLSHardening(ctx); err != nil {
		t.Fatalf("harden: %v", err)
	}

	tenantA, tenantB := uuid.NewString(), uuid.NewString()

	// Acting as tenant A, try to write a row belonging to tenant B.
	err = b.withTenantTx(ctx, tenantA, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`INSERT INTO _collections (id, tenant_id, name, fields) VALUES ($1,$2,$3,'[]')`,
			uuid.NewString(), tenantB, "smuggled")
		return err
	})
	if err == nil {
		t.Fatal("WITH CHECK is not enforced: tenant A wrote a row owned by tenant B")
	}
}

// TestTenantSettingDoesNotLeakAcrossPooledConnections guards the specific bug
// the original helper had: a session-level SET that survives connection reuse.
func TestTenantSettingDoesNotLeakAcrossPooledConnections(t *testing.T) {
	dsn := testDSN(t)
	ctx := context.Background()

	b, err := NewPostgresBackend(dsn, 4, 1)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer b.Close()

	tenantA := uuid.NewString()
	if err := b.withTenantTx(ctx, tenantA, func(tx pgx.Tx) error { return nil }); err != nil {
		t.Fatalf("tx: %v", err)
	}

	// A fresh statement outside any tenant tx must see an empty setting.
	var got string
	if err := b.pool.QueryRow(ctx,
		`SELECT current_setting('app.current_tenant', true)`).Scan(&got); err != nil {
		t.Fatalf("read setting: %v", err)
	}
	if got == tenantA {
		t.Fatalf("tenant leaked onto a pooled connection: got %q", got)
	}
}

func seed(t *testing.T, b *PostgresBackend, tenantID, name string) {
	t.Helper()
	ctx := context.Background()
	err := b.withTenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`INSERT INTO _collections (id, tenant_id, name, fields) VALUES ($1,$2,$3,'[]')`,
			uuid.NewString(), tenantID, name)
		return err
	})
	if err != nil {
		t.Fatalf("seed %s: %v", name, err)
	}
}
