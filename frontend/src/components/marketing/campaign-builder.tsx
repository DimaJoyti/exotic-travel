'use client'

import React, { useState, useEffect } from 'react'
import { 
  Target, 
  Users, 
  Calendar, 
  DollarSign, 
  Settings, 
  Plus,
  Trash2,
  Edit,
  Play,
  Pause,
  BarChart3,
  CheckCircle,
  AlertCircle,
  Clock
} from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { FadeIn, StaggerContainer, StaggerItem } from '@/components/ui/animated'
import { Campaign, CampaignType, CampaignStatus } from '@/lib/marketing-api'

interface CampaignBuilderProps {
  campaign?: Campaign
  onSave?: (campaign: Campaign) => void
  onCancel?: () => void
}

interface CampaignFormData {
  name: string
  description: string
  type: CampaignType | ''
  budget: number
  startDate: string
  endDate: string
  targetAudience: {
    demographics: Record<string, any>
    interests: string[]
    locations: string[]
  }
  objectives: string[]
  platforms: string[]
}

const campaignTypes: { value: CampaignType; label: string; description: string; icon: string }[] = [
  { value: 'social', label: 'Social Media', description: 'Facebook, Instagram, Twitter campaigns', icon: '📱' },
  { value: 'email', label: 'Email Marketing', description: 'Newsletter and email campaigns', icon: '📧' },
  { value: 'display', label: 'Display Advertising', description: 'Banner and display ads', icon: '🖼️' },
  { value: 'search', label: 'Search Marketing', description: 'Google Ads and search campaigns', icon: '🔍' },
  { value: 'video', label: 'Video Marketing', description: 'YouTube and video campaigns', icon: '🎬' },
  { value: 'influencer', label: 'Influencer Marketing', description: 'Influencer partnerships', icon: '⭐' }
]

const platforms = [
  { value: 'facebook', label: 'Facebook', color: 'bg-blue-600' },
  { value: 'instagram', label: 'Instagram', color: 'bg-pink-600' },
  { value: 'twitter', label: 'Twitter', color: 'bg-sky-500' },
  { value: 'linkedin', label: 'LinkedIn', color: 'bg-blue-700' },
  { value: 'youtube', label: 'YouTube', color: 'bg-red-600' },
  { value: 'google', label: 'Google Ads', color: 'bg-green-600' },
  { value: 'tiktok', label: 'TikTok', color: 'bg-black' },
  { value: 'email', label: 'Email', color: 'bg-gray-600' }
]

const objectives = [
  'Brand Awareness',
  'Lead Generation',
  'Sales Conversion',
  'Website Traffic',
  'Engagement',
  'App Downloads',
  'Event Promotion',
  'Customer Retention'
]

export default function CampaignBuilder({ campaign, onSave, onCancel }: CampaignBuilderProps) {
  const [formData, setFormData] = useState<CampaignFormData>({
    name: '',
    description: '',
    type: '',
    budget: 1000,
    startDate: new Date().toISOString().split('T')[0],
    endDate: '',
    targetAudience: {
      demographics: {},
      interests: [],
      locations: []
    },
    objectives: [],
    platforms: []
  })

  const [activeTab, setActiveTab] = useState('basic')
  const [isLoading, setIsLoading] = useState(false)
  const [errors, setErrors] = useState<Record<string, string>>({})

  // Initialize form with existing campaign data
  useEffect(() => {
    if (campaign) {
      setFormData({
        name: campaign.name,
        description: campaign.description,
        type: campaign.type,
        budget: campaign.budget,
        startDate: campaign.startDate.split('T')[0],
        endDate: campaign.endDate ? campaign.endDate.split('T')[0] : '',
        targetAudience: campaign.targetAudience as any || { demographics: {}, interests: [], locations: [] },
        objectives: (campaign.objectives as any)?.objectives || [],
        platforms: (campaign.platforms as any)?.platforms || []
      })
    }
  }, [campaign])

  const handleInputChange = (field: keyof CampaignFormData, value: any) => {
    setFormData(prev => ({ ...prev, [field]: value }))
    // Clear error when user starts typing
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: '' }))
    }
  }

  const handleArrayAdd = (field: 'objectives' | 'platforms', value: string) => {
    if (value && !formData[field].includes(value)) {
      setFormData(prev => ({
        ...prev,
        [field]: [...prev[field], value]
      }))
    }
  }

  const handleArrayRemove = (field: 'objectives' | 'platforms', value: string) => {
    setFormData(prev => ({
      ...prev,
      [field]: prev[field].filter(item => item !== value)
    }))
  }

  const handleInterestAdd = (interest: string) => {
    if (interest && !formData.targetAudience.interests.includes(interest)) {
      setFormData(prev => ({
        ...prev,
        targetAudience: {
          ...prev.targetAudience,
          interests: [...prev.targetAudience.interests, interest]
        }
      }))
    }
  }

  const handleInterestRemove = (interest: string) => {
    setFormData(prev => ({
      ...prev,
      targetAudience: {
        ...prev.targetAudience,
        interests: prev.targetAudience.interests.filter(i => i !== interest)
      }
    }))
  }

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {}

    if (!formData.name.trim()) {
      newErrors.name = 'Campaign name is required'
    }

    if (!formData.type) {
      newErrors.type = 'Campaign type is required'
    }

    if (formData.budget <= 0) {
      newErrors.budget = 'Budget must be greater than 0'
    }

    if (!formData.startDate) {
      newErrors.startDate = 'Start date is required'
    }

    if (formData.platforms.length === 0) {
      newErrors.platforms = 'At least one platform is required'
    }

    if (formData.objectives.length === 0) {
      newErrors.objectives = 'At least one objective is required'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSave = async () => {
    if (!validateForm()) {
      return
    }

    setIsLoading(true)

    try {
      // Convert form data to campaign format
      const campaignData: Partial<Campaign> = {
        ...campaign,
        name: formData.name,
        description: formData.description,
        type: formData.type as CampaignType,
        budget: formData.budget,
        startDate: formData.startDate,
        endDate: formData.endDate || undefined,
        targetAudience: formData.targetAudience,
        objectives: { objectives: formData.objectives },
        platforms: { platforms: formData.platforms }
      }

      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000))

      onSave?.(campaignData as Campaign)
    } catch (error) {
      console.error('Failed to save campaign:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const getStatusColor = (status?: CampaignStatus) => {
    switch (status) {
      case 'active': return 'text-green-600 bg-green-50'
      case 'paused': return 'text-yellow-600 bg-yellow-50'
      case 'completed': return 'text-blue-600 bg-blue-50'
      case 'cancelled': return 'text-red-600 bg-red-50'
      default: return 'text-gray-600 bg-gray-50'
    }
  }

  const getStatusIcon = (status?: CampaignStatus) => {
    switch (status) {
      case 'active': return <Play className="h-4 w-4" />
      case 'paused': return <Pause className="h-4 w-4" />
      case 'completed': return <CheckCircle className="h-4 w-4" />
      case 'cancelled': return <AlertCircle className="h-4 w-4" />
      default: return <Clock className="h-4 w-4" />
    }
  }

  return (
    <div className="max-w-4xl mx-auto p-6 space-y-6">
      <FadeIn>
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-3xl font-bold text-gray-900 flex items-center">
              <Target className="h-8 w-8 mr-3 text-blue-600" />
              {campaign ? 'Edit Campaign' : 'Create New Campaign'}
            </h1>
            {campaign && (
              <div className="flex items-center mt-2">
                <Badge className={`${getStatusColor(campaign.status)} mr-2`}>
                  {getStatusIcon(campaign.status)}
                  <span className="ml-1 capitalize">{campaign.status}</span>
                </Badge>
                <span className="text-sm text-gray-600">
                  Created {new Date(campaign.createdAt).toLocaleDateString()}
                </span>
              </div>
            )}
          </div>
          <div className="flex space-x-3">
            <Button variant="outline" onClick={onCancel}>
              Cancel
            </Button>
            <Button 
              onClick={handleSave} 
              disabled={isLoading}
              className="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700"
            >
              {isLoading ? 'Saving...' : campaign ? 'Update Campaign' : 'Create Campaign'}
            </Button>
          </div>
        </div>
      </FadeIn>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="basic" className="flex items-center">
            <Settings className="h-4 w-4 mr-2" />
            Basic Info
          </TabsTrigger>
          <TabsTrigger value="audience" className="flex items-center">
            <Users className="h-4 w-4 mr-2" />
            Audience
          </TabsTrigger>
          <TabsTrigger value="objectives" className="flex items-center">
            <Target className="h-4 w-4 mr-2" />
            Objectives
          </TabsTrigger>
          <TabsTrigger value="budget" className="flex items-center">
            <DollarSign className="h-4 w-4 mr-2" />
            Budget & Schedule
          </TabsTrigger>
        </TabsList>

        <TabsContent value="basic" className="space-y-6">
          <StaggerContainer>
            <StaggerItem>
              <Card>
                <CardHeader>
                  <CardTitle>Campaign Details</CardTitle>
                  <CardDescription>
                    Basic information about your marketing campaign
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div>
                    <Label htmlFor="name">Campaign Name *</Label>
                    <Input
                      id="name"
                      value={formData.name}
                      onChange={(e) => handleInputChange('name', e.target.value)}
                      placeholder="Enter campaign name"
                      className={errors.name ? 'border-red-500' : ''}
                    />
                    {errors.name && <p className="text-sm text-red-600 mt-1">{errors.name}</p>}
                  </div>

                  <div>
                    <Label htmlFor="description">Description</Label>
                    <Textarea
                      id="description"
                      value={formData.description}
                      onChange={(e) => handleInputChange('description', e.target.value)}
                      placeholder="Describe your campaign goals and strategy"
                      rows={3}
                    />
                  </div>

                  <div>
                    <Label htmlFor="type">Campaign Type *</Label>
                    <Select value={formData.type} onValueChange={(value) => handleInputChange('type', value)}>
                      <SelectTrigger className={errors.type ? 'border-red-500' : ''}>
                        <SelectValue placeholder="Select campaign type" />
                      </SelectTrigger>
                      <SelectContent>
                        {campaignTypes.map((type) => (
                          <SelectItem key={type.value} value={type.value}>
                            <div className="flex items-center">
                              <span className="mr-2">{type.icon}</span>
                              <div>
                                <div className="font-medium">{type.label}</div>
                                <div className="text-sm text-gray-600">{type.description}</div>
                              </div>
                            </div>
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    {errors.type && <p className="text-sm text-red-600 mt-1">{errors.type}</p>}
                  </div>

                  <div>
                    <Label>Platforms *</Label>
                    <div className="grid grid-cols-2 md:grid-cols-4 gap-2 mt-2">
                      {platforms.map((platform) => (
                        <Button
                          key={platform.value}
                          variant={formData.platforms.includes(platform.value) ? "default" : "outline"}
                          size="sm"
                          onClick={() => {
                            if (formData.platforms.includes(platform.value)) {
                              handleArrayRemove('platforms', platform.value)
                            } else {
                              handleArrayAdd('platforms', platform.value)
                            }
                          }}
                          className="justify-start"
                        >
                          <div className={`w-3 h-3 rounded-full ${platform.color} mr-2`} />
                          {platform.label}
                        </Button>
                      ))}
                    </div>
                    {errors.platforms && <p className="text-sm text-red-600 mt-1">{errors.platforms}</p>}
                  </div>
                </CardContent>
              </Card>
            </StaggerItem>
          </StaggerContainer>
        </TabsContent>

        <TabsContent value="audience" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Target Audience</CardTitle>
              <CardDescription>
                Define who you want to reach with this campaign
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <Label>Interests</Label>
                <div className="flex flex-wrap gap-2 mt-2">
                  {formData.targetAudience.interests.map((interest) => (
                    <Badge
                      key={interest}
                      variant="secondary"
                      className="cursor-pointer"
                      onClick={() => handleInterestRemove(interest)}
                    >
                      {interest} ×
                    </Badge>
                  ))}
                </div>
                <Input
                  placeholder="Add interest and press Enter"
                  className="mt-2"
                  onKeyPress={(e) => {
                    if (e.key === 'Enter') {
                      handleInterestAdd(e.currentTarget.value)
                      e.currentTarget.value = ''
                    }
                  }}
                />
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="objectives" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Campaign Objectives</CardTitle>
              <CardDescription>
                Select the primary goals for this campaign
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
                {objectives.map((objective) => (
                  <Button
                    key={objective}
                    variant={formData.objectives.includes(objective) ? "default" : "outline"}
                    size="sm"
                    onClick={() => {
                      if (formData.objectives.includes(objective)) {
                        handleArrayRemove('objectives', objective)
                      } else {
                        handleArrayAdd('objectives', objective)
                      }
                    }}
                    className="justify-start"
                  >
                    {objective}
                  </Button>
                ))}
              </div>
              {errors.objectives && <p className="text-sm text-red-600 mt-2">{errors.objectives}</p>}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="budget" className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <Card>
              <CardHeader>
                <CardTitle>Budget</CardTitle>
                <CardDescription>
                  Set your campaign budget
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div>
                  <Label htmlFor="budget">Total Budget ($) *</Label>
                  <Input
                    id="budget"
                    type="number"
                    value={formData.budget}
                    onChange={(e) => handleInputChange('budget', parseFloat(e.target.value) || 0)}
                    placeholder="1000"
                    className={errors.budget ? 'border-red-500' : ''}
                  />
                  {errors.budget && <p className="text-sm text-red-600 mt-1">{errors.budget}</p>}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Schedule</CardTitle>
                <CardDescription>
                  Set campaign start and end dates
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <Label htmlFor="startDate">Start Date *</Label>
                  <Input
                    id="startDate"
                    type="date"
                    value={formData.startDate}
                    onChange={(e) => handleInputChange('startDate', e.target.value)}
                    className={errors.startDate ? 'border-red-500' : ''}
                  />
                  {errors.startDate && <p className="text-sm text-red-600 mt-1">{errors.startDate}</p>}
                </div>

                <div>
                  <Label htmlFor="endDate">End Date (Optional)</Label>
                  <Input
                    id="endDate"
                    type="date"
                    value={formData.endDate}
                    onChange={(e) => handleInputChange('endDate', e.target.value)}
                  />
                </div>
              </CardContent>
            </Card>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}
