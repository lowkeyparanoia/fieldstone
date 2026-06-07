-- Healthcare Pack Schema for Fieldstone (100xLongevity model)
-- Enables: patient profiles, consents, health metrics, consultations, scoring, audit

-- Users extension (extends _users)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE REFERENCES _users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'patient' CHECK (role IN ('patient','physician','admin','super_admin')),
    full_name TEXT,
    phone TEXT,
    age INTEGER,
    sex TEXT,
    height_cm NUMERIC,
    weight_kg NUMERIC,
    bmi_category TEXT GENERATED ALWAYS AS (
        CASE
            WHEN weight_kg IS NULL OR height_cm IS NULL THEN NULL
            WHEN weight_kg / ((height_cm/100)^2) < 18.5 THEN 'underweight'
            WHEN weight_kg / ((height_cm/100)^2) < 25 THEN 'normal'
            WHEN weight_kg / ((height_cm/100)^2) < 30 THEN 'overweight'
            ELSE 'obese'
        END
    ) STORED,
    onboarding_complete BOOLEAN DEFAULT FALSE,
    zoom_user_id TEXT,
    specialty TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Consents (DPDP/HIPAA)
CREATE TABLE IF NOT EXISTS consents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    version TEXT NOT NULL,
    accepted_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    ip_hash TEXT,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Health metrics
CREATE TABLE IF NOT EXISTS health_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    metric_key TEXT NOT NULL,
    value NUMERIC NOT NULL,
    unit TEXT,
    source TEXT NOT NULL DEFAULT 'manual' CHECK (source IN ('manual','fitbit')),
    recorded_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Metric definitions
CREATE TABLE IF NOT EXISTS metric_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_key TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    unit TEXT,
    organ_system TEXT,
    ref_min NUMERIC,
    ref_max NUMERIC,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Consultation slots
CREATE TABLE IF NOT EXISTS consultation_slots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    physician_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    booked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Consultations
CREATE TABLE IF NOT EXISTS consultations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    physician_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    slot_id UUID REFERENCES consultation_slots(id),
    status TEXT NOT NULL DEFAULT 'confirmed' CHECK (status IN ('confirmed','in_progress','completed','cancelled','no_show')),
    reschedule_count INTEGER DEFAULT 0,
    zoom_meeting_id TEXT,
    zoom_join_url TEXT,
    zoom_start_url TEXT,
    video_type TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Audit log (immutable)
CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id UUID,
    action TEXT NOT NULL,
    target_table TEXT NOT NULL,
    target_id TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Device tokens (Fitbit etc)
CREATE TABLE IF NOT EXISTS device_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Brain checkins
CREATE TABLE IF NOT EXISTS brain_checkins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    brain_score INTEGER NOT NULL CHECK (brain_score BETWEEN 0 AND 100),
    responses JSONB DEFAULT '{}',
    source TEXT DEFAULT 'manual',
    notes TEXT,
    recorded_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Sleep sessions
CREATE TABLE IF NOT EXISTS sleep_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    night_of DATE NOT NULL,
    duration_min INTEGER,
    quality INTEGER CHECK (quality BETWEEN 1 AND 5),
    bedtime TIMESTAMPTZ,
    wake_time TIMESTAMPTZ,
    stages JSONB DEFAULT '{}',
    source TEXT DEFAULT 'manual',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Metric baselines (scoring engine)
CREATE TABLE IF NOT EXISTS metric_baselines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    metric_key TEXT NOT NULL,
    rolling_mean NUMERIC,
    rolling_std NUMERIC,
    n_days INTEGER,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, metric_key)
);

-- Scores
CREATE TABLE IF NOT EXISTS scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    score_type TEXT NOT NULL CHECK (score_type IN ('recovery','strain','stress','longevity')),
    value NUMERIC NOT NULL,
    confidence NUMERIC,
    components JSONB DEFAULT '{}',
    computed_for DATE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Biological ages
CREATE TABLE IF NOT EXISTS biological_ages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    domain TEXT NOT NULL,
    age_years NUMERIC NOT NULL,
    age_low NUMERIC,
    age_high NUMERIC,
    confidence NUMERIC,
    inputs JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Score weights
CREATE TABLE IF NOT EXISTS score_weights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    score_type TEXT NOT NULL,
    component TEXT NOT NULL,
    weight NUMERIC NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    approved_by UUID REFERENCES users(user_id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Score release gate
CREATE TABLE IF NOT EXISTS score_release (
    id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    released BOOLEAN DEFAULT FALSE,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
