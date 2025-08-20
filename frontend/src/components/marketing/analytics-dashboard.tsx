'use client'

import React, { useState, useEffect } from 'react'
import { 
  BarChart3, 
  TrendingUp, 
  TrendingDown, 
  DollarSign, 
  Users, 
  Eye, 
  MousePointer, 
  Target,
  Calendar,
  Filter,
  Download,
  RefreshCw,
  ArrowUpRight,
  ArrowDownRight
} from 'lucide-react'
import { motion } from 'framer-motion'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { FadeIn, StaggerContainer, StaggerItem } from '@/components/ui/animated'

interface AnalyticsDashboardProps {
  campaignId?: number
  timeRange?: string
}

interface MetricCard {
  title: string
  value: string | number
  change: number
  changeType: 'increase' | 'decrease' | 'neutral'
  icon: React.ReactNode
  color: string
  description?: string
}

interface ChartData {
  date: string
  impressions: number
  clicks: number
  conversions: number
  spend: number
  revenue: number
}

interface PlatformPerformance {
  platform: string
  impressions: number
  clicks: number
  ctr: number
  spend: number
  roas: number
  color: string
}

export default function AnalyticsDashboard({ campaignId, timeRange = '7d' }: AnalyticsDashboardProps) {
  const [selectedTimeRange, setSelectedTimeRange] = useState(timeRange)
  const [selectedMetric, setSelectedMetric] = useState('impressions')
  const [isLoading, setIsLoading] = useState(false)
  const [activeTab, setActiveTab] = useState('overview')

  // Mock data - in real implementation, this would come from API
  const [metrics, setMetrics] = useState<MetricCard[]>([
    {
      title: 'Total Impressions',
      value: '2.4M',
      change: 12.5,
      changeType: 'increase',
      icon: <Eye className="h-5 w-5" />,
      color: 'bg-blue-500',
      description: 'Total ad views across all platforms'
    },
    {
      title: 'Click-Through Rate',
      value: '3.2%',
      change: 8.3,
      changeType: 'increase',
      icon: <MousePointer className="h-5 w-5" />,
      color: 'bg-green-500',
      description: 'Percentage of impressions that resulted in clicks'
    },
    {
      title: 'Conversions',
      value: '1,247',
      change: -2.1,
      changeType: 'decrease',
      icon: <Target className="h-5 w-5" />,
      color: 'bg-purple-500',
      description: 'Total conversions from all campaigns'
    },
    {
      title: 'ROAS',
      value: '4.2x',
      change: 15.7,
      changeType: 'increase',
      icon: <DollarSign className="h-5 w-5" />,
      color: 'bg-orange-500',
      description: 'Return on advertising spend'
    },
    {
      title: 'Cost Per Click',
      value: '$1.23',
      change: -5.4,
      changeType: 'increase',
      icon: <TrendingDown className="h-5 w-5" />,
      color: 'bg-red-500',
      description: 'Average cost per click across campaigns'
    },
    {
      title: 'Reach',
      value: '847K',
      change: 22.1,
      changeType: 'increase',
      icon: <Users className="h-5 w-5" />,
      color: 'bg-indigo-500',
      description: 'Unique users reached by campaigns'
    }
  ])

  const [chartData, setChartData] = useState<ChartData[]>([
    { date: '2024-01-01', impressions: 45000, clicks: 1440, conversions: 87, spend: 1200, revenue: 5040 },
    { date: '2024-01-02', impressions: 52000, clicks: 1664, conversions: 95, spend: 1350, revenue: 5700 },
    { date: '2024-01-03', impressions: 48000, clicks: 1536, conversions: 92, spend: 1280, revenue: 5520 },
    { date: '2024-01-04', impressions: 61000, clicks: 1952, conversions: 118, spend: 1580, revenue: 7080 },
    { date: '2024-01-05', impressions: 58000, clicks: 1856, conversions: 112, spend: 1520, revenue: 6720 },
    { date: '2024-01-06', impressions: 55000, clicks: 1760, conversions: 106, spend: 1450, revenue: 6360 },
    { date: '2024-01-07', impressions: 63000, clicks: 2016, conversions: 125, spend: 1680, revenue: 7500 }
  ])

  const [platformData, setPlatformData] = useState<PlatformPerformance[]>([
    { platform: 'Facebook', impressions: 850000, clicks: 27200, ctr: 3.2, spend: 4200, roas: 4.8, color: 'bg-blue-600' },
    { platform: 'Instagram', impressions: 720000, clicks: 25200, ctr: 3.5, spend: 3800, roas: 5.2, color: 'bg-pink-600' },
    { platform: 'Google Ads', impressions: 620000, clicks: 18600, ctr: 3.0, spend: 5200, roas: 3.9, color: 'bg-green-600' },
    { platform: 'LinkedIn', impressions: 180000, clicks: 5400, ctr: 3.0, spend: 2100, roas: 3.2, color: 'bg-blue-700' },
    { platform: 'Twitter', impressions: 95000, clicks: 2850, ctr: 3.0, spend: 1200, roas: 2.8, color: 'bg-sky-500' }
  ])

  const timeRanges = [
    { value: '1d', label: 'Last 24 hours' },
    { value: '7d', label: 'Last 7 days' },
    { value: '30d', label: 'Last 30 days' },
    { value: '90d', label: 'Last 90 days' },
    { value: 'custom', label: 'Custom range' }
  ]

  const handleTimeRangeChange = (value: string) => {
    setSelectedTimeRange(value)
    // In real implementation, fetch new data based on time range
  }

  const handleRefresh = async () => {
    setIsLoading(true)
    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 1000))
    setIsLoading(false)
  }

  const formatNumber = (num: number): string => {
    if (num >= 1000000) {
      return `${(num / 1000000).toFixed(1)}M`
    }
    if (num >= 1000) {
      return `${(num / 1000).toFixed(1)}K`
    }
    return num.toString()
  }

  const formatCurrency = (amount: number): string => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(amount)
  }

  const getChangeIcon = (changeType: 'increase' | 'decrease' | 'neutral') => {
    switch (changeType) {
      case 'increase':
        return <ArrowUpRight className="h-4 w-4 text-green-600" />
      case 'decrease':
        return <ArrowDownRight className="h-4 w-4 text-red-600" />
      default:
        return null
    }
  }

  const getChangeColor = (changeType: 'increase' | 'decrease' | 'neutral') => {
    switch (changeType) {
      case 'increase':
        return 'text-green-600'
      case 'decrease':
        return 'text-red-600'
      default:
        return 'text-gray-600'
    }
  }

  return (
    <div className="max-w-7xl mx-auto p-6 space-y-6">
      {/* Header */}
      <FadeIn>
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900 flex items-center">
              <BarChart3 className="h-8 w-8 mr-3 text-blue-600" />
              Marketing Analytics
            </h1>
            <p className="text-lg text-gray-600 mt-1">
              Real-time performance insights and campaign analytics
            </p>
          </div>
          <div className="flex items-center space-x-3">
            <Select value={selectedTimeRange} onValueChange={handleTimeRangeChange}>
              <SelectTrigger className="w-48">
                <Calendar className="h-4 w-4 mr-2" />
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {timeRanges.map((range) => (
                  <SelectItem key={range.value} value={range.value}>
                    {range.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button variant="outline" onClick={handleRefresh} disabled={isLoading}>
              <RefreshCw className={`h-4 w-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
            <Button variant="outline">
              <Download className="h-4 w-4 mr-2" />
              Export
            </Button>
          </div>
        </div>
      </FadeIn>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="platforms">Platforms</TabsTrigger>
          <TabsTrigger value="campaigns">Campaigns</TabsTrigger>
          <TabsTrigger value="audience">Audience</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-6">
          {/* Key Metrics Grid */}
          <StaggerContainer>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {metrics.map((metric, index) => (
                <StaggerItem key={metric.title}>
                  <Card className="bg-white/80 backdrop-blur-sm border-0 shadow-lg hover:shadow-xl transition-all duration-300">
                    <CardContent className="p-6">
                      <div className="flex items-center justify-between">
                        <div className="flex-1">
                          <div className="flex items-center justify-between mb-2">
                            <p className="text-sm font-medium text-gray-600">{metric.title}</p>
                            <div className={`p-2 rounded-lg ${metric.color} text-white`}>
                              {metric.icon}
                            </div>
                          </div>
                          <p className="text-3xl font-bold text-gray-900 mb-1">{metric.value}</p>
                          <div className="flex items-center">
                            {getChangeIcon(metric.changeType)}
                            <span className={`text-sm font-medium ml-1 ${getChangeColor(metric.changeType)}`}>
                              {Math.abs(metric.change)}%
                            </span>
                            <span className="text-sm text-gray-500 ml-1">vs last period</span>
                          </div>
                          {metric.description && (
                            <p className="text-xs text-gray-500 mt-2">{metric.description}</p>
                          )}
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </StaggerItem>
              ))}
            </div>
          </StaggerContainer>

          {/* Performance Chart */}
          <FadeIn delay={0.2}>
            <Card className="bg-white/80 backdrop-blur-sm border-0 shadow-lg">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle>Performance Trends</CardTitle>
                    <CardDescription>
                      Campaign performance over the selected time period
                    </CardDescription>
                  </div>
                  <Select value={selectedMetric} onValueChange={setSelectedMetric}>
                    <SelectTrigger className="w-48">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="impressions">Impressions</SelectItem>
                      <SelectItem value="clicks">Clicks</SelectItem>
                      <SelectItem value="conversions">Conversions</SelectItem>
                      <SelectItem value="spend">Spend</SelectItem>
                      <SelectItem value="revenue">Revenue</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </CardHeader>
              <CardContent>
                <div className="h-80 flex items-center justify-center bg-gray-50 rounded-lg">
                  <div className="text-center">
                    <BarChart3 className="h-16 w-16 text-gray-400 mx-auto mb-4" />
                    <p className="text-gray-600 font-medium">Interactive Chart Component</p>
                    <p className="text-sm text-gray-500">
                      Chart showing {selectedMetric} trends over {selectedTimeRange}
                    </p>
                    <div className="mt-4 grid grid-cols-7 gap-2">
                      {chartData.map((data, index) => (
                        <div key={index} className="text-center">
                          <div 
                            className="bg-blue-500 rounded-t"
                            style={{ 
                              height: `${(data[selectedMetric as keyof ChartData] as number / Math.max(...chartData.map(d => d[selectedMetric as keyof ChartData] as number))) * 60}px`,
                              minHeight: '4px'
                            }}
                          />
                          <p className="text-xs text-gray-500 mt-1">
                            {new Date(data.date).toLocaleDateString('en-US', { weekday: 'short' })}
                          </p>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </FadeIn>
        </TabsContent>

        <TabsContent value="platforms" className="space-y-6">
          <FadeIn>
            <Card className="bg-white/80 backdrop-blur-sm border-0 shadow-lg">
              <CardHeader>
                <CardTitle>Platform Performance</CardTitle>
                <CardDescription>
                  Compare performance across different marketing platforms
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {platformData.map((platform) => (
                    <div key={platform.platform} className="p-4 border rounded-lg hover:bg-gray-50 transition-colors">
                      <div className="flex items-center justify-between mb-3">
                        <div className="flex items-center">
                          <div className={`w-4 h-4 rounded-full ${platform.color} mr-3`} />
                          <h3 className="font-semibold text-gray-900">{platform.platform}</h3>
                        </div>
                        <Badge variant="outline" className="text-green-600 border-green-200">
                          {platform.roas}x ROAS
                        </Badge>
                      </div>
                      <div className="grid grid-cols-2 md:grid-cols-5 gap-4 text-sm">
                        <div>
                          <p className="text-gray-600">Impressions</p>
                          <p className="font-semibold">{formatNumber(platform.impressions)}</p>
                        </div>
                        <div>
                          <p className="text-gray-600">Clicks</p>
                          <p className="font-semibold">{formatNumber(platform.clicks)}</p>
                        </div>
                        <div>
                          <p className="text-gray-600">CTR</p>
                          <p className="font-semibold">{platform.ctr}%</p>
                        </div>
                        <div>
                          <p className="text-gray-600">Spend</p>
                          <p className="font-semibold">{formatCurrency(platform.spend)}</p>
                        </div>
                        <div>
                          <p className="text-gray-600">ROAS</p>
                          <p className="font-semibold text-green-600">{platform.roas}x</p>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </FadeIn>
        </TabsContent>

        <TabsContent value="campaigns" className="space-y-6">
          <FadeIn>
            <Card className="bg-white/80 backdrop-blur-sm border-0 shadow-lg">
              <CardHeader>
                <CardTitle>Campaign Performance</CardTitle>
                <CardDescription>
                  Individual campaign metrics and performance comparison
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="text-center py-12">
                  <Target className="h-16 w-16 text-gray-400 mx-auto mb-4" />
                  <p className="text-gray-600 font-medium">Campaign Analytics</p>
                  <p className="text-sm text-gray-500">
                    Detailed campaign performance metrics and comparisons
                  </p>
                </div>
              </CardContent>
            </Card>
          </FadeIn>
        </TabsContent>

        <TabsContent value="audience" className="space-y-6">
          <FadeIn>
            <Card className="bg-white/80 backdrop-blur-sm border-0 shadow-lg">
              <CardHeader>
                <CardTitle>Audience Insights</CardTitle>
                <CardDescription>
                  Demographics, interests, and behavior analysis
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="text-center py-12">
                  <Users className="h-16 w-16 text-gray-400 mx-auto mb-4" />
                  <p className="text-gray-600 font-medium">Audience Analytics</p>
                  <p className="text-sm text-gray-500">
                    Demographic breakdowns and audience behavior insights
                  </p>
                </div>
              </CardContent>
            </Card>
          </FadeIn>
        </TabsContent>
      </Tabs>
    </div>
  )
}
