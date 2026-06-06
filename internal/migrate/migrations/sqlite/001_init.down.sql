-- Migration: 001_init
-- Description: Rollback initial schema

DROP TRIGGER IF EXISTS tenants_updated_at;
DROP TRIGGER IF EXISTS users_updated_at;
DROP TRIGGER IF EXISTS records_updated_at;
DROP TRIGGER IF EXISTS collections_updated_at;

DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_tenant;
DROP INDEX IF EXISTS idx_records_tenant;
DROP INDEX IF EXISTS idx_records_collection;
DROP INDEX IF EXISTS idx_collections_tenant;

DROP TABLE IF EXISTS tenants;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS records;
DROP TABLE IF EXISTS collections;
