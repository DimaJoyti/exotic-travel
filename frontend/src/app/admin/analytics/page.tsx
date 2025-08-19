'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { 
  TrendingUp, 
  TrendingDown,
  Users, 
  Calendar, 
  DollarSign, 
  MapPin,
  ArrowUpRight,
  ArrowDownRight,
  Download,
  RefreshCw
} from 'lucide-react'
import { useAuth } from '@/contexts/auth-context'
import { formatCurrency } from '@/lib/utils'
import AdminLayout from '@/components/admin/admin-layout'

interface AnalyticsData {
  revenue: {
    current: number
    previous: number
    growth: number
    monthly_data: Array<{ month: string; amount: number }>
  }
  bookings: {
    current: number
    previous: number
    growth: number
    monthly_data: Array<{ month: string; count: number }>
  }
  users: {
    current: number
    previous: number
    growth: number
    monthly_data: Array<{ month: string; count: number }>
  }
  destinations: {
    top_performing: Array<{
      id: number
      name: string
      bookings: number
      revenue: number
      growth: number
    }>
  }
  conversion: {
    rate: number
    previous_rate: number
    funnel: Array<{
      stage: string
      count: number
      percentage: number
    }>
  }
}

export default function AdminAnalyticsPage() {
  const { user } = useAuth()
  const router = useRouter()
  const [analytics, setAnalytics] = useState<AnalyticsData | null>(null)
  const [loading, setLoading] = useState(true)
  const [timeRange, setTimeRange] = useState('30d')

  useEffect(() => {
    if (!user) {
      router.push('/auth/login')
      return
    }

    if (user.role !== 'admin') {
      router.push('/dashboard')
      return
    }

    loadAnalytics()
  }, [user, router, timeRange])

  const loadAnalytics = async () => {
    try {
      // Mock analytics data
      const mockAnalytics: AnalyticsData = {
        revenue: {
          current: 487200,
          previous: 412800,
          growth: 18.0,
          monthly_data: [
            { month: 'Jan', amount: 320000 },
            { month: 'Feb', amount: 380000 },
            { month: 'Mar', amount: 450000 },
            { month: 'Apr', amount: 487200 },
          ]
        },
        bookings: {
          current: 156,
          previous: 132,
          growth: 18.2,
          monthly_data: [
            { month: 'Jan', count: 98 },
            { month: 'Feb', count: 115 },
            { month: 'Mar', count: 142 },
            { month: 'Apr', count: 156 },
          ]
        },
        users: {
          current: 1247,
          previous: 1089,
          growth: 14.5,
          monthly_data: [
            { month: 'Jan', count: 950 },
            { month: 'Feb', count: 1050 },
            { month: 'Mar', count: 1180 },
            { month: 'Apr', count: 1247 },
          ]
        },
        destinations: {
          top_performing: [
            { id: 1, name: 'Maldives Paradise Resort', bookings: 89, revenue: 445000, growth: 25.3 },
            { id: 2, name: 'Amazon Rainforest Adventure', bookings: 67, revenue: 268000, growth: 12.8 },
            { id: 3, name: 'Sahara Desert Glamping', bookings: 54, revenue: 162000, growth: -5.2 },
            { id: 4, name: 'Antarctic Expedition', bookings: 23, revenue: 230000, growth: 45.6 },
          ]
        },
        conversion: {
          rate: 3.2,
          previous_rate: 2.8,
          funnel: [
            { stage: 'Visitors', count: 15420, percentage: 100 },
            { stage: 'Destination Views', count: 8234, percentage: 53.4 },
            { stage: 'Booking Started', count: 1876, percentage: 12.2 },
            { stage: 'Payment Initiated', count: 687, percentage: 4.5 },
            { stage: 'Booking Completed', count: 493, percentage: 3.2 },
          ]
        }
      }
      
      setAnalytics(mockAnalytics)
    } catch (error) {
      console.error('Error loading analytics:', error)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <AdminLayout>
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      </AdminLayout>
    )
  }

  if (!analytics) {
    return (
      <AdminLayout>
        <div className="text-center py-12">
          <p className="text-gray-600">Failed to load analytics data</p>
        </div>
      </AdminLayout>
    )
  }

  return (
    <AdminLayout>
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Analytics</h1>
            <p className="text-gray-600 mt-1">Track your business performance and insights</p>
          </div>
          <div className="flex space-x-3">
            <select
              value={timeRange}
              onChange={(e) => setTimeRange(e.target.value)}
              className="px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
            >
              <option value="7d">Last 7 days</option>
              <option value="30d">Last 30 days</option>
              <option value="90d">Last 90 days</option>
              <option value="1y">Last year</option>
            </select>
            <button
              onClick={loadAnalytics}
              className="flex items-center px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
            >
              <RefreshCw className="h-4 w-4 mr-2" />
              Refresh
            </button>
            <button className="flex items-center px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors">
              <Download className="h-4 w-4 mr-2" />
              Export Report
            </button>
          </div>
        </div>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <div className="p-2 bg-green-100 rounded-lg">
                <DollarSign className="h-6 w-6 text-green-600" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">Revenue</p>
                <p className="text-2xl font-bold text-gray-900">{formatCurrency(analytics.revenue.current)}</p>
              </div>
            </div>
            <div className={`flex items-center ${analytics.revenue.growth >= 0 ? 'text-green-600' : 'text-red-600'}`}>
              {analytics.revenue.growth >= 0 ? (
                <ArrowUpRight className="h-4 w-4 mr-1" />
              ) : (
                <ArrowDownRight className="h-4 w-4 mr-1" />
              )}
              <span className="text-sm font-medium">{Math.abs(analytics.revenue.growth)}%</span>
            </div>
          </div>
          <div className="mt-4">
            <div className="text-xs text-gray-500">vs previous period: {formatCurrency(analytics.revenue.previous)}</div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <div className="p-2 bg-blue-100 rounded-lg">
                <Calendar className="h-6 w-6 text-blue-600" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">Bookings</p>
                <p className="text-2xl font-bold text-gray-900">{analytics.bookings.current}</p>
              </div>
            </div>
            <div className={`flex items-center ${analytics.bookings.growth >= 0 ? 'text-green-600' : 'text-red-600'}`}>
              {analytics.bookings.growth >= 0 ? (
                <ArrowUpRight className="h-4 w-4 mr-1" />
              ) : (
                <ArrowDownRight className="h-4 w-4 mr-1" />
              )}
              <span className="text-sm font-medium">{Math.abs(analytics.bookings.growth)}%</span>
            </div>
          </div>
          <div className="mt-4">
            <div className="text-xs text-gray-500">vs previous period: {analytics.bookings.previous}</div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <div className="p-2 bg-purple-100 rounded-lg">
                <Users className="h-6 w-6 text-purple-600" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">Total Users</p>
                <p className="text-2xl font-bold text-gray-900">{analytics.users.current}</p>
              </div>
            </div>
            <div className={`flex items-center ${analytics.users.growth >= 0 ? 'text-green-600' : 'text-red-600'}`}>
              {analytics.users.growth >= 0 ? (
                <ArrowUpRight className="h-4 w-4 mr-1" />
              ) : (
                <ArrowDownRight className="h-4 w-4 mr-1" />
              )}
              <span className="text-sm font-medium">{Math.abs(analytics.users.growth)}%</span>
            </div>
          </div>
          <div className="mt-4">
            <div className="text-xs text-gray-500">vs previous period: {analytics.users.previous}</div>
          </div>
        </div>
      </div>

      {/* Charts and Detailed Analytics */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
        {/* Revenue Trend */}
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Revenue Trend</h3>
          <div className="space-y-3">
            {analytics.revenue.monthly_data.map((data, index) => (
              <div key={data.month} className="flex items-center justify-between">
                <span className="text-sm text-gray-600">{data.month}</span>
                <div className="flex items-center space-x-3">
                  <div className="w-32 bg-gray-200 rounded-full h-2">
                    <div 
                      className="bg-green-600 h-2 rounded-full" 
                      style={{ width: `${(data.amount / Math.max(...analytics.revenue.monthly_data.map(d => d.amount))) * 100}%` }}
                    ></div>
                  </div>
                  <span className="text-sm font-medium text-gray-900 w-20 text-right">
                    {formatCurrency(data.amount)}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Conversion Funnel */}
        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Conversion Funnel</h3>
          <div className="space-y-3">
            {analytics.conversion.funnel.map((stage, index) => (
              <div key={stage.stage} className="flex items-center justify-between">
                <span className="text-sm text-gray-600">{stage.stage}</span>
                <div className="flex items-center space-x-3">
                  <div className="w-32 bg-gray-200 rounded-full h-2">
                    <div 
                      className="bg-blue-600 h-2 rounded-full" 
                      style={{ width: `${stage.percentage}%` }}
                    ></div>
                  </div>
                  <span className="text-sm font-medium text-gray-900 w-16 text-right">
                    {stage.percentage}%
                  </span>
                  <span className="text-xs text-gray-500 w-16 text-right">
                    {stage.count.toLocaleString()}
                  </span>
                </div>
              </div>
            ))}
          </div>
          <div className="mt-4 pt-4 border-t border-gray-200">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium text-gray-900">Overall Conversion Rate</span>
              <div className="flex items-center">
                <span className="text-lg font-bold text-gray-900">{analytics.conversion.rate}%</span>
                <div className={`ml-2 flex items-center ${analytics.conversion.rate >= analytics.conversion.previous_rate ? 'text-green-600' : 'text-red-600'}`}>
                  {analytics.conversion.rate >= analytics.conversion.previous_rate ? (
                    <TrendingUp className="h-4 w-4" />
                  ) : (
                    <TrendingDown className="h-4 w-4" />
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Top Performing Destinations */}
      <div className="bg-white rounded-lg shadow p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-6">Top Performing Destinations</h3>
        <div className="overflow-x-auto">
          <table className="min-w-full">
            <thead>
              <tr className="border-b border-gray-200">
                <th className="text-left py-3 px-4 font-medium text-gray-600">Destination</th>
                <th className="text-left py-3 px-4 font-medium text-gray-600">Bookings</th>
                <th className="text-left py-3 px-4 font-medium text-gray-600">Revenue</th>
                <th className="text-left py-3 px-4 font-medium text-gray-600">Growth</th>
              </tr>
            </thead>
            <tbody>
              {analytics.destinations.top_performing.map((destination, index) => (
                <tr key={destination.id} className="border-b border-gray-100">
                  <td className="py-4 px-4">
                    <div className="flex items-center">
                      <span className="w-6 h-6 bg-primary text-primary-foreground rounded-full flex items-center justify-center text-xs font-medium mr-3">
                        {index + 1}
                      </span>
                      <span className="font-medium text-gray-900">{destination.name}</span>
                    </div>
                  </td>
                  <td className="py-4 px-4 text-gray-900">{destination.bookings}</td>
                  <td className="py-4 px-4 text-gray-900">{formatCurrency(destination.revenue)}</td>
                  <td className="py-4 px-4">
                    <div className={`flex items-center ${destination.growth >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                      {destination.growth >= 0 ? (
                        <ArrowUpRight className="h-4 w-4 mr-1" />
                      ) : (
                        <ArrowDownRight className="h-4 w-4 mr-1" />
                      )}
                      <span className="text-sm font-medium">{Math.abs(destination.growth)}%</span>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </AdminLayout>
  )
}
