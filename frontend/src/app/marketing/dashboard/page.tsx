'use client'

import React, { useState, useEffect } from 'react'
import { 
  BarChart3, 
  PenTool, 
  Target, 
  TrendingUp, 
  Users, 
  DollarSign,
  Calendar,
  Zap,
  Brain,
  Image,
  Mail,
  Share2,
  Plus,
  ArrowRight
} from 'lucide-react'
import { motion } from 'framer-motion'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { FadeIn, StaggerContainer, StaggerItem } from '@/components/ui/animated'
import PerformanceMonitor from '@/components/marketing/performance-monitor'

interface DashboardStats {
  totalCampaigns: number
  activeCampaigns: number
  contentGenerated: number
  totalReach: number
  engagement: number
  roi: number
  monthlySpend: number
  conversions: number
}

interface RecentActivity {
  id: string
  type: 'content_generated' | 'campaign_created' | 'performance_alert'
  title: string
  description: string
  timestamp: string
  status: 'success' | 'warning' | 'info'
}

interface QuickAction {
  id: string
  title: string
  description: string
  icon: React.ReactNode
  href: string
  color: string
}

export default function MarketingDashboard() {
  const [stats, setStats] = useState<DashboardStats>({
    totalCampaigns: 24,
    activeCampaigns: 8,
    contentGenerated: 156,
    totalReach: 2400000,
    engagement: 4.2,
    roi: 285,
    monthlySpend: 15420,
    conversions: 1240
  })

  const [recentActivity, setRecentActivity] = useState<RecentActivity[]>([
    {
      id: '1',
      type: 'content_generated',
      title: 'AI Content Generated',
      description: 'Created 5 social media posts for Summer Campaign',
      timestamp: '2 minutes ago',
      status: 'success'
    },
    {
      id: '2',
      type: 'campaign_created',
      title: 'New Campaign Launched',
      description: 'Holiday Promotion campaign is now live',
      timestamp: '1 hour ago',
      status: 'info'
    },
    {
      id: '3',
      type: 'performance_alert',
      title: 'Performance Alert',
      description: 'CTR increased by 15% in the last 24 hours',
      timestamp: '3 hours ago',
      status: 'success'
    }
  ])

  const quickActions: QuickAction[] = [
    {
      id: 'generate-content',
      title: 'Generate Content',
      description: 'Create AI-powered marketing content',
      icon: <PenTool className="h-6 w-6" />,
      href: '/marketing/content/generate',
      color: 'bg-blue-500'
    },
    {
      id: 'create-campaign',
      title: 'Create Campaign',
      description: 'Launch a new marketing campaign',
      icon: <Target className="h-6 w-6" />,
      href: '/marketing/campaigns/create',
      color: 'bg-green-500'
    },
    {
      id: 'analyze-performance',
      title: 'View Analytics',
      description: 'Analyze campaign performance',
      icon: <BarChart3 className="h-6 w-6" />,
      href: '/marketing/analytics',
      color: 'bg-purple-500'
    },
    {
      id: 'manage-audience',
      title: 'Audience Insights',
      description: 'Explore audience segments',
      icon: <Users className="h-6 w-6" />,
      href: '/marketing/audience',
      color: 'bg-orange-500'
    }
  ]

  const formatNumber = (num: number): string => {
    if (num >= 1000000) {
      return `${(num / 1000000).toFixed(1)}M`
    }
    if (num >= 1000) {
      return `${(num / 1000).toFixed(1)}K`
    }
    return num.toString()
  }

  const getActivityIcon = (type: RecentActivity['type']) => {
    switch (type) {
      case 'content_generated':
        return <Brain className="h-4 w-4" />
      case 'campaign_created':
        return <Target className="h-4 w-4" />
      case 'performance_alert':
        return <TrendingUp className="h-4 w-4" />
      default:
        return <Zap className="h-4 w-4" />
    }
  }

  const getStatusColor = (status: RecentActivity['status']) => {
    switch (status) {
      case 'success':
        return 'text-green-600 bg-green-50'
      case 'warning':
        return 'text-yellow-600 bg-yellow-50'
      case 'info':
        return 'text-blue-600 bg-blue-50'
      default:
        return 'text-gray-600 bg-gray-50'
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-blue-50 p-6">
      <div className="max-w-7xl mx-auto space-y-8">
        {/* Header */}
        <FadeIn>
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-4xl font-bold text-gray-900 mb-2">
                Marketing AI Dashboard
              </h1>
              <p className="text-lg text-gray-600">
                Powered by Generative AI • Real-time insights and automation
              </p>
            </div>
            <div className="flex items-center space-x-4">
              <Button className="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700">
                <Plus className="h-4 w-4 mr-2" />
                New Campaign
              </Button>
            </div>
          </div>
        </FadeIn>

        {/* Stats Grid */}
        <StaggerContainer>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <StaggerItem>
              <Card className="bg-white/80 backdrop-blur-sm border-0 shadow-lg hover:shadow-xl transition-all duration-300">
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-gray-600">Total Campaigns</p>
                      <p className="text-3xl font-bold text-gray-900">{stats.totalCampaigns}</p>
                      <p className="text-sm text-green-600 mt-1">
                        {stats.activeCampaigns} active
                      </p>
                    </div>
                    <div className="p-3 bg-blue-100 rounded-full">
                      <Target className="h-6 w-6 text-blue-600" />
                    </div>
                  </div>
                </CardContent>
              </Card>
            </StaggerItem>

            <StaggerItem>
              <Card className="bg-white/80 backdrop-blur-sm border-0 shadow-lg hover:shadow-xl transition-all duration-300">
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-gray-600">Content Generated</p>
                      <p className="text-3xl font-bold text-gray-900">{stats.contentGenerated}</p>
                      <p className="text-sm text-green-600 mt-1">+23 this week</p>
                    </div>
                    <div className="p-3 bg-green-100 rounded-full">
                      <PenTool className="h-6 w-6 text-green-600" />
                    </div>
                  </div>
                </CardContent>
              </Card>
            </StaggerItem>

            <StaggerItem>
              <Card className="bg-white/80 backdrop-blur-sm border-0 shadow-lg hover:shadow-xl transition-all duration-300">
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-gray-600">Total Reach</p>
                      <p className="text-3xl font-bold text-gray-900">{formatNumber(stats.totalReach)}</p>
                      <p className="text-sm text-green-600 mt-1">+12% vs last month</p>
                    </div>
                    <div className="p-3 bg-purple-100 rounded-full">
                      <Users className="h-6 w-6 text-purple-600" />
                    </div>
                  </div>
                </CardContent>
              </Card>
            </StaggerItem>

            <StaggerItem>
              <Card className="bg-white/80 backdrop-blur-sm border-0 shadow-lg hover:shadow-xl transition-all duration-300">
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-gray-600">ROI</p>
                      <p className="text-3xl font-bold text-gray-900">{stats.roi}%</p>
                      <p className="text-sm text-green-600 mt-1">+8% improvement</p>
                    </div>
                    <div className="p-3 bg-orange-100 rounded-full">
                      <TrendingUp className="h-6 w-6 text-orange-600" />
                    </div>
                  </div>
                </CardContent>
              </Card>
            </StaggerItem>
          </div>
        </StaggerContainer>

        {/* Quick Actions */}
        <FadeIn delay={0.2}>
          <Card className="bg-white/80 backdrop-blur-sm border-0 shadow-lg">
            <CardHeader>
              <CardTitle className="flex items-center">
                <Zap className="h-5 w-5 mr-2 text-yellow-500" />
                Quick Actions
              </CardTitle>
              <CardDescription>
                Get started with AI-powered marketing tools
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                {quickActions.map((action) => (
                  <motion.div
                    key={action.id}
                    whileHover={{ scale: 1.02 }}
                    whileTap={{ scale: 0.98 }}
                  >
                    <Button
                      variant="outline"
                      className="h-auto p-4 flex flex-col items-start space-y-2 w-full hover:shadow-md transition-all duration-200"
                      onClick={() => window.location.href = action.href}
                    >
                      <div className={`p-2 rounded-lg ${action.color} text-white`}>
                        {action.icon}
                      </div>
                      <div className="text-left">
                        <p className="font-semibold text-gray-900">{action.title}</p>
                        <p className="text-sm text-gray-600">{action.description}</p>
                      </div>
                      <ArrowRight className="h-4 w-4 text-gray-400 self-end" />
                    </Button>
                  </motion.div>
                ))}
              </div>
            </CardContent>
          </Card>
        </FadeIn>

        {/* Performance Monitor */}
        <FadeIn delay={0.2}>
          <PerformanceMonitor realTime={true} />
        </FadeIn>

        {/* Recent Activity */}
        <FadeIn delay={0.3}>
          <Card className="bg-white/80 backdrop-blur-sm border-0 shadow-lg">
            <CardHeader>
              <CardTitle>Recent Activity</CardTitle>
              <CardDescription>
                Latest updates from your marketing campaigns
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {recentActivity.map((activity) => (
                  <div
                    key={activity.id}
                    className="flex items-center space-x-4 p-3 rounded-lg hover:bg-gray-50 transition-colors"
                  >
                    <div className={`p-2 rounded-full ${getStatusColor(activity.status)}`}>
                      {getActivityIcon(activity.type)}
                    </div>
                    <div className="flex-1">
                      <p className="font-medium text-gray-900">{activity.title}</p>
                      <p className="text-sm text-gray-600">{activity.description}</p>
                    </div>
                    <p className="text-sm text-gray-500">{activity.timestamp}</p>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </FadeIn>
      </div>
    </div>
  )
}
