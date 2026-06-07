-- LakeB2B Pack RLS Policies

ALTER TABLE bulk_uploads ENABLE ROW LEVEL SECURITY;
ALTER TABLE entities ENABLE ROW LEVEL SECURITY;
ALTER TABLE dedup_rules ENABLE ROW LEVEL SECURITY;
ALTER TABLE enrichments ENABLE ROW LEVEL SECURITY;
ALTER TABLE entity_relations ENABLE ROW LEVEL SECURITY;
ALTER TABLE cdc_events ENABLE ROW LEVEL SECURITY;

-- Similar to marketing, LakeB2B usually operates at tenant scope.
-- Add tenant_id columns and tenant isolation policies in production.
