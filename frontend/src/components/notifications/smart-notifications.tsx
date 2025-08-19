'use client'

import React, { useState, useEffect } from 'react'
import { Bell, X, Check, Star, TrendingUp, DollarSign, MapPin, Calendar, Zap, Brain, Heart, AlertCircle } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { Button } from '@/components/ui/button'
import { FadeIn } from '@/components/ui/animated'

interface SmartNotification {
  id: string
  type: 'recommendation' | 'price_alert' | 'trending' | 'personalized_offer' | 'booking_reminder' | 'travel_tip'
  title: string
  message: string
  action_text?: string
  action_url?: string
  priority: 'low' | 'medium' | 'high'
  timestamp: string
  read: boolean
  data?: Record<string, any>
  expires_at?: string
}

interface NotificationPreferences {
  recommendations: boolean
  price_alerts: boolean
  trending_destinations: boolean
  personalized_offers: boolean
  booking_reminders: boolean
  travel_tips: boolean
  frequency: 'immediate' | 'daily' | 'weekly'
  quiet_hours: { start: string; end: string }
}

interface SmartNotificationsProps {
  userId: string
  className?: string
}

export default function SmartNotifications({ userId, className = '' }: SmartNotificationsProps) {
  const [notifications, setNotifications] = useState<SmartNotification[]>([])
  const [showNotifications, setShowNotifications] = useState(false)
  const [preferences, setPreferences] = useState<NotificationPreferences | null>(null)
  const [unreadCount, setUnreadCount] = useState(0)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadNotifications()
    loadPreferences()
    
    // Set up real-time notification updates
    const interval = setInterval(checkForNewNotifications, 30000) // Check every 30 seconds
    
    return () => clearInterval(interval)
  }, [userId])

  useEffect(() => {
    const unread = notifications.filter(n => !n.read).length
    setUnreadCount(unread)
  }, [notifications])

  const loadNotifications = async () => {
    setLoading(true)
    try {
      // In a real app, this would fetch from API
      const mockNotifications = generateMockNotifications()
      setNotifications(mockNotifications)
    } catch (error) {
      console.error('Error loading notifications:', error)
    } finally {
      setLoading(false)
    }
  }

  const loadPreferences = async () => {
    try {
      const stored = localStorage.getItem(`notification_preferences_${userId}`)
      if (stored) {
        setPreferences(JSON.parse(stored))
      } else {
        setPreferences(getDefaultPreferences())
      }
    } catch (error) {
      console.error('Error loading preferences:', error)
      setPreferences(getDefaultPreferences())
    }
  }

  const checkForNewNotifications = async () => {
    // Simulate checking for new notifications
    const hasNewNotifications = Math.random() > 0.8 // 20% chance of new notification
    
    if (hasNewNotifications) {
      const newNotification = generatePersonalizedNotification()
      setNotifications(prev => [newNotification, ...prev])
      
      // Show browser notification if permission granted
      if (Notification.permission === 'granted') {
        new Notification(newNotification.title, {
          body: newNotification.message,
          icon: '/favicon.ico',
          tag: newNotification.id
        })
      }
    }
  }

  const generateMockNotifications = (): SmartNotification[] => [
    {
      id: '1',
      type: 'recommendation',
      title: 'Perfect Match Found!',
      message: 'Based on your preferences, we found 3 destinations that are 95% match for your travel style.',
      action_text: 'View Recommendations',
      action_url: '/recommendations',
      priority: 'high',
      timestamp: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(), // 2 hours ago
      read: false,
      data: { destination_count: 3, match_score: 0.95 }
    },
    {
      id: '2',
      type: 'price_alert',
      title: 'Price Drop Alert',
      message: 'Santorini, Greece dropped by 15% - now $2,125. Perfect time to book!',
      action_text: 'Book Now',
      action_url: '/destinations/santorini',
      priority: 'high',
      timestamp: new Date(Date.now() - 4 * 60 * 60 * 1000).toISOString(), // 4 hours ago
      read: false,
      data: { destination_id: 'santorini', old_price: 2500, new_price: 2125, discount: 15 }
    },
    {
      id: '3',
      type: 'trending',
      title: 'Trending Destination',
      message: 'Iceland is trending among travelers with your profile. 89% satisfaction rate.',
      action_text: 'Explore Iceland',
      action_url: '/destinations/iceland',
      priority: 'medium',
      timestamp: new Date(Date.now() - 6 * 60 * 60 * 1000).toISOString(), // 6 hours ago
      read: true,
      data: { destination_id: 'iceland', satisfaction_rate: 0.89, trend_score: 0.92 }
    },
    {
      id: '4',
      type: 'personalized_offer',
      title: 'Exclusive Offer for You',
      message: 'Get 20% off your next cultural destination booking. Offer expires in 48 hours.',
      action_text: 'Claim Offer',
      action_url: '/offers/cultural-20',
      priority: 'medium',
      timestamp: new Date(Date.now() - 12 * 60 * 60 * 1000).toISOString(), // 12 hours ago
      read: false,
      data: { discount: 20, category: 'cultural', expires_in_hours: 48 },
      expires_at: new Date(Date.now() + 48 * 60 * 60 * 1000).toISOString()
    },
    {
      id: '5',
      type: 'travel_tip',
      title: 'Smart Travel Tip',
      message: 'Book flights 6-8 weeks in advance for your preferred destinations to save up to 25%.',
      priority: 'low',
      timestamp: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(), // 1 day ago
      read: true,
      data: { tip_category: 'booking_timing', potential_savings: 25 }
    }
  ]

  const generatePersonalizedNotification = (): SmartNotification => {
    const types: SmartNotification['type'][] = ['recommendation', 'price_alert', 'trending', 'travel_tip']
    const type = types[Math.floor(Math.random() * types.length)]
    
    const notifications = {
      recommendation: {
        title: 'New Recommendations Available',
        message: 'We found 2 new destinations that match your recent searches.',
        action_text: 'View Now',
        priority: 'medium' as const
      },
      price_alert: {
        title: 'Price Alert',
        message: 'A destination in your wishlist just dropped in price by 12%.',
        action_text: 'Check Price',
        priority: 'high' as const
      },
      trending: {
        title: 'Trending Now',
        message: 'Morocco is gaining popularity among adventure travelers.',
        action_text: 'Explore',
        priority: 'low' as const
      },
      travel_tip: {
        title: 'Travel Tip',
        message: 'Consider traveling in shoulder season for better prices and fewer crowds.',
        action_text: 'Learn More',
        priority: 'low' as const
      },
      booking_reminder: {
        title: 'Booking Reminder',
        message: 'Don\'t forget to complete your booking for Santorini, Greece.',
        action_text: 'Complete Booking',
        priority: 'high' as const
      },
      personalized_offer: {
        title: 'Special Offer',
        message: 'Exclusive 15% discount on your next adventure booking.',
        action_text: 'View Offer',
        priority: 'medium' as const
      }
    }

    const notification = notifications[type]
    
    return {
      id: `notif_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      type,
      title: notification.title,
      message: notification.message,
      action_text: notification.action_text,
      priority: notification.priority,
      timestamp: new Date().toISOString(),
      read: false
    }
  }

  const getDefaultPreferences = (): NotificationPreferences => ({
    recommendations: true,
    price_alerts: true,
    trending_destinations: false,
    personalized_offers: true,
    booking_reminders: true,
    travel_tips: false,
    frequency: 'daily',
    quiet_hours: { start: '22:00', end: '08:00' }
  })

  const markAsRead = (notificationId: string) => {
    setNotifications(prev => 
      prev.map(n => n.id === notificationId ? { ...n, read: true } : n)
    )
  }

  const markAllAsRead = () => {
    setNotifications(prev => prev.map(n => ({ ...n, read: true })))
  }

  const deleteNotification = (notificationId: string) => {
    setNotifications(prev => prev.filter(n => n.id !== notificationId))
  }

  const getNotificationIcon = (type: SmartNotification['type']) => {
    switch (type) {
      case 'recommendation':
        return <Brain className="h-5 w-5" />
      case 'price_alert':
        return <DollarSign className="h-5 w-5" />
      case 'trending':
        return <TrendingUp className="h-5 w-5" />
      case 'personalized_offer':
        return <Star className="h-5 w-5" />
      case 'booking_reminder':
        return <Calendar className="h-5 w-5" />
      case 'travel_tip':
        return <Zap className="h-5 w-5" />
      default:
        return <Bell className="h-5 w-5" />
    }
  }

  const getNotificationColor = (type: SmartNotification['type'], priority: SmartNotification['priority']) => {
    if (priority === 'high') return 'text-red-600 bg-red-100'
    if (priority === 'medium') return 'text-yellow-600 bg-yellow-100'
    
    switch (type) {
      case 'recommendation':
        return 'text-blue-600 bg-blue-100'
      case 'price_alert':
        return 'text-green-600 bg-green-100'
      case 'trending':
        return 'text-purple-600 bg-purple-100'
      case 'personalized_offer':
        return 'text-orange-600 bg-orange-100'
      default:
        return 'text-gray-600 bg-gray-100'
    }
  }

  const formatTimeAgo = (timestamp: string): string => {
    const now = new Date()
    const time = new Date(timestamp)
    const diffMs = now.getTime() - time.getTime()
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60))
    const diffDays = Math.floor(diffHours / 24)

    if (diffHours < 1) return 'Just now'
    if (diffHours < 24) return `${diffHours}h ago`
    if (diffDays < 7) return `${diffDays}d ago`
    return time.toLocaleDateString()
  }

  const requestNotificationPermission = async () => {
    if ('Notification' in window && Notification.permission === 'default') {
      await Notification.requestPermission()
    }
  }

  return (
    <div className={`relative ${className}`}>
      {/* Notification Bell */}
      <button
        onClick={() => {
          setShowNotifications(!showNotifications)
          requestNotificationPermission()
        }}
        className="relative p-2 text-gray-600 hover:text-gray-900 transition-colors"
      >
        <Bell className="h-6 w-6" />
        {unreadCount > 0 && (
          <motion.span
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            className="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full h-5 w-5 flex items-center justify-center font-medium"
          >
            {unreadCount > 9 ? '9+' : unreadCount}
          </motion.span>
        )}
      </button>

      {/* Notifications Dropdown */}
      <AnimatePresence>
        {showNotifications && (
          <motion.div
            initial={{ opacity: 0, y: -10, scale: 0.95 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: -10, scale: 0.95 }}
            className="absolute top-full right-0 mt-2 w-96 bg-white rounded-lg shadow-xl border border-gray-200 z-50 max-h-96 overflow-hidden"
          >
            {/* Header */}
            <div className="p-4 border-b border-gray-200">
              <div className="flex items-center justify-between">
                <h3 className="text-lg font-semibold text-gray-900">Notifications</h3>
                <div className="flex items-center space-x-2">
                  {unreadCount > 0 && (
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={markAllAsRead}
                      className="text-xs"
                    >
                      Mark all read
                    </Button>
                  )}
                  <button
                    onClick={() => setShowNotifications(false)}
                    className="text-gray-400 hover:text-gray-600"
                  >
                    <X className="h-5 w-5" />
                  </button>
                </div>
              </div>
            </div>

            {/* Notifications List */}
            <div className="max-h-80 overflow-y-auto">
              {loading ? (
                <div className="p-8 text-center">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-brand-500 mx-auto mb-2"></div>
                  <p className="text-gray-600">Loading notifications...</p>
                </div>
              ) : notifications.length === 0 ? (
                <div className="p-8 text-center">
                  <Bell className="h-12 w-12 text-gray-300 mx-auto mb-3" />
                  <p className="text-gray-600">No notifications yet</p>
                  <p className="text-sm text-gray-500">We'll notify you about personalized recommendations and deals</p>
                </div>
              ) : (
                <div className="divide-y divide-gray-100">
                  {notifications.map((notification) => (
                    <motion.div
                      key={notification.id}
                      initial={{ opacity: 0, x: -20 }}
                      animate={{ opacity: 1, x: 0 }}
                      className={`p-4 hover:bg-gray-50 transition-colors ${
                        !notification.read ? 'bg-blue-50' : ''
                      }`}
                    >
                      <div className="flex items-start space-x-3">
                        <div className={`p-2 rounded-lg ${getNotificationColor(notification.type, notification.priority)}`}>
                          {getNotificationIcon(notification.type)}
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center justify-between mb-1">
                            <h4 className={`text-sm font-medium ${!notification.read ? 'text-gray-900' : 'text-gray-700'}`}>
                              {notification.title}
                            </h4>
                            <button
                              onClick={() => deleteNotification(notification.id)}
                              className="text-gray-400 hover:text-gray-600"
                            >
                              <X className="h-4 w-4" />
                            </button>
                          </div>
                          <p className="text-sm text-gray-600 mb-2">{notification.message}</p>
                          <div className="flex items-center justify-between">
                            <span className="text-xs text-gray-500">
                              {formatTimeAgo(notification.timestamp)}
                            </span>
                            <div className="flex items-center space-x-2">
                              {!notification.read && (
                                <button
                                  onClick={() => markAsRead(notification.id)}
                                  className="text-xs text-blue-600 hover:text-blue-800"
                                >
                                  Mark read
                                </button>
                              )}
                              {notification.action_text && (
                                <Button
                                  variant="outline"
                                  size="sm"
                                  className="text-xs"
                                  onClick={() => {
                                    markAsRead(notification.id)
                                    if (notification.action_url) {
                                      window.location.href = notification.action_url
                                    }
                                  }}
                                >
                                  {notification.action_text}
                                </Button>
                              )}
                            </div>
                          </div>
                        </div>
                      </div>
                    </motion.div>
                  ))}
                </div>
              )}
            </div>

            {/* Footer */}
            {notifications.length > 0 && (
              <div className="p-3 border-t border-gray-200 bg-gray-50">
                <Button
                  variant="ghost"
                  size="sm"
                  className="w-full text-sm"
                  onClick={() => {
                    setShowNotifications(false)
                    // Navigate to notifications page
                  }}
                >
                  View All Notifications
                </Button>
              </div>
            )}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}
