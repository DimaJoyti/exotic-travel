/**
 * AI-Powered Recommendation Engine for Exotic Travel Booking Platform
 * 
 * Provides personalized destination recommendations using machine learning algorithms
 * and user behavior analysis
 */

import { Destination, User } from '@/types'
import { DestinationsService } from './destinations'

export interface UserPreferences {
  budget_range: [number, number]
  preferred_duration: number[]
  travel_style: 'luxury' | 'adventure' | 'cultural' | 'relaxation' | 'family' | 'romantic'
  activity_preferences: string[]
  climate_preferences: 'tropical' | 'temperate' | 'cold' | 'desert' | 'any'
  accommodation_type: 'hotel' | 'resort' | 'villa' | 'apartment' | 'any'
  group_size: number
  accessibility_needs: string[]
  dietary_restrictions: string[]
  language_preferences: string[]
  previous_destinations: string[]
  interests: string[]
}

export interface UserBehavior {
  search_history: SearchEvent[]
  view_history: ViewEvent[]
  booking_history: BookingEvent[]
  wishlist_items: string[]
  time_spent_on_pages: Record<string, number>
  interaction_patterns: InteractionPattern[]
  seasonal_preferences: SeasonalPreference[]
}

export interface SearchEvent {
  query: string
  filters: Record<string, any>
  results_clicked: string[]
  timestamp: string
  session_id: string
}

export interface ViewEvent {
  destination_id: string
  duration: number
  scroll_depth: number
  interactions: string[]
  timestamp: string
  referrer: string
}

export interface BookingEvent {
  destination_id: string
  booking_date: string
  travel_date: string
  duration: number
  group_size: number
  total_amount: number
  booking_source: string
}

export interface InteractionPattern {
  pattern_type: 'scroll' | 'click' | 'hover' | 'search' | 'filter'
  frequency: number
  context: Record<string, any>
  effectiveness_score: number
}

export interface SeasonalPreference {
  season: 'spring' | 'summer' | 'fall' | 'winter'
  preference_score: number
  destinations: string[]
}

export interface RecommendationRequest {
  user_id: string
  preferences: UserPreferences
  behavior: UserBehavior
  context: RecommendationContext
  limit?: number
  exclude_visited?: boolean
}

export interface RecommendationContext {
  current_season: string
  user_location: { lat: number; lng: number; country: string }
  trending_destinations: string[]
  weather_conditions: Record<string, any>
  special_events: SpecialEvent[]
  market_conditions: MarketCondition[]
}

export interface SpecialEvent {
  name: string
  location: string
  start_date: string
  end_date: string
  category: string
  impact_score: number
}

export interface MarketCondition {
  destination_id: string
  demand_level: 'low' | 'medium' | 'high'
  price_trend: 'decreasing' | 'stable' | 'increasing'
  availability: number
  seasonal_factor: number
}

export interface RecommendationResult {
  destination: Destination
  confidence_score: number
  reasoning: RecommendationReason[]
  personalization_factors: PersonalizationFactor[]
  predicted_satisfaction: number
  optimal_booking_window: { start: string; end: string }
  estimated_total_cost: number
  similar_users_booked: number
}

export interface RecommendationReason {
  factor: string
  weight: number
  explanation: string
  evidence: any
}

export interface PersonalizationFactor {
  type: 'preference' | 'behavior' | 'context' | 'social'
  name: string
  value: any
  impact: number
  confidence: number
}

class RecommendationEngine {
  private static readonly CACHE_DURATION = 30 * 60 * 1000 // 30 minutes
  private static cache = new Map<string, { data: any; timestamp: number }>()

  // Main recommendation method
  static async getPersonalizedRecommendations(
    request: RecommendationRequest
  ): Promise<RecommendationResult[]> {
    const cacheKey = this.generateCacheKey(request)
    const cached = this.getFromCache(cacheKey)
    
    if (cached) {
      return cached
    }

    try {
      // Get all destinations
      const allDestinations = await DestinationsService.getDestinations({})
      
      // Apply ML-based scoring
      const scoredDestinations = await this.scoreDestinations(
        allDestinations,
        request
      )

      // Sort by confidence score and apply limit
      const recommendations = scoredDestinations
        .sort((a, b) => b.confidence_score - a.confidence_score)
        .slice(0, request.limit || 10)

      // Cache results
      this.setCache(cacheKey, recommendations)

      return recommendations
    } catch (error) {
      console.error('Error generating recommendations:', error)
      return this.getFallbackRecommendations(request)
    }
  }

  // Score destinations based on user preferences and behavior
  private static async scoreDestinations(
    destinations: Destination[],
    request: RecommendationRequest
  ): Promise<RecommendationResult[]> {
    const results: RecommendationResult[] = []

    for (const destination of destinations) {
      // Skip if user has already visited (if requested)
      if (request.exclude_visited && 
          request.preferences.previous_destinations.includes(destination.id.toString())) {
        continue
      }

      const score = await this.calculateDestinationScore(destination, request)
      
      if (score.confidence_score > 0.3) { // Minimum threshold
        results.push(score)
      }
    }

    return results
  }

  // Calculate comprehensive destination score
  private static async calculateDestinationScore(
    destination: Destination,
    request: RecommendationRequest
  ): Promise<RecommendationResult> {
    const factors: PersonalizationFactor[] = []
    const reasons: RecommendationReason[] = []
    let totalScore = 0
    let totalWeight = 0

    // Budget compatibility (25% weight)
    const budgetScore = this.calculateBudgetScore(destination, request.preferences)
    totalScore += budgetScore.score * 0.25
    totalWeight += 0.25
    factors.push(budgetScore.factor)
    if (budgetScore.score > 0.7) {
      reasons.push(budgetScore.reason)
    }

    // Duration compatibility (15% weight)
    const durationScore = this.calculateDurationScore(destination, request.preferences)
    totalScore += durationScore.score * 0.15
    totalWeight += 0.15
    factors.push(durationScore.factor)

    // Activity preferences (20% weight)
    const activityScore = this.calculateActivityScore(destination, request.preferences)
    totalScore += activityScore.score * 0.20
    totalWeight += 0.20
    factors.push(activityScore.factor)

    // Behavioral patterns (25% weight)
    const behaviorScore = this.calculateBehaviorScore(destination, request.behavior)
    totalScore += behaviorScore.score * 0.25
    totalWeight += 0.25
    factors.push(behaviorScore.factor)

    // Contextual factors (15% weight)
    const contextScore = this.calculateContextScore(destination, request.context)
    totalScore += contextScore.score * 0.15
    totalWeight += 0.15
    factors.push(contextScore.factor)

    const confidence_score = totalWeight > 0 ? totalScore / totalWeight : 0

    return {
      destination,
      confidence_score,
      reasoning: reasons,
      personalization_factors: factors,
      predicted_satisfaction: this.predictSatisfaction(confidence_score, factors),
      optimal_booking_window: this.calculateOptimalBookingWindow(destination, request.context),
      estimated_total_cost: this.estimateTotalCost(destination, request.preferences),
      similar_users_booked: this.getSimilarUsersCount(destination, request.preferences)
    }
  }

  // Calculate budget compatibility score
  private static calculateBudgetScore(
    destination: Destination,
    preferences: UserPreferences
  ): { score: number; factor: PersonalizationFactor; reason: RecommendationReason } {
    const [minBudget, maxBudget] = preferences.budget_range
    const price = destination.price
    
    let score = 0
    if (price >= minBudget && price <= maxBudget) {
      score = 1.0
    } else if (price < minBudget) {
      score = Math.max(0, 1 - (minBudget - price) / minBudget)
    } else {
      score = Math.max(0, 1 - (price - maxBudget) / maxBudget)
    }

    return {
      score,
      factor: {
        type: 'preference',
        name: 'Budget Compatibility',
        value: price,
        impact: score,
        confidence: 0.9
      },
      reason: {
        factor: 'budget',
        weight: 0.25,
        explanation: `Price of $${price} ${score > 0.8 ? 'fits perfectly' : score > 0.5 ? 'is close to' : 'is outside'} your budget range of $${minBudget}-$${maxBudget}`,
        evidence: { price, budget_range: preferences.budget_range, score }
      }
    }
  }

  // Calculate duration compatibility score
  private static calculateDurationScore(
    destination: Destination,
    preferences: UserPreferences
  ): { score: number; factor: PersonalizationFactor } {
    const preferredDurations = preferences.preferred_duration
    const destinationDuration = destination.duration

    let score = 0
    if (preferredDurations.includes(destinationDuration)) {
      score = 1.0
    } else {
      const closestDuration = preferredDurations.reduce((prev, curr) => 
        Math.abs(curr - destinationDuration) < Math.abs(prev - destinationDuration) ? curr : prev
      )
      const difference = Math.abs(closestDuration - destinationDuration)
      score = Math.max(0, 1 - difference / 7) // Penalize by days difference
    }

    return {
      score,
      factor: {
        type: 'preference',
        name: 'Duration Match',
        value: destinationDuration,
        impact: score,
        confidence: 0.8
      }
    }
  }

  // Calculate activity preferences score
  private static calculateActivityScore(
    destination: Destination,
    preferences: UserPreferences
  ): { score: number; factor: PersonalizationFactor } {
    const userActivities = preferences.activity_preferences
    const destinationFeatures = destination.features

    if (userActivities.length === 0) {
      return {
        score: 0.5,
        factor: {
          type: 'preference',
          name: 'Activity Match',
          value: 'No preferences set',
          impact: 0.5,
          confidence: 0.3
        }
      }
    }

    const matches = userActivities.filter(activity => 
      destinationFeatures.some(feature => 
        feature.toLowerCase().includes(activity.toLowerCase()) ||
        activity.toLowerCase().includes(feature.toLowerCase())
      )
    )

    const score = matches.length / userActivities.length

    return {
      score,
      factor: {
        type: 'preference',
        name: 'Activity Match',
        value: `${matches.length}/${userActivities.length} activities match`,
        impact: score,
        confidence: 0.85
      }
    }
  }

  // Calculate behavior-based score
  private static calculateBehaviorScore(
    destination: Destination,
    behavior: UserBehavior
  ): { score: number; factor: PersonalizationFactor } {
    let score = 0.5 // Base score

    // Check if user has viewed this destination before
    const hasViewed = behavior.view_history.some(view => view.destination_id === destination.id.toString())
    if (hasViewed) {
      score += 0.2
    }

    // Check if destination is in wishlist
    const inWishlist = behavior.wishlist_items.includes(destination.id.toString())
    if (inWishlist) {
      score += 0.3
    }

    // Analyze search patterns
    const relevantSearches = behavior.search_history.filter(search =>
      search.query.toLowerCase().includes(destination.country.toLowerCase()) ||
      search.query.toLowerCase().includes(destination.city.toLowerCase()) ||
      destination.features.some(feature => 
        search.query.toLowerCase().includes(feature.toLowerCase())
      )
    )

    if (relevantSearches.length > 0) {
      score += Math.min(0.2, relevantSearches.length * 0.05)
    }

    return {
      score: Math.min(1, score),
      factor: {
        type: 'behavior',
        name: 'User Behavior Match',
        value: `Viewed: ${hasViewed}, Wishlist: ${inWishlist}, Searches: ${relevantSearches.length}`,
        impact: score,
        confidence: 0.7
      }
    }
  }

  // Calculate contextual score
  private static calculateContextScore(
    destination: Destination,
    context: RecommendationContext
  ): { score: number; factor: PersonalizationFactor } {
    let score = 0.5

    // Check if destination is trending
    if (context.trending_destinations.includes(destination.id.toString())) {
      score += 0.2
    }

    // Seasonal appropriateness (simplified)
    const seasonalBonus = this.getSeasonalBonus(destination, context.current_season)
    score += seasonalBonus

    // Market conditions
    const marketCondition = context.market_conditions.find(mc => mc.destination_id === destination.id.toString())
    if (marketCondition) {
      if (marketCondition.demand_level === 'low') score += 0.1
      if (marketCondition.price_trend === 'decreasing') score += 0.1
    }

    return {
      score: Math.min(1, score),
      factor: {
        type: 'context',
        name: 'Market & Seasonal Factors',
        value: `Trending: ${context.trending_destinations.includes(destination.id.toString())}, Season: ${context.current_season}`,
        impact: score,
        confidence: 0.6
      }
    }
  }

  // Helper methods
  private static getSeasonalBonus(destination: Destination, season: string): number {
    // Simplified seasonal logic - in reality this would be more sophisticated
    const seasonalMap: Record<string, string[]> = {
      'spring': ['Europe', 'Japan', 'Turkey'],
      'summer': ['Mediterranean', 'Scandinavia', 'Canada'],
      'fall': ['New England', 'Germany', 'India'],
      'winter': ['Southeast Asia', 'Australia', 'Caribbean']
    }

    const seasonalDestinations = seasonalMap[season] || []
    const isSeasonalMatch = seasonalDestinations.some(region => 
      destination.country.includes(region) || destination.city.includes(region)
    )

    return isSeasonalMatch ? 0.15 : 0
  }

  private static predictSatisfaction(score: number, factors: PersonalizationFactor[]): number {
    // Simple satisfaction prediction based on confidence and factor diversity
    const factorDiversity = new Set(factors.map(f => f.type)).size / 4 // 4 possible types
    return Math.min(1, score * 0.8 + factorDiversity * 0.2)
  }

  private static calculateOptimalBookingWindow(
    destination: Destination,
    context: RecommendationContext
  ): { start: string; end: string } {
    // Simplified booking window calculation
    const now = new Date()
    const start = new Date(now.getTime() + 30 * 24 * 60 * 60 * 1000) // 30 days from now
    const end = new Date(now.getTime() + 90 * 24 * 60 * 60 * 1000) // 90 days from now

    return {
      start: start.toISOString().split('T')[0],
      end: end.toISOString().split('T')[0]
    }
  }

  private static estimateTotalCost(destination: Destination, preferences: UserPreferences): number {
    // Estimate total cost including flights, accommodation, activities
    const baseCost = destination.price * preferences.group_size
    const flightEstimate = baseCost * 0.3 // Rough flight cost estimate
    const extrasEstimate = baseCost * 0.2 // Activities, meals, etc.
    
    return Math.round(baseCost + flightEstimate + extrasEstimate)
  }

  private static getSimilarUsersCount(destination: Destination, preferences: UserPreferences): number {
    // Mock similar users count - in reality this would query user behavior data
    return Math.floor(Math.random() * 50) + 10
  }

  // Cache management
  private static generateCacheKey(request: RecommendationRequest): string {
    return `rec_${request.user_id}_${JSON.stringify(request.preferences).slice(0, 50)}`
  }

  private static getFromCache(key: string): any {
    const cached = this.cache.get(key)
    if (cached && Date.now() - cached.timestamp < this.CACHE_DURATION) {
      return cached.data
    }
    return null
  }

  private static setCache(key: string, data: any): void {
    this.cache.set(key, { data, timestamp: Date.now() })
  }

  // Fallback recommendations
  private static async getFallbackRecommendations(request: RecommendationRequest): Promise<RecommendationResult[]> {
    try {
      const destinations = await DestinationsService.getDestinations({})
      return destinations.slice(0, request.limit || 5).map(destination => ({
        destination,
        confidence_score: 0.5,
        reasoning: [{
          factor: 'fallback',
          weight: 1,
          explanation: 'Popular destination recommendation',
          evidence: {}
        }],
        personalization_factors: [],
        predicted_satisfaction: 0.6,
        optimal_booking_window: this.calculateOptimalBookingWindow(destination, request.context),
        estimated_total_cost: this.estimateTotalCost(destination, request.preferences),
        similar_users_booked: 25
      }))
    } catch (error) {
      console.error('Error getting fallback recommendations:', error)
      return []
    }
  }

  // User behavior tracking
  static trackUserBehavior(event: SearchEvent | ViewEvent | BookingEvent): void {
    // In a real implementation, this would send data to analytics service
    console.log('Tracking user behavior:', event)
  }

  // Get trending destinations
  static async getTrendingDestinations(limit: number = 10): Promise<string[]> {
    // Mock trending destinations - in reality this would analyze booking/view data
    const mockTrending = [
      'dest_1', 'dest_2', 'dest_3', 'dest_4', 'dest_5',
      'dest_6', 'dest_7', 'dest_8', 'dest_9', 'dest_10'
    ]
    return mockTrending.slice(0, limit)
  }
}

export { RecommendationEngine }
