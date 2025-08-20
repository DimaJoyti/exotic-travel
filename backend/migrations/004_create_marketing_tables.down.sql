-- Drop marketing tables in reverse order to handle foreign key constraints

-- Drop triggers first
DROP TRIGGER IF EXISTS update_integrations_updated_at ON integrations;
DROP TRIGGER IF EXISTS update_audiences_updated_at ON audiences;
DROP TRIGGER IF EXISTS update_assets_updated_at ON assets;
DROP TRIGGER IF EXISTS update_contents_updated_at ON contents;
DROP TRIGGER IF EXISTS update_campaigns_updated_at ON campaigns;
DROP TRIGGER IF EXISTS update_brands_updated_at ON brands;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_integrations_created_by;
DROP INDEX IF EXISTS idx_integrations_status;
DROP INDEX IF EXISTS idx_integrations_platform;

DROP INDEX IF EXISTS idx_metrics_recorded_at;
DROP INDEX IF EXISTS idx_metrics_platform;
DROP INDEX IF EXISTS idx_metrics_campaign_id;

DROP INDEX IF EXISTS idx_audiences_created_by;

DROP INDEX IF EXISTS idx_assets_type;
DROP INDEX IF EXISTS idx_assets_brand_id;
DROP INDEX IF EXISTS idx_assets_content_id;

DROP INDEX IF EXISTS idx_contents_parent_content_id;
DROP INDEX IF EXISTS idx_contents_variation_group;
DROP INDEX IF EXISTS idx_contents_status;
DROP INDEX IF EXISTS idx_contents_platform;
DROP INDEX IF EXISTS idx_contents_type;
DROP INDEX IF EXISTS idx_contents_campaign_id;

DROP INDEX IF EXISTS idx_campaigns_start_date;
DROP INDEX IF EXISTS idx_campaigns_created_by;
DROP INDEX IF EXISTS idx_campaigns_brand_id;
DROP INDEX IF EXISTS idx_campaigns_type;
DROP INDEX IF EXISTS idx_campaigns_status;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS integrations;
DROP TABLE IF EXISTS metrics;
DROP TABLE IF EXISTS audiences;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS contents;
DROP TABLE IF EXISTS campaigns;
DROP TABLE IF EXISTS brands;
