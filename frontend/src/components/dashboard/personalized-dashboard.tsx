'use client'

import React, { useState, useEffect } from 'react'
import { User, Heart, MapPin, Calendar, TrendingUp, Star, Clock, Zap, Brain, Target, Award, Compass } from 'lucide-react'
import { motion } from 'framer-motion'
import { FadeIn, StaggerContainer, StaggerItem, HoverAnimation } from '@/components/ui/animated'
import { Button } from '@/components/ui/button'
import { Destination, User as UserType } from '@/types'
import { RecommendationEngine, RecommendationResult } from '@/lib/recommendations'
import { UserPreferencesService, TravelProfile } from '@/lib/user-preferences'

interface PersonalizedDashboardProps {
  user: UserType
  className?: string
}

interface DashboardStats {
  destinations_viewed: number
  wishlist_items: number
  bookings_made: number
  countries_explored: number
  total_savings: number
  travel_score: number
}

interface UpcomingTrip {
  id: string
  destination: string
  date: string
  days_until: number
  status: 'confirmed' | 'pending' | 'planning'
}

interface TravelInsight {
  type: 'recommendation' | 'trend' | 'saving' | 'achievement'
  title: string
  description: string
  action?: string
  icon: React.ReactNode
  priority: 'high' | 'medium' | 'low'
}

export default function PersonalizedDashboard({ user, className = '' }: PersonalizedDashboardProps) {
  const [recommendations, setRecommendations] = useState<RecommendationResult[]>([])
  const [travelProfile, setTravelProfile] = useState<TravelProfile | null>(null)
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [upcomingTrips, setUpcomingTrips] = useState<UpcomingTrip[]>([])
  const [insights, setInsights] = useState<TravelInsight[]>([])
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState<'overview' | 'recommendations' | 'insights' | 'profile'>('overview')

  useEffect(() => {
    loadDashboardData()
  }, [user.id])

  const loadDashboardData = async () => {
    setLoading(true)
    try {
      // Load user profile and preferences
      const profiles = await UserPreferencesService.getTravelProfiles(user.id.toString())
      const defaultProfile = profiles.find(p => p.is_default) || profiles[0]
      setTravelProfile(defaultProfile)

      // Load personalized recommendations
      if (defaultProfile) {
        const recommendationRequest = {
          user_id: user.id.toString(),
          preferences: defaultProfile.preferences,
          behavior: {
            search_history: [],
            view_history: [],
            booking_history: [],
            wishlist_items: [],
            time_spent_on_pages: {},
            interaction_patterns: [],
            seasonal_preferences: []
          },
          context: {
            current_season: getCurrentSeason(),
            user_location: { lat: 40.7128, lng: -74.0060, country: 'US' },
            trending_destinations: await RecommendationEngine.getTrendingDestinations(),
            weather_conditions: {},
            special_events: [],
            market_conditions: []
          },
          limit: 6
        }

        const recs = await RecommendationEngine.getPersonalizedRecommendations(recommendationRequest)
        setRecommendations(recs)
      }

      // Load dashboard stats
      setStats(getMockStats())
      
      // Load upcoming trips
      setUpcomingTrips(getMockUpcomingTrips())
      
      // Generate insights
      setInsights(generateInsights(defaultProfile))

    } catch (error) {
      console.error('Error loading dashboard data:', error)
    } finally {
      setLoading(false)
    }
  }

  const getCurrentSeason = (): string => {
    const month = new Date().getMonth()
    if (month >= 2 && month <= 4) return 'spring'
    if (month >= 5 && month <= 7) return 'summer'
    if (month >= 8 && month <= 10) return 'fall'
    return 'winter'
  }

  const getMockStats = (): DashboardStats => ({
    destinations_viewed: 47,
    wishlist_items: 12,
    bookings_made: 3,
    countries_explored: 8,
    total_savings: 1250,
    travel_score: 85
  })

  const getMockUpcomingTrips = (): UpcomingTrip[] => [
    {
      id: '1',
      destination: 'Santorini, Greece',
      date: '2024-06-15',
      days_until: 45,
      status: 'confirmed'
    },
    {
      id: '2',
      destination: 'Tokyo, Japan',
      date: '2024-09-20',
      days_until: 142,
      status: 'planning'
    }
  ]

  const generateInsights = (profile: TravelProfile | null): TravelInsight[] => {
    if (!profile) return []

    return [
      {
        type: 'recommendation',
        title: 'Perfect Season for Your Style',
        description: 'Spring is ideal for cultural destinations in Europe. Book now for 20% savings.',
        action: 'View Recommendations',
        icon: <Compass className="h-5 w-5" />,
        priority: 'high'
      },
      {
        type: 'trend',
        title: 'Trending in Your Budget',
        description: 'Eastern Europe destinations are gaining popularity and fit your budget perfectly.',
        action: 'Explore Trends',
        icon: <TrendingUp className="h-5 w-5" />,
        priority: 'medium'
      },
      {
        type: 'saving',
        title: 'Smart Booking Window',
        description: 'Book your next trip 6-8 weeks in advance to save up to 25%.',
        icon: <Target className="h-5 w-5" />,
        priority: 'medium'
      },
      {
        type: 'achievement',
        title: 'Travel Explorer Badge',
        description: 'You\'ve explored 8 countries! Visit 2 more to unlock the Global Explorer badge.',
        action: 'View Achievements',
        icon: <Award className="h-5 w-5" />,
        priority: 'low'
      }
    ]
  }

  const formatCurrency = (amount: number): string => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(amount)
  }

  if (loading) {
    return (
      <div className={`flex items-center justify-center py-12 ${className}`}>
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-brand-500 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading your personalized dashboard...</p>
        </div>
      </div>
    )
  }

  return (
    <div className={`space-y-8 ${className}`}>
      {/* Welcome Header */}
      <FadeIn>
        <div className="bg-gradient-to-r from-brand-500 to-accent-500 rounded-2xl p-8 text-white">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold mb-2">
                Welcome back, {user.first_name}!
              </h1>
              <p className="text-white/90 text-lg">
                Your next adventure awaits. Here's what we've discovered for you.
              </p>
            </div>
            <div className="text-right">
              <div className="text-2xl font-bold">{stats?.travel_score}/100</div>
              <div className="text-white/80 text-sm">Travel Score</div>
            </div>
          </div>
        </div>
      </FadeIn>

      {/* Tab Navigation */}
      <div className="border-b border-gray-200">
        <nav className="-mb-px flex space-x-8">
          {[
            { id: 'overview', label: 'Overview', icon: User },
            { id: 'recommendations', label: 'For You', icon: Brain },
            { id: 'insights', label: 'Insights', icon: Zap },
            { id: 'profile', label: 'Profile', icon: Target }
          ].map((tab) => {
            const Icon = tab.icon
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as any)}
                className={`
                  flex items-center space-x-2 py-4 px-1 border-b-2 font-medium text-sm transition-colors
                  ${activeTab === tab.id
                    ? 'border-brand-500 text-brand-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                  }
                `}
              >
                <Icon className="h-4 w-4" />
                <span>{tab.label}</span>
              </button>
            )
          })}
        </nav>
      </div>

      {/* Tab Content */}
      {activeTab === 'overview' && (
        <div className="space-y-8">
          {/* Stats Grid */}
          <StaggerContainer staggerDelay={0.1}>
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
              {stats && [
                { label: 'Destinations Viewed', value: stats.destinations_viewed, icon: MapPin },
                { label: 'Wishlist Items', value: stats.wishlist_items, icon: Heart },
                { label: 'Bookings Made', value: stats.bookings_made, icon: Calendar },
                { label: 'Countries Explored', value: stats.countries_explored, icon: Compass },
                { label: 'Total Savings', value: formatCurrency(stats.total_savings), icon: Target },
                { label: 'Travel Score', value: `${stats.travel_score}/100`, icon: Star }
              ].map((stat, index) => {
                const Icon = stat.icon
                return (
                  <StaggerItem key={index}>
                    <HoverAnimation hoverY={-4}>
                      <div className="bg-white rounded-lg p-4 border border-gray-200 text-center">
                        <Icon className="h-6 w-6 text-brand-500 mx-auto mb-2" />
                        <div className="text-2xl font-bold text-gray-900">{stat.value}</div>
                        <div className="text-xs text-gray-600">{stat.label}</div>
                      </div>
                    </HoverAnimation>
                  </StaggerItem>
                )
              })}
            </div>
          </StaggerContainer>

          {/* Upcoming Trips */}
          {upcomingTrips.length > 0 && (
            <div>
              <h2 className="text-xl font-bold text-gray-900 mb-4">Upcoming Adventures</h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {upcomingTrips.map((trip) => (
                  <motion.div
                    key={trip.id}
                    whileHover={{ scale: 1.02 }}
                    className="bg-white rounded-lg p-6 border border-gray-200 shadow-sm"
                  >
                    <div className="flex items-center justify-between mb-3">
                      <h3 className="font-semibold text-gray-900">{trip.destination}</h3>
                      <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                        trip.status === 'confirmed' ? 'bg-green-100 text-green-700' :
                        trip.status === 'pending' ? 'bg-yellow-100 text-yellow-700' :
                        'bg-blue-100 text-blue-700'
                      }`}>
                        {trip.status}
                      </span>
                    </div>
                    <div className="flex items-center text-gray-600 mb-2">
                      <Calendar className="h-4 w-4 mr-2" />
                      <span>{new Date(trip.date).toLocaleDateString()}</span>
                    </div>
                    <div className="flex items-center text-gray-600">
                      <Clock className="h-4 w-4 mr-2" />
                      <span>{trip.days_until} days to go</span>
                    </div>
                  </motion.div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {activeTab === 'recommendations' && (
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-bold text-gray-900">Personalized for You</h2>
            <Button variant="outline" size="sm">
              Refresh Recommendations
            </Button>
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {recommendations.map((rec, index) => (
              <motion.div
                key={rec.destination.id}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: index * 0.1 }}
                className="bg-white rounded-lg overflow-hidden border border-gray-200 shadow-sm hover:shadow-md transition-shadow"
              >
                <div className="aspect-video bg-gray-200">
                  {rec.destination.images[0] && (
                    <img
                      src={rec.destination.images[0]}
                      alt={rec.destination.name}
                      className="w-full h-full object-cover"
                    />
                  )}
                </div>
                <div className="p-4">
                  <div className="flex items-center justify-between mb-2">
                    <h3 className="font-semibold text-gray-900">{rec.destination.name}</h3>
                    <div className="text-sm text-brand-600 font-medium">
                      {Math.round(rec.confidence_score * 100)}% match
                    </div>
                  </div>
                  <p className="text-gray-600 text-sm mb-3">{rec.destination.city}, {rec.destination.country}</p>
                  
                  {/* AI insights would go here */}
                  
                  <div className="flex items-center justify-between">
                    <span className="text-lg font-bold text-gray-900">
                      {formatCurrency(rec.destination.price)}
                    </span>
                    <Button size="sm">View Details</Button>
                  </div>
                </div>
              </motion.div>
            ))}
          </div>
        </div>
      )}

      {activeTab === 'insights' && (
        <div className="space-y-6">
          <h2 className="text-xl font-bold text-gray-900">Travel Insights</h2>
          
          <div className="space-y-4">
            {insights.map((insight, index) => (
              <motion.div
                key={index}
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: index * 0.1 }}
                className={`p-4 rounded-lg border-l-4 ${
                  insight.priority === 'high' ? 'border-red-500 bg-red-50' :
                  insight.priority === 'medium' ? 'border-yellow-500 bg-yellow-50' :
                  'border-blue-500 bg-blue-50'
                }`}
              >
                <div className="flex items-start space-x-3">
                  <div className={`p-2 rounded-lg ${
                    insight.priority === 'high' ? 'bg-red-100 text-red-600' :
                    insight.priority === 'medium' ? 'bg-yellow-100 text-yellow-600' :
                    'bg-blue-100 text-blue-600'
                  }`}>
                    {insight.icon}
                  </div>
                  <div className="flex-1">
                    <h3 className="font-semibold text-gray-900 mb-1">{insight.title}</h3>
                    <p className="text-gray-700 mb-2">{insight.description}</p>
                    {insight.action && (
                      <Button variant="outline" size="sm">
                        {insight.action}
                      </Button>
                    )}
                  </div>
                </div>
              </motion.div>
            ))}
          </div>
        </div>
      )}

      {activeTab === 'profile' && travelProfile && (
        <div className="space-y-6">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-bold text-gray-900">Travel Profile</h2>
            <Button variant="outline" size="sm">
              Edit Profile
            </Button>
          </div>
          
          <div className="bg-white rounded-lg p-6 border border-gray-200">
            <h3 className="font-semibold text-gray-900 mb-4">{travelProfile.name}</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Travel Style</label>
                <div className="capitalize text-gray-900">{travelProfile.preferences.travel_style}</div>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Budget Range</label>
                <div className="text-gray-900">
                  {formatCurrency(travelProfile.preferences.budget_range[0])} - {formatCurrency(travelProfile.preferences.budget_range[1])}
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Preferred Duration</label>
                <div className="text-gray-900">{travelProfile.preferences.preferred_duration.join(', ')} days</div>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Group Size</label>
                <div className="text-gray-900">{travelProfile.preferences.group_size} people</div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
