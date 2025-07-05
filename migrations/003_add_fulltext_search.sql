-- Add full-text search to operations table

-- Add a tsvector column for full-text search
ALTER TABLE operations 
ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- Create a function to update the search vector
CREATE OR REPLACE FUNCTION operations_search_vector_trigger() RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('english', COALESCE(NEW.id, '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(NEW.type::text, '')), 'B') ||
    setweight(to_tsvector('english', COALESCE(NEW.error_message, '')), 'C') ||
    setweight(to_tsvector('english', COALESCE(NEW.correlation_id, '')), 'D');
  RETURN NEW;
END
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update search vector
DROP TRIGGER IF EXISTS operations_search_vector_update ON operations;
CREATE TRIGGER operations_search_vector_update
  BEFORE INSERT OR UPDATE ON operations
  FOR EACH ROW
  EXECUTE FUNCTION operations_search_vector_trigger();

-- Update existing rows
UPDATE operations 
SET search_vector = 
  setweight(to_tsvector('english', COALESCE(id, '')), 'A') ||
  setweight(to_tsvector('english', COALESCE(type::text, '')), 'B') ||
  setweight(to_tsvector('english', COALESCE(error_message, '')), 'C') ||
  setweight(to_tsvector('english', COALESCE(correlation_id, '')), 'D');

-- Create GIN index for fast full-text search
CREATE INDEX IF NOT EXISTS idx_operations_search_vector 
ON operations USING GIN(search_vector);

-- Also create a trigram index for fuzzy matching (requires pg_trgm extension)
-- This is optional but provides better fuzzy matching
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Create trigram indexes for fuzzy search on specific columns
CREATE INDEX IF NOT EXISTS idx_operations_id_trgm ON operations USING GIN (id gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_operations_type_trgm ON operations USING GIN ((type::text) gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_operations_error_trgm ON operations USING GIN (error_message gin_trgm_ops) WHERE error_message IS NOT NULL;