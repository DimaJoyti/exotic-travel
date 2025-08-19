'use client'

import React, { useState, useEffect } from 'react'
import { TrendingUp, TrendingDown, Eye, Users, Zap, Wifi, WifiOff, Clock } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { RealTimeService, RealTimeUpdate, PriceUpdate, BookingUpdate, ViewCountUpdate } from '@/lib/real-time'
import { Destination } from '@/types'

interface RealTimeIndicatorProps {
  destination: Destination
  showPriceUpdates?: boolean
  showBookingUpdates?: boolean
  showViewCount?: boolean
  className?: string
}

export default function RealTimeIndicator({
  destination,
  showPriceUpdates = true,
  showBookingUpdates = true,
  showViewCount = true,
  className = ''
}: RealTimeIndicatorProps) {
  const [isConnected, setIsConnected] = useState(false)
  const [lastUpdate, setLastUpdate] = useState<RealTimeUpdate | null>(null)
  const [priceData, setPriceData] = useState<PriceUpdate | null>(null)
  const [bookingData, setBookingData] = useState<BookingUpdate | null>(null)
  const [viewData, setViewData] = useState<ViewCountUpdate | null>(null)
  const [showUpdateAnimation, setShowUpdateAnimation] = useState(false)

  useEffect(() => {
    // Connect to real-time service
    RealTimeService.connect([destination])
    setIsConnected(true)

    // Listen for updates
    const unsubscribe = RealTimeService.addListener((update: RealTimeUpdate) => {
      if (update.destination_id === destination.id.toString() || update.destination_id === 'system') {
        setLastUpdate(update)
        setShowUpdateAnimation(true)

        // Process different types of updates
        switch (update.type) {
          case 'price':
            if (showPriceUpdates) {
              setPriceData(update.data as PriceUpdate)
            }
            break
          case 'booking':
            if (showBookingUpdates) {
              setBookingData(update.data as BookingUpdate)
            }
            break
          case 'view_count':
            if (showViewCount) {
              setViewData(update.data as ViewCountUpdate)
            }
            break
        }

        // Hide animation after 3 seconds
        setTimeout(() => setShowUpdateAnimation(false), 3000)
      }
    })

    return () => {
      unsubscribe()
      RealTimeService.disconnect()
    }
  }, [destination.id, showPriceUpdates, showBookingUpdates, showViewCount])

  const formatTimeAgo = (timestamp: string): string => {
    const now = new Date()
    const updateTime = new Date(timestamp)
    const diffMs = now.getTime() - updateTime.getTime()
    const diffSeconds = Math.floor(diffMs / 1000)
    const diffMinutes = Math.floor(diffSeconds / 60)

    if (diffSeconds < 60) {
      return `${diffSeconds}s ago`
    } else if (diffMinutes < 60) {
      return `${diffMinutes}m ago`
    } else {
      return updateTime.toLocaleTimeString()
    }
  }

  return (
    <div className={`space-y-3 ${className}`}>
      {/* Connection Status */}
      <div className="flex items-center space-x-2">
        <motion.div
          animate={{ 
            scale: isConnected ? [1, 1.2, 1] : 1,
            opacity: isConnected ? 1 : 0.5
          }}
          transition={{ 
            duration: 2, 
            repeat: isConnected ? Infinity : 0,
            repeatType: 'loop'
          }}
          className={`w-2 h-2 rounded-full ${
            isConnected ? 'bg-green-500' : 'bg-red-500'
          }`}
        />
        <span className="text-xs text-gray-500">
          {isConnected ? 'Live updates' : 'Disconnected'}
        </span>
        {lastUpdate && (
          <span className="text-xs text-gray-400">
            • {formatTimeAgo(lastUpdate.timestamp)}
          </span>
        )}
      </div>

      {/* Price Updates */}
      <AnimatePresence>
        {showPriceUpdates && priceData && (
          <motion.div
            initial={{ opacity: 0, y: 20, scale: 0.9 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: -20, scale: 0.9 }}
            className={`
              flex items-center space-x-2 p-3 rounded-lg border
              ${priceData.change_percentage > 0 
                ? 'bg-red-50 border-red-200 text-red-700' 
                : 'bg-green-50 border-green-200 text-green-700'
              }
            `}
          >
            {priceData.change_percentage > 0 ? (
              <TrendingUp className="h-4 w-4" />
            ) : (
              <TrendingDown className="h-4 w-4" />
            )}
            <div className="flex-1">
              <div className="font-medium text-sm">
                Price {priceData.change_percentage > 0 ? 'increased' : 'decreased'}
              </div>
              <div className="text-xs opacity-75">
                ${priceData.old_price} → ${priceData.new_price} 
                ({priceData.change_percentage > 0 ? '+' : ''}{priceData.change_percentage}%)
              </div>
            </div>
            {showUpdateAnimation && (
              <motion.div
                initial={{ scale: 0 }}
                animate={{ scale: [0, 1.5, 1] }}
                className="w-2 h-2 bg-current rounded-full"
              />
            )}
          </motion.div>
        )}
      </AnimatePresence>

      {/* Booking Updates */}
      <AnimatePresence>
        {showBookingUpdates && bookingData && (
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: 20 }}
            className="flex items-center space-x-2 p-3 bg-blue-50 border border-blue-200 rounded-lg text-blue-700"
          >
            <Users className="h-4 w-4" />
            <div className="flex-1">
              <div className="font-medium text-sm">
                {bookingData.recent_bookings} recent booking{bookingData.recent_bookings !== 1 ? 's' : ''}
              </div>
              <div className="text-xs opacity-75">
                {bookingData.total_bookings_today} bookings today
              </div>
            </div>
            {bookingData.trending_score > 70 && (
              <motion.div
                animate={{ rotate: [0, 10, -10, 0] }}
                transition={{ duration: 0.5, repeat: 2 }}
                className="text-orange-500"
              >
                <Zap className="h-4 w-4" />
              </motion.div>
            )}
          </motion.div>
        )}
      </AnimatePresence>

      {/* View Count Updates */}
      <AnimatePresence>
        {showViewCount && viewData && (
          <motion.div
            initial={{ opacity: 0, scale: 0.8 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.8 }}
            className="flex items-center space-x-2 p-3 bg-purple-50 border border-purple-200 rounded-lg text-purple-700"
          >
            <Eye className="h-4 w-4" />
            <div className="flex-1">
              <div className="font-medium text-sm">
                {viewData.current_viewers} viewing now
              </div>
              <div className="text-xs opacity-75">
                {viewData.total_views_today} views today • #{viewData.popularity_rank} trending
              </div>
            </div>
            <motion.div
              animate={{ 
                scale: [1, 1.2, 1],
                opacity: [0.5, 1, 0.5]
              }}
              transition={{ 
                duration: 2, 
                repeat: Infinity,
                repeatType: 'loop'
              }}
              className="w-2 h-2 bg-purple-500 rounded-full"
            />
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}

// Connection Status Component
interface ConnectionStatusProps {
  className?: string
}

export function ConnectionStatus({ className = '' }: ConnectionStatusProps) {
  const [isConnected, setIsConnected] = useState(false)
  const [stats, setStats] = useState<any>(null)

  useEffect(() => {
    const checkConnection = () => {
      setIsConnected(RealTimeService.isConnectedStatus())
      setStats(RealTimeService.getUpdateStats())
    }

    checkConnection()
    const interval = setInterval(checkConnection, 5000)

    return () => clearInterval(interval)
  }, [])

  return (
    <div className={`flex items-center space-x-3 ${className}`}>
      <div className="flex items-center space-x-2">
        {isConnected ? (
          <Wifi className="h-4 w-4 text-green-500" />
        ) : (
          <WifiOff className="h-4 w-4 text-red-500" />
        )}
        <span className={`text-sm font-medium ${
          isConnected ? 'text-green-700' : 'text-red-700'
        }`}>
          {isConnected ? 'Connected' : 'Disconnected'}
        </span>
      </div>

      {stats && isConnected && (
        <div className="flex items-center space-x-4 text-xs text-gray-500">
          <div className="flex items-center space-x-1">
            <Clock className="h-3 w-3" />
            <span>{stats.totalUpdates} updates</span>
          </div>
          <div>
            Avg: {(stats.averageUpdateInterval / 1000).toFixed(1)}s
          </div>
        </div>
      )}
    </div>
  )
}

// Live Activity Feed Component
interface LiveActivityFeedProps {
  destinations: Destination[]
  maxItems?: number
  className?: string
}

export function LiveActivityFeed({ 
  destinations, 
  maxItems = 10, 
  className = '' 
}: LiveActivityFeedProps) {
  const [activities, setActivities] = useState<RealTimeUpdate[]>([])

  useEffect(() => {
    RealTimeService.connect(destinations)

    const unsubscribe = RealTimeService.addListener((update: RealTimeUpdate) => {
      setActivities(prev => {
        const newActivities = [update, ...prev].slice(0, maxItems)
        return newActivities
      })
    })

    return () => {
      unsubscribe()
    }
  }, [destinations, maxItems])

  const getActivityIcon = (type: string) => {
    switch (type) {
      case 'price':
        return <TrendingUp className="h-4 w-4 text-blue-500" />
      case 'booking':
        return <Users className="h-4 w-4 text-green-500" />
      case 'view_count':
        return <Eye className="h-4 w-4 text-purple-500" />
      default:
        return <Zap className="h-4 w-4 text-gray-500" />
    }
  }

  const getActivityMessage = (update: RealTimeUpdate): string => {
    const destination = destinations.find(d => d.id.toString() === update.destination_id)
    const destinationName = destination?.name || 'Unknown destination'

    switch (update.type) {
      case 'price':
        const priceData = update.data as PriceUpdate
        return `${destinationName} price ${priceData.change_percentage > 0 ? 'increased' : 'decreased'} by ${Math.abs(priceData.change_percentage)}%`
      case 'booking':
        const bookingData = update.data as BookingUpdate
        return `${bookingData.recent_bookings} new booking${bookingData.recent_bookings !== 1 ? 's' : ''} for ${destinationName}`
      case 'view_count':
        const viewData = update.data as ViewCountUpdate
        return `${viewData.current_viewers} people viewing ${destinationName}`
      default:
        return `Activity on ${destinationName}`
    }
  }

  return (
    <div className={`space-y-3 ${className}`}>
      <h3 className="font-semibold text-gray-900 flex items-center space-x-2">
        <Zap className="h-4 w-4 text-yellow-500" />
        <span>Live Activity</span>
      </h3>

      <div className="space-y-2 max-h-64 overflow-y-auto">
        <AnimatePresence>
          {activities.map((activity, index) => (
            <motion.div
              key={`${activity.destination_id}-${activity.timestamp}-${index}`}
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: 20 }}
              transition={{ delay: index * 0.05 }}
              className="flex items-center space-x-3 p-2 bg-gray-50 rounded-lg"
            >
              {getActivityIcon(activity.type)}
              <div className="flex-1 min-w-0">
                <p className="text-sm text-gray-700 truncate">
                  {getActivityMessage(activity)}
                </p>
                <p className="text-xs text-gray-500">
                  {formatTimeAgo(activity.timestamp)}
                </p>
              </div>
            </motion.div>
          ))}
        </AnimatePresence>

        {activities.length === 0 && (
          <div className="text-center py-8 text-gray-500">
            <Zap className="h-8 w-8 mx-auto mb-2 opacity-50" />
            <p className="text-sm">Waiting for live updates...</p>
          </div>
        )}
      </div>
    </div>
  )
}

function formatTimeAgo(timestamp: string): string {
  const now = new Date()
  const updateTime = new Date(timestamp)
  const diffMs = now.getTime() - updateTime.getTime()
  const diffSeconds = Math.floor(diffMs / 1000)
  const diffMinutes = Math.floor(diffSeconds / 60)

  if (diffSeconds < 60) {
    return `${diffSeconds}s ago`
  } else if (diffMinutes < 60) {
    return `${diffMinutes}m ago`
  } else {
    return updateTime.toLocaleTimeString()
  }
}
