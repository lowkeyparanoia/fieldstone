-- Marketing Pack RLS Policies

ALTER TABLE companies ENABLE ROW LEVEL SECURITY;
ALTER TABLE prospects ENABLE ROW LEVEL SECURITY;
ALTER TABLE signals ENABLE ROW LEVEL SECURITY;
ALTER TABLE campaigns ENABLE ROW LEVEL SECURITY;
ALTER TABLE sequences ENABLE ROW LEVEL SECURITY;
ALTER TABLE engagements ENABLE ROW LEVEL SECURITY;
ALTER TABLE suppression ENABLE ROW LEVEL SECURITY;
ALTER TABLE sending_domains ENABLE ROW LEVEL SECURITY;

-- For marketing, we use a simple tenant isolation model via tenant_id column
-- (In real deployment, add tenant_id to tables or use app.current_tenant)

-- Example: if tenant_id is added, policies would be:
-- CREATE POLICY "tenant isolation" ON companies
--     FOR ALL USING (tenant_id = current_setting('app.current_tenant', TRUE)::UUID);

-- For this pack, we leave policies open (marketing data is usually agency-managed)
-- and recommend adding tenant_id + RLS in production.
