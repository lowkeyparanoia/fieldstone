-- PGMQ Extension for Fieldstone
-- Provides queue capabilities via Postgres

CREATE EXTENSION IF NOT EXISTS pgmq;

-- Create standard queues for different verticals
SELECT pgmq.create('emails');
SELECT pgmq.create('webhooks');
SELECT pgmq.create('enrichment');
SELECT pgmq.create('cdc_events');
SELECT pgmq.create('scoring');

-- Example: create a queue with retention
SELECT pgmq.create('high_priority');

-- Helper function to enqueue with tenant context
CREATE OR REPLACE FUNCTION pgmq_enqueue(
    queue_name TEXT,
    payload JSONB,
    tenant_id UUID DEFAULT NULL,
    delay INTEGER DEFAULT 0
) RETURNS BIGINT AS $$
DECLARE
    msg_id BIGINT;
    full_payload JSONB;
BEGIN
    full_payload := payload || jsonb_build_object(
        '_tenant_id', tenant_id,
        '_enqueued_at', NOW()
    );
    
    IF delay > 0 THEN
        SELECT * INTO msg_id FROM pgmq.send(queue_name, full_payload, delay);
    ELSE
        SELECT * INTO msg_id FROM pgmq.send(queue_name, full_payload);
    END IF;
    
    RETURN msg_id;
END;
$$ LANGUAGE plpgsql;

-- Helper function to dequeue with tenant filter
CREATE OR REPLACE FUNCTION pgmq_dequeue_tenant(
    queue_name TEXT,
    p_tenant_id UUID,
    vt INTEGER DEFAULT 30
) RETURNS TABLE (
    msg_id BIGINT,
    read_ct INTEGER,
    enqueued_at TIMESTAMPTZ,
    archived_at TIMESTAMPTZ,
    visible_at TIMESTAMPTZ,
    tenant_id UUID,
    payload JSONB
) AS $$
DECLARE
    msg RECORD;
BEGIN
    SELECT * INTO msg FROM pgmq.read(queue_name, vt, 1);
    
    IF msg IS NULL OR msg.message->>'_tenant_id' IS DISTINCT FROM p_tenant_id::TEXT THEN
        RETURN;
    END IF;
    
    msg_id := msg.msg_id;
    read_ct := msg.read_ct;
    enqueued_at := msg.enqueued_at;
    archived_at := msg.archived_at;
    visible_at := msg.visible_at;
    tenant_id := (msg.message->>'_tenant_id')::UUID;
    payload := msg.message;
    RETURN NEXT;
END;
$$ LANGUAGE plpgsql;
