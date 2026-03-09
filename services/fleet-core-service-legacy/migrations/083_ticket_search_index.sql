-- Migration: 083_ticket_search_index
-- Description: Add GIN-based full-text search index to support_tickets

-- Create a functional GIN index for full-text search across title, description and tags
-- We use a combination of title (weighted 'A'), description (weighted 'B'), and tags (weighted 'C')
CREATE INDEX IF NOT EXISTS idx_support_tickets_search_gin ON support_tickets USING GIN (
    (
        setweight(to_tsvector('english', COALESCE(title, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(description, '')), 'B') ||
        setweight(to_tsvector('english', array_to_string(tags, ' ')), 'C')
    )
);

-- Note: Functional indexes can be used directly with the same expression in the WHERE clause.
-- Example query:
-- SELECT * FROM support_tickets 
-- WHERE (
--     setweight(to_tsvector('english', COALESCE(title, '')), 'A') ||
--     setweight(to_tsvector('english', COALESCE(description, '')), 'B') ||
--     setweight(to_tsvector('english', array_to_string(tags, ' ')), 'C')
-- ) @@ websearch_to_tsquery('english', 'search query');
