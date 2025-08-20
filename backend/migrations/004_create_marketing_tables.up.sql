-- Create marketing-related tables for the Generative AI marketing system

-- Brands table
CREATE TABLE IF NOT EXISTS brands (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    voice_guidelines JSONB DEFAULT '{}',
    visual_identity JSONB DEFAULT '{}',
    color_palette JSONB DEFAULT '{}',
    typography JSONB DEFAULT '{}',
    logo_url VARCHAR(500),
    brand_assets JSONB DEFAULT '{}',
    company_id INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Campaigns table
CREATE TABLE IF NOT EXISTS campaigns (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL CHECK (type IN ('social', 'email', 'display', 'search', 'video', 'influencer')),
    status VARCHAR(50) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'paused', 'completed', 'cancelled')),
    budget DECIMAL(12,2) NOT NULL DEFAULT 0,
    spent_budget DECIMAL(12,2) NOT NULL DEFAULT 0,
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE,
    target_audience JSONB DEFAULT '{}',
    objectives JSONB DEFAULT '{}',
    platforms JSONB DEFAULT '{}',
    created_by INTEGER NOT NULL REFERENCES users(id),
    brand_id INTEGER NOT NULL REFERENCES brands(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Contents table
CREATE TABLE IF NOT EXISTS contents (
    id SERIAL PRIMARY KEY,
    campaign_id INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL CHECK (type IN ('ad', 'social_post', 'email', 'blog', 'landing', 'video')),
    title VARCHAR(500) NOT NULL,
    body TEXT NOT NULL,
    platform VARCHAR(100) NOT NULL,
    brand_voice VARCHAR(255),
    seo_data JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'review', 'approved', 'published', 'archived')),
    variation_group VARCHAR(255),
    parent_content_id INTEGER REFERENCES contents(id),
    created_by INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Assets table
CREATE TABLE IF NOT EXISTS assets (
    id SERIAL PRIMARY KEY,
    content_id INTEGER REFERENCES contents(id) ON DELETE CASCADE,
    brand_id INTEGER REFERENCES brands(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL CHECK (type IN ('image', 'video', 'audio', 'logo', 'banner')),
    name VARCHAR(255) NOT NULL,
    url VARCHAR(1000) NOT NULL,
    metadata JSONB DEFAULT '{}',
    usage_rights JSONB DEFAULT '{}',
    created_by INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Audiences table
CREATE TABLE IF NOT EXISTS audiences (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    demographics JSONB DEFAULT '{}',
    interests JSONB DEFAULT '{}',
    behaviors JSONB DEFAULT '{}',
    platform_data JSONB DEFAULT '{}',
    size INTEGER DEFAULT 0,
    created_by INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Metrics table
CREATE TABLE IF NOT EXISTS metrics (
    id SERIAL PRIMARY KEY,
    campaign_id INTEGER NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    platform VARCHAR(100) NOT NULL,
    impressions BIGINT DEFAULT 0,
    clicks BIGINT DEFAULT 0,
    conversions BIGINT DEFAULT 0,
    spend DECIMAL(12,2) DEFAULT 0,
    revenue DECIMAL(12,2) DEFAULT 0,
    ctr DECIMAL(5,4) DEFAULT 0,
    cpc DECIMAL(8,2) DEFAULT 0,
    roas DECIMAL(8,2) DEFAULT 0,
    metric_data JSONB DEFAULT '{}',
    recorded_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Integrations table
CREATE TABLE IF NOT EXISTS integrations (
    id SERIAL PRIMARY KEY,
    platform VARCHAR(100) NOT NULL,
    account_id VARCHAR(255) NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'error', 'expired')),
    config JSONB DEFAULT '{}',
    last_sync TIMESTAMP WITH TIME ZONE,
    created_by INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_campaigns_status ON campaigns(status);
CREATE INDEX IF NOT EXISTS idx_campaigns_type ON campaigns(type);
CREATE INDEX IF NOT EXISTS idx_campaigns_brand_id ON campaigns(brand_id);
CREATE INDEX IF NOT EXISTS idx_campaigns_created_by ON campaigns(created_by);
CREATE INDEX IF NOT EXISTS idx_campaigns_start_date ON campaigns(start_date);

CREATE INDEX IF NOT EXISTS idx_contents_campaign_id ON contents(campaign_id);
CREATE INDEX IF NOT EXISTS idx_contents_type ON contents(type);
CREATE INDEX IF NOT EXISTS idx_contents_platform ON contents(platform);
CREATE INDEX IF NOT EXISTS idx_contents_status ON contents(status);
CREATE INDEX IF NOT EXISTS idx_contents_variation_group ON contents(variation_group);
CREATE INDEX IF NOT EXISTS idx_contents_parent_content_id ON contents(parent_content_id);

CREATE INDEX IF NOT EXISTS idx_assets_content_id ON assets(content_id);
CREATE INDEX IF NOT EXISTS idx_assets_brand_id ON assets(brand_id);
CREATE INDEX IF NOT EXISTS idx_assets_type ON assets(type);

CREATE INDEX IF NOT EXISTS idx_audiences_created_by ON audiences(created_by);

CREATE INDEX IF NOT EXISTS idx_metrics_campaign_id ON metrics(campaign_id);
CREATE INDEX IF NOT EXISTS idx_metrics_platform ON metrics(platform);
CREATE INDEX IF NOT EXISTS idx_metrics_recorded_at ON metrics(recorded_at);

CREATE INDEX IF NOT EXISTS idx_integrations_platform ON integrations(platform);
CREATE INDEX IF NOT EXISTS idx_integrations_status ON integrations(status);
CREATE INDEX IF NOT EXISTS idx_integrations_created_by ON integrations(created_by);

-- Create triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_brands_updated_at BEFORE UPDATE ON brands
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_campaigns_updated_at BEFORE UPDATE ON campaigns
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_contents_updated_at BEFORE UPDATE ON contents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_assets_updated_at BEFORE UPDATE ON assets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_audiences_updated_at BEFORE UPDATE ON audiences
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_integrations_updated_at BEFORE UPDATE ON integrations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert sample data for development
INSERT INTO brands (name, description, voice_guidelines, visual_identity, color_palette, company_id) VALUES
('TechCorp', 'Innovative technology solutions', 
 '{"personality": ["innovative", "professional", "approachable"], "values": ["innovation", "quality", "customer-focus"], "tone": "professional yet friendly"}',
 '{"style": "modern", "elements": ["clean lines", "minimalist", "tech-focused"]}',
 '{"primary": "#2563eb", "secondary": "#64748b", "accent": "#f59e0b"}',
 1),
('EcoLife', 'Sustainable living products',
 '{"personality": ["eco-friendly", "caring", "authentic"], "values": ["sustainability", "health", "community"], "tone": "warm and inspiring"}',
 '{"style": "natural", "elements": ["organic shapes", "earth tones", "nature-inspired"]}',
 '{"primary": "#16a34a", "secondary": "#84cc16", "accent": "#eab308"}',
 1);

-- Insert sample campaigns
INSERT INTO campaigns (name, description, type, status, budget, start_date, target_audience, objectives, platforms, created_by, brand_id) VALUES
('Summer Tech Launch', 'Launch campaign for new tech product', 'social', 'active', 10000.00, NOW(), 
 '{"age_range": "25-45", "interests": ["technology", "innovation"], "location": "US"}',
 '{"primary": "brand_awareness", "secondary": "lead_generation"}',
 '["facebook", "instagram", "linkedin"]',
 1, 1),
('Eco Product Awareness', 'Raise awareness for sustainable products', 'display', 'draft', 5000.00, NOW() + INTERVAL '1 week',
 '{"age_range": "30-55", "interests": ["sustainability", "health"], "location": "Global"}',
 '{"primary": "education", "secondary": "conversion"}',
 '["google", "facebook", "instagram"]',
 1, 2);

-- Insert sample content
INSERT INTO contents (campaign_id, type, title, body, platform, brand_voice, status, created_by) VALUES
(1, 'social_post', 'Introducing Revolutionary Tech', 
 'Discover the future of technology with our latest innovation. Built for professionals who demand excellence. #TechInnovation #FutureTech',
 'instagram', 'professional yet friendly', 'published', 1),
(2, 'ad', 'Go Green with EcoLife', 
 'Make a difference with sustainable choices. Our eco-friendly products help you live better while protecting the planet. Shop now!',
 'facebook', 'warm and inspiring', 'draft', 1);

-- Insert sample audiences
INSERT INTO audiences (name, description, demographics, interests, behaviors, created_by) VALUES
('Tech Enthusiasts', 'Early adopters of technology',
 '{"age": "25-45", "income": "high", "education": "college+"}',
 '["technology", "gadgets", "innovation", "startups"]',
 '{"online_behavior": "active_social_media", "purchase_behavior": "early_adopter"}',
 1),
('Eco-Conscious Consumers', 'Environmentally aware shoppers',
 '{"age": "30-55", "income": "medium-high", "lifestyle": "health-conscious"}',
 '["sustainability", "environment", "health", "organic"]',
 '{"online_behavior": "research_focused", "purchase_behavior": "value_conscious"}',
 1);
