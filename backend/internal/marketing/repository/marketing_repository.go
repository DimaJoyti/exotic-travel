package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/exotic-travel-booking/backend/internal/models"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// MarketingRepository handles database operations for marketing entities
type MarketingRepository struct {
	db     *sqlx.DB
	tracer trace.Tracer
}

// NewMarketingRepository creates a new marketing repository
func NewMarketingRepository(db *sqlx.DB) *MarketingRepository {
	return &MarketingRepository{
		db:     db,
		tracer: otel.Tracer("marketing.repository"),
	}
}

// Campaign operations
func (r *MarketingRepository) CreateCampaign(ctx context.Context, campaign *models.Campaign) error {
	ctx, span := r.tracer.Start(ctx, "repository.create_campaign")
	defer span.End()

	query := `
		INSERT INTO campaigns (name, description, type, status, budget, start_date, end_date, 
			target_audience, objectives, platforms, created_by, brand_id)
		VALUES (:name, :description, :type, :status, :budget, :start_date, :end_date,
			:target_audience, :objectives, :platforms, :created_by, :brand_id)
		RETURNING id, created_at, updated_at`

	rows, err := r.db.NamedQueryContext(ctx, query, campaign)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create campaign: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&campaign.ID, &campaign.CreatedAt, &campaign.UpdatedAt)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to scan campaign result: %w", err)
		}
	}

	span.SetAttributes(attribute.Int("campaign.id", campaign.ID))
	return nil
}

func (r *MarketingRepository) GetCampaignByID(ctx context.Context, id int) (*models.Campaign, error) {
	ctx, span := r.tracer.Start(ctx, "repository.get_campaign_by_id")
	defer span.End()

	span.SetAttributes(attribute.Int("campaign.id", id))

	query := `
		SELECT c.*, b.name as brand_name, u.first_name, u.last_name
		FROM campaigns c
		LEFT JOIN brands b ON c.brand_id = b.id
		LEFT JOIN users u ON c.created_by = u.id
		WHERE c.id = $1`

	var campaign models.Campaign
	err := r.db.GetContext(ctx, &campaign, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("campaign not found")
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}

	return &campaign, nil
}

func (r *MarketingRepository) GetCampaigns(ctx context.Context, filters map[string]interface{}) ([]models.Campaign, error) {
	ctx, span := r.tracer.Start(ctx, "repository.get_campaigns")
	defer span.End()

	query := `
		SELECT c.*, b.name as brand_name, u.first_name, u.last_name
		FROM campaigns c
		LEFT JOIN brands b ON c.brand_id = b.id
		LEFT JOIN users u ON c.created_by = u.id
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	if status, ok := filters["status"]; ok {
		query += fmt.Sprintf(" AND c.status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	if campaignType, ok := filters["type"]; ok {
		query += fmt.Sprintf(" AND c.type = $%d", argIndex)
		args = append(args, campaignType)
		argIndex++
	}

	if createdBy, ok := filters["created_by"]; ok {
		query += fmt.Sprintf(" AND c.created_by = $%d", argIndex)
		args = append(args, createdBy)
		argIndex++
	}

	query += " ORDER BY c.created_at DESC"

	if limit, ok := filters["limit"]; ok {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, limit)
	}

	var campaigns []models.Campaign
	err := r.db.SelectContext(ctx, &campaigns, query, args...)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get campaigns: %w", err)
	}

	span.SetAttributes(attribute.Int("campaigns.count", len(campaigns)))
	return campaigns, nil
}

func (r *MarketingRepository) UpdateCampaign(ctx context.Context, campaign *models.Campaign) error {
	ctx, span := r.tracer.Start(ctx, "repository.update_campaign")
	defer span.End()

	span.SetAttributes(attribute.Int("campaign.id", campaign.ID))

	query := `
		UPDATE campaigns 
		SET name = :name, description = :description, type = :type, status = :status,
			budget = :budget, start_date = :start_date, end_date = :end_date,
			target_audience = :target_audience, objectives = :objectives, platforms = :platforms,
			updated_at = NOW()
		WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, campaign)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update campaign: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("campaign not found")
	}

	return nil
}

// Content operations
func (r *MarketingRepository) CreateContent(ctx context.Context, content *models.Content) error {
	ctx, span := r.tracer.Start(ctx, "repository.create_content")
	defer span.End()

	query := `
		INSERT INTO contents (campaign_id, type, title, body, platform, brand_voice, 
			seo_data, metadata, status, variation_group, parent_content_id, created_by)
		VALUES (:campaign_id, :type, :title, :body, :platform, :brand_voice,
			:seo_data, :metadata, :status, :variation_group, :parent_content_id, :created_by)
		RETURNING id, created_at, updated_at`

	rows, err := r.db.NamedQueryContext(ctx, query, content)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create content: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&content.ID, &content.CreatedAt, &content.UpdatedAt)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to scan content result: %w", err)
		}
	}

	span.SetAttributes(attribute.Int("content.id", content.ID))
	return nil
}

func (r *MarketingRepository) GetContentByID(ctx context.Context, id int) (*models.Content, error) {
	ctx, span := r.tracer.Start(ctx, "repository.get_content_by_id")
	defer span.End()

	span.SetAttributes(attribute.Int("content.id", id))

	query := `
		SELECT c.*, u.first_name, u.last_name
		FROM contents c
		LEFT JOIN users u ON c.created_by = u.id
		WHERE c.id = $1`

	var content models.Content
	err := r.db.GetContext(ctx, &content, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("content not found")
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get content: %w", err)
	}

	return &content, nil
}

func (r *MarketingRepository) GetContentByCampaign(ctx context.Context, campaignID int) ([]models.Content, error) {
	ctx, span := r.tracer.Start(ctx, "repository.get_content_by_campaign")
	defer span.End()

	span.SetAttributes(attribute.Int("campaign.id", campaignID))

	query := `
		SELECT c.*, u.first_name, u.last_name
		FROM contents c
		LEFT JOIN users u ON c.created_by = u.id
		WHERE c.campaign_id = $1
		ORDER BY c.created_at DESC`

	var contents []models.Content
	err := r.db.SelectContext(ctx, &contents, query, campaignID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get content by campaign: %w", err)
	}

	span.SetAttributes(attribute.Int("contents.count", len(contents)))
	return contents, nil
}

func (r *MarketingRepository) UpdateContent(ctx context.Context, content *models.Content) error {
	ctx, span := r.tracer.Start(ctx, "repository.update_content")
	defer span.End()

	span.SetAttributes(attribute.Int("content.id", content.ID))

	query := `
		UPDATE contents 
		SET title = :title, body = :body, platform = :platform, brand_voice = :brand_voice,
			seo_data = :seo_data, metadata = :metadata, status = :status, updated_at = NOW()
		WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, content)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update content: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("content not found")
	}

	return nil
}

// Brand operations
func (r *MarketingRepository) GetBrandByID(ctx context.Context, id int) (*models.Brand, error) {
	ctx, span := r.tracer.Start(ctx, "repository.get_brand_by_id")
	defer span.End()

	span.SetAttributes(attribute.Int("brand.id", id))

	query := `SELECT * FROM brands WHERE id = $1`

	var brand models.Brand
	err := r.db.GetContext(ctx, &brand, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("brand not found")
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get brand: %w", err)
	}

	return &brand, nil
}

func (r *MarketingRepository) GetBrands(ctx context.Context) ([]models.Brand, error) {
	ctx, span := r.tracer.Start(ctx, "repository.get_brands")
	defer span.End()

	query := `SELECT * FROM brands ORDER BY name`

	var brands []models.Brand
	err := r.db.SelectContext(ctx, &brands, query)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get brands: %w", err)
	}

	span.SetAttributes(attribute.Int("brands.count", len(brands)))
	return brands, nil
}

// Audience operations
func (r *MarketingRepository) CreateAudience(ctx context.Context, audience *models.Audience) error {
	ctx, span := r.tracer.Start(ctx, "repository.create_audience")
	defer span.End()

	query := `
		INSERT INTO audiences (name, description, demographics, interests, behaviors, 
			platform_data, size, created_by)
		VALUES (:name, :description, :demographics, :interests, :behaviors,
			:platform_data, :size, :created_by)
		RETURNING id, created_at, updated_at`

	rows, err := r.db.NamedQueryContext(ctx, query, audience)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create audience: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&audience.ID, &audience.CreatedAt, &audience.UpdatedAt)
		if err != nil {
			span.RecordError(err)
			return fmt.Errorf("failed to scan audience result: %w", err)
		}
	}

	span.SetAttributes(attribute.Int("audience.id", audience.ID))
	return nil
}

func (r *MarketingRepository) GetAudiences(ctx context.Context) ([]models.Audience, error) {
	ctx, span := r.tracer.Start(ctx, "repository.get_audiences")
	defer span.End()

	query := `
		SELECT a.*, u.first_name, u.last_name
		FROM audiences a
		LEFT JOIN users u ON a.created_by = u.id
		ORDER BY a.name`

	var audiences []models.Audience
	err := r.db.SelectContext(ctx, &audiences, query)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get audiences: %w", err)
	}

	span.SetAttributes(attribute.Int("audiences.count", len(audiences)))
	return audiences, nil
}
