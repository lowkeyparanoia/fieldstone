-- Healthcare Pack RLS Policies
-- Enforces: patient self-access, physician active-consult access, admin read

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE consents ENABLE ROW LEVEL SECURITY;
ALTER TABLE health_metrics ENABLE ROW LEVEL SECURITY;
ALTER TABLE consultation_slots ENABLE ROW LEVEL SECURITY;
ALTER TABLE consultations ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;
ALTER TABLE device_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE brain_checkins ENABLE ROW LEVEL SECURITY;
ALTER TABLE sleep_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE metric_baselines ENABLE ROW LEVEL SECURITY;
ALTER TABLE scores ENABLE ROW LEVEL SECURITY;
ALTER TABLE biological_ages ENABLE ROW LEVEL SECURITY;
ALTER TABLE score_weights ENABLE ROW LEVEL SECURITY;

-- Helper: current user id from JWT
CREATE OR REPLACE FUNCTION auth.uid() RETURNS UUID AS $$
BEGIN
    RETURN NULLIF(current_setting('request.jwt.claims', TRUE)::json->>'sub', '')::UUID;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Helper: current user role from JWT
CREATE OR REPLACE FUNCTION auth.role() RETURNS TEXT AS $$
BEGIN
    RETURN COALESCE(current_setting('request.jwt.claims', TRUE)::json->>'role', 'authenticated');
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- users: patient sees self, physician sees self, admin sees all
CREATE POLICY "patient self-access" ON users
    FOR ALL USING (user_id = auth.uid()) WITH CHECK (user_id = auth.uid());

CREATE POLICY "admin read users" ON users
    FOR SELECT USING (auth.role() IN ('admin','super_admin'));

-- consents: patient only
CREATE POLICY "patient self-access" ON consents
    FOR ALL USING (user_id = auth.uid()) WITH CHECK (user_id = auth.uid());

-- health_metrics: patient self + physician during active consult + admin read
CREATE POLICY "patient self-access" ON health_metrics
    FOR ALL USING (user_id = auth.uid()) WITH CHECK (user_id = auth.uid());

CREATE POLICY "physician reads health metrics during active consult" ON health_metrics
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM consultations c
            WHERE c.user_id = health_metrics.user_id
              AND c.physician_id = auth.uid()
              AND c.status IN ('confirmed','in_progress')
        )
    );

CREATE POLICY "admin read health metrics" ON health_metrics
    FOR SELECT USING (auth.role() IN ('admin','super_admin'));

-- consultation_slots: physician sees own, patient sees available
CREATE POLICY "physician sees own slots" ON consultation_slots
    FOR ALL USING (physician_id = auth.uid()) WITH CHECK (physician_id = auth.uid());

CREATE POLICY "patient sees available slots" ON consultation_slots
    FOR SELECT USING (booked = FALSE);

-- consultations: patient self + physician self + admin read
CREATE POLICY "patient self-access" ON consultations
    FOR ALL USING (user_id = auth.uid()) WITH CHECK (user_id = auth.uid());

CREATE POLICY "physician sees own consultations" ON consultations
    FOR ALL USING (physician_id = auth.uid()) WITH CHECK (physician_id = auth.uid());

CREATE POLICY "admin read consultations" ON consultations
    FOR SELECT USING (auth.role() IN ('admin','super_admin'));

-- audit_log: admin read only (append-only via trigger)
CREATE POLICY "admin read audit" ON audit_log
    FOR SELECT USING (auth.role() IN ('admin','super_admin'));

-- device_tokens: patient only (highly sensitive)
CREATE POLICY "patient self-access" ON device_tokens
    FOR ALL USING (user_id = auth.uid()) WITH CHECK (user_id = auth.uid());

-- brain_checkins: patient self + physician during consult + admin read
CREATE POLICY "patient self-access" ON brain_checkins
    FOR ALL USING (user_id = auth.uid()) WITH CHECK (user_id = auth.uid());

CREATE POLICY "physician reads during consult" ON brain_checkins
    FOR SELECT USING (
        EXISTS (
            SELECT 1 FROM consultations c
            WHERE c.user_id = brain_checkins.user_id
              AND c.physician_id = auth.uid()
              AND c.status IN ('confirmed','in_progress')
        )
    );

-- sleep_sessions: same pattern
CREATE POLICY "patient self-access" ON sleep_sessions
    FOR ALL USING (user_id = auth.uid()) WITH CHECK (user_id = auth.uid());

-- metric_baselines, scores, biological_ages: patient self
CREATE POLICY "patient self-access" ON metric_baselines
    FOR ALL USING (user_id = auth.uid()) WITH CHECK (user_id = auth.uid());

CREATE POLICY "patient self-access" ON scores
    FOR ALL USING (user_id = auth.uid()) WITH CHECK (user_id = auth.uid());

CREATE POLICY "patient self-access" ON biological_ages
    FOR ALL USING (user_id = auth.uid()) WITH CHECK (user_id = auth.uid());

-- score_weights: admin read
CREATE POLICY "admin read score weights" ON score_weights
    FOR SELECT USING (auth.role() IN ('admin','super_admin'));

-- Auto-create profile trigger on _users insert
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER LANGUAGE plpgsql SECURITY DEFINER SET search_path = public AS $$
BEGIN
    INSERT INTO public.users (user_id, role, full_name, phone, email)
    VALUES (
        NEW.id,
        COALESCE(NEW.metadata->>'role', 'patient'),
        NULLIF(TRIM(COALESCE(NEW.metadata->>'full_name', '')), ''),
        NULLIF(TRIM(COALESCE(NEW.phone, NEW.metadata->>'phone', '')), ''),
        NEW.email
    )
    ON CONFLICT (user_id) DO UPDATE SET email = excluded.email;
    RETURN NEW;
END;
$$;

-- Note: attach trigger manually if desired:
-- CREATE TRIGGER on_auth_user_created
--   AFTER INSERT ON _users
--   FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();
