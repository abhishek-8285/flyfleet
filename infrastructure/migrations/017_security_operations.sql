-- Security Operations Module Schema
-- IP Rules, Access Reviews, Review Items

-- IP Access Rules
CREATE TABLE IF NOT EXISTS ip_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    cidr VARCHAR(45) NOT NULL,  -- e.g., "192.168.1.0/24" or single IP as "/32"
    action VARCHAR(10) NOT NULL CHECK (action IN ('ALLOW', 'BLOCK')),
    description TEXT,
    added_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_ip_rules_org ON ip_rules(org_id);
CREATE INDEX idx_ip_rules_action ON ip_rules(org_id, action);

-- Access Reviews
CREATE TABLE IF NOT EXISTS access_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    title VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'IN_PROGRESS', 'COMPLETED')),
    due_date TIMESTAMP WITH TIME ZONE,
    total_users INTEGER DEFAULT 0,
    reviewed_users INTEGER DEFAULT 0,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_access_reviews_org ON access_reviews(org_id);
CREATE INDEX idx_access_reviews_status ON access_reviews(org_id, status);

-- Review Items (individual user/role combinations to review)
CREATE TABLE IF NOT EXISTS review_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id UUID NOT NULL REFERENCES access_reviews(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    role_id UUID REFERENCES roles(id),
    current_status VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (current_status IN ('PENDING', 'APPROVED', 'REVOKED')),
    reviewer_id UUID REFERENCES users(id),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_review_items_review ON review_items(review_id);
CREATE INDEX idx_review_items_user ON review_items(user_id);
CREATE INDEX idx_review_items_status ON review_items(review_id, current_status);

-- Add device_id to auth_tokens if not exists (for session management)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'auth_tokens' AND column_name = 'device_id') THEN
        ALTER TABLE auth_tokens ADD COLUMN device_id VARCHAR(255);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'auth_tokens' AND column_name = 'created_at') THEN
        ALTER TABLE auth_tokens ADD COLUMN created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();
    END IF;
END $$;
