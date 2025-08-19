'use client'

import React, { useState, useEffect, useRef, useCallback } from 'react'
import { Search, Sparkles, Filter, X, MapPin, Calendar, Users, DollarSign, Zap, Brain, TrendingUp } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { FadeIn, StaggerContainer, StaggerItem } from '@/components/ui/animated'
import { Destination } from '@/types'
import { RecommendationEngine, UserPreferences } from '@/lib/recommendations'
import { UserPreferencesService } from '@/lib/user-preferences'

interface AISearchProps {
  onResults: (results: SearchResult[]) => void
  onSuggestionSelect: (suggestion: SearchSuggestion) => void
  className?: string
  userId?: string
}

interface SearchResult {
  destination: Destination
  relevance_score: number
  ai_insights: string[]
  match_reasons: string[]
  personalization_score: number
}

interface SearchSuggestion {
  type: 'destination' | 'activity' | 'style' | 'budget' | 'duration'
  text: string
  confidence: number
  icon: React.ReactNode
  action: () => void
}

interface SmartFilter {
  id: string
  name: string
  type: 'range' | 'select' | 'multi-select' | 'toggle'
  value: any
  options?: { label: string; value: any }[]
  ai_suggested: boolean
  impact_score: number
}

export default function AIPoweredSearch({ onResults, onSuggestionSelect, className = '', userId }: AISearchProps) {
  const [query, setQuery] = useState('')
  const [isSearching, setIsSearching] = useState(false)
  const [suggestions, setSuggestions] = useState<SearchSuggestion[]>([])
  const [showSuggestions, setShowSuggestions] = useState(false)
  const [smartFilters, setSmartFilters] = useState<SmartFilter[]>([])
  const [userPreferences, setUserPreferences] = useState<UserPreferences | null>(null)
  const [searchHistory, setSearchHistory] = useState<string[]>([])
  const [aiInsights, setAiInsights] = useState<string[]>([])
  const [isAiMode, setIsAiMode] = useState(true)
  
  const searchRef = useRef<HTMLDivElement>(null)
  const debounceRef = useRef<NodeJS.Timeout>()

  // Load user preferences and search history
  useEffect(() => {
    if (userId) {
      loadUserData()
    }
  }, [userId])

  const loadUserData = async () => {
    if (!userId) return
    
    try {
      const preferences = await UserPreferencesService.getUserPreferences(userId)
      setUserPreferences(preferences)
      
      // Load search history from localStorage
      const history = JSON.parse(localStorage.getItem(`search_history_${userId}`) || '[]')
      setSearchHistory(history.slice(0, 5)) // Keep last 5 searches
      
      // Generate smart filters based on preferences
      generateSmartFilters(preferences)
    } catch (error) {
      console.error('Error loading user data:', error)
    }
  }

  // Generate AI-powered search suggestions
  const generateSuggestions = useCallback(async (searchQuery: string) => {
    if (searchQuery.length < 2) {
      setSuggestions([])
      return
    }

    const newSuggestions: SearchSuggestion[] = []

    // Destination suggestions
    if (searchQuery.toLowerCase().includes('beach') || searchQuery.toLowerCase().includes('ocean')) {
      newSuggestions.push({
        type: 'destination',
        text: 'Tropical beach destinations',
        confidence: 0.9,
        icon: <MapPin className="h-4 w-4" />,
        action: () => handleSuggestionClick('tropical beach destinations')
      })
    }

    // Activity suggestions
    if (searchQuery.toLowerCase().includes('adventure') || searchQuery.toLowerCase().includes('hiking')) {
      newSuggestions.push({
        type: 'activity',
        text: 'Adventure & outdoor activities',
        confidence: 0.85,
        icon: <TrendingUp className="h-4 w-4" />,
        action: () => handleSuggestionClick('adventure activities')
      })
    }

    // Budget suggestions based on user preferences
    if (userPreferences && searchQuery.toLowerCase().includes('budget')) {
      const [minBudget, maxBudget] = userPreferences.budget_range
      newSuggestions.push({
        type: 'budget',
        text: `Destinations within $${minBudget}-$${maxBudget}`,
        confidence: 0.95,
        icon: <DollarSign className="h-4 w-4" />,
        action: () => applyBudgetFilter(minBudget, maxBudget)
      })
    }

    // Duration suggestions
    if (searchQuery.toLowerCase().includes('week') || searchQuery.toLowerCase().includes('day')) {
      newSuggestions.push({
        type: 'duration',
        text: 'Perfect for your preferred trip length',
        confidence: 0.8,
        icon: <Calendar className="h-4 w-4" />,
        action: () => applyDurationFilter()
      })
    }

    // AI-powered contextual suggestions
    if (isAiMode && userPreferences) {
      const aiSuggestion = await generateAISuggestion(searchQuery, userPreferences)
      if (aiSuggestion) {
        newSuggestions.push(aiSuggestion)
      }
    }

    setSuggestions(newSuggestions)
    setShowSuggestions(true)
  }, [userPreferences, isAiMode])

  // Debounced search
  useEffect(() => {
    if (debounceRef.current) {
      clearTimeout(debounceRef.current)
    }

    debounceRef.current = setTimeout(() => {
      if (query) {
        generateSuggestions(query)
      } else {
        setSuggestions([])
        setShowSuggestions(false)
      }
    }, 300)

    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current)
      }
    }
  }, [query, generateSuggestions])

  // Generate smart filters based on user preferences
  const generateSmartFilters = (preferences: UserPreferences) => {
    const filters: SmartFilter[] = [
      {
        id: 'budget',
        name: 'Budget Range',
        type: 'range',
        value: preferences.budget_range,
        ai_suggested: true,
        impact_score: 0.9
      },
      {
        id: 'duration',
        name: 'Trip Duration',
        type: 'multi-select',
        value: preferences.preferred_duration,
        options: [
          { label: '3-5 days', value: [3, 5] },
          { label: '1 week', value: [7] },
          { label: '2 weeks', value: [14] },
          { label: '3+ weeks', value: [21, 30] }
        ],
        ai_suggested: true,
        impact_score: 0.7
      },
      {
        id: 'travel_style',
        name: 'Travel Style',
        type: 'select',
        value: preferences.travel_style,
        options: [
          { label: 'Luxury', value: 'luxury' },
          { label: 'Adventure', value: 'adventure' },
          { label: 'Cultural', value: 'cultural' },
          { label: 'Relaxation', value: 'relaxation' },
          { label: 'Family', value: 'family' },
          { label: 'Romantic', value: 'romantic' }
        ],
        ai_suggested: false,
        impact_score: 0.8
      },
      {
        id: 'group_size',
        name: 'Group Size',
        type: 'range',
        value: [1, preferences.group_size],
        ai_suggested: false,
        impact_score: 0.6
      }
    ]

    setSmartFilters(filters)
  }

  // AI suggestion generation
  const generateAISuggestion = async (query: string, preferences: UserPreferences): Promise<SearchSuggestion | null> => {
    // Simulate AI analysis
    await new Promise(resolve => setTimeout(resolve, 200))

    const insights = [
      'Based on your travel history, you might enjoy cultural destinations',
      'Your budget suggests premium experiences in emerging markets',
      'Consider shoulder season travel for better value',
      'Your group size is perfect for boutique accommodations'
    ]

    const randomInsight = insights[Math.floor(Math.random() * insights.length)]

    return {
      type: 'style',
      text: `AI Insight: ${randomInsight}`,
      confidence: 0.75,
      icon: <Brain className="h-4 w-4" />,
      action: () => setAiInsights([randomInsight])
    }
  }

  // Handle search execution
  const handleSearch = async () => {
    if (!query.trim()) return

    setIsSearching(true)
    setShowSuggestions(false)

    try {
      // Add to search history
      const newHistory = [query, ...searchHistory.filter(h => h !== query)].slice(0, 5)
      setSearchHistory(newHistory)
      if (userId) {
        localStorage.setItem(`search_history_${userId}`, JSON.stringify(newHistory))
      }

      // Perform AI-powered search
      const results = await performAISearch(query)
      onResults(results)

      // Generate AI insights
      const insights = generateSearchInsights(query, results)
      setAiInsights(insights)

    } catch (error) {
      console.error('Search error:', error)
    } finally {
      setIsSearching(false)
    }
  }

  // AI-powered search logic
  const performAISearch = async (searchQuery: string): Promise<SearchResult[]> => {
    // This would integrate with the recommendation engine
    // For now, we'll simulate AI-powered results
    
    const mockResults: SearchResult[] = [
      {
        destination: {
          id: 1,
          name: 'Santorini, Greece',
          country: 'Greece',
          city: 'Santorini',
          price: 2500,
          duration: 7,
          max_guests: 4,
          images: ['/images/santorini.jpg'],
          features: ['Beach Access', 'Cultural Sites', 'Romantic Setting'],
          description: 'Beautiful Greek island with stunning sunsets',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z'
        } as Destination,
        relevance_score: 0.95,
        ai_insights: [
          'Perfect match for your romantic travel style',
          'Price aligns with your budget preferences',
          'Duration matches your preferred trip length'
        ],
        match_reasons: [
          'Romantic setting matches your travel style',
          'Mediterranean climate fits your preferences',
          'Cultural activities align with your interests'
        ],
        personalization_score: 0.9
      }
    ]

    return mockResults
  }

  // Generate search insights
  const generateSearchInsights = (query: string, results: SearchResult[]): string[] => {
    const insights: string[] = []

    if (results.length > 0) {
      insights.push(`Found ${results.length} destinations matching your search`)
      
      const avgScore = results.reduce((sum, r) => sum + r.relevance_score, 0) / results.length
      if (avgScore > 0.8) {
        insights.push('High-quality matches found based on your preferences')
      }
      
      if (results.some(r => r.ai_insights.length > 0)) {
        insights.push('AI has identified personalized recommendations for you')
      }
    }

    return insights
  }

  // Handle suggestion clicks
  const handleSuggestionClick = (suggestion: string) => {
    setQuery(suggestion)
    setShowSuggestions(false)
    onSuggestionSelect({
      type: 'destination',
      text: suggestion,
      confidence: 0.8,
      icon: <MapPin className="h-4 w-4" />,
      action: () => {}
    })
  }

  // Filter actions
  const applyBudgetFilter = (min: number, max: number) => {
    const filter = smartFilters.find(f => f.id === 'budget')
    if (filter) {
      filter.value = [min, max]
      setSmartFilters([...smartFilters])
    }
  }

  const applyDurationFilter = () => {
    if (userPreferences) {
      const filter = smartFilters.find(f => f.id === 'duration')
      if (filter) {
        filter.value = userPreferences.preferred_duration
        setSmartFilters([...smartFilters])
      }
    }
  }

  // Close suggestions when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (searchRef.current && !searchRef.current.contains(event.target as Node)) {
        setShowSuggestions(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  return (
    <div className={`space-y-6 ${className}`}>
      {/* AI Mode Toggle */}
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-gray-900">Smart Search</h2>
        <div className="flex items-center space-x-3">
          <span className="text-sm text-gray-600">AI-Powered</span>
          <button
            onClick={() => setIsAiMode(!isAiMode)}
            className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
              isAiMode ? 'bg-brand-500' : 'bg-gray-300'
            }`}
          >
            <span
              className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                isAiMode ? 'translate-x-6' : 'translate-x-1'
              }`}
            />
          </button>
          {isAiMode && <Sparkles className="h-5 w-5 text-brand-500" />}
        </div>
      </div>

      {/* Search Input */}
      <div className="relative" ref={searchRef}>
        <div className="relative">
          <Input
            placeholder={isAiMode ? "Describe your dream destination..." : "Search destinations..."}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
            leftIcon={isAiMode ? <Brain className="h-5 w-5" /> : <Search className="h-5 w-5" />}
            rightIcon={
              isSearching ? (
                <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-brand-500" />
              ) : query ? (
                <X 
                  className="h-5 w-5 cursor-pointer hover:text-gray-700" 
                  onClick={() => setQuery('')}
                />
              ) : null
            }
            className="text-lg h-14 pr-12"
            size="lg"
          />
        </div>

        {/* AI Suggestions */}
        <AnimatePresence>
          {showSuggestions && suggestions.length > 0 && (
            <motion.div
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              className="absolute top-full left-0 right-0 bg-white border border-gray-200 rounded-lg shadow-xl z-50 mt-2"
            >
              <div className="p-3 border-b border-gray-100">
                <div className="flex items-center space-x-2 text-sm text-gray-600">
                  <Sparkles className="h-4 w-4" />
                  <span>AI Suggestions</span>
                </div>
              </div>
              {suggestions.map((suggestion, index) => (
                <motion.div
                  key={index}
                  whileHover={{ backgroundColor: '#f8fafc' }}
                  className="flex items-center p-4 cursor-pointer border-b border-gray-100 last:border-b-0"
                  onClick={suggestion.action}
                >
                  <div className="flex items-center space-x-3 flex-1">
                    <div className="text-brand-500">{suggestion.icon}</div>
                    <div>
                      <div className="font-medium text-gray-900">{suggestion.text}</div>
                      <div className="text-xs text-gray-500 capitalize">{suggestion.type}</div>
                    </div>
                  </div>
                  <div className="text-xs text-gray-400">
                    {Math.round(suggestion.confidence * 100)}% match
                  </div>
                </motion.div>
              ))}
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      {/* Search History */}
      {searchHistory.length > 0 && (
        <div>
          <h3 className="text-sm font-medium text-gray-700 mb-2">Recent Searches</h3>
          <div className="flex flex-wrap gap-2">
            {searchHistory.map((search, index) => (
              <button
                key={index}
                onClick={() => setQuery(search)}
                className="px-3 py-1 bg-gray-100 text-gray-700 rounded-full text-sm hover:bg-gray-200 transition-colors"
              >
                {search}
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Smart Filters */}
      {smartFilters.length > 0 && (
        <div>
          <div className="flex items-center space-x-2 mb-4">
            <Filter className="h-5 w-5 text-gray-600" />
            <h3 className="text-lg font-semibold text-gray-900">Smart Filters</h3>
            <span className="text-xs bg-brand-100 text-brand-700 px-2 py-1 rounded-full">
              AI Optimized
            </span>
          </div>
          
          <StaggerContainer staggerDelay={0.1}>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {smartFilters.map((filter) => (
                <StaggerItem key={filter.id}>
                  <div className={`p-4 rounded-lg border ${
                    filter.ai_suggested ? 'border-brand-200 bg-brand-50' : 'border-gray-200 bg-white'
                  }`}>
                    <div className="flex items-center justify-between mb-2">
                      <label className="font-medium text-gray-900">{filter.name}</label>
                      {filter.ai_suggested && (
                        <span className="text-xs bg-brand-500 text-white px-2 py-1 rounded-full">
                          AI
                        </span>
                      )}
                    </div>
                    {/* Filter controls would go here based on filter.type */}
                    <div className="text-sm text-gray-600">
                      Impact: {Math.round(filter.impact_score * 100)}%
                    </div>
                  </div>
                </StaggerItem>
              ))}
            </div>
          </StaggerContainer>
        </div>
      )}

      {/* AI Insights */}
      {aiInsights.length > 0 && (
        <FadeIn>
          <div className="bg-gradient-to-r from-brand-50 to-accent-50 rounded-lg p-6">
            <div className="flex items-center space-x-2 mb-3">
              <Brain className="h-5 w-5 text-brand-600" />
              <h3 className="font-semibold text-brand-900">AI Insights</h3>
            </div>
            <div className="space-y-2">
              {aiInsights.map((insight, index) => (
                <div key={index} className="flex items-start space-x-2">
                  <Zap className="h-4 w-4 text-brand-500 mt-0.5 flex-shrink-0" />
                  <p className="text-brand-800">{insight}</p>
                </div>
              ))}
            </div>
          </div>
        </FadeIn>
      )}

      {/* Search Button */}
      <Button
        onClick={handleSearch}
        disabled={!query.trim() || isSearching}
        className="w-full h-12 text-lg font-semibold"
        rightIcon={isAiMode ? <Sparkles className="h-5 w-5" /> : <Search className="h-5 w-5" />}
      >
        {isSearching ? 'Searching...' : isAiMode ? 'Search with AI' : 'Search'}
      </Button>
    </div>
  )
}
