/**
 * User Preferences Management System
 * 
 * Handles user preferences, travel profiles, and personalization settings
 */

import { UserPreferences } from './recommendations'

export interface TravelProfile {
  id: string
  user_id: string
  name: string
  description: string
  preferences: UserPreferences
  is_default: boolean
  created_at: string
  updated_at: string
}

export interface PersonalizationSettings {
  show_recommendations: boolean
  recommendation_frequency: 'high' | 'medium' | 'low'
  privacy_level: 'public' | 'friends' | 'private'
  data_sharing: {
    analytics: boolean
    marketing: boolean
    recommendations: boolean
  }
  notification_preferences: {
    new_recommendations: boolean
    price_alerts: boolean
    trending_destinations: boolean
    personalized_offers: boolean
  }
  display_preferences: {
    currency: string
    language: string
    date_format: string
    temperature_unit: 'celsius' | 'fahrenheit'
    distance_unit: 'metric' | 'imperial'
  }
}

export interface UserInterests {
  categories: string[]
  activities: string[]
  cuisines: string[]
  accommodation_types: string[]
  transportation_modes: string[]
  budget_categories: string[]
  travel_companions: string[]
  special_occasions: string[]
}

export interface TravelHistory {
  destinations_visited: string[]
  favorite_destinations: string[]
  travel_frequency: number
  average_trip_duration: number
  preferred_seasons: string[]
  booking_patterns: BookingPattern[]
  spending_patterns: SpendingPattern[]
}

export interface BookingPattern {
  advance_booking_days: number
  preferred_booking_time: string
  booking_channel: string
  decision_factors: string[]
  research_duration_days: number
}

export interface SpendingPattern {
  category: string
  average_amount: number
  percentage_of_budget: number
  seasonal_variation: number
}

class UserPreferencesService {
  private static readonly STORAGE_KEY = 'user_preferences'
  private static readonly PROFILES_KEY = 'travel_profiles'
  private static readonly SETTINGS_KEY = 'personalization_settings'
  private static readonly INTERESTS_KEY = 'user_interests'
  private static readonly HISTORY_KEY = 'travel_history'

  // User Preferences Management
  static async getUserPreferences(userId: string): Promise<UserPreferences> {
    try {
      // In a real app, this would fetch from API
      const stored = localStorage.getItem(`${this.STORAGE_KEY}_${userId}`)
      if (stored) {
        return JSON.parse(stored)
      }
      return this.getDefaultPreferences()
    } catch (error) {
      console.error('Error getting user preferences:', error)
      return this.getDefaultPreferences()
    }
  }

  static async updateUserPreferences(userId: string, preferences: UserPreferences): Promise<void> {
    try {
      localStorage.setItem(`${this.STORAGE_KEY}_${userId}`, JSON.stringify(preferences))
      
      // Simulate API call
      await new Promise(resolve => setTimeout(resolve, 500))
      
      console.log('User preferences updated:', preferences)
    } catch (error) {
      console.error('Error updating user preferences:', error)
      throw error
    }
  }

  static getDefaultPreferences(): UserPreferences {
    return {
      budget_range: [1000, 5000],
      preferred_duration: [7, 10, 14],
      travel_style: 'cultural',
      activity_preferences: ['sightseeing', 'local cuisine', 'cultural experiences'],
      climate_preferences: 'temperate',
      accommodation_type: 'hotel',
      group_size: 2,
      accessibility_needs: [],
      dietary_restrictions: [],
      language_preferences: ['English'],
      previous_destinations: [],
      interests: ['history', 'art', 'nature']
    }
  }

  // Travel Profiles Management
  static async getTravelProfiles(userId: string): Promise<TravelProfile[]> {
    try {
      const stored = localStorage.getItem(`${this.PROFILES_KEY}_${userId}`)
      if (stored) {
        return JSON.parse(stored)
      }
      return this.getDefaultProfiles(userId)
    } catch (error) {
      console.error('Error getting travel profiles:', error)
      return []
    }
  }

  static async createTravelProfile(userId: string, profile: Omit<TravelProfile, 'id' | 'user_id' | 'created_at' | 'updated_at'>): Promise<TravelProfile> {
    const newProfile: TravelProfile = {
      id: `profile_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      user_id: userId,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      ...profile
    }

    const profiles = await this.getTravelProfiles(userId)
    
    // If this is set as default, unset others
    if (newProfile.is_default) {
      profiles.forEach(p => p.is_default = false)
    }
    
    profiles.push(newProfile)
    localStorage.setItem(`${this.PROFILES_KEY}_${userId}`, JSON.stringify(profiles))
    
    return newProfile
  }

  static async updateTravelProfile(userId: string, profileId: string, updates: Partial<TravelProfile>): Promise<TravelProfile> {
    const profiles = await this.getTravelProfiles(userId)
    const profileIndex = profiles.findIndex(p => p.id === profileId)
    
    if (profileIndex === -1) {
      throw new Error('Profile not found')
    }

    // If setting as default, unset others
    if (updates.is_default) {
      profiles.forEach(p => p.is_default = false)
    }

    profiles[profileIndex] = {
      ...profiles[profileIndex],
      ...updates,
      updated_at: new Date().toISOString()
    }

    localStorage.setItem(`${this.PROFILES_KEY}_${userId}`, JSON.stringify(profiles))
    return profiles[profileIndex]
  }

  static async deleteTravelProfile(userId: string, profileId: string): Promise<void> {
    const profiles = await this.getTravelProfiles(userId)
    const filteredProfiles = profiles.filter(p => p.id !== profileId)
    
    // If we deleted the default profile, make the first one default
    if (filteredProfiles.length > 0 && !filteredProfiles.some(p => p.is_default)) {
      filteredProfiles[0].is_default = true
    }
    
    localStorage.setItem(`${this.PROFILES_KEY}_${userId}`, JSON.stringify(filteredProfiles))
  }

  static getDefaultProfiles(userId: string): TravelProfile[] {
    return [
      {
        id: 'default_profile',
        user_id: userId,
        name: 'My Travel Style',
        description: 'Default travel preferences',
        preferences: this.getDefaultPreferences(),
        is_default: true,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }
    ]
  }

  // Personalization Settings
  static async getPersonalizationSettings(userId: string): Promise<PersonalizationSettings> {
    try {
      const stored = localStorage.getItem(`${this.SETTINGS_KEY}_${userId}`)
      if (stored) {
        return JSON.parse(stored)
      }
      return this.getDefaultSettings()
    } catch (error) {
      console.error('Error getting personalization settings:', error)
      return this.getDefaultSettings()
    }
  }

  static async updatePersonalizationSettings(userId: string, settings: PersonalizationSettings): Promise<void> {
    try {
      localStorage.setItem(`${this.SETTINGS_KEY}_${userId}`, JSON.stringify(settings))
      console.log('Personalization settings updated:', settings)
    } catch (error) {
      console.error('Error updating personalization settings:', error)
      throw error
    }
  }

  static getDefaultSettings(): PersonalizationSettings {
    return {
      show_recommendations: true,
      recommendation_frequency: 'medium',
      privacy_level: 'private',
      data_sharing: {
        analytics: true,
        marketing: false,
        recommendations: true
      },
      notification_preferences: {
        new_recommendations: true,
        price_alerts: true,
        trending_destinations: false,
        personalized_offers: true
      },
      display_preferences: {
        currency: 'USD',
        language: 'en',
        date_format: 'MM/DD/YYYY',
        temperature_unit: 'fahrenheit',
        distance_unit: 'imperial'
      }
    }
  }

  // User Interests Management
  static async getUserInterests(userId: string): Promise<UserInterests> {
    try {
      const stored = localStorage.getItem(`${this.INTERESTS_KEY}_${userId}`)
      if (stored) {
        return JSON.parse(stored)
      }
      return this.getDefaultInterests()
    } catch (error) {
      console.error('Error getting user interests:', error)
      return this.getDefaultInterests()
    }
  }

  static async updateUserInterests(userId: string, interests: UserInterests): Promise<void> {
    try {
      localStorage.setItem(`${this.INTERESTS_KEY}_${userId}`, JSON.stringify(interests))
      console.log('User interests updated:', interests)
    } catch (error) {
      console.error('Error updating user interests:', error)
      throw error
    }
  }

  static getDefaultInterests(): UserInterests {
    return {
      categories: ['culture', 'nature', 'adventure'],
      activities: ['sightseeing', 'hiking', 'museums'],
      cuisines: ['local', 'international'],
      accommodation_types: ['hotel', 'resort'],
      transportation_modes: ['flight', 'car'],
      budget_categories: ['mid-range'],
      travel_companions: ['partner', 'friends'],
      special_occasions: ['vacation', 'anniversary']
    }
  }

  // Travel History Management
  static async getTravelHistory(userId: string): Promise<TravelHistory> {
    try {
      const stored = localStorage.getItem(`${this.HISTORY_KEY}_${userId}`)
      if (stored) {
        return JSON.parse(stored)
      }
      return this.getDefaultHistory()
    } catch (error) {
      console.error('Error getting travel history:', error)
      return this.getDefaultHistory()
    }
  }

  static async updateTravelHistory(userId: string, history: TravelHistory): Promise<void> {
    try {
      localStorage.setItem(`${this.HISTORY_KEY}_${userId}`, JSON.stringify(history))
      console.log('Travel history updated:', history)
    } catch (error) {
      console.error('Error updating travel history:', error)
      throw error
    }
  }

  static getDefaultHistory(): TravelHistory {
    return {
      destinations_visited: [],
      favorite_destinations: [],
      travel_frequency: 2,
      average_trip_duration: 7,
      preferred_seasons: ['spring', 'fall'],
      booking_patterns: [{
        advance_booking_days: 60,
        preferred_booking_time: 'evening',
        booking_channel: 'online',
        decision_factors: ['price', 'reviews', 'location'],
        research_duration_days: 14
      }],
      spending_patterns: [{
        category: 'accommodation',
        average_amount: 150,
        percentage_of_budget: 40,
        seasonal_variation: 0.2
      }]
    }
  }

  // Preference Analysis
  static analyzePreferences(preferences: UserPreferences): {
    travel_persona: string
    recommendations: string[]
    insights: string[]
  } {
    const persona = this.determineTravelPersona(preferences)
    const recommendations = this.generatePreferenceRecommendations(preferences)
    const insights = this.generateInsights(preferences)

    return { travel_persona: persona, recommendations, insights }
  }

  private static determineTravelPersona(preferences: UserPreferences): string {
    const { travel_style, budget_range, activity_preferences } = preferences
    const [minBudget, maxBudget] = budget_range
    const avgBudget = (minBudget + maxBudget) / 2

    if (avgBudget > 4000 && travel_style === 'luxury') {
      return 'Luxury Traveler'
    } else if (activity_preferences.includes('adventure') || activity_preferences.includes('hiking')) {
      return 'Adventure Seeker'
    } else if (travel_style === 'cultural' && activity_preferences.includes('museums')) {
      return 'Cultural Explorer'
    } else if (travel_style === 'relaxation') {
      return 'Relaxation Seeker'
    } else if (travel_style === 'family') {
      return 'Family Traveler'
    } else {
      return 'Balanced Traveler'
    }
  }

  private static generatePreferenceRecommendations(preferences: UserPreferences): string[] {
    const recommendations: string[] = []
    
    if (preferences.budget_range[0] < 1500) {
      recommendations.push('Consider budget-friendly destinations in Southeast Asia or Eastern Europe')
    }
    
    if (preferences.activity_preferences.includes('adventure')) {
      recommendations.push('Explore destinations with outdoor activities like New Zealand or Costa Rica')
    }
    
    if (preferences.climate_preferences === 'tropical') {
      recommendations.push('Perfect time to visit Caribbean or Pacific islands')
    }

    return recommendations
  }

  private static generateInsights(preferences: UserPreferences): string[] {
    const insights: string[] = []
    
    if (preferences.preferred_duration.every(d => d <= 7)) {
      insights.push('You prefer shorter trips - consider nearby destinations to maximize your time')
    }
    
    if (preferences.group_size === 1) {
      insights.push('Solo travel opens up unique opportunities for personal growth and flexibility')
    }
    
    if (preferences.dietary_restrictions.length > 0) {
      insights.push('Research local cuisine options in advance for the best dining experience')
    }

    return insights
  }

  // Export/Import functionality
  static async exportUserData(userId: string): Promise<string> {
    const data = {
      preferences: await this.getUserPreferences(userId),
      profiles: await this.getTravelProfiles(userId),
      settings: await this.getPersonalizationSettings(userId),
      interests: await this.getUserInterests(userId),
      history: await this.getTravelHistory(userId),
      exported_at: new Date().toISOString()
    }
    
    return JSON.stringify(data, null, 2)
  }

  static async importUserData(userId: string, jsonData: string): Promise<void> {
    try {
      const data = JSON.parse(jsonData)
      
      if (data.preferences) {
        await this.updateUserPreferences(userId, data.preferences)
      }
      
      if (data.settings) {
        await this.updatePersonalizationSettings(userId, data.settings)
      }
      
      if (data.interests) {
        await this.updateUserInterests(userId, data.interests)
      }
      
      if (data.history) {
        await this.updateTravelHistory(userId, data.history)
      }
      
      console.log('User data imported successfully')
    } catch (error) {
      console.error('Error importing user data:', error)
      throw new Error('Invalid data format')
    }
  }
}

export { UserPreferencesService }
