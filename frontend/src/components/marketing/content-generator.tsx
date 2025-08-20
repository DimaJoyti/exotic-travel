'use client'

import React, { useState, useEffect } from 'react'
import { 
  PenTool, 
  Brain, 
  Sparkles, 
  Copy, 
  Download, 
  RefreshCw, 
  Settings,
  Target,
  Hash,
  Type,
  Palette,
  Zap,
  CheckCircle,
  AlertCircle
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

interface ContentGeneratorProps {
  campaignId?: number
  onContentGenerated?: (content: GeneratedContent) => void
}

interface GeneratedContent {
  id: string
  title: string
  body: string
  hashtags?: string[]
  metadata: {
    platform: string
    type: string
    tone: string
    wordCount: number
    readabilityScore?: number
  }
  variations?: ContentVariation[]
}

interface ContentVariation {
  id: string
  title: string
  body: string
  testFocus: string
  description: string
}

interface GenerationRequest {
  campaignId: number
  contentType: string
  platform: string
  brief: string
  keywords: string[]
  tone: string
  length: string
  callToAction: string
  generateVariations: boolean
}

const contentTypes = [
  { value: 'social_post', label: 'Social Media Post', icon: '📱' },
  { value: 'ad', label: 'Advertisement', icon: '🎯' },
  { value: 'email', label: 'Email Campaign', icon: '📧' },
  { value: 'blog', label: 'Blog Post', icon: '📝' },
  { value: 'landing', label: 'Landing Page', icon: '🌐' },
  { value: 'video', label: 'Video Script', icon: '🎬' }
]

const platforms = [
  { value: 'facebook', label: 'Facebook', color: 'bg-blue-600' },
  { value: 'instagram', label: 'Instagram', color: 'bg-pink-600' },
  { value: 'twitter', label: 'Twitter', color: 'bg-sky-500' },
  { value: 'linkedin', label: 'LinkedIn', color: 'bg-blue-700' },
  { value: 'tiktok', label: 'TikTok', color: 'bg-black' },
  { value: 'youtube', label: 'YouTube', color: 'bg-red-600' },
  { value: 'google', label: 'Google Ads', color: 'bg-green-600' },
  { value: 'email', label: 'Email', color: 'bg-gray-600' }
]

const tones = [
  'Professional', 'Casual', 'Friendly', 'Authoritative', 
  'Playful', 'Urgent', 'Inspirational', 'Educational'
]

const lengths = [
  { value: 'short', label: 'Short (50-100 words)' },
  { value: 'medium', label: 'Medium (100-300 words)' },
  { value: 'long', label: 'Long (300+ words)' }
]

export default function ContentGenerator({ campaignId = 1, onContentGenerated }: ContentGeneratorProps) {
  const [isGenerating, setIsGenerating] = useState(false)
  const [generatedContent, setGeneratedContent] = useState<GeneratedContent | null>(null)
  const [activeTab, setActiveTab] = useState('form')

  const [formData, setFormData] = useState<GenerationRequest>({
    campaignId,
    contentType: '',
    platform: '',
    brief: '',
    keywords: [],
    tone: '',
    length: 'medium',
    callToAction: '',
    generateVariations: true
  })

  const [keywordInput, setKeywordInput] = useState('')

  const handleInputChange = (field: keyof GenerationRequest, value: any) => {
    setFormData(prev => ({ ...prev, [field]: value }))
  }

  const addKeyword = () => {
    if (keywordInput.trim() && !formData.keywords.includes(keywordInput.trim())) {
      setFormData(prev => ({
        ...prev,
        keywords: [...prev.keywords, keywordInput.trim()]
      }))
      setKeywordInput('')
    }
  }

  const removeKeyword = (keyword: string) => {
    setFormData(prev => ({
      ...prev,
      keywords: prev.keywords.filter(k => k !== keyword)
    }))
  }

  const handleGenerate = async () => {
    if (!formData.contentType || !formData.platform || !formData.brief) {
      alert('Please fill in all required fields')
      return
    }

    setIsGenerating(true)
    setActiveTab('result')

    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 3000))

      const mockContent: GeneratedContent = {
        id: `content_${Date.now()}`,
        title: generateMockTitle(formData.contentType, formData.platform),
        body: generateMockBody(formData.brief, formData.tone),
        hashtags: formData.platform === 'instagram' || formData.platform === 'twitter' 
          ? ['#marketing', '#AI', '#innovation', '#business'] 
          : undefined,
        metadata: {
          platform: formData.platform,
          type: formData.contentType,
          tone: formData.tone,
          wordCount: Math.floor(Math.random() * 200) + 100,
          readabilityScore: Math.floor(Math.random() * 30) + 70
        },
        variations: formData.generateVariations ? [
          {
            id: 'var_1',
            title: 'Variation A: Emotional Appeal',
            body: 'Alternative version focusing on emotional connection...',
            testFocus: 'Emotional Appeal',
            description: 'Tests emotional triggers and storytelling'
          },
          {
            id: 'var_2',
            title: 'Variation B: Data-Driven',
            body: 'Alternative version emphasizing statistics and facts...',
            testFocus: 'Data-Driven',
            description: 'Tests logical appeal with data and statistics'
          }
        ] : undefined
      }

      setGeneratedContent(mockContent)
      onContentGenerated?.(mockContent)
    } catch (error) {
      console.error('Failed to generate content:', error)
    } finally {
      setIsGenerating(false)
    }
  }

  const generateMockTitle = (type: string, platform: string): string => {
    const titles = {
      social_post: `Engaging ${platform} post that drives interaction`,
      ad: `High-converting ${platform} advertisement`,
      email: 'Compelling email subject line that increases opens',
      blog: 'SEO-optimized blog post title',
      landing: 'Conversion-focused landing page headline',
      video: 'Engaging video script outline'
    }
    return titles[type as keyof typeof titles] || 'AI-Generated Content'
  }

  const generateMockBody = (brief: string, tone: string): string => {
    return `This is AI-generated content based on your brief: "${brief}". 

The content is crafted with a ${tone.toLowerCase()} tone to resonate with your target audience. Our advanced AI has analyzed your requirements and created compelling copy that:

✨ Captures attention from the first line
🎯 Speaks directly to your audience's needs
💡 Incorporates your key messaging
🚀 Drives the desired action

This content is optimized for engagement and conversion, following best practices for your selected platform and content type.`
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    // You could add a toast notification here
  }

  return (
    <div className="max-w-6xl mx-auto p-6 space-y-6">
      <FadeIn>
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-gray-900 mb-4 flex items-center justify-center">
            <Brain className="h-10 w-10 mr-3 text-blue-600" />
            AI Content Generator
          </h1>
          <p className="text-lg text-gray-600">
            Create compelling marketing content powered by advanced AI
          </p>
        </div>
      </FadeIn>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="form" className="flex items-center">
            <Settings className="h-4 w-4 mr-2" />
            Configure
          </TabsTrigger>
          <TabsTrigger value="result" className="flex items-center">
            <Sparkles className="h-4 w-4 mr-2" />
            Generated Content
          </TabsTrigger>
        </TabsList>

        <TabsContent value="form" className="space-y-6">
          <StaggerContainer>
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* Content Type & Platform */}
              <StaggerItem>
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center">
                      <Type className="h-5 w-5 mr-2" />
                      Content Type & Platform
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div>
                      <Label htmlFor="contentType">Content Type</Label>
                      <Select value={formData.contentType} onValueChange={(value) => handleInputChange('contentType', value)}>
                        <SelectTrigger>
                          <SelectValue placeholder="Select content type" />
                        </SelectTrigger>
                        <SelectContent>
                          {contentTypes.map((type) => (
                            <SelectItem key={type.value} value={type.value}>
                              {type.icon} {type.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>

                    <div>
                      <Label htmlFor="platform">Platform</Label>
                      <Select value={formData.platform} onValueChange={(value) => handleInputChange('platform', value)}>
                        <SelectTrigger>
                          <SelectValue placeholder="Select platform" />
                        </SelectTrigger>
                        <SelectContent>
                          {platforms.map((platform) => (
                            <SelectItem key={platform.value} value={platform.value}>
                              <div className="flex items-center">
                                <div className={`w-3 h-3 rounded-full ${platform.color} mr-2`} />
                                {platform.label}
                              </div>
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  </CardContent>
                </Card>
              </StaggerItem>

              {/* Content Details */}
              <StaggerItem>
                <Card>
                  <CardHeader>
                    <CardTitle className="flex items-center">
                      <Target className="h-5 w-5 mr-2" />
                      Content Details
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div>
                      <Label htmlFor="tone">Tone</Label>
                      <Select value={formData.tone} onValueChange={(value) => handleInputChange('tone', value)}>
                        <SelectTrigger>
                          <SelectValue placeholder="Select tone" />
                        </SelectTrigger>
                        <SelectContent>
                          {tones.map((tone) => (
                            <SelectItem key={tone} value={tone.toLowerCase()}>
                              {tone}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>

                    <div>
                      <Label htmlFor="length">Length</Label>
                      <Select value={formData.length} onValueChange={(value) => handleInputChange('length', value)}>
                        <SelectTrigger>
                          <SelectValue placeholder="Select length" />
                        </SelectTrigger>
                        <SelectContent>
                          {lengths.map((length) => (
                            <SelectItem key={length.value} value={length.value}>
                              {length.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  </CardContent>
                </Card>
              </StaggerItem>
            </div>

            {/* Brief & Keywords */}
            <StaggerItem>
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center">
                    <PenTool className="h-5 w-5 mr-2" />
                    Content Brief
                  </CardTitle>
                  <CardDescription>
                    Describe what you want the content to achieve
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div>
                    <Label htmlFor="brief">Brief</Label>
                    <Textarea
                      id="brief"
                      placeholder="Describe your content goals, target audience, key messages..."
                      value={formData.brief}
                      onChange={(e) => handleInputChange('brief', e.target.value)}
                      rows={4}
                    />
                  </div>

                  <div>
                    <Label htmlFor="callToAction">Call to Action</Label>
                    <Input
                      id="callToAction"
                      placeholder="e.g., Shop Now, Learn More, Sign Up..."
                      value={formData.callToAction}
                      onChange={(e) => handleInputChange('callToAction', e.target.value)}
                    />
                  </div>

                  <div>
                    <Label htmlFor="keywords">Keywords</Label>
                    <div className="flex space-x-2 mb-2">
                      <Input
                        placeholder="Add keyword..."
                        value={keywordInput}
                        onChange={(e) => setKeywordInput(e.target.value)}
                        onKeyPress={(e) => e.key === 'Enter' && addKeyword()}
                      />
                      <Button onClick={addKeyword} variant="outline">
                        <Hash className="h-4 w-4" />
                      </Button>
                    </div>
                    <div className="flex flex-wrap gap-2">
                      {formData.keywords.map((keyword) => (
                        <Badge
                          key={keyword}
                          variant="secondary"
                          className="cursor-pointer"
                          onClick={() => removeKeyword(keyword)}
                        >
                          {keyword} ×
                        </Badge>
                      ))}
                    </div>
                  </div>
                </CardContent>
              </Card>
            </StaggerItem>

            {/* Generate Button */}
            <StaggerItem>
              <div className="flex justify-center">
                <Button
                  onClick={handleGenerate}
                  disabled={isGenerating || !formData.contentType || !formData.platform || !formData.brief}
                  className="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white px-8 py-3 text-lg"
                >
                  {isGenerating ? (
                    <>
                      <RefreshCw className="h-5 w-5 mr-2 animate-spin" />
                      Generating...
                    </>
                  ) : (
                    <>
                      <Sparkles className="h-5 w-5 mr-2" />
                      Generate Content
                    </>
                  )}
                </Button>
              </div>
            </StaggerItem>
          </StaggerContainer>
        </TabsContent>

        <TabsContent value="result" className="space-y-6">
          <AnimatePresence mode="wait">
            {isGenerating ? (
              <motion.div
                key="loading"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="text-center py-12"
              >
                <div className="inline-flex items-center justify-center w-16 h-16 bg-blue-100 rounded-full mb-4">
                  <Brain className="h-8 w-8 text-blue-600 animate-pulse" />
                </div>
                <h3 className="text-xl font-semibold text-gray-900 mb-2">
                  AI is crafting your content...
                </h3>
                <p className="text-gray-600">
                  This may take a few moments while we analyze your requirements
                </p>
              </motion.div>
            ) : generatedContent ? (
              <motion.div
                key="content"
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -20 }}
                className="space-y-6"
              >
                {/* Generated Content */}
                <Card>
                  <CardHeader>
                    <div className="flex items-center justify-between">
                      <CardTitle className="flex items-center">
                        <CheckCircle className="h-5 w-5 mr-2 text-green-600" />
                        Generated Content
                      </CardTitle>
                      <div className="flex space-x-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => copyToClipboard(generatedContent.body)}
                        >
                          <Copy className="h-4 w-4 mr-1" />
                          Copy
                        </Button>
                        <Button variant="outline" size="sm">
                          <Download className="h-4 w-4 mr-1" />
                          Export
                        </Button>
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div>
                      <Label className="text-sm font-medium text-gray-700">Title</Label>
                      <div className="mt-1 p-3 bg-gray-50 rounded-lg">
                        <p className="font-semibold">{generatedContent.title}</p>
                      </div>
                    </div>

                    <div>
                      <Label className="text-sm font-medium text-gray-700">Content</Label>
                      <div className="mt-1 p-4 bg-gray-50 rounded-lg">
                        <p className="whitespace-pre-wrap">{generatedContent.body}</p>
                      </div>
                    </div>

                    {generatedContent.hashtags && (
                      <div>
                        <Label className="text-sm font-medium text-gray-700">Hashtags</Label>
                        <div className="mt-1 flex flex-wrap gap-2">
                          {generatedContent.hashtags.map((hashtag) => (
                            <Badge key={hashtag} variant="outline">
                              {hashtag}
                            </Badge>
                          ))}
                        </div>
                      </div>
                    )}

                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4 pt-4 border-t">
                      <div className="text-center">
                        <p className="text-sm text-gray-600">Platform</p>
                        <p className="font-semibold capitalize">{generatedContent.metadata.platform}</p>
                      </div>
                      <div className="text-center">
                        <p className="text-sm text-gray-600">Word Count</p>
                        <p className="font-semibold">{generatedContent.metadata.wordCount}</p>
                      </div>
                      <div className="text-center">
                        <p className="text-sm text-gray-600">Tone</p>
                        <p className="font-semibold capitalize">{generatedContent.metadata.tone}</p>
                      </div>
                      <div className="text-center">
                        <p className="text-sm text-gray-600">Readability</p>
                        <p className="font-semibold">{generatedContent.metadata.readabilityScore}/100</p>
                      </div>
                    </div>
                  </CardContent>
                </Card>

                {/* Variations */}
                {generatedContent.variations && generatedContent.variations.length > 0 && (
                  <Card>
                    <CardHeader>
                      <CardTitle className="flex items-center">
                        <Zap className="h-5 w-5 mr-2 text-yellow-600" />
                        A/B Test Variations
                      </CardTitle>
                      <CardDescription>
                        Alternative versions for testing different approaches
                      </CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        {generatedContent.variations.map((variation) => (
                          <div key={variation.id} className="p-4 border rounded-lg">
                            <div className="flex items-center justify-between mb-2">
                              <h4 className="font-semibold">{variation.title}</h4>
                              <Badge variant="outline">{variation.testFocus}</Badge>
                            </div>
                            <p className="text-sm text-gray-600 mb-3">{variation.description}</p>
                            <div className="p-3 bg-gray-50 rounded text-sm">
                              {variation.body}
                            </div>
                          </div>
                        ))}
                      </div>
                    </CardContent>
                  </Card>
                )}
              </motion.div>
            ) : (
              <motion.div
                key="empty"
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="text-center py-12"
              >
                <AlertCircle className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                <h3 className="text-xl font-semibold text-gray-900 mb-2">
                  No content generated yet
                </h3>
                <p className="text-gray-600 mb-4">
                  Configure your content requirements and click generate
                </p>
                <Button onClick={() => setActiveTab('form')} variant="outline">
                  Go to Configuration
                </Button>
              </motion.div>
            )}
          </AnimatePresence>
        </TabsContent>
      </Tabs>
    </div>
  )
}
