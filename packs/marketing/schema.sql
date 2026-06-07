-- Marketing Pack Schema for Fieldstone (ChampIQ SDR / outreach model)
-- Enables: prospects, companies, signals, campaigns, sequences, email sends, suppression

CREATE TABLE IF NOT EXISTS companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    domain TEXT,
    industry TEXT,
    size TEXT,
    fit_score INTEGER CHECK (fit_score BETWEEN 0 AND 100),
    intent_band TEXT CHECK (intent_band IN ('hot','warm','cold')),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS prospects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID REFERENCES companies(id) ON DELETE SET NULL,
    email TEXT NOT NULL,
    first_name TEXT,
    last_name TEXT,
    title TEXT,
    seniority TEXT,
    phone TEXT,
    linkedin_url TEXT,
    consent_email BOOLEAN DEFAULT FALSE,
    consent_voice BOOLEAN DEFAULT FALSE,
    fit_score INTEGER CHECK (fit_score BETWEEN 0 AND 100),
    intent_band TEXT CHECK (intent_band IN ('hot','warm','cold')),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS signals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    prospect_id UUID REFERENCES prospects(id) ON DELETE CASCADE,
    signal_type TEXT NOT NULL CHECK (signal_type IN ('funding','hiring_sdrs','exec_change','compliance_hiring','tech_install','engagement')),
    description TEXT,
    source TEXT,
    occurred_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS campaigns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','active','paused','archived')),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sequences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID REFERENCES campaigns(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    step_order INTEGER NOT NULL DEFAULT 0,
    channel TEXT NOT NULL DEFAULT 'email' CHECK (channel IN ('email','linkedin','voice')),
    template_subject TEXT,
    template_body TEXT,
    delay_hours INTEGER DEFAULT 24,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS engagements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    prospect_id UUID NOT NULL REFERENCES prospects(id) ON DELETE CASCADE,
    sequence_id UUID REFERENCES sequences(id) ON DELETE SET NULL,
    campaign_id UUID REFERENCES campaigns(id) ON DELETE SET NULL,
    channel TEXT NOT NULL,
    direction TEXT NOT NULL CHECK (direction IN ('outbound','inbound')),
    status TEXT NOT NULL DEFAULT 'sent' CHECK (status IN ('sent','delivered','opened','clicked','replied','bounced')),
    content TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS suppression (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    domain TEXT,
    reason TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sending_domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    domain TEXT UNIQUE NOT NULL,
    dkim_verified BOOLEAN DEFAULT FALSE,
    spf_verified BOOLEAN DEFAULT FALSE,
    daily_cap INTEGER DEFAULT 100,
    warmed_up BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
