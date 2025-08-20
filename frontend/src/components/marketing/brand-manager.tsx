'use client'

import React, { useState, useEffect } from 'react'
import { 
  Palette, 
  Type, 
  Image, 
  Upload, 
  Save, 
  Plus, 
  Trash2, 
  Edit, 
  Eye,
  Download,
  Copy,
  CheckCircle,
  AlertCircle
} from 'lucide-react'
import { motion } from 'framer-motion'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { FadeIn, StaggerContainer, StaggerItem } from '@/components/ui/animated'
import { Brand } from '@/lib/marketing-api'

interface BrandManagerProps {
  brandId?: number
  onSave?: (brand: Brand) => void
  onCancel?: () => void
}

interface ColorPalette {
  primary: string
  secondary: string
  accent: string
  neutral: string
  success: string
  warning: string
  error: string
}

interface Typography {
  primary: string
  secondary: string
  headings: string
  body: string
}

interface VoiceGuidelines {
  personality: string[]
  values: string[]
  tone: string
  doList: string[]
  dontList: string[]
  exampleContent: string[]
}

export default function BrandManager({ brandId, onSave, onCancel }: BrandManagerProps) {
  const [brand, setBrand] = useState<Brand | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [isSaving, setIsSaving] = useState(false)
  const [activeTab, setActiveTab] = useState('identity')
  const [errors, setErrors] = useState<Record<string, string>>({})

  const [formData, setFormData] = useState({
    name: '',
    description: '',
    logoUrl: '',
    colorPalette: {
      primary: '#2563eb',
      secondary: '#64748b',
      accent: '#f59e0b',
      neutral: '#6b7280',
      success: '#10b981',
      warning: '#f59e0b',
      error: '#ef4444'
    } as ColorPalette,
    typography: {
      primary: 'Inter',
      secondary: 'Inter',
      headings: 'Inter',
      body: 'Inter'
    } as Typography,
    voiceGuidelines: {
      personality: ['Professional', 'Innovative', 'Approachable'],
      values: ['Quality', 'Innovation', 'Customer Focus'],
      tone: 'Professional yet friendly',
      doList: ['Use clear, concise language', 'Focus on benefits', 'Be authentic'],
      dontList: ['Use jargon', 'Make unrealistic claims', 'Be overly promotional'],
      exampleContent: []
    } as VoiceGuidelines
  })

  // Load brand data if editing
  useEffect(() => {
    if (brandId) {
      loadBrand(brandId)
    }
  }, [brandId])

  const loadBrand = async (id: number) => {
    setIsLoading(true)
    try {
      // Mock API call - replace with actual API
      const mockBrand: Brand = {
        id,
        name: 'TechCorp',
        description: 'Innovative technology solutions for modern businesses',
        logoUrl: '/api/placeholder/logo.png',
        colorPalette: {
          primary: '#2563eb',
          secondary: '#64748b',
          accent: '#f59e0b'
        },
        typography: {
          primary: 'Inter',
          headings: 'Inter'
        },
        visualIdentity: {
          style: 'modern',
          elements: ['clean lines', 'minimalist']
        },
        voiceGuidelines: {
          personality: ['Professional', 'Innovative'],
          values: ['Quality', 'Innovation'],
          tone: 'Professional yet friendly'
        },
        brandAssets: {},
        companyId: 1,
        createdAt: '2024-01-01T00:00:00Z',
        updatedAt: '2024-01-01T00:00:00Z'
      }

      setBrand(mockBrand)
      setFormData({
        name: mockBrand.name,
        description: mockBrand.description,
        logoUrl: mockBrand.logoUrl,
        colorPalette: mockBrand.colorPalette as ColorPalette,
        typography: mockBrand.typography as Typography,
        voiceGuidelines: mockBrand.voiceGuidelines as VoiceGuidelines
      })
    } catch (error) {
      console.error('Failed to load brand:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const handleInputChange = (field: string, value: any) => {
    setFormData(prev => ({ ...prev, [field]: value }))
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: '' }))
    }
  }

  const handleColorChange = (colorKey: keyof ColorPalette, value: string) => {
    setFormData(prev => ({
      ...prev,
      colorPalette: { ...prev.colorPalette, [colorKey]: value }
    }))
  }

  const handleArrayAdd = (field: keyof VoiceGuidelines, value: string) => {
    if (value.trim()) {
      setFormData(prev => ({
        ...prev,
        voiceGuidelines: {
          ...prev.voiceGuidelines,
          [field]: [...(prev.voiceGuidelines[field] as string[]), value.trim()]
        }
      }))
    }
  }

  const handleArrayRemove = (field: keyof VoiceGuidelines, index: number) => {
    setFormData(prev => ({
      ...prev,
      voiceGuidelines: {
        ...prev.voiceGuidelines,
        [field]: (prev.voiceGuidelines[field] as string[]).filter((_, i) => i !== index)
      }
    }))
  }

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {}

    if (!formData.name.trim()) {
      newErrors.name = 'Brand name is required'
    }

    if (!formData.description.trim()) {
      newErrors.description = 'Brand description is required'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSave = async () => {
    if (!validateForm()) return

    setIsSaving(true)
    try {
      const brandData: Partial<Brand> = {
        ...brand,
        name: formData.name,
        description: formData.description,
        logoUrl: formData.logoUrl,
        colorPalette: formData.colorPalette,
        typography: formData.typography,
        voiceGuidelines: formData.voiceGuidelines
      }

      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 1000))

      onSave?.(brandData as Brand)
    } catch (error) {
      console.error('Failed to save brand:', error)
    } finally {
      setIsSaving(false)
    }
  }

  const copyColorToClipboard = (color: string) => {
    navigator.clipboard.writeText(color)
    // You could add a toast notification here
  }

  if (isLoading) {
    return (
      <div className="max-w-4xl mx-auto p-6">
        <div className="text-center py-12">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading brand information...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="max-w-4xl mx-auto p-6 space-y-6">
      {/* Header */}
      <FadeIn>
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900 flex items-center">
              <Palette className="h-8 w-8 mr-3 text-blue-600" />
              {brand ? 'Edit Brand' : 'Create Brand'}
            </h1>
            <p className="text-lg text-gray-600 mt-1">
              Define your brand identity, voice, and visual guidelines
            </p>
          </div>
          <div className="flex space-x-3">
            <Button variant="outline" onClick={onCancel}>
              Cancel
            </Button>
            <Button 
              onClick={handleSave} 
              disabled={isSaving}
              className="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700"
            >
              <Save className="h-4 w-4 mr-2" />
              {isSaving ? 'Saving...' : 'Save Brand'}
            </Button>
          </div>
        </div>
      </FadeIn>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="identity" className="flex items-center">
            <Palette className="h-4 w-4 mr-2" />
            Identity
          </TabsTrigger>
          <TabsTrigger value="colors" className="flex items-center">
            <Palette className="h-4 w-4 mr-2" />
            Colors
          </TabsTrigger>
          <TabsTrigger value="typography" className="flex items-center">
            <Type className="h-4 w-4 mr-2" />
            Typography
          </TabsTrigger>
          <TabsTrigger value="voice" className="flex items-center">
            <Edit className="h-4 w-4 mr-2" />
            Voice & Tone
          </TabsTrigger>
        </TabsList>

        <TabsContent value="identity" className="space-y-6">
          <StaggerContainer>
            <StaggerItem>
              <Card>
                <CardHeader>
                  <CardTitle>Brand Identity</CardTitle>
                  <CardDescription>
                    Basic information about your brand
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div>
                    <Label htmlFor="name">Brand Name *</Label>
                    <Input
                      id="name"
                      value={formData.name}
                      onChange={(e) => handleInputChange('name', e.target.value)}
                      placeholder="Enter brand name"
                      className={errors.name ? 'border-red-500' : ''}
                    />
                    {errors.name && <p className="text-sm text-red-600 mt-1">{errors.name}</p>}
                  </div>

                  <div>
                    <Label htmlFor="description">Description *</Label>
                    <Textarea
                      id="description"
                      value={formData.description}
                      onChange={(e) => handleInputChange('description', e.target.value)}
                      placeholder="Describe your brand, mission, and values"
                      rows={3}
                      className={errors.description ? 'border-red-500' : ''}
                    />
                    {errors.description && <p className="text-sm text-red-600 mt-1">{errors.description}</p>}
                  </div>

                  <div>
                    <Label htmlFor="logo">Brand Logo</Label>
                    <div className="mt-2 flex items-center space-x-4">
                      {formData.logoUrl && (
                        <div className="w-16 h-16 bg-gray-100 rounded-lg flex items-center justify-center">
                          <img 
                            src={formData.logoUrl} 
                            alt="Brand logo" 
                            className="max-w-full max-h-full object-contain"
                            onError={(e) => {
                              e.currentTarget.style.display = 'none'
                              e.currentTarget.nextElementSibling!.style.display = 'flex'
                            }}
                          />
                          <div className="hidden items-center justify-center w-full h-full">
                            <Image className="h-6 w-6 text-gray-400" />
                          </div>
                        </div>
                      )}
                      <div className="flex-1">
                        <Input
                          placeholder="Logo URL or upload new logo"
                          value={formData.logoUrl}
                          onChange={(e) => handleInputChange('logoUrl', e.target.value)}
                        />
                      </div>
                      <Button variant="outline">
                        <Upload className="h-4 w-4 mr-2" />
                        Upload
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </StaggerItem>
          </StaggerContainer>
        </TabsContent>

        <TabsContent value="colors" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Color Palette</CardTitle>
              <CardDescription>
                Define your brand's color scheme for consistent visual identity
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {Object.entries(formData.colorPalette).map(([key, color]) => (
                  <div key={key} className="space-y-2">
                    <Label className="capitalize">{key}</Label>
                    <div className="flex items-center space-x-3">
                      <div 
                        className="w-12 h-12 rounded-lg border-2 border-gray-200 cursor-pointer"
                        style={{ backgroundColor: color }}
                        onClick={() => copyColorToClipboard(color)}
                      />
                      <div className="flex-1">
                        <Input
                          type="color"
                          value={color}
                          onChange={(e) => handleColorChange(key as keyof ColorPalette, e.target.value)}
                          className="h-12"
                        />
                      </div>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => copyColorToClipboard(color)}
                      >
                        <Copy className="h-4 w-4" />
                      </Button>
                    </div>
                    <p className="text-sm text-gray-600 font-mono">{color}</p>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="typography" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Typography</CardTitle>
              <CardDescription>
                Choose fonts that represent your brand personality
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <Label htmlFor="primaryFont">Primary Font</Label>
                  <Input
                    id="primaryFont"
                    value={formData.typography.primary}
                    onChange={(e) => setFormData(prev => ({
                      ...prev,
                      typography: { ...prev.typography, primary: e.target.value }
                    }))}
                    placeholder="e.g., Inter, Roboto, Arial"
                  />
                </div>
                <div>
                  <Label htmlFor="secondaryFont">Secondary Font</Label>
                  <Input
                    id="secondaryFont"
                    value={formData.typography.secondary}
                    onChange={(e) => setFormData(prev => ({
                      ...prev,
                      typography: { ...prev.typography, secondary: e.target.value }
                    }))}
                    placeholder="e.g., Inter, Roboto, Arial"
                  />
                </div>
              </div>
              
              <div className="mt-6 p-6 bg-gray-50 rounded-lg">
                <h3 className="text-lg font-semibold mb-4">Typography Preview</h3>
                <div className="space-y-4">
                  <div style={{ fontFamily: formData.typography.primary }}>
                    <h1 className="text-3xl font-bold">Heading 1 - Primary Font</h1>
                    <h2 className="text-2xl font-semibold">Heading 2 - Primary Font</h2>
                    <p className="text-base">Body text using primary font. This is how your content will look.</p>
                  </div>
                  <div style={{ fontFamily: formData.typography.secondary }}>
                    <p className="text-base">Secondary font example for supporting text and captions.</p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="voice" className="space-y-6">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <Card>
              <CardHeader>
                <CardTitle>Brand Personality</CardTitle>
                <CardDescription>
                  Define your brand's personality traits
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <Label>Personality Traits</Label>
                  <div className="flex flex-wrap gap-2 mt-2">
                    {formData.voiceGuidelines.personality.map((trait, index) => (
                      <Badge
                        key={index}
                        variant="secondary"
                        className="cursor-pointer"
                        onClick={() => handleArrayRemove('personality', index)}
                      >
                        {trait} ×
                      </Badge>
                    ))}
                  </div>
                  <Input
                    placeholder="Add personality trait and press Enter"
                    className="mt-2"
                    onKeyPress={(e) => {
                      if (e.key === 'Enter') {
                        handleArrayAdd('personality', e.currentTarget.value)
                        e.currentTarget.value = ''
                      }
                    }}
                  />
                </div>

                <div>
                  <Label>Core Values</Label>
                  <div className="flex flex-wrap gap-2 mt-2">
                    {formData.voiceGuidelines.values.map((value, index) => (
                      <Badge
                        key={index}
                        variant="secondary"
                        className="cursor-pointer"
                        onClick={() => handleArrayRemove('values', index)}
                      >
                        {value} ×
                      </Badge>
                    ))}
                  </div>
                  <Input
                    placeholder="Add core value and press Enter"
                    className="mt-2"
                    onKeyPress={(e) => {
                      if (e.key === 'Enter') {
                        handleArrayAdd('values', e.currentTarget.value)
                        e.currentTarget.value = ''
                      }
                    }}
                  />
                </div>

                <div>
                  <Label htmlFor="tone">Brand Tone</Label>
                  <Textarea
                    id="tone"
                    value={formData.voiceGuidelines.tone}
                    onChange={(e) => setFormData(prev => ({
                      ...prev,
                      voiceGuidelines: { ...prev.voiceGuidelines, tone: e.target.value }
                    }))}
                    placeholder="Describe your brand's tone of voice"
                    rows={3}
                  />
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Voice Guidelines</CardTitle>
                <CardDescription>
                  Do's and don'ts for brand communication
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <Label className="flex items-center text-green-600">
                    <CheckCircle className="h-4 w-4 mr-2" />
                    Do's
                  </Label>
                  <div className="flex flex-wrap gap-2 mt-2">
                    {formData.voiceGuidelines.doList.map((item, index) => (
                      <Badge
                        key={index}
                        variant="outline"
                        className="cursor-pointer border-green-200 text-green-700"
                        onClick={() => handleArrayRemove('doList', index)}
                      >
                        {item} ×
                      </Badge>
                    ))}
                  </div>
                  <Input
                    placeholder="Add guideline and press Enter"
                    className="mt-2"
                    onKeyPress={(e) => {
                      if (e.key === 'Enter') {
                        handleArrayAdd('doList', e.currentTarget.value)
                        e.currentTarget.value = ''
                      }
                    }}
                  />
                </div>

                <div>
                  <Label className="flex items-center text-red-600">
                    <AlertCircle className="h-4 w-4 mr-2" />
                    Don'ts
                  </Label>
                  <div className="flex flex-wrap gap-2 mt-2">
                    {formData.voiceGuidelines.dontList.map((item, index) => (
                      <Badge
                        key={index}
                        variant="outline"
                        className="cursor-pointer border-red-200 text-red-700"
                        onClick={() => handleArrayRemove('dontList', index)}
                      >
                        {item} ×
                      </Badge>
                    ))}
                  </div>
                  <Input
                    placeholder="Add guideline and press Enter"
                    className="mt-2"
                    onKeyPress={(e) => {
                      if (e.key === 'Enter') {
                        handleArrayAdd('dontList', e.currentTarget.value)
                        e.currentTarget.value = ''
                      }
                    }}
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
