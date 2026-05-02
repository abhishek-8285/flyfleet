-- Migration: 083_ticket_search_index
-- Description: Add GIN-based full-text search index to support_tickets

-- Ensure tags is TEXT[] (it might be JSONB from legacy infrastructure migrations)
DO $$ 
BEGIN 
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'support_tickets' AND column_name = 'tags' AND data_type = 'jsonb') THEN
        -- 1. Add temp column
        ALTER TABLE support_tickets ADD COLUMN tags_temp TEXT[] DEFAULT '{}';
        
        -- 2. Migrate data (UPDATE allows subqueries)
        UPDATE support_tickets SET tags_temp = ARRAY(
            SELECT jsonb_array_elements_text(tags)
        ) WHERE tags IS NOT NULL AND jsonb_typeof(tags) = 'array';
        
        -- 3. Swap columns
        ALTER TABLE support_tickets DROP COLUMN tags;
        ALTER TABLE support_tickets RENAME COLUMN tags_temp TO tags;
    END IF;
END $$;

-- Create an IMMUTABLE function for the search vector
CREATE OR REPLACE FUNCTION support_ticket_to_tsvector(sub TEXT, descr TEXT, tgs TEXT[])
RETURNS tsvector AS $$
BEGIN
    RETURN (
        setweight(to_tsvector('english', COALESCE(sub, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(descr, '')), 'B') ||
        setweight(to_tsvector('english', array_to_string(tgs, ' ')), 'C')
    );
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Create a functional GIN index using the immutable function
CREATE INDEX IF NOT EXISTS idx_support_tickets_search_gin ON support_tickets USING GIN (support_ticket_to_tsvector(subject, description, tags));

-- Note: Functional indexes can be used directly with the same expression in the WHERE clause.
-- Example query:
-- SELECT * FROM support_tickets 
-- WHERE support_ticket_to_tsvector(subject, description, tags) @@ websearch_to_tsquery('english', 'search query');
