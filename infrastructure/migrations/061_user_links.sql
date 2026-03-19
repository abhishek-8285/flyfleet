-- 061_user_links.sql

-- Step 1: Create the many-to-many relationship table for Linked Accounts
CREATE TABLE IF NOT EXISTS user_links (
    parent_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    linked_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    PRIMARY KEY (parent_user_id, linked_user_id),
    CONSTRAINT no_self_link CHECK (parent_user_id != linked_user_id)
);

-- Index the linked_user_id for faster reverse lookups if a user wants to sever all parents
CREATE INDEX IF NOT EXISTS idx_user_links_linked_user_id ON user_links(linked_user_id);

-- Step 2: Empower Security Events for Admin Shadowing and Impersonation Audits
-- This structurally answers "Who actually did this?" vs "Who were they acting as?"
ALTER TABLE security_events ADD COLUMN IF NOT EXISTS actor_id UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE security_events ADD COLUMN IF NOT EXISTS on_behalf_of UUID REFERENCES users(id) ON DELETE SET NULL;
