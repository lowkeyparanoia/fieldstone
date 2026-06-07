-- Postgres Extensions Pack for Fieldstone
-- Enables: pgvector (AI/RAG), pgmq (queues), PostGIS (geospatial)

-- pgvector: vector similarity search for AI/RAG
CREATE EXTENSION IF NOT EXISTS vector;

-- Create a generic vector table that vertical packs can extend
CREATE TABLE IF NOT EXISTS vector_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    table_name TEXT NOT NULL,
    record_id UUID NOT NULL,
    embedding vector(1536), -- OpenAI ada-002 dimension; adjust if needed
    content TEXT,
    metadata JSONB DEFAULT '{}',
    model TEXT DEFAULT 'text-embedding-ada-002',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- HNSW index for fast similarity search
CREATE INDEX IF NOT EXISTS idx_vector_embeddings_embedding 
ON vector_embeddings USING hnsw (embedding vector_cosine_ops);

-- Index for tenant-scoped queries
CREATE INDEX IF NOT EXISTS idx_vector_embeddings_tenant 
ON vector_embeddings(tenant_id, table_name);

-- RLS for vector embeddings
ALTER TABLE vector_embeddings ENABLE ROW LEVEL SECURITY;

CREATE POLICY "tenant isolation vector" ON vector_embeddings
    FOR ALL USING (tenant_id = current_setting('app.current_tenant', TRUE)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant', TRUE)::UUID);

-- Function: semantic search
CREATE OR REPLACE FUNCTION semantic_search(
    p_tenant_id UUID,
    p_table_name TEXT,
    p_query_embedding vector(1536),
    p_limit INTEGER DEFAULT 10,
    p_threshold FLOAT DEFAULT 0.7
) RETURNS TABLE (
    id UUID,
    record_id UUID,
    content TEXT,
    similarity FLOAT
) AS $$
BEGIN
    SET LOCAL app.current_tenant = p_tenant_id::TEXT;
    RETURN QUERY
    SELECT 
        ve.id,
        ve.record_id,
        ve.content,
        1 - (ve.embedding <=> p_query_embedding) AS similarity
    FROM vector_embeddings ve
    WHERE ve.tenant_id = p_tenant_id
      AND ve.table_name = p_table_name
      AND 1 - (ve.embedding <=> p_query_embedding) >= p_threshold
    ORDER BY ve.embedding <=> p_query_embedding
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
