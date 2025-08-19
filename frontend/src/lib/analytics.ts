/**
 * User Behavior Analytics and Tracking System
 * 
 * Tracks user interactions, preferences, and behavior patterns for personalization
 */

export interface UserEvent {
  id: string
  user_id: string
  session_id: string
  event_type: string
  event_data: Record<string, any>
  timestamp: string
  page_url: string
  user_agent: string
  ip_address?: string
}

export interface PageView {
  page: string
  duration: number
  scroll_depth: number
  interactions: string[]
  referrer: string
  timestamp: string
}

export interface SearchEvent {
  query: string
  filters: Record<string, any>
  results_count: number
  results_clicked: string[]
  time_to_first_click: number
  session_duration: number
}

export interface DestinationInteraction {
  destination_id: string
  interaction_type: 'view' | 'wishlist_add' | 'wishlist_remove' | 'share' | 'compare' | 'book'
  duration?: number
  context: Record<string, any>
}

export interface UserPreferenceSignal {
  signal_type: 'explicit' | 'implicit'
  category: string
  value: any
  confidence: number
  source: string
}

export interface BehaviorPattern {
  pattern_id: string
  pattern_type: 'search' | 'browse' | 'booking' | 'engagement'
  frequency: number
  last_occurrence: string
  trend: 'increasing' | 'stable' | 'decreasing'
  seasonal_variation: number
}

export interface UserSegment {
  segment_id: string
  name: string
  description: string
  criteria: Record<string, any>
  user_count: number
  characteristics: string[]
}

class AnalyticsService {
  private static readonly STORAGE_KEY = 'user_analytics'
  private static readonly SESSION_KEY = 'analytics_session'
  private static readonly BATCH_SIZE = 10
  private static eventQueue: UserEvent[] = []
  private static sessionId: string = ''
  private static userId: string = ''

  // Initialize analytics
  static initialize(userId: string): void {
    this.userId = userId
    this.sessionId = this.generateSessionId()
    this.startSession()
    
    // Set up periodic batch sending
    setInterval(() => {
      this.flushEventQueue()
    }, 30000) // Send events every 30 seconds

    // Set up page visibility change tracking
    document.addEventListener('visibilitychange', () => {
      if (document.hidden) {
        this.flushEventQueue()
      }
    })

    // Set up beforeunload tracking
    window.addEventListener('beforeunload', () => {
      this.flushEventQueue()
    })
  }

  // Track user events
  static trackEvent(eventType: string, eventData: Record<string, any> = {}): void {
    const event: UserEvent = {
      id: this.generateEventId(),
      user_id: this.userId,
      session_id: this.sessionId,
      event_type: eventType,
      event_data: eventData,
      timestamp: new Date().toISOString(),
      page_url: window.location.href,
      user_agent: navigator.userAgent
    }

    this.eventQueue.push(event)

    // Flush queue if it reaches batch size
    if (this.eventQueue.length >= this.BATCH_SIZE) {
      this.flushEventQueue()
    }

    // Store locally for offline capability
    this.storeEventLocally(event)
  }

  // Track page views
  static trackPageView(page: string, additionalData: Record<string, any> = {}): void {
    const pageViewData = {
      page,
      referrer: document.referrer,
      timestamp: new Date().toISOString(),
      ...additionalData
    }

    this.trackEvent('page_view', pageViewData)

    // Start tracking page engagement
    this.startPageEngagementTracking(page)
  }

  // Track search events
  static trackSearch(query: string, filters: Record<string, any>, resultsCount: number): void {
    this.trackEvent('search', {
      query,
      filters,
      results_count: resultsCount,
      search_timestamp: new Date().toISOString()
    })
  }

  // Track destination interactions
  static trackDestinationInteraction(destinationId: string, interactionType: string, context: Record<string, any> = {}): void {
    this.trackEvent('destination_interaction', {
      destination_id: destinationId,
      interaction_type: interactionType,
      context,
      timestamp: new Date().toISOString()
    })
  }

  // Track user preferences (explicit)
  static trackPreferenceChange(category: string, value: any, source: string = 'user_input'): void {
    this.trackEvent('preference_change', {
      signal_type: 'explicit',
      category,
      value,
      confidence: 1.0,
      source,
      timestamp: new Date().toISOString()
    })
  }

  // Track implicit preference signals
  static trackImplicitPreference(category: string, value: any, confidence: number, source: string): void {
    this.trackEvent('preference_signal', {
      signal_type: 'implicit',
      category,
      value,
      confidence,
      source,
      timestamp: new Date().toISOString()
    })
  }

  // Track booking funnel
  static trackBookingStep(step: string, destinationId: string, additionalData: Record<string, any> = {}): void {
    this.trackEvent('booking_funnel', {
      step,
      destination_id: destinationId,
      funnel_timestamp: new Date().toISOString(),
      ...additionalData
    })
  }

  // Track errors and issues
  static trackError(errorType: string, errorMessage: string, context: Record<string, any> = {}): void {
    this.trackEvent('error', {
      error_type: errorType,
      error_message: errorMessage,
      context,
      timestamp: new Date().toISOString()
    })
  }

  // Get user behavior patterns
  static async getBehaviorPatterns(userId: string): Promise<BehaviorPattern[]> {
    try {
      const events = this.getStoredEvents(userId)
      return this.analyzeBehaviorPatterns(events)
    } catch (error) {
      console.error('Error getting behavior patterns:', error)
      return []
    }
  }

  // Get user segment
  static async getUserSegment(userId: string): Promise<UserSegment | null> {
    try {
      const patterns = await this.getBehaviorPatterns(userId)
      return this.determineUserSegment(patterns)
    } catch (error) {
      console.error('Error determining user segment:', error)
      return null
    }
  }

  // Get analytics insights
  static async getAnalyticsInsights(userId: string): Promise<{
    total_events: number
    session_count: number
    avg_session_duration: number
    top_pages: Array<{ page: string; views: number }>
    search_patterns: Array<{ query: string; frequency: number }>
    engagement_score: number
  }> {
    try {
      const events = this.getStoredEvents(userId)
      return this.generateInsights(events)
    } catch (error) {
      console.error('Error generating insights:', error)
      return {
        total_events: 0,
        session_count: 0,
        avg_session_duration: 0,
        top_pages: [],
        search_patterns: [],
        engagement_score: 0
      }
    }
  }

  // Private methods
  private static generateSessionId(): string {
    return `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  private static generateEventId(): string {
    return `event_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  private static startSession(): void {
    this.trackEvent('session_start', {
      session_id: this.sessionId,
      user_agent: navigator.userAgent,
      screen_resolution: `${screen.width}x${screen.height}`,
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      language: navigator.language
    })
  }

  private static startPageEngagementTracking(page: string): void {
    const startTime = Date.now()
    let maxScrollDepth = 0
    const interactions: string[] = []

    // Track scroll depth
    const handleScroll = () => {
      const scrollDepth = Math.round((window.scrollY / (document.body.scrollHeight - window.innerHeight)) * 100)
      maxScrollDepth = Math.max(maxScrollDepth, scrollDepth)
    }

    // Track interactions
    const handleInteraction = (event: Event) => {
      interactions.push(event.type)
    }

    window.addEventListener('scroll', handleScroll)
    document.addEventListener('click', handleInteraction)
    document.addEventListener('keydown', handleInteraction)

    // Clean up and send data when leaving page
    const cleanup = () => {
      const duration = Date.now() - startTime
      
      this.trackEvent('page_engagement', {
        page,
        duration,
        scroll_depth: maxScrollDepth,
        interactions: Array.from(new Set(interactions)), // Remove duplicates
        interaction_count: interactions.length
      })

      window.removeEventListener('scroll', handleScroll)
      document.removeEventListener('click', handleInteraction)
      document.removeEventListener('keydown', handleInteraction)
    }

    // Set up cleanup
    window.addEventListener('beforeunload', cleanup)
    document.addEventListener('visibilitychange', () => {
      if (document.hidden) cleanup()
    })
  }

  private static storeEventLocally(event: UserEvent): void {
    try {
      const stored = JSON.parse(localStorage.getItem(this.STORAGE_KEY) || '[]')
      stored.push(event)
      
      // Keep only last 1000 events to prevent storage bloat
      if (stored.length > 1000) {
        stored.splice(0, stored.length - 1000)
      }
      
      localStorage.setItem(this.STORAGE_KEY, JSON.stringify(stored))
    } catch (error) {
      console.error('Error storing event locally:', error)
    }
  }

  private static getStoredEvents(userId: string): UserEvent[] {
    try {
      const stored = JSON.parse(localStorage.getItem(this.STORAGE_KEY) || '[]')
      return stored.filter((event: UserEvent) => event.user_id === userId)
    } catch (error) {
      console.error('Error getting stored events:', error)
      return []
    }
  }

  private static flushEventQueue(): void {
    if (this.eventQueue.length === 0) return

    // In a real implementation, this would send events to analytics server
    console.log('Sending analytics events:', this.eventQueue)
    
    // Simulate API call
    this.sendEventsToServer(this.eventQueue)
    
    // Clear queue
    this.eventQueue = []
  }

  private static async sendEventsToServer(events: UserEvent[]): Promise<void> {
    try {
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 100))
      console.log(`Sent ${events.length} events to analytics server`)
    } catch (error) {
      console.error('Error sending events to server:', error)
      // Re-queue events for retry
      this.eventQueue.unshift(...events)
    }
  }

  private static analyzeBehaviorPatterns(events: UserEvent[]): BehaviorPattern[] {
    const patterns: BehaviorPattern[] = []

    // Analyze search patterns
    const searchEvents = events.filter(e => e.event_type === 'search')
    if (searchEvents.length > 0) {
      patterns.push({
        pattern_id: 'search_frequency',
        pattern_type: 'search',
        frequency: searchEvents.length,
        last_occurrence: searchEvents[searchEvents.length - 1]?.timestamp || '',
        trend: 'stable',
        seasonal_variation: 0
      })
    }

    // Analyze browsing patterns
    const pageViews = events.filter(e => e.event_type === 'page_view')
    if (pageViews.length > 0) {
      patterns.push({
        pattern_id: 'browse_frequency',
        pattern_type: 'browse',
        frequency: pageViews.length,
        last_occurrence: pageViews[pageViews.length - 1]?.timestamp || '',
        trend: 'stable',
        seasonal_variation: 0
      })
    }

    return patterns
  }

  private static determineUserSegment(patterns: BehaviorPattern[]): UserSegment | null {
    // Simplified segmentation logic
    const searchPattern = patterns.find(p => p.pattern_type === 'search')
    const browsePattern = patterns.find(p => p.pattern_type === 'browse')

    if (searchPattern && searchPattern.frequency > 10) {
      return {
        segment_id: 'active_searcher',
        name: 'Active Searcher',
        description: 'Users who frequently search for destinations',
        criteria: { search_frequency: '>10' },
        user_count: 1250,
        characteristics: ['High search activity', 'Research-oriented', 'Comparison shopping']
      }
    }

    if (browsePattern && browsePattern.frequency > 20) {
      return {
        segment_id: 'active_browser',
        name: 'Active Browser',
        description: 'Users who frequently browse destinations',
        criteria: { browse_frequency: '>20' },
        user_count: 890,
        characteristics: ['High engagement', 'Exploration-focused', 'Visual learner']
      }
    }

    return null
  }

  private static generateInsights(events: UserEvent[]): any {
    const sessions = Array.from(new Set(events.map(e => e.session_id)))
    const pageViews = events.filter(e => e.event_type === 'page_view')
    const searches = events.filter(e => e.event_type === 'search')

    // Calculate page view counts
    const pageViewCounts = pageViews.reduce((acc, event) => {
      const page = event.event_data.page || 'unknown'
      acc[page] = (acc[page] || 0) + 1
      return acc
    }, {} as Record<string, number>)

    const topPages = Object.entries(pageViewCounts)
      .sort(([,a], [,b]) => b - a)
      .slice(0, 5)
      .map(([page, views]) => ({ page, views }))

    // Calculate search patterns
    const searchCounts = searches.reduce((acc, event) => {
      const query = event.event_data.query || 'unknown'
      acc[query] = (acc[query] || 0) + 1
      return acc
    }, {} as Record<string, number>)

    const searchPatterns = Object.entries(searchCounts)
      .sort(([,a], [,b]) => b - a)
      .slice(0, 5)
      .map(([query, frequency]) => ({ query, frequency }))

    return {
      total_events: events.length,
      session_count: sessions.length,
      avg_session_duration: 0, // Would calculate from session events
      top_pages: topPages,
      search_patterns: searchPatterns,
      engagement_score: Math.min(100, events.length * 2) // Simplified score
    }
  }

  // Utility methods for common tracking scenarios
  static trackDestinationView(destinationId: string, source: string = 'browse'): void {
    this.trackDestinationInteraction(destinationId, 'view', { source })
  }

  static trackWishlistAdd(destinationId: string): void {
    this.trackDestinationInteraction(destinationId, 'wishlist_add')
  }

  static trackWishlistRemove(destinationId: string): void {
    this.trackDestinationInteraction(destinationId, 'wishlist_remove')
  }

  static trackDestinationShare(destinationId: string, platform: string): void {
    this.trackDestinationInteraction(destinationId, 'share', { platform })
  }

  static trackDestinationCompare(destinationIds: string[]): void {
    this.trackEvent('destination_compare', {
      destination_ids: destinationIds,
      comparison_count: destinationIds.length
    })
  }

  static trackFilterUsage(filterType: string, filterValue: any): void {
    this.trackEvent('filter_usage', {
      filter_type: filterType,
      filter_value: filterValue
    })
  }

  static trackRecommendationClick(destinationId: string, recommendationType: string, position: number): void {
    this.trackEvent('recommendation_click', {
      destination_id: destinationId,
      recommendation_type: recommendationType,
      position,
      timestamp: new Date().toISOString()
    })
  }
}

export { AnalyticsService }
