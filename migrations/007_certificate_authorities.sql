-- Create trigger function if it doesn't exist
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add certificate authorities table
CREATE TABLE IF NOT EXISTS certificate_authorities (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    certificate TEXT NOT NULL, -- PEM encoded certificate
    fingerprint TEXT NOT NULL, -- SHA256 fingerprint for quick lookup
    subject TEXT NOT NULL, -- Certificate subject DN
    issuer TEXT NOT NULL, -- Certificate issuer DN
    not_before TIMESTAMP WITH TIME ZONE NOT NULL,
    not_after TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT NOT NULL,
    updated_by TEXT NOT NULL,
    CONSTRAINT certificate_authorities_name_unique UNIQUE (name),
    CONSTRAINT certificate_authorities_fingerprint_unique UNIQUE (fingerprint)
);

-- Add index for active certificates
CREATE INDEX idx_certificate_authorities_active ON certificate_authorities(is_active) WHERE is_active = true;

-- Add index for expiry monitoring
CREATE INDEX idx_certificate_authorities_expiry ON certificate_authorities(not_after) WHERE is_active = true;

-- Add trigger to update updated_at timestamp
CREATE TRIGGER update_certificate_authorities_updated_at BEFORE UPDATE ON certificate_authorities
    FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();