-- SaaS Pack RLS Policies

ALTER TABLE orgs ENABLE ROW LEVEL SECURITY;
ALTER TABLE org_memberships ENABLE ROW LEVEL SECURITY;
ALTER TABLE subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;
ALTER TABLE usage_records ENABLE ROW LEVEL SECURITY;
ALTER TABLE saas_audit ENABLE ROW LEVEL SECURITY;

-- Helper reused from healthcare pack (auth.uid, auth.role)

-- orgs: members can read their org; owners/admins can update
CREATE POLICY "members read org" ON orgs
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM org_memberships m
            WHERE m.org_id = orgs.id AND m.user_id = auth.uid()
        )
    );

CREATE POLICY "owner admin update org" ON orgs
    FOR UPDATE USING (
        EXISTS (
            SELECT 1 FROM org_memberships m
            WHERE m.org_id = orgs.id AND m.user_id = auth.uid()
              AND m.role IN ('owner','admin')
        )
    );

-- org_memberships: members can read memberships in their org; self can leave; owner can manage
CREATE POLICY "members read memberships" ON org_memberships
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM org_memberships m
            WHERE m.org_id = org_memberships.org_id AND m.user_id = auth.uid()
        )
    );

CREATE POLICY "owner admin manage memberships" ON org_memberships
    FOR ALL USING (
        EXISTS (
            SELECT 1 FROM org_memberships m
            WHERE m.org_id = org_memberships.org_id AND m.user_id = auth.uid()
              AND m.role IN ('owner','admin')
        )
    ) WITH CHECK (
        EXISTS (
            SELECT 1 FROM org_memberships m
            WHERE m.org_id = org_memberships.org_id AND m.user_id = auth.uid()
              AND m.role IN ('owner','admin')
        )
    );

-- subscriptions: members read; billing+owner can update
CREATE POLICY "members read subscriptions" ON subscriptions
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM org_memberships m
            WHERE m.org_id = subscriptions.org_id AND m.user_id = auth.uid()
        )
    );

-- invoices: same
CREATE POLICY "members read invoices" ON invoices
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM org_memberships m
            WHERE m.org_id = invoices.org_id AND m.user_id = auth.uid()
        )
    );

-- usage_records: members read
CREATE POLICY "members read usage" ON usage_records
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM org_memberships m
            WHERE m.org_id = usage_records.org_id AND m.user_id = auth.uid()
        )
    );

-- saas_audit: members read org audit
CREATE POLICY "members read audit" ON saas_audit
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM org_memberships m
            WHERE m.org_id = saas_audit.org_id AND m.user_id = auth.uid()
        )
    );
