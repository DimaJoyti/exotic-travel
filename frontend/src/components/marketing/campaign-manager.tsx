'use client'

import React, { useState, useEffect } from 'react'
import { 
  Target, 
  Plus, 
  Search, 
  Filter, 
  MoreHorizontal, 
  Play, 
  Pause, 
  Edit, 
  Trash2, 
  Copy, 
  BarChart3,
  Calendar,
  DollarSign,
  Users,
  TrendingUp,
  CheckCircle,
  AlertCircle,
  Clock,
  Eye
} from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { FadeIn, StaggerContainer, StaggerItem } from '@/components/ui/animated'
import { Campaign, CampaignStatus, CampaignType } from '@/lib/marketing-api'

interface CampaignManagerProps {
  onCreateCampaign?: () => void
  onEditCampaign?: (campaign: Campaign) => void
  onViewAnalytics?: (campaign: Campaign) => void
}

interface CampaignWithMetrics extends Campaign {
  metrics: {
    impressions: number
    clicks: number
    conversions: number
    spend: number
    roas: number
    ctr: number
  }
}

export default function CampaignManager({ onCreateCampaign, onEditCampaign, onViewAnalytics }: CampaignManagerProps) {
  const [campaigns, setCampaigns] = useState<CampaignWithMetrics[]>([])
  const [filteredCampaigns, setFilteredCampaigns] = useState<CampaignWithMetrics[]>([])
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<CampaignStatus | 'all'>('all')
  const [typeFilter, setTypeFilter] = useState<CampaignType | 'all'>('all')
  const [sortBy, setSortBy] = useState<'name' | 'created' | 'spend' | 'roas'>('created')
  const [isLoading, setIsLoading] = useState(false)
  const [activeTab, setActiveTab] = useState('all')

  // Mock data - in real implementation, this would come from API
  useEffect(() => {
    const mockCampaigns: CampaignWithMetrics[] = [
      {
        id: 1,
        name: 'Summer Product Launch',
        description: 'Launch campaign for new summer collection',
        type: 'social',
        status: 'active',
        budget: 15000,
        spentBudget: 8500,
        startDate: '2024-01-15T00:00:00Z',
        endDate: '2024-02-15T00:00:00Z',
        targetAudience: { demographics: { age: '25-45' }, interests: ['fashion', 'lifestyle'] },
        objectives: { primary: 'brand_awareness', secondary: 'conversions' },
        platforms: { platforms: ['facebook', 'instagram', 'google'] },
        createdBy: 1,
        brandId: 1,
        createdAt: '2024-01-10T00:00:00Z',
        updatedAt: '2024-01-20T00:00:00Z',
        metrics: {
          impressions: 1250000,
          clicks: 42500,
          conversions: 1275,
          spend: 8500,
          roas: 4.2,
          ctr: 3.4
        }
      },
      {
        id: 2,
        name: 'Holiday Email Campaign',
        description: 'Email marketing for holiday season',
        type: 'email',
        status: 'completed',
        budget: 5000,
        spentBudget: 4800,
        startDate: '2023-12-01T00:00:00Z',
        endDate: '2023-12-31T00:00:00Z',
        targetAudience: { demographics: { age: '30-55' }, interests: ['shopping', 'gifts'] },
        objectives: { primary: 'conversions', secondary: 'retention' },
        platforms: { platforms: ['email', 'facebook'] },
        createdBy: 1,
        brandId: 1,
        createdAt: '2023-11-25T00:00:00Z',
        updatedAt: '2024-01-05T00:00:00Z',
        metrics: {
          impressions: 850000,
          clicks: 25500,
          conversions: 892,
          spend: 4800,
          roas: 5.8,
          ctr: 3.0
        }
      },
      {
        id: 3,
        name: 'Brand Awareness Video',
        description: 'YouTube and social video campaign',
        type: 'video',
        status: 'paused',
        budget: 12000,
        spentBudget: 3200,
        startDate: '2024-01-20T00:00:00Z',
        endDate: '2024-03-20T00:00:00Z',
        targetAudience: { demographics: { age: '18-35' }, interests: ['entertainment', 'technology'] },
        objectives: { primary: 'brand_awareness', secondary: 'engagement' },
        platforms: { platforms: ['youtube', 'tiktok', 'instagram'] },
        createdBy: 1,
        brandId: 1,
        createdAt: '2024-01-18T00:00:00Z',
        updatedAt: '2024-01-25T00:00:00Z',
        metrics: {
          impressions: 420000,
          clicks: 12600,
          conversions: 156,
          spend: 3200,
          roas: 2.1,
          ctr: 3.0
        }
      },
      {
        id: 4,
        name: 'Search Marketing Q1',
        description: 'Google Ads search campaign for Q1',
        type: 'search',
        status: 'draft',
        budget: 20000,
        spentBudget: 0,
        startDate: '2024-02-01T00:00:00Z',
        endDate: '2024-04-30T00:00:00Z',
        targetAudience: { demographics: { age: '25-50' }, interests: ['business', 'productivity'] },
        objectives: { primary: 'lead_generation', secondary: 'conversions' },
        platforms: { platforms: ['google'] },
        createdBy: 1,
        brandId: 1,
        createdAt: '2024-01-28T00:00:00Z',
        updatedAt: '2024-01-28T00:00:00Z',
        metrics: {
          impressions: 0,
          clicks: 0,
          conversions: 0,
          spend: 0,
          roas: 0,
          ctr: 0
        }
      }
    ]
    setCampaigns(mockCampaigns)
    setFilteredCampaigns(mockCampaigns)
  }, [])

  // Filter and search logic
  useEffect(() => {
    let filtered = campaigns

    // Search filter
    if (searchQuery) {
      filtered = filtered.filter(campaign =>
        campaign.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        campaign.description.toLowerCase().includes(searchQuery.toLowerCase())
      )
    }

    // Status filter
    if (statusFilter !== 'all') {
      filtered = filtered.filter(campaign => campaign.status === statusFilter)
    }

    // Type filter
    if (typeFilter !== 'all') {
      filtered = filtered.filter(campaign => campaign.type === typeFilter)
    }

    // Sort
    filtered.sort((a, b) => {
      switch (sortBy) {
        case 'name':
          return a.name.localeCompare(b.name)
        case 'spend':
          return b.spentBudget - a.spentBudget
        case 'roas':
          return b.metrics.roas - a.metrics.roas
        case 'created':
        default:
          return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
      }
    })

    setFilteredCampaigns(filtered)
  }, [campaigns, searchQuery, statusFilter, typeFilter, sortBy])

  const getStatusColor = (status: CampaignStatus) => {
    switch (status) {
      case 'active': return 'text-green-600 bg-green-50 border-green-200'
      case 'paused': return 'text-yellow-600 bg-yellow-50 border-yellow-200'
      case 'completed': return 'text-blue-600 bg-blue-50 border-blue-200'
      case 'cancelled': return 'text-red-600 bg-red-50 border-red-200'
      default: return 'text-gray-600 bg-gray-50 border-gray-200'
    }
  }

  const getStatusIcon = (status: CampaignStatus) => {
    switch (status) {
      case 'active': return <Play className="h-3 w-3" />
      case 'paused': return <Pause className="h-3 w-3" />
      case 'completed': return <CheckCircle className="h-3 w-3" />
      case 'cancelled': return <AlertCircle className="h-3 w-3" />
      default: return <Clock className="h-3 w-3" />
    }
  }

  const getTypeIcon = (type: CampaignType) => {
    switch (type) {
      case 'social': return '📱'
      case 'email': return '📧'
      case 'display': return '🖼️'
      case 'search': return '🔍'
      case 'video': return '🎬'
      case 'influencer': return '⭐'
      default: return '🎯'
    }
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 0
    }).format(amount)
  }

  const formatNumber = (num: number) => {
    if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`
    if (num >= 1000) return `${(num / 1000).toFixed(1)}K`
    return num.toString()
  }

  const handleCampaignAction = async (action: string, campaign: CampaignWithMetrics) => {
    setIsLoading(true)
    
    try {
      switch (action) {
        case 'play':
          // Update campaign status to active
          setCampaigns(prev => prev.map(c => 
            c.id === campaign.id ? { ...c, status: 'active' as CampaignStatus } : c
          ))
          break
        case 'pause':
          // Update campaign status to paused
          setCampaigns(prev => prev.map(c => 
            c.id === campaign.id ? { ...c, status: 'paused' as CampaignStatus } : c
          ))
          break
        case 'edit':
          onEditCampaign?.(campaign)
          break
        case 'analytics':
          onViewAnalytics?.(campaign)
          break
        case 'duplicate':
          // Create a copy of the campaign
          const newCampaign = {
            ...campaign,
            id: Date.now(),
            name: `${campaign.name} (Copy)`,
            status: 'draft' as CampaignStatus,
            spentBudget: 0,
            createdAt: new Date().toISOString(),
            metrics: { ...campaign.metrics, spend: 0, impressions: 0, clicks: 0, conversions: 0 }
          }
          setCampaigns(prev => [newCampaign, ...prev])
          break
        case 'delete':
          setCampaigns(prev => prev.filter(c => c.id !== campaign.id))
          break
      }
    } catch (error) {
      console.error('Campaign action failed:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const getTabCounts = () => {
    return {
      all: campaigns.length,
      active: campaigns.filter(c => c.status === 'active').length,
      paused: campaigns.filter(c => c.status === 'paused').length,
      draft: campaigns.filter(c => c.status === 'draft').length,
      completed: campaigns.filter(c => c.status === 'completed').length
    }
  }

  const tabCounts = getTabCounts()

  return (
    <div className="max-w-7xl mx-auto p-6 space-y-6">
      {/* Header */}
      <FadeIn>
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900 flex items-center">
              <Target className="h-8 w-8 mr-3 text-blue-600" />
              Campaign Manager
            </h1>
            <p className="text-lg text-gray-600 mt-1">
              Manage and monitor your marketing campaigns
            </p>
          </div>
          <Button 
            onClick={onCreateCampaign}
            className="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700"
          >
            <Plus className="h-4 w-4 mr-2" />
            Create Campaign
          </Button>
        </div>
      </FadeIn>

      {/* Filters and Search */}
      <FadeIn delay={0.1}>
        <Card className="bg-white/80 backdrop-blur-sm border-0 shadow-lg">
          <CardContent className="p-6">
            <div className="flex flex-col md:flex-row gap-4">
              <div className="flex-1">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
                  <Input
                    placeholder="Search campaigns..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-10"
                  />
                </div>
              </div>
              <div className="flex gap-3">
                <Select value={statusFilter} onValueChange={(value) => setStatusFilter(value as CampaignStatus | 'all')}>
                  <SelectTrigger className="w-40">
                    <SelectValue placeholder="Status" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Status</SelectItem>
                    <SelectItem value="active">Active</SelectItem>
                    <SelectItem value="paused">Paused</SelectItem>
                    <SelectItem value="draft">Draft</SelectItem>
                    <SelectItem value="completed">Completed</SelectItem>
                  </SelectContent>
                </Select>
                <Select value={typeFilter} onValueChange={(value) => setTypeFilter(value as CampaignType | 'all')}>
                  <SelectTrigger className="w-40">
                    <SelectValue placeholder="Type" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Types</SelectItem>
                    <SelectItem value="social">Social</SelectItem>
                    <SelectItem value="email">Email</SelectItem>
                    <SelectItem value="display">Display</SelectItem>
                    <SelectItem value="search">Search</SelectItem>
                    <SelectItem value="video">Video</SelectItem>
                    <SelectItem value="influencer">Influencer</SelectItem>
                  </SelectContent>
                </Select>
                <Select value={sortBy} onValueChange={(value) => setSortBy(value as any)}>
                  <SelectTrigger className="w-40">
                    <SelectValue placeholder="Sort by" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="created">Created Date</SelectItem>
                    <SelectItem value="name">Name</SelectItem>
                    <SelectItem value="spend">Spend</SelectItem>
                    <SelectItem value="roas">ROAS</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          </CardContent>
        </Card>
      </FadeIn>

      {/* Campaign Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid w-full grid-cols-5">
          <TabsTrigger value="all">All ({tabCounts.all})</TabsTrigger>
          <TabsTrigger value="active">Active ({tabCounts.active})</TabsTrigger>
          <TabsTrigger value="paused">Paused ({tabCounts.paused})</TabsTrigger>
          <TabsTrigger value="draft">Draft ({tabCounts.draft})</TabsTrigger>
          <TabsTrigger value="completed">Completed ({tabCounts.completed})</TabsTrigger>
        </TabsList>

        <TabsContent value={activeTab} className="space-y-4">
          <StaggerContainer>
            <AnimatePresence>
              {filteredCampaigns
                .filter(campaign => activeTab === 'all' || campaign.status === activeTab)
                .map((campaign) => (
                <StaggerItem key={campaign.id}>
                  <motion.div
                    layout
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: -20 }}
                    transition={{ duration: 0.2 }}
                  >
                    <Card className="bg-white/80 backdrop-blur-sm border-0 shadow-lg hover:shadow-xl transition-all duration-300">
                      <CardContent className="p-6">
                        <div className="flex items-center justify-between mb-4">
                          <div className="flex items-center space-x-3">
                            <div className="text-2xl">{getTypeIcon(campaign.type)}</div>
                            <div>
                              <h3 className="text-lg font-semibold text-gray-900">{campaign.name}</h3>
                              <p className="text-sm text-gray-600">{campaign.description}</p>
                            </div>
                          </div>
                          <div className="flex items-center space-x-3">
                            <Badge className={`${getStatusColor(campaign.status)} border`}>
                              {getStatusIcon(campaign.status)}
                              <span className="ml-1 capitalize">{campaign.status}</span>
                            </Badge>
                            <DropdownMenu>
                              <DropdownMenuTrigger asChild>
                                <Button variant="ghost" size="sm">
                                  <MoreHorizontal className="h-4 w-4" />
                                </Button>
                              </DropdownMenuTrigger>
                              <DropdownMenuContent align="end">
                                {campaign.status === 'active' ? (
                                  <DropdownMenuItem onClick={() => handleCampaignAction('pause', campaign)}>
                                    <Pause className="h-4 w-4 mr-2" />
                                    Pause Campaign
                                  </DropdownMenuItem>
                                ) : campaign.status !== 'completed' ? (
                                  <DropdownMenuItem onClick={() => handleCampaignAction('play', campaign)}>
                                    <Play className="h-4 w-4 mr-2" />
                                    Start Campaign
                                  </DropdownMenuItem>
                                ) : null}
                                <DropdownMenuItem onClick={() => handleCampaignAction('edit', campaign)}>
                                  <Edit className="h-4 w-4 mr-2" />
                                  Edit Campaign
                                </DropdownMenuItem>
                                <DropdownMenuItem onClick={() => handleCampaignAction('analytics', campaign)}>
                                  <BarChart3 className="h-4 w-4 mr-2" />
                                  View Analytics
                                </DropdownMenuItem>
                                <DropdownMenuItem onClick={() => handleCampaignAction('duplicate', campaign)}>
                                  <Copy className="h-4 w-4 mr-2" />
                                  Duplicate
                                </DropdownMenuItem>
                                <DropdownMenuItem 
                                  onClick={() => handleCampaignAction('delete', campaign)}
                                  className="text-red-600"
                                >
                                  <Trash2 className="h-4 w-4 mr-2" />
                                  Delete
                                </DropdownMenuItem>
                              </DropdownMenuContent>
                            </DropdownMenu>
                          </div>
                        </div>

                        <div className="grid grid-cols-2 md:grid-cols-6 gap-4 mb-4">
                          <div>
                            <p className="text-xs text-gray-600 uppercase tracking-wide">Budget</p>
                            <p className="text-sm font-semibold">{formatCurrency(campaign.budget)}</p>
                            <p className="text-xs text-gray-500">
                              {formatCurrency(campaign.spentBudget)} spent
                            </p>
                          </div>
                          <div>
                            <p className="text-xs text-gray-600 uppercase tracking-wide">Impressions</p>
                            <p className="text-sm font-semibold">{formatNumber(campaign.metrics.impressions)}</p>
                          </div>
                          <div>
                            <p className="text-xs text-gray-600 uppercase tracking-wide">Clicks</p>
                            <p className="text-sm font-semibold">{formatNumber(campaign.metrics.clicks)}</p>
                            <p className="text-xs text-gray-500">{campaign.metrics.ctr}% CTR</p>
                          </div>
                          <div>
                            <p className="text-xs text-gray-600 uppercase tracking-wide">Conversions</p>
                            <p className="text-sm font-semibold">{formatNumber(campaign.metrics.conversions)}</p>
                          </div>
                          <div>
                            <p className="text-xs text-gray-600 uppercase tracking-wide">ROAS</p>
                            <p className="text-sm font-semibold text-green-600">{campaign.metrics.roas}x</p>
                          </div>
                          <div>
                            <p className="text-xs text-gray-600 uppercase tracking-wide">Platforms</p>
                            <div className="flex flex-wrap gap-1">
                              {(campaign.platforms as any)?.platforms?.slice(0, 3).map((platform: string) => (
                                <Badge key={platform} variant="outline" className="text-xs">
                                  {platform}
                                </Badge>
                              ))}
                              {(campaign.platforms as any)?.platforms?.length > 3 && (
                                <Badge variant="outline" className="text-xs">
                                  +{(campaign.platforms as any).platforms.length - 3}
                                </Badge>
                              )}
                            </div>
                          </div>
                        </div>

                        <div className="flex items-center justify-between text-xs text-gray-500">
                          <span>
                            Created {new Date(campaign.createdAt).toLocaleDateString()}
                          </span>
                          <span>
                            {campaign.startDate && `Runs ${new Date(campaign.startDate).toLocaleDateString()}`}
                            {campaign.endDate && ` - ${new Date(campaign.endDate).toLocaleDateString()}`}
                          </span>
                        </div>
                      </CardContent>
                    </Card>
                  </motion.div>
                </StaggerItem>
              ))}
            </AnimatePresence>
          </StaggerContainer>

          {filteredCampaigns.filter(campaign => activeTab === 'all' || campaign.status === activeTab).length === 0 && (
            <FadeIn>
              <Card className="bg-white/80 backdrop-blur-sm border-0 shadow-lg">
                <CardContent className="p-12 text-center">
                  <Target className="h-16 w-16 text-gray-400 mx-auto mb-4" />
                  <h3 className="text-lg font-semibold text-gray-900 mb-2">No campaigns found</h3>
                  <p className="text-gray-600 mb-6">
                    {searchQuery || statusFilter !== 'all' || typeFilter !== 'all'
                      ? 'Try adjusting your filters or search terms'
                      : 'Create your first campaign to get started'
                    }
                  </p>
                  {!searchQuery && statusFilter === 'all' && typeFilter === 'all' && (
                    <Button onClick={onCreateCampaign}>
                      <Plus className="h-4 w-4 mr-2" />
                      Create Campaign
                    </Button>
                  )}
                </CardContent>
              </Card>
            </FadeIn>
          )}
        </TabsContent>
      </Tabs>
    </div>
  )
}
