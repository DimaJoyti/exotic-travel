'use client'

import { useAuth } from '@/contexts/auth-context'
import { useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import Link from 'next/link'
import { User, MapPin, Calendar, CreditCard, Settings, BookOpen, Star, Award, Plane, Clock } from 'lucide-react'
import { UsersService, UserStats } from '@/lib/users'
import { BookingsService } from '@/lib/bookings'
import { Booking } from '@/types'
import { formatCurrency } from '@/lib/utils'

export default function DashboardPage() {
  const { user, loading } = useAuth()
  const router = useRouter()
  const [userStats, setUserStats] = useState<UserStats | null>(null)
  const [recentBookings, setRecentBookings] = useState<Booking[]>([])
  const [statsLoading, setStatsLoading] = useState(true)

  useEffect(() => {
    if (!loading && !user) {
      router.push('/auth/login')
    }
  }, [user, loading, router])

  useEffect(() => {
    const loadUserData = async () => {
      if (!user) return

      try {
        // Load user stats
        let stats: UserStats
        try {
          stats = await UsersService.getUserStats(user.id)
        } catch (apiError) {
          console.warn('API not available, using mock data:', apiError)
          stats = UsersService.getMockUserStats()
        }
        setUserStats(stats)

        // Load recent bookings
        try {
          const bookings = await BookingsService.getUserBookings(user.id)
          setRecentBookings(bookings.slice(0, 3)) // Show only 3 most recent
        } catch (apiError) {
          console.warn('API not available, using mock bookings')
          // Mock recent bookings data
          setRecentBookings([])
        }
      } catch (error) {
        console.error('Error loading user data:', error)
      } finally {
        setStatsLoading(false)
      }
    }

    loadUserData()
  }, [user])

  if (loading || statsLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  if (!user) {
    return null
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-8">
          <div className="flex flex-col md:flex-row md:items-center md:justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">
                Welcome back, {user.first_name}!
              </h1>
              <p className="text-gray-600 mt-2">
                Manage your bookings and explore new destinations
              </p>
            </div>
            <div className="mt-4 md:mt-0 flex space-x-3">
              <Link
                href="/profile"
                className="flex items-center px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
              >
                <Settings className="h-4 w-4 mr-2" />
                Settings
              </Link>
              <Link
                href="/destinations"
                className="flex items-center px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors"
              >
                <Plane className="h-4 w-4 mr-2" />
                Book Trip
              </Link>
            </div>
          </div>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-2 bg-blue-100 rounded-lg">
                <BookOpen className="h-6 w-6 text-blue-600" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">Total Bookings</p>
                <p className="text-2xl font-bold text-gray-900">{userStats?.total_bookings || 0}</p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-2 bg-green-100 rounded-lg">
                <MapPin className="h-6 w-6 text-green-600" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">Destinations Visited</p>
                <p className="text-2xl font-bold text-gray-900">{userStats?.destinations_visited || 0}</p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-2 bg-purple-100 rounded-lg">
                <CreditCard className="h-6 w-6 text-purple-600" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">Total Spent</p>
                <p className="text-2xl font-bold text-gray-900">
                  {userStats ? formatCurrency(userStats.total_spent) : '$0'}
                </p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-2 bg-orange-100 rounded-lg">
                <Award className="h-6 w-6 text-orange-600" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">Loyalty Points</p>
                <p className="text-2xl font-bold text-gray-900">{userStats?.loyalty_points || 0}</p>
              </div>
            </div>
          </div>
        </div>

        {/* Next Trip Card */}
        {userStats?.next_trip && (
          <div className="bg-gradient-to-r from-primary to-primary/80 text-white rounded-lg p-6 mb-8">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-lg font-semibold mb-2">Your Next Adventure</h3>
                <p className="text-xl font-bold">{userStats.next_trip.destination}</p>
                <div className="flex items-center mt-2 text-primary-foreground/90">
                  <Calendar className="h-4 w-4 mr-2" />
                  {new Date(userStats.next_trip.date).toLocaleDateString()}
                </div>
              </div>
              <div className="text-right">
                <div className="text-3xl font-bold">{userStats.next_trip.days_until}</div>
                <div className="text-sm text-primary-foreground/90">days to go</div>
              </div>
            </div>
          </div>
        )}

        {/* Quick Actions */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Recent Bookings */}
          <div className="bg-white rounded-lg shadow">
            <div className="p-6 border-b border-gray-200">
              <h2 className="text-lg font-semibold text-gray-900">Recent Bookings</h2>
            </div>
            <div className="p-6">
              <div className="space-y-4">
                <div className="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
                  <div>
                    <h3 className="font-medium text-gray-900">Maldives Paradise Resort</h3>
                    <p className="text-sm text-gray-600">March 15-22, 2024</p>
                  </div>
                  <span className="px-2 py-1 text-xs font-medium bg-green-100 text-green-800 rounded-full">
                    Confirmed
                  </span>
                </div>
                <div className="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
                  <div>
                    <h3 className="font-medium text-gray-900">Amazon Rainforest Adventure</h3>
                    <p className="text-sm text-gray-600">April 10-20, 2024</p>
                  </div>
                  <span className="px-2 py-1 text-xs font-medium bg-yellow-100 text-yellow-800 rounded-full">
                    Pending
                  </span>
                </div>
              </div>
              <div className="mt-6">
                <a
                  href="/bookings"
                  className="text-primary hover:text-primary/80 font-medium text-sm"
                >
                  View all bookings →
                </a>
              </div>
            </div>
          </div>

          {/* Quick Actions */}
          <div className="bg-white rounded-lg shadow">
            <div className="p-6 border-b border-gray-200">
              <h2 className="text-lg font-semibold text-gray-900">Quick Actions</h2>
            </div>
            <div className="p-6">
              <div className="space-y-4">
                <a
                  href="/destinations"
                  className="block p-4 border border-gray-200 rounded-lg hover:border-primary hover:bg-primary/5 transition-colors"
                >
                  <h3 className="font-medium text-gray-900 mb-1">Browse Destinations</h3>
                  <p className="text-sm text-gray-600">Discover new exotic locations</p>
                </a>
                <a
                  href="/profile"
                  className="block p-4 border border-gray-200 rounded-lg hover:border-primary hover:bg-primary/5 transition-colors"
                >
                  <h3 className="font-medium text-gray-900 mb-1">Update Profile</h3>
                  <p className="text-sm text-gray-600">Manage your account settings</p>
                </a>
                <a
                  href="/contact"
                  className="block p-4 border border-gray-200 rounded-lg hover:border-primary hover:bg-primary/5 transition-colors"
                >
                  <h3 className="font-medium text-gray-900 mb-1">Contact Support</h3>
                  <p className="text-sm text-gray-600">Get help with your bookings</p>
                </a>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
