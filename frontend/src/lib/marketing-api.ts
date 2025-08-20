import { api } from './api'
import { User } from '@/types'

// Types
export interface Campaign {
  id: number
  name: string
  description: string
  type: CampaignType
  status: CampaignStatus
  budget: number
  spentBudget: number
  startDate: string
  endDate?: string
  targetAudience: Record<string, any>
  objectives: Record<string, any>
  platforms: Record<string, any>
  createdBy: number
  brandId: number
  createdAt: string
  updatedAt: string
  brand?: Brand
  creator?: User
  contents?: Content[]
  metrics?: Metrics
}

export interface Content {
  id: number
  campaignId: number
  type: ContentType
  title: string
  body: string
  platform: string
  brandVoice: string
  seoData: Record<string, any>
  metadata: Record<string, any>
  status: ContentStatus
  variationGroup?: string
  parentContentId?: number
  createdBy: number
  createdAt: string
  updatedAt: string
  campaign?: Campaign
  creator?: User
  assets?: Asset[]
}

export interface Brand {
  id: number
  name: string
  description: string
  voiceGuidelines: Record<string, any>
  visualIdentity: Record<string, any>
  colorPalette: Record<string, any>
  typography: Record<string, any>
  logoUrl: string
  brandAssets: Record<string, any>
  companyId: number
  createdAt: string
  updatedAt: string
}

export interface Asset {
  id: number
  contentId?: number
  brandId?: number
  type: AssetType
  name: string
  url: string
  metadata: Record<string, any>
  usageRights: Record<string, any>
  createdBy: number
  createdAt: string
  updatedAt: string
}

export interface Audience {
  id: number
  name: string
  description: string
  demographics: Record<string, any>
  interests: Record<string, any>
  behaviors: Record<string, any>
  platformData: Record<string, any>
  size: number
  createdBy: number
  createdAt: string
  updatedAt: string
}

export interface Metrics {
  id: number
  campaignId: number
  platform: string
  impressions: number
  clicks: number
  conversions: number
  spend: number
  revenue: number
  ctr: number
  cpc: number
  roas: number
  metricData: Record<string, any>
  recordedAt: string
  createdAt: string
}

export interface Integration {
  id: number
  platform: string
  accountId: string
  status: IntegrationStatus
  config: Record<string, any>
  lastSync?: string
  createdBy: number
  createdAt: string
  updatedAt: string
}

// Enums
export type CampaignType = 'social' | 'email' | 'display' | 'search' | 'video' | 'influencer'
export type CampaignStatus = 'draft' | 'active' | 'paused' | 'completed' | 'cancelled'
export type ContentType = 'ad' | 'social_post' | 'email' | 'blog' | 'landing' | 'video'
export type ContentStatus = 'draft' | 'review' | 'approved' | 'published' | 'archived'
export type AssetType = 'image' | 'video' | 'audio' | 'logo' | 'banner'
export type IntegrationStatus = 'active' | 'inactive' | 'error' | 'expired'

// Request/Response types
export interface GenerateContentRequest {
  campaignId: number
  contentType: ContentType
  platform: string
  title?: string
  brief: string
  keywords: string[]
  tone: string
  length: string
  callToAction: string
  generateVariations: boolean
  variationCount?: number
}

export interface GenerateContentResponse {
  success: boolean
  content: Content
  variations?: Content[]
  metadata: Record<string, any>
  message?: string
}

export interface CreateCampaignRequest {
  name: string
  description: string
  type: CampaignType
  budget: number
  startDate: string
  endDate?: string
  targetAudience: Record<string, any>
  objectives: Record<string, any>
  platforms: string[]
  brandId: number
}

export interface UpdateCampaignRequest extends Partial<CreateCampaignRequest> {
  status?: CampaignStatus
}

export interface CreateAudienceRequest {
  name: string
  description: string
  demographics: Record<string, any>
  interests: string[]
  behaviors: Record<string, any>
  platformData: Record<string, any>
}

// API Client
export class MarketingAPI {
  private baseUrl = '/api/v1/marketing'

  // Content Generation
  async generateContent(request: GenerateContentRequest): Promise<GenerateContentResponse> {
    const response = await api.post(`${this.baseUrl}/content/generate`, request)
    return response.data
  }

  async getContentHistory(campaignId: number): Promise<{ success: boolean; contents: Content[]; count: number }> {
    const response = await api.get(`${this.baseUrl}/content/history/${campaignId}`)
    return response.data
  }

  async regenerateContent(contentId: number, modifications: Record<string, any>): Promise<{ success: boolean; content: Content }> {
    const response = await api.post(`${this.baseUrl}/content/${contentId}/regenerate`, modifications)
    return response.data
  }

  async updateContentStatus(contentId: number, status: ContentStatus): Promise<{ success: boolean; content: Content }> {
    const response = await api.patch(`${this.baseUrl}/content/${contentId}`, { status })
    return response.data
  }

  // Campaign Management
  async createCampaign(request: CreateCampaignRequest): Promise<{ success: boolean; campaign: Campaign }> {
    const response = await api.post(`${this.baseUrl}/campaigns`, request)
    return response.data
  }

  async getCampaigns(params?: { status?: CampaignStatus; type?: CampaignType; limit?: number }): Promise<{ success: boolean; campaigns: Campaign[]; total: number }> {
    const response = await api.get(`${this.baseUrl}/campaigns`, { params })
    return response.data
  }

  async getCampaign(campaignId: number): Promise<{ success: boolean; campaign: Campaign }> {
    const response = await api.get(`${this.baseUrl}/campaigns/${campaignId}`)
    return response.data
  }

  async updateCampaign(campaignId: number, request: UpdateCampaignRequest): Promise<{ success: boolean; campaign: Campaign }> {
    const response = await api.patch(`${this.baseUrl}/campaigns/${campaignId}`, request)
    return response.data
  }

  async deleteCampaign(campaignId: number): Promise<{ success: boolean }> {
    const response = await api.delete(`${this.baseUrl}/campaigns/${campaignId}`)
    return response.data
  }

  // Brand Management
  async getBrands(): Promise<{ success: boolean; brands: Brand[] }> {
    const response = await api.get(`${this.baseUrl}/brands`)
    return response.data
  }

  async getBrand(brandId: number): Promise<{ success: boolean; brand: Brand }> {
    const response = await api.get(`${this.baseUrl}/brands/${brandId}`)
    return response.data
  }

  async updateBrand(brandId: number, updates: Partial<Brand>): Promise<{ success: boolean; brand: Brand }> {
    const response = await api.patch(`${this.baseUrl}/brands/${brandId}`, updates)
    return response.data
  }

  // Audience Management
  async createAudience(request: CreateAudienceRequest): Promise<{ success: boolean; audience: Audience }> {
    const response = await api.post(`${this.baseUrl}/audiences`, request)
    return response.data
  }

  async getAudiences(): Promise<{ success: boolean; audiences: Audience[] }> {
    const response = await api.get(`${this.baseUrl}/audiences`)
    return response.data
  }

  async getAudience(audienceId: number): Promise<{ success: boolean; audience: Audience }> {
    const response = await api.get(`${this.baseUrl}/audiences/${audienceId}`)
    return response.data
  }

  // Analytics
  async getCampaignMetrics(campaignId: number, timeRange?: { start: string; end: string }): Promise<{ success: boolean; metrics: Metrics[] }> {
    const response = await api.get(`${this.baseUrl}/campaigns/${campaignId}/metrics`, { params: timeRange })
    return response.data
  }

  async getDashboardStats(): Promise<{ 
    success: boolean
    stats: {
      totalCampaigns: number
      activeCampaigns: number
      contentGenerated: number
      totalReach: number
      engagement: number
      roi: number
      monthlySpend: number
      conversions: number
    }
  }> {
    const response = await api.get(`${this.baseUrl}/dashboard/stats`)
    return response.data
  }

  // Platform Integrations
  async getIntegrations(): Promise<{ success: boolean; integrations: Integration[] }> {
    const response = await api.get(`${this.baseUrl}/integrations`)
    return response.data
  }

  async createIntegration(platform: string, config: Record<string, any>): Promise<{ success: boolean; integration: Integration }> {
    const response = await api.post(`${this.baseUrl}/integrations`, { platform, config })
    return response.data
  }

  async updateIntegration(integrationId: number, config: Record<string, any>): Promise<{ success: boolean; integration: Integration }> {
    const response = await api.patch(`${this.baseUrl}/integrations/${integrationId}`, { config })
    return response.data
  }

  async syncIntegration(integrationId: number): Promise<{ success: boolean; lastSync: string }> {
    const response = await api.post(`${this.baseUrl}/integrations/${integrationId}/sync`)
    return response.data
  }

  // Asset Management
  async uploadAsset(file: File, metadata: { type: AssetType; contentId?: number; brandId?: number }): Promise<{ success: boolean; asset: Asset }> {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('metadata', JSON.stringify(metadata))

    const response = await api.post(`${this.baseUrl}/assets/upload`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
    return response.data
  }

  async getAssets(params?: { type?: AssetType; contentId?: number; brandId?: number }): Promise<{ success: boolean; assets: Asset[] }> {
    const response = await api.get(`${this.baseUrl}/assets`, { params })
    return response.data
  }

  // Health Check
  async healthCheck(): Promise<{ success: boolean; status: string; timestamp: string }> {
    const response = await api.get(`${this.baseUrl}/health`)
    return response.data
  }
}

// Export singleton instance
export const marketingAPI = new MarketingAPI()
