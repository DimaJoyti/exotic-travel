'use client'

import React, { useState, useEffect } from 'react'
import { 
  Activity, 
  TrendingUp, 
  TrendingDown, 
  AlertTriangle, 
  CheckCircle, 
  Clock, 
  Zap,
  Target,
  DollarSign,
  Users,
  Eye,
  MousePointer
} from 'lucide-react'
import { motion } from 'framer-motion'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import { FadeIn, StaggerContainer, StaggerItem } from '@/components/ui/animated'

interface PerformanceMonitorProps {
  campaignId?: number
  realTime?: boolean
}

interface Alert {
  id: string
  type: 'success' | 'warning' | 'error' | 'info'
  title: string
  message: string
  timestamp: string
  campaignId?: number
  metric?: string
}

interface RealtimeMetric {
  name: string
  value: number
  change: number
  target?: number
  unit: string
  icon: React.ReactNode
  color: string
}

export default function PerformanceMonitor({ campaignId, realTime = true }: PerformanceMonitorProps) {
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [metrics, setMetrics] = useState<RealtimeMetric[]>([])
  const [isConnected, setIsConnected] = useState(false)
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date())

  // Mock real-time data updates
  useEffect(() => {
    if (!realTime) return

    const interval = setInterval(() => {
      updateMetrics()
      checkForAlerts()
      setLastUpdate(new Date())
    }, 5000) // Update every 5 seconds

    setIsConnected(true)

    return () => {
      clearInterval(interval)
      setIsConnected(false)
    }
  }, [realTime])

  // Initialize mock data
  useEffect(() => {
    setMetrics([
      {
        name: 'Impressions/Hour',
        value: 12500,
        change: 8.3,
        target: 15000,
        unit: '',
        icon: <Eye className="h-4 w-4" />,
        color: 'bg-blue-500'
      },
      {
        name: 'Click Rate',
        value: 3.2,
        change: -0.5,
        target: 3.5,
        unit: '%',
        icon: <MousePointer className="h-4 w-4" />,
        color: 'bg-green-500'
      },
      {
        name: 'Cost/Click',
        value: 1.23,
        change: -5.2,
        target: 1.50,
        unit: '$',
        icon: <DollarSign className="h-4 w-4" />,
        color: 'bg-orange-500'
      },
      {
        name: 'Conversions/Hour',
        value: 47,
        change: 12.1,
        target: 50,
        unit: '',
        icon: <Target className="h-4 w-4" />,
        color: 'bg-purple-500'
      }
    ])

    setAlerts([
      {
        id: '1',
        type: 'success',
        title: 'Performance Improvement',
        message: 'CTR increased by 15% in the last hour',
        timestamp: new Date(Date.now() - 300000).toISOString(),
        campaignId: 1,
        metric: 'ctr'
      },
      {
        id: '2',
        type: 'warning',
        title: 'Budget Alert',
        message: 'Campaign approaching 80% of daily budget',
        timestamp: new Date(Date.now() - 600000).toISOString(),
        campaignId: 1,
        metric: 'budget'
      },
      {
        id: '3',
        type: 'info',
        title: 'Optimization Suggestion',
        message: 'Consider increasing bid for high-performing keywords',
        timestamp: new Date(Date.now() - 900000).toISOString(),
        campaignId: 1,
        metric: 'optimization'
      }
    ])
  }, [])

  const updateMetrics = () => {
    setMetrics(prev => prev.map(metric => ({
      ...metric,
      value: metric.value + (Math.random() - 0.5) * metric.value * 0.1,
      change: (Math.random() - 0.5) * 20
    })))
  }

  const checkForAlerts = () => {
    // Simulate new alerts
    if (Math.random() < 0.3) { // 30% chance of new alert
      const alertTypes = ['success', 'warning', 'error', 'info'] as const
      const messages = [
        'Conversion rate spike detected',
        'Unusual traffic pattern observed',
        'Budget threshold reached',
        'New optimization opportunity identified'
      ]
      
      const newAlert: Alert = {
        id: Date.now().toString(),
        type: alertTypes[Math.floor(Math.random() * alertTypes.length)],
        title: 'Real-time Alert',
        message: messages[Math.floor(Math.random() * messages.length)],
        timestamp: new Date().toISOString(),
        campaignId: campaignId || 1
      }

      setAlerts(prev => [newAlert, ...prev.slice(0, 9)]) // Keep only 10 most recent
    }
  }

  const getAlertIcon = (type: Alert['type']) => {
    switch (type) {
      case 'success':
        return <CheckCircle className="h-4 w-4 text-green-600" />
      case 'warning':
        return <AlertTriangle className="h-4 w-4 text-yellow-600" />
      case 'error':
        return <AlertTriangle className="h-4 w-4 text-red-600" />
      default:
        return <Activity className="h-4 w-4 text-blue-600" />
    }
  }

  const getAlertColor = (type: Alert['type']) => {
    switch (type) {
      case 'success':
        return 'border-green-200 bg-green-50'
      case 'warning':
        return 'border-yellow-200 bg-yellow-50'
      case 'error':
        return 'border-red-200 bg-red-50'
      default:
        return 'border-blue-200 bg-blue-50'
    }
  }

  const formatNumber = (num: number, unit: string) => {
    const formatted = num >= 1000 ? `${(num / 1000).toFixed(1)}K` : num.toFixed(unit === '%' || unit === '$' ? 2 : 0)
    return unit === '$' ? `$${formatted}` : `${formatted}${unit}`
  }

  const getProgressValue = (current: number, target?: number) => {
    if (!target) return 0
    return Math.min((current / target) * 100, 100)
  }

  const getProgressColor = (current: number, target?: number) => {
    if (!target) return 'bg-gray-400'
    const percentage = (current / target) * 100
    if (percentage >= 90) return 'bg-green-500'
    if (percentage >= 70) return 'bg-yellow-500'
    return 'bg-red-500'
  }

  return (
    <div className="space-y-6">
      {/* Connection Status */}
      <FadeIn>
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className={`w-3 h-3 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'} animate-pulse`} />
            <span className="text-sm font-medium text-gray-700">
              {isConnected ? 'Real-time monitoring active' : 'Monitoring disconnected'}
            </span>
            <span className="text-xs text-gray-500">
              Last update: {lastUpdate.toLocaleTimeString()}
            </span>
          </div>
          <Button variant="outline" size="sm">
            <Activity className="h-4 w-4 mr-2" />
            View Details
          </Button>
        </div>
      </FadeIn>

      {/* Real-time Metrics */}
      <StaggerContainer>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {metrics.map((metric, index) => (
            <StaggerItem key={metric.name}>
              <Card className="bg-white/80 backdrop-blur-sm border-0 shadow-lg">
                <CardContent className="p-4">
                  <div className="flex items-center justify-between mb-2">
                    <div className={`p-2 rounded-lg ${metric.color} text-white`}>
                      {metric.icon}
                    </div>
                    <div className={`flex items-center text-sm ${
                      metric.change >= 0 ? 'text-green-600' : 'text-red-600'
                    }`}>
                      {metric.change >= 0 ? (
                        <TrendingUp className="h-3 w-3 mr-1" />
                      ) : (
                        <TrendingDown className="h-3 w-3 mr-1" />
                      )}
                      {Math.abs(metric.change).toFixed(1)}%
                    </div>
                  </div>
                  <div>
                    <p className="text-xs text-gray-600 mb-1">{metric.name}</p>
                    <p className="text-lg font-bold text-gray-900">
                      {formatNumber(metric.value, metric.unit)}
                    </p>
                    {metric.target && (
                      <div className="mt-2">
                        <div className="flex items-center justify-between text-xs text-gray-600 mb-1">
                          <span>Target: {formatNumber(metric.target, metric.unit)}</span>
                          <span>{getProgressValue(metric.value, metric.target).toFixed(0)}%</span>
                        </div>
                        <Progress 
                          value={getProgressValue(metric.value, metric.target)} 
                          className="h-1"
                        />
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>
            </StaggerItem>
          ))}
        </div>
      </StaggerContainer>

      {/* Real-time Alerts */}
      <FadeIn delay={0.2}>
        <Card className="bg-white/80 backdrop-blur-sm border-0 shadow-lg">
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle className="flex items-center">
                  <Zap className="h-5 w-5 mr-2 text-yellow-500" />
                  Real-time Alerts
                </CardTitle>
                <CardDescription>
                  Live performance notifications and optimization suggestions
                </CardDescription>
              </div>
              <Badge variant="outline" className="text-blue-600 border-blue-200">
                {alerts.length} Active
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-3 max-h-80 overflow-y-auto">
              {alerts.map((alert) => (
                <motion.div
                  key={alert.id}
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  className={`p-3 rounded-lg border ${getAlertColor(alert.type)} transition-all duration-200 hover:shadow-md`}
                >
                  <div className="flex items-start space-x-3">
                    <div className="mt-0.5">
                      {getAlertIcon(alert.type)}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between">
                        <p className="text-sm font-medium text-gray-900">{alert.title}</p>
                        <span className="text-xs text-gray-500">
                          {new Date(alert.timestamp).toLocaleTimeString()}
                        </span>
                      </div>
                      <p className="text-sm text-gray-600 mt-1">{alert.message}</p>
                      {alert.campaignId && (
                        <Badge variant="outline" className="mt-2 text-xs">
                          Campaign {alert.campaignId}
                        </Badge>
                      )}
                    </div>
                  </div>
                </motion.div>
              ))}
              
              {alerts.length === 0 && (
                <div className="text-center py-8">
                  <CheckCircle className="h-12 w-12 text-green-500 mx-auto mb-3" />
                  <p className="text-gray-600 font-medium">All systems running smoothly</p>
                  <p className="text-sm text-gray-500">No alerts at this time</p>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      </FadeIn>
    </div>
  )
}
