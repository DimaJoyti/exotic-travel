'use client'

import React, { useState, useEffect } from 'react'
import { 
  Link, 
  Unlink, 
  CheckCircle, 
  AlertCircle, 
  Clock, 
  RefreshCw, 
  Settings, 
  ExternalLink,
  Shield,
  Zap,
  Activity,
  Plus,
  Trash2
} from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Progress } from '@/components/ui/progress'
import { FadeIn, StaggerContainer, StaggerItem } from '@/components/ui/animated'

interface IntegrationManagerProps {
  onIntegrationChange?: (integrations: Integration[]) => void
}

interface Integration {
  id: number
  platform: string
  accountId: string
  accountName?: string
  status: 'active' | 'inactive' | 'error' | 'expired'
  lastSync?: string
  connectedAt: string
  features: string[]
  metrics?: {
    campaigns: number
    spend: number
    impressions: number
  }
}

interface PlatformInfo {
  id: string
  name: string
  description: string
  icon: string
  color: string
  features: string[]
  isConnected: boolean
  integration?: Integration
}

export default function IntegrationManager({ onIntegrationChange }: IntegrationManagerProps) {
  const [integrations, setIntegrations] = useState<Integration[]>([])
  const [availablePlatforms, setAvailablePlatforms] = useState<PlatformInfo[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const [connectingPlatform, setConnectingPlatform] = useState<string | null>(null)

  // Mock data - in real implementation, this would come from API
  useEffect(() => {
    const mockIntegrations: Integration[] = [
      {
        id: 1,
        platform: 'google_ads',
        accountId: '123-456-7890',
        accountName: 'TechCorp Ads Account',
        status: 'active',
        lastSync: '2024-01-20T10:30:00Z',
        connectedAt: '2024-01-15T09:00:00Z',
        features: ['campaigns', 'keywords', 'audiences', 'metrics'],
        metrics: {
          campaigns: 12,
          spend: 15420,
          impressions: 2400000
        }
      },
      {
        id: 2,
        platform: 'facebook_ads',
        accountId: 'act_987654321',
        accountName: 'TechCorp Facebook Ads',
        status: 'active',
        lastSync: '2024-01-20T11:15:00Z',
        connectedAt: '2024-01-16T14:30:00Z',
        features: ['campaigns', 'audiences', 'metrics', 'lookalike_audiences'],
        metrics: {
          campaigns: 8,
          spend: 12800,
          impressions: 1800000
        }
      },
      {
        id: 3,
        platform: 'mailchimp',
        accountId: 'mc_abc123',
        accountName: 'TechCorp Email Marketing',
        status: 'expired',
        lastSync: '2024-01-18T16:45:00Z',
        connectedAt: '2024-01-10T11:20:00Z',
        features: ['email_campaigns', 'audiences', 'automation'],
        metrics: {
          campaigns: 5,
          spend: 299,
          impressions: 45000
        }
      }
    ]

    const mockPlatforms: PlatformInfo[] = [
      {
        id: 'google_ads',
        name: 'Google Ads',
        description: 'Search and display advertising on Google',
        icon: '🔍',
        color: 'bg-green-500',
        features: ['Search Ads', 'Display Ads', 'YouTube Ads', 'Shopping Ads'],
        isConnected: true,
        integration: mockIntegrations[0]
      },
      {
        id: 'facebook_ads',
        name: 'Meta Ads',
        description: 'Facebook and Instagram advertising',
        icon: '📘',
        color: 'bg-blue-600',
        features: ['Facebook Ads', 'Instagram Ads', 'Audience Network', 'Messenger Ads'],
        isConnected: true,
        integration: mockIntegrations[1]
      },
      {
        id: 'mailchimp',
        name: 'Mailchimp',
        description: 'Email marketing and automation',
        icon: '📧',
        color: 'bg-yellow-500',
        features: ['Email Campaigns', 'Automation', 'Landing Pages', 'Postcards'],
        isConnected: true,
        integration: mockIntegrations[2]
      },
      {
        id: 'linkedin_ads',
        name: 'LinkedIn Ads',
        description: 'Professional network advertising',
        icon: '💼',
        color: 'bg-blue-700',
        features: ['Sponsored Content', 'Message Ads', 'Dynamic Ads', 'Lead Gen Forms'],
        isConnected: false
      },
      {
        id: 'twitter_ads',
        name: 'Twitter Ads',
        description: 'Social media advertising on Twitter',
        icon: '🐦',
        color: 'bg-sky-500',
        features: ['Promoted Tweets', 'Promoted Accounts', 'Promoted Trends', 'Website Cards'],
        isConnected: false
      },
      {
        id: 'tiktok_ads',
        name: 'TikTok Ads',
        description: 'Video advertising on TikTok',
        icon: '🎵',
        color: 'bg-black',
        features: ['In-Feed Ads', 'Brand Takeover', 'TopView', 'Branded Effects'],
        isConnected: false
      }
    ]

    setIntegrations(mockIntegrations)
    setAvailablePlatforms(mockPlatforms)
  }, [])

  const getStatusColor = (status: Integration['status']) => {
    switch (status) {
      case 'active': return 'text-green-600 bg-green-50 border-green-200'
      case 'expired': return 'text-yellow-600 bg-yellow-50 border-yellow-200'
      case 'error': return 'text-red-600 bg-red-50 border-red-200'
      case 'inactive': return 'text-gray-600 bg-gray-50 border-gray-200'
      default: return 'text-gray-600 bg-gray-50 border-gray-200'
    }
  }

  const getStatusIcon = (status: Integration['status']) => {
    switch (status) {
      case 'active': return <CheckCircle className="h-4 w-4" />
      case 'expired': return <Clock className="h-4 w-4" />
      case 'error': return <AlertCircle className="h-4 w-4" />
      case 'inactive': return <AlertCircle className="h-4 w-4" />
      default: return <Clock className="h-4 w-4" />
    }
  }

  const handleConnect = async (platformId: string) => {
    setConnectingPlatform(platformId)
    setIsLoading(true)

    try {
      // Simulate OAuth flow
      await new Promise(resolve => setTimeout(resolve, 2000))

      // Mock successful connection
      const newIntegration: Integration = {
        id: Date.now(),
        platform: platformId,
        accountId: `acc_${Date.now()}`,
        accountName: `${platformId} Account`,
        status: 'active',
        connectedAt: new Date().toISOString(),
        features: ['campaigns', 'metrics'],
        metrics: {
          campaigns: 0,
          spend: 0,
          impressions: 0
        }
      }

      setIntegrations(prev => [...prev, newIntegration])
      setAvailablePlatforms(prev => prev.map(p => 
        p.id === platformId 
          ? { ...p, isConnected: true, integration: newIntegration }
          : p
      ))

      onIntegrationChange?.([...integrations, newIntegration])
    } catch (error) {
      console.error('Failed to connect platform:', error)
    } finally {
      setIsLoading(false)
      setConnectingPlatform(null)
    }
  }

  const handleDisconnect = async (platformId: string) => {
    setIsLoading(true)

    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000))

      setIntegrations(prev => prev.filter(i => i.platform !== platformId))
      setAvailablePlatforms(prev => prev.map(p => 
        p.id === platformId 
          ? { ...p, isConnected: false, integration: undefined }
          : p
      ))

      const updatedIntegrations = integrations.filter(i => i.platform !== platformId)
      onIntegrationChange?.(updatedIntegrations)
    } catch (error) {
      console.error('Failed to disconnect platform:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const handleRefresh = async (integrationId: number) => {
    setIsLoading(true)

    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000))

      setIntegrations(prev => prev.map(i => 
        i.id === integrationId 
          ? { ...i, status: 'active', lastSync: new Date().toISOString() }
          : i
      ))
    } catch (error) {
      console.error('Failed to refresh integration:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const formatNumber = (num: number) => {
    if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`
    if (num >= 1000) return `${(num / 1000).toFixed(1)}K`
    return num.toString()
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 0
    }).format(amount)
  }

  const connectedIntegrations = availablePlatforms.filter(p => p.isConnected)
  const availableIntegrations = availablePlatforms.filter(p => !p.isConnected)

  return (
    <div className="max-w-6xl mx-auto p-6 space-y-8">
      {/* Header */}
      <FadeIn>
        <div className="text-center">
          <h1 className="text-4xl font-bold text-gray-900 mb-4 flex items-center justify-center">
            <Link className="h-10 w-10 mr-3 text-blue-600" />
            Platform Integrations
          </h1>
          <p className="text-lg text-gray-600 max-w-2xl mx-auto">
            Connect your marketing platforms to sync campaigns, audiences, and metrics automatically
          </p>
        </div>
      </FadeIn>

      {/* Connected Integrations */}
      {connectedIntegrations.length > 0 && (
        <FadeIn delay={0.1}>
          <Card className="bg-white/80 backdrop-blur-sm border-0 shadow-lg">
            <CardHeader>
              <CardTitle className="flex items-center">
                <CheckCircle className="h-5 w-5 mr-2 text-green-600" />
                Connected Platforms ({connectedIntegrations.length})
              </CardTitle>
              <CardDescription>
                Your active platform integrations and their performance
              </CardDescription>
            </CardHeader>
            <CardContent>
              <StaggerContainer>
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                  {connectedIntegrations.map((platform) => (
                    <StaggerItem key={platform.id}>
                      <Card className="border border-gray-200 hover:shadow-md transition-all duration-200">
                        <CardContent className="p-6">
                          <div className="flex items-center justify-between mb-4">
                            <div className="flex items-center space-x-3">
                              <div className={`w-12 h-12 rounded-lg ${platform.color} flex items-center justify-center text-white text-xl`}>
                                {platform.icon}
                              </div>
                              <div>
                                <h3 className="font-semibold text-gray-900">{platform.name}</h3>
                                <p className="text-sm text-gray-600">{platform.integration?.accountName}</p>
                              </div>
                            </div>
                            <div className="flex items-center space-x-2">
                              <Badge className={`${getStatusColor(platform.integration!.status)} border`}>
                                {getStatusIcon(platform.integration!.status)}
                                <span className="ml-1 capitalize">{platform.integration!.status}</span>
                              </Badge>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => handleRefresh(platform.integration!.id)}
                                disabled={isLoading}
                              >
                                <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
                              </Button>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => handleDisconnect(platform.id)}
                                disabled={isLoading}
                              >
                                <Unlink className="h-4 w-4" />
                              </Button>
                            </div>
                          </div>

                          {platform.integration?.metrics && (
                            <div className="grid grid-cols-3 gap-4 mb-4">
                              <div className="text-center">
                                <p className="text-2xl font-bold text-gray-900">
                                  {platform.integration.metrics.campaigns}
                                </p>
                                <p className="text-xs text-gray-600">Campaigns</p>
                              </div>
                              <div className="text-center">
                                <p className="text-2xl font-bold text-gray-900">
                                  {formatCurrency(platform.integration.metrics.spend)}
                                </p>
                                <p className="text-xs text-gray-600">Spend</p>
                              </div>
                              <div className="text-center">
                                <p className="text-2xl font-bold text-gray-900">
                                  {formatNumber(platform.integration.metrics.impressions)}
                                </p>
                                <p className="text-xs text-gray-600">Impressions</p>
                              </div>
                            </div>
                          )}

                          <div className="flex flex-wrap gap-2 mb-4">
                            {platform.features.slice(0, 3).map((feature) => (
                              <Badge key={feature} variant="outline" className="text-xs">
                                {feature}
                              </Badge>
                            ))}
                            {platform.features.length > 3 && (
                              <Badge variant="outline" className="text-xs">
                                +{platform.features.length - 3} more
                              </Badge>
                            )}
                          </div>

                          {platform.integration?.lastSync && (
                            <p className="text-xs text-gray-500">
                              Last sync: {new Date(platform.integration.lastSync).toLocaleString()}
                            </p>
                          )}
                        </CardContent>
                      </Card>
                    </StaggerItem>
                  ))}
                </div>
              </StaggerContainer>
            </CardContent>
          </Card>
        </FadeIn>
      )}

      {/* Available Integrations */}
      {availableIntegrations.length > 0 && (
        <FadeIn delay={0.2}>
          <Card className="bg-white/80 backdrop-blur-sm border-0 shadow-lg">
            <CardHeader>
              <CardTitle className="flex items-center">
                <Plus className="h-5 w-5 mr-2 text-blue-600" />
                Available Platforms ({availableIntegrations.length})
              </CardTitle>
              <CardDescription>
                Connect additional platforms to expand your marketing reach
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {availableIntegrations.map((platform) => (
                  <motion.div
                    key={platform.id}
                    whileHover={{ scale: 1.02 }}
                    whileTap={{ scale: 0.98 }}
                  >
                    <Card className="border border-gray-200 hover:shadow-md transition-all duration-200 cursor-pointer">
                      <CardContent className="p-6">
                        <div className="text-center">
                          <div className={`w-16 h-16 rounded-lg ${platform.color} flex items-center justify-center text-white text-2xl mx-auto mb-4`}>
                            {platform.icon}
                          </div>
                          <h3 className="font-semibold text-gray-900 mb-2">{platform.name}</h3>
                          <p className="text-sm text-gray-600 mb-4">{platform.description}</p>
                          
                          <div className="flex flex-wrap gap-1 justify-center mb-4">
                            {platform.features.slice(0, 2).map((feature) => (
                              <Badge key={feature} variant="outline" className="text-xs">
                                {feature}
                              </Badge>
                            ))}
                            {platform.features.length > 2 && (
                              <Badge variant="outline" className="text-xs">
                                +{platform.features.length - 2}
                              </Badge>
                            )}
                          </div>

                          <Button
                            onClick={() => handleConnect(platform.id)}
                            disabled={isLoading || connectingPlatform === platform.id}
                            className="w-full"
                          >
                            {connectingPlatform === platform.id ? (
                              <>
                                <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
                                Connecting...
                              </>
                            ) : (
                              <>
                                <Link className="h-4 w-4 mr-2" />
                                Connect
                              </>
                            )}
                          </Button>
                        </div>
                      </CardContent>
                    </Card>
                  </motion.div>
                ))}
              </div>
            </CardContent>
          </Card>
        </FadeIn>
      )}

      {/* Integration Health Summary */}
      <FadeIn delay={0.3}>
        <Card className="bg-gradient-to-r from-blue-50 to-purple-50 border-0 shadow-lg">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-lg font-semibold text-gray-900 mb-2">Integration Health</h3>
                <p className="text-sm text-gray-600">
                  {connectedIntegrations.filter(p => p.integration?.status === 'active').length} of {connectedIntegrations.length} integrations are healthy
                </p>
              </div>
              <div className="flex items-center space-x-4">
                <div className="text-center">
                  <div className="text-2xl font-bold text-green-600">
                    {connectedIntegrations.filter(p => p.integration?.status === 'active').length}
                  </div>
                  <div className="text-xs text-gray-600">Active</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-yellow-600">
                    {connectedIntegrations.filter(p => p.integration?.status === 'expired').length}
                  </div>
                  <div className="text-xs text-gray-600">Expired</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-red-600">
                    {connectedIntegrations.filter(p => p.integration?.status === 'error').length}
                  </div>
                  <div className="text-xs text-gray-600">Error</div>
                </div>
              </div>
            </div>
            
            <div className="mt-4">
              <Progress 
                value={(connectedIntegrations.filter(p => p.integration?.status === 'active').length / Math.max(connectedIntegrations.length, 1)) * 100} 
                className="h-2"
              />
            </div>
          </CardContent>
        </Card>
      </FadeIn>
    </div>
  )
}
