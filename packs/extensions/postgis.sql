-- PostGIS Extension for Fieldstone
-- Provides geospatial capabilities

CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS postgis_topology;

-- Generic locations table for any vertical
CREATE TABLE IF NOT EXISTS locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    table_name TEXT NOT NULL,
    record_id UUID NOT NULL,
    name TEXT,
    geom GEOMETRY(POINT, 4326),
    latitude NUMERIC,
    longitude NUMERIC,
    address TEXT,
    city TEXT,
    country TEXT,
    postal_code TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Spatial index
CREATE INDEX IF NOT EXISTS idx_locations_geom ON locations USING GIST(geom);

-- Regular index for queries
CREATE INDEX IF NOT EXISTS idx_locations_tenant ON locations(tenant_id, table_name);

-- RLS
ALTER TABLE locations ENABLE ROW LEVEL SECURITY;

CREATE POLICY "tenant isolation locations" ON locations
    FOR ALL USING (tenant_id = current_setting('app.current_tenant', TRUE)::UUID)
    WITH CHECK (tenant_id = current_setting('app.current_tenant', TRUE)::UUID);

-- Function: find nearby locations
CREATE OR REPLACE FUNCTION nearby_locations(
    p_tenant_id UUID,
    p_table_name TEXT,
    p_lat NUMERIC,
    p_lon NUMERIC,
    p_radius_meters INTEGER DEFAULT 5000,
    p_limit INTEGER DEFAULT 20
) RETURNS TABLE (
    id UUID,
    record_id UUID,
    name TEXT,
    distance_meters FLOAT,
    latitude NUMERIC,
    longitude NUMERIC
) AS $$
BEGIN
    SET LOCAL app.current_tenant = p_tenant_id::TEXT;
    RETURN QUERY
    SELECT 
        l.id,
        l.record_id,
        l.name,
        ST_Distance(
            l.geom::geography,
            ST_SetSRID(ST_MakePoint(p_lon, p_lat), 4326)::geography
        )::FLOAT AS distance_meters,
        l.latitude,
        l.longitude
    FROM locations l
    WHERE l.tenant_id = p_tenant_id
      AND l.table_name = p_table_name
      AND ST_DWithin(
          l.geom::geography,
          ST_SetSRID(ST_MakePoint(p_lon, p_lat), 4326)::geography,
          p_radius_meters
      )
    ORDER BY distance_meters
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function: geocode an address (placeholder for external service integration)
CREATE OR REPLACE FUNCTION geocode_address(
    p_address TEXT
) RETURNS TABLE (
    latitude NUMERIC,
    longitude NUMERIC,
    formatted_address TEXT
) AS $$
BEGIN
    -- In production, integrate with a geocoding service
    -- This is a stub that returns NULL
    RETURN QUERY SELECT NULL::NUMERIC, NULL::NUMERIC, p_address;
END;
$$ LANGUAGE plpgsql;
