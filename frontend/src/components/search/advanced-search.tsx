'use client'

import { useState, useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { Search, Filter, X, Calendar, Users, DollarSign, MapPin, Star } from 'lucide-react'
import { DestinationsService } from '@/lib/destinations'
import { Destination } from '@/types'

interface SearchFilters {
  query: string
  country: string
  minPrice: number
  maxPrice: number
  duration: string
  maxGuests: number
  features: string[]
  rating: number
  sortBy: string
  sortOrder: 'asc' | 'desc'
}

interface AdvancedSearchProps {
  onResults: (destinations: Destination[]) => void
  onFiltersChange: (filters: SearchFilters) => void
  initialFilters?: Partial<SearchFilters>
}

export default function AdvancedSearch({
  onResults,
  onFiltersChange,
  initialFilters = {},
}: AdvancedSearchProps) {
  const router = useRouter()
  const searchParams = useSearchParams()
  
  const [filters, setFilters] = useState<SearchFilters>({
    query: '',
    country: '',
    minPrice: 0,
    maxPrice: 10000,
    duration: '',
    maxGuests: 1,
    features: [],
    rating: 0,
    sortBy: 'price',
    sortOrder: 'asc',
    ...initialFilters,
  })
  
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [loading, setLoading] = useState(false)
  const [countries, setCountries] = useState<string[]>([])
  const [availableFeatures, setAvailableFeatures] = useState<string[]>([])

  useEffect(() => {
    // Load initial data
    loadFilterOptions()
    
    // Initialize from URL params
    const urlFilters = getFiltersFromURL()
    if (Object.keys(urlFilters).length > 0) {
      setFilters(prev => ({ ...prev, ...urlFilters }))
      setShowAdvanced(true)
    }
  }, [])

  useEffect(() => {
    // Perform search when filters change
    performSearch()
    onFiltersChange(filters)
    updateURL()
  }, [filters])

  const loadFilterOptions = async () => {
    try {
      const destinations = DestinationsService.getMockDestinations()
      
      // Extract unique countries
      const uniqueCountries = Array.from(new Set(destinations.map(d => d.country))).sort()
      setCountries(uniqueCountries)
      
      // Extract unique features
      const allFeatures = destinations.flatMap(d => d.features)
      const uniqueFeatures = Array.from(new Set(allFeatures)).sort()
      setAvailableFeatures(uniqueFeatures)
    } catch (error) {
      console.error('Error loading filter options:', error)
    }
  }

  const getFiltersFromURL = (): Partial<SearchFilters> => {
    const params: Partial<SearchFilters> = {}
    
    if (searchParams?.get('q')) params.query = searchParams.get('q')!
    if (searchParams?.get('country')) params.country = searchParams.get('country')!
    if (searchParams?.get('minPrice')) params.minPrice = Number(searchParams.get('minPrice'))
    if (searchParams?.get('maxPrice')) params.maxPrice = Number(searchParams.get('maxPrice'))
    if (searchParams?.get('duration')) params.duration = searchParams.get('duration')!
    if (searchParams?.get('guests')) params.maxGuests = Number(searchParams.get('guests'))
    if (searchParams?.get('features')) params.features = searchParams.get('features')!.split(',')
    if (searchParams?.get('rating')) params.rating = Number(searchParams.get('rating'))
    if (searchParams?.get('sort')) params.sortBy = searchParams.get('sort')!
    if (searchParams?.get('order')) params.sortOrder = searchParams.get('order') as 'asc' | 'desc'
    
    return params
  }

  const updateURL = () => {
    const params = new URLSearchParams()
    
    if (filters.query) params.set('q', filters.query)
    if (filters.country) params.set('country', filters.country)
    if (filters.minPrice > 0) params.set('minPrice', filters.minPrice.toString())
    if (filters.maxPrice < 10000) params.set('maxPrice', filters.maxPrice.toString())
    if (filters.duration) params.set('duration', filters.duration)
    if (filters.maxGuests > 1) params.set('guests', filters.maxGuests.toString())
    if (filters.features.length > 0) params.set('features', filters.features.join(','))
    if (filters.rating > 0) params.set('rating', filters.rating.toString())
    if (filters.sortBy !== 'price') params.set('sort', filters.sortBy)
    if (filters.sortOrder !== 'asc') params.set('order', filters.sortOrder)
    
    const newURL = `${window.location.pathname}?${params.toString()}`
    window.history.replaceState({}, '', newURL)
  }

  const performSearch = async () => {
    setLoading(true)
    
    try {
      // Get all destinations
      let destinations = DestinationsService.getMockDestinations()
      
      // Apply filters
      destinations = destinations.filter(destination => {
        // Text search
        if (filters.query) {
          const searchText = filters.query.toLowerCase()
          const matchesText = 
            destination.name.toLowerCase().includes(searchText) ||
            destination.description.toLowerCase().includes(searchText) ||
            destination.country.toLowerCase().includes(searchText) ||
            destination.city.toLowerCase().includes(searchText)
          
          if (!matchesText) return false
        }
        
        // Country filter
        if (filters.country && destination.country !== filters.country) {
          return false
        }
        
        // Price range
        if (destination.price < filters.minPrice || destination.price > filters.maxPrice) {
          return false
        }
        
        // Duration
        if (filters.duration) {
          const targetDuration = parseInt(filters.duration)
          if (destination.duration !== targetDuration) {
            return false
          }
        }
        
        // Max guests
        if (destination.max_guests < filters.maxGuests) {
          return false
        }
        
        // Features
        if (filters.features.length > 0) {
          const hasAllFeatures = filters.features.every(feature =>
            destination.features.includes(feature)
          )
          if (!hasAllFeatures) return false
        }
        
        // Rating (mock - in real app this would come from reviews)
        const mockRating = 4.5 // Mock rating
        if (mockRating < filters.rating) {
          return false
        }
        
        return true
      })
      
      // Apply sorting
      destinations.sort((a, b) => {
        let comparison = 0
        
        switch (filters.sortBy) {
          case 'price':
            comparison = a.price - b.price
            break
          case 'name':
            comparison = a.name.localeCompare(b.name)
            break
          case 'duration':
            comparison = a.duration - b.duration
            break
          case 'rating':
            // Mock rating comparison
            comparison = 0 // In real app, compare actual ratings
            break
          default:
            comparison = 0
        }
        
        return filters.sortOrder === 'desc' ? -comparison : comparison
      })
      
      onResults(destinations)
    } catch (error) {
      console.error('Search error:', error)
      onResults([])
    } finally {
      setLoading(false)
    }
  }

  const updateFilter = (key: keyof SearchFilters, value: any) => {
    setFilters(prev => ({ ...prev, [key]: value }))
  }

  const toggleFeature = (feature: string) => {
    setFilters(prev => ({
      ...prev,
      features: prev.features.includes(feature)
        ? prev.features.filter(f => f !== feature)
        : [...prev.features, feature]
    }))
  }

  const clearFilters = () => {
    setFilters({
      query: '',
      country: '',
      minPrice: 0,
      maxPrice: 10000,
      duration: '',
      maxGuests: 1,
      features: [],
      rating: 0,
      sortBy: 'price',
      sortOrder: 'asc',
    })
  }

  const hasActiveFilters = () => {
    return (
      filters.query ||
      filters.country ||
      filters.minPrice > 0 ||
      filters.maxPrice < 10000 ||
      filters.duration ||
      filters.maxGuests > 1 ||
      filters.features.length > 0 ||
      filters.rating > 0
    )
  }

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
      {/* Basic Search */}
      <div className="flex flex-col md:flex-row gap-4 mb-4">
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
          <input
            type="text"
            placeholder="Search destinations, countries, or activities..."
            value={filters.query}
            onChange={(e) => updateFilter('query', e.target.value)}
            className="w-full pl-10 pr-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          />
        </div>
        
        <button
          onClick={() => setShowAdvanced(!showAdvanced)}
          className={`flex items-center px-4 py-3 border rounded-lg transition-colors ${
            showAdvanced || hasActiveFilters()
              ? 'border-primary bg-primary text-primary-foreground'
              : 'border-gray-300 text-gray-700 hover:bg-gray-50'
          }`}
        >
          <Filter className="h-5 w-5 mr-2" />
          Filters
          {hasActiveFilters() && (
            <span className="ml-2 bg-white text-primary rounded-full w-5 h-5 flex items-center justify-center text-xs font-medium">
              {filters.features.length + (filters.country ? 1 : 0) + (filters.duration ? 1 : 0)}
            </span>
          )}
        </button>
      </div>

      {/* Advanced Filters */}
      {showAdvanced && (
        <div className="border-t border-gray-200 pt-6 space-y-6">
          {/* Location and Duration */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Country
              </label>
              <select
                value={filters.country}
                onChange={(e) => updateFilter('country', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
              >
                <option value="">All Countries</option>
                {countries.map(country => (
                  <option key={country} value={country}>{country}</option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Duration
              </label>
              <select
                value={filters.duration}
                onChange={(e) => updateFilter('duration', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
              >
                <option value="">Any Duration</option>
                <option value="3">3 days</option>
                <option value="5">5 days</option>
                <option value="7">7 days</option>
                <option value="10">10 days</option>
                <option value="14">14 days</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Max Guests
              </label>
              <select
                value={filters.maxGuests}
                onChange={(e) => updateFilter('maxGuests', Number(e.target.value))}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
              >
                {Array.from({ length: 10 }, (_, i) => i + 1).map(num => (
                  <option key={num} value={num}>{num} guest{num !== 1 ? 's' : ''}</option>
                ))}
              </select>
            </div>
          </div>

          {/* Price Range */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Price Range: ${filters.minPrice.toLocaleString()} - ${filters.maxPrice.toLocaleString()}
            </label>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <input
                  type="range"
                  min="0"
                  max="10000"
                  step="100"
                  value={filters.minPrice}
                  onChange={(e) => updateFilter('minPrice', Number(e.target.value))}
                  className="w-full"
                />
                <div className="text-xs text-gray-500 mt-1">Min: ${filters.minPrice.toLocaleString()}</div>
              </div>
              <div>
                <input
                  type="range"
                  min="0"
                  max="10000"
                  step="100"
                  value={filters.maxPrice}
                  onChange={(e) => updateFilter('maxPrice', Number(e.target.value))}
                  className="w-full"
                />
                <div className="text-xs text-gray-500 mt-1">Max: ${filters.maxPrice.toLocaleString()}</div>
              </div>
            </div>
          </div>

          {/* Features */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Features & Amenities
            </label>
            <div className="flex flex-wrap gap-2">
              {availableFeatures.map(feature => (
                <button
                  key={feature}
                  onClick={() => toggleFeature(feature)}
                  className={`px-3 py-1 text-sm rounded-full border transition-colors ${
                    filters.features.includes(feature)
                      ? 'border-primary bg-primary text-primary-foreground'
                      : 'border-gray-300 text-gray-700 hover:border-primary hover:text-primary'
                  }`}
                >
                  {feature}
                </button>
              ))}
            </div>
          </div>

          {/* Sorting */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Sort By
              </label>
              <select
                value={filters.sortBy}
                onChange={(e) => updateFilter('sortBy', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
              >
                <option value="price">Price</option>
                <option value="name">Name</option>
                <option value="duration">Duration</option>
                <option value="rating">Rating</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Order
              </label>
              <select
                value={filters.sortOrder}
                onChange={(e) => updateFilter('sortOrder', e.target.value as 'asc' | 'desc')}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
              >
                <option value="asc">Low to High</option>
                <option value="desc">High to Low</option>
              </select>
            </div>
          </div>

          {/* Clear Filters */}
          {hasActiveFilters() && (
            <div className="flex justify-between items-center pt-4 border-t border-gray-200">
              <span className="text-sm text-gray-600">
                {hasActiveFilters() ? 'Active filters applied' : 'No filters applied'}
              </span>
              <button
                onClick={clearFilters}
                className="flex items-center text-sm text-gray-600 hover:text-primary transition-colors"
              >
                <X className="h-4 w-4 mr-1" />
                Clear all filters
              </button>
            </div>
          )}
        </div>
      )}

      {/* Loading indicator */}
      {loading && (
        <div className="flex items-center justify-center py-4">
          <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
          <span className="ml-2 text-gray-600">Searching...</span>
        </div>
      )}
    </div>
  )
}
