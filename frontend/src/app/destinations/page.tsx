'use client'

import { useState, useEffect, useMemo, Suspense } from 'react'
import { useSearchParams } from 'next/navigation'
import Image from 'next/image'
import Link from 'next/link'
import { Search, Filter, MapPin, Clock, Users, Star, Grid, List } from 'lucide-react'
import { Destination } from '@/types'
import { DestinationsService, DestinationFilter } from '@/lib/destinations'
import { formatCurrency } from '@/lib/utils'
import StarRating from '@/components/reviews/star-rating'
import OptimizedImage from '@/components/images/optimized-image'

function DestinationsPageContent() {
  const searchParams = useSearchParams()
  const [destinations, setDestinations] = useState<Destination[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')
  const [showFilters, setShowFilters] = useState(false)

  // Filter state - Initialize with empty values first
  const [filters, setFilters] = useState<DestinationFilter>({
    search: '',
    country: '',
    min_price: undefined,
    max_price: undefined,
    duration: undefined,
    max_guests: undefined,
  })

  // Initialize filters from search params after component mounts
  useEffect(() => {
    if (searchParams) {
      setFilters({
        search: searchParams.get('search') || '',
        country: searchParams.get('country') || '',
        min_price: searchParams.get('min_price') ? Number(searchParams.get('min_price')) : undefined,
        max_price: searchParams.get('max_price') ? Number(searchParams.get('max_price')) : undefined,
        duration: searchParams.get('duration') ? Number(searchParams.get('duration')) : undefined,
        max_guests: searchParams.get('max_guests') ? Number(searchParams.get('max_guests')) : undefined,
      })
    }
  }, [searchParams])

  // Load destinations
  useEffect(() => {
    const loadDestinations = async () => {
      setLoading(true)
      setError('')
      
      try {
        // Try to load from API, fallback to mock data
        let data: Destination[]
        try {
          data = await DestinationsService.getDestinations(filters)
        } catch (apiError) {
          console.warn('API not available, using mock data:', apiError)
          data = DestinationsService.getMockDestinations()
        }
        
        setDestinations(data)
      } catch (err) {
        setError('Failed to load destinations. Please try again.')
        console.error('Error loading destinations:', err)
      } finally {
        setLoading(false)
      }
    }

    loadDestinations()
  }, [filters])

  // Filter destinations locally for mock data
  const filteredDestinations = useMemo(() => {
    return destinations.filter(destination => {
      if (filters.search && !destination.name.toLowerCase().includes(filters.search.toLowerCase()) &&
          !destination.description.toLowerCase().includes(filters.search.toLowerCase()) &&
          !destination.country.toLowerCase().includes(filters.search.toLowerCase())) {
        return false
      }
      
      if (filters.country && destination.country !== filters.country) {
        return false
      }
      
      if (filters.min_price && destination.price < filters.min_price) {
        return false
      }
      
      if (filters.max_price && destination.price > filters.max_price) {
        return false
      }
      
      if (filters.duration && destination.duration !== filters.duration) {
        return false
      }
      
      if (filters.max_guests && destination.max_guests < filters.max_guests) {
        return false
      }
      
      return true
    })
  }, [destinations, filters])

  const handleFilterChange = (key: keyof DestinationFilter, value: any) => {
    setFilters(prev => ({
      ...prev,
      [key]: value === '' ? undefined : value
    }))
  }

  const clearFilters = () => {
    setFilters({
      search: '',
      country: '',
      min_price: undefined,
      max_price: undefined,
      duration: undefined,
      max_guests: undefined,
    })
  }

  const countries = Array.from(new Set(destinations.map(d => d.country))).sort()

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Hero Section */}
      <section className="bg-gradient-to-r from-primary to-primary/80 text-white py-16">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center">
            <h1 className="text-4xl md:text-5xl font-bold mb-4">
              Exotic Destinations
            </h1>
            <p className="text-xl text-primary-foreground/90 max-w-3xl mx-auto">
              Discover extraordinary places that will create memories to last a lifetime
            </p>
          </div>
        </div>
      </section>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Search and Filters */}
        <div className="bg-white rounded-lg shadow-sm p-6 mb-8">
          {/* Search Bar */}
          <div className="flex flex-col lg:flex-row gap-4 mb-4">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
              <input
                type="text"
                placeholder="Search destinations..."
                value={filters.search || ''}
                onChange={(e) => handleFilterChange('search', e.target.value)}
                className="w-full pl-10 pr-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
              />
            </div>
            <button
              onClick={() => setShowFilters(!showFilters)}
              className="flex items-center px-4 py-3 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
            >
              <Filter className="h-5 w-5 mr-2" />
              Filters
            </button>
          </div>

          {/* Filters */}
          {showFilters && (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 pt-4 border-t">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Country</label>
                <select
                  value={filters.country || ''}
                  onChange={(e) => handleFilterChange('country', e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                >
                  <option value="">All Countries</option>
                  {countries.map(country => (
                    <option key={country} value={country}>{country}</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Max Price</label>
                <select
                  value={filters.max_price || ''}
                  onChange={(e) => handleFilterChange('max_price', e.target.value ? Number(e.target.value) : undefined)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                >
                  <option value="">Any Price</option>
                  <option value="1000">Under $1,000</option>
                  <option value="2000">Under $2,000</option>
                  <option value="5000">Under $5,000</option>
                  <option value="10000">Under $10,000</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Duration</label>
                <select
                  value={filters.duration || ''}
                  onChange={(e) => handleFilterChange('duration', e.target.value ? Number(e.target.value) : undefined)}
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
                <label className="block text-sm font-medium text-gray-700 mb-2">Guests</label>
                <select
                  value={filters.max_guests || ''}
                  onChange={(e) => handleFilterChange('max_guests', e.target.value ? Number(e.target.value) : undefined)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                >
                  <option value="">Any Size</option>
                  <option value="2">2 guests</option>
                  <option value="4">4 guests</option>
                  <option value="6">6 guests</option>
                  <option value="8">8+ guests</option>
                </select>
              </div>

              <div className="md:col-span-2 lg:col-span-4 flex justify-between items-center pt-4">
                <button
                  onClick={clearFilters}
                  className="text-gray-600 hover:text-primary transition-colors"
                >
                  Clear all filters
                </button>
                <div className="flex items-center space-x-2">
                  <span className="text-sm text-gray-600">View:</span>
                  <button
                    onClick={() => setViewMode('grid')}
                    className={`p-2 rounded ${viewMode === 'grid' ? 'bg-primary text-white' : 'bg-gray-100 text-gray-600'}`}
                  >
                    <Grid className="h-4 w-4" />
                  </button>
                  <button
                    onClick={() => setViewMode('list')}
                    className={`p-2 rounded ${viewMode === 'list' ? 'bg-primary text-white' : 'bg-gray-100 text-gray-600'}`}
                  >
                    <List className="h-4 w-4" />
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Results */}
        <div className="mb-6">
          <p className="text-gray-600">
            {filteredDestinations.length} destination{filteredDestinations.length !== 1 ? 's' : ''} found
          </p>
        </div>

        {/* Error State */}
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-8">
            <p className="text-red-700">{error}</p>
          </div>
        )}

        {/* Destinations Grid/List */}
        {viewMode === 'grid' ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {filteredDestinations.map((destination) => (
              <DestinationCard key={destination.id} destination={destination} />
            ))}
          </div>
        ) : (
          <div className="space-y-6">
            {filteredDestinations.map((destination) => (
              <DestinationListItem key={destination.id} destination={destination} />
            ))}
          </div>
        )}

        {/* No Results */}
        {filteredDestinations.length === 0 && !loading && (
          <div className="text-center py-12">
            <div className="text-gray-400 mb-4">
              <MapPin className="h-16 w-16 mx-auto" />
            </div>
            <h3 className="text-xl font-semibold text-gray-900 mb-2">No destinations found</h3>
            <p className="text-gray-600 mb-4">Try adjusting your search criteria or filters</p>
            <button
              onClick={clearFilters}
              className="bg-primary text-primary-foreground px-6 py-2 rounded-lg hover:bg-primary/90 transition-colors"
            >
              Clear Filters
            </button>
          </div>
        )}
      </div>
    </div>
  )
}

// Destination Card Component
function DestinationCard({ destination }: { destination: Destination }) {
  return (
    <div className="bg-white rounded-2xl shadow-lg overflow-hidden hover:shadow-xl transition-shadow duration-300 group">
      <div className="relative h-64 overflow-hidden">
        <OptimizedImage
          src={destination.images[0] || '/placeholder-destination.jpg'}
          alt={destination.name}
          width={400}
          height={256}
          quality={0.8}
          format="webp"
          fit="cover"
          lazy={true}
          className="w-full h-full object-cover group-hover:scale-110 transition-transform duration-300"
        />
        <div className="absolute top-4 right-4 bg-white/90 backdrop-blur-sm rounded-full px-3 py-1">
          <StarRating rating={4.2} size="sm" showValue />
        </div>
      </div>

      <div className="p-6">
        <div className="flex items-center text-sm text-gray-500 mb-2">
          <MapPin className="h-4 w-4 mr-1" />
          {destination.city}, {destination.country}
        </div>
        
        <h3 className="text-xl font-bold text-gray-900 mb-2 group-hover:text-primary transition-colors">
          {destination.name}
        </h3>
        
        <p className="text-gray-600 mb-4 line-clamp-2">
          {destination.description}
        </p>

        <div className="flex flex-wrap gap-2 mb-4">
          {destination.features.slice(0, 2).map((feature, index) => (
            <span
              key={index}
              className="bg-primary/10 text-primary text-xs px-2 py-1 rounded-full"
            >
              {feature}
            </span>
          ))}
          {destination.features.length > 2 && (
            <span className="text-xs text-gray-500">
              +{destination.features.length - 2} more
            </span>
          )}
        </div>

        <div className="flex items-center justify-between text-sm text-gray-500 mb-4">
          <div className="flex items-center">
            <Clock className="h-4 w-4 mr-1" />
            {destination.duration} days
          </div>
          <div className="flex items-center">
            <Users className="h-4 w-4 mr-1" />
            Up to {destination.max_guests} guests
          </div>
        </div>

        <div className="flex items-center justify-between">
          <div>
            <span className="text-2xl font-bold text-gray-900">
              {formatCurrency(destination.price)}
            </span>
            <span className="text-gray-500 text-sm ml-1">per person</span>
          </div>
          <Link
            href={`/destinations/${destination.id}`}
            className="bg-primary text-primary-foreground hover:bg-primary/90 px-4 py-2 rounded-lg font-medium transition-colors"
          >
            View Details
          </Link>
        </div>
      </div>
    </div>
  )
}

// Destination List Item Component
function DestinationListItem({ destination }: { destination: Destination }) {
  return (
    <div className="bg-white rounded-lg shadow-sm overflow-hidden hover:shadow-md transition-shadow duration-300">
      <div className="flex flex-col md:flex-row">
        <div className="relative h-48 md:h-32 md:w-48 flex-shrink-0">
          <Image
            src={destination.images[0] || '/placeholder-destination.jpg'}
            alt={destination.name}
            fill
            className="object-cover"
          />
        </div>
        
        <div className="flex-1 p-6">
          <div className="flex flex-col md:flex-row md:items-center md:justify-between">
            <div className="flex-1">
              <div className="flex items-center text-sm text-gray-500 mb-1">
                <MapPin className="h-4 w-4 mr-1" />
                {destination.city}, {destination.country}
              </div>
              
              <h3 className="text-xl font-bold text-gray-900 mb-2">
                {destination.name}
              </h3>
              
              <p className="text-gray-600 mb-3 line-clamp-2">
                {destination.description}
              </p>
              
              <div className="flex items-center space-x-4 text-sm text-gray-500">
                <div className="flex items-center">
                  <Clock className="h-4 w-4 mr-1" />
                  {destination.duration} days
                </div>
                <div className="flex items-center">
                  <Users className="h-4 w-4 mr-1" />
                  Up to {destination.max_guests} guests
                </div>
                <div className="flex items-center">
                  <Star className="h-4 w-4 mr-1 text-yellow-400 fill-current" />
                  4.8
                </div>
              </div>
            </div>
            
            <div className="mt-4 md:mt-0 md:ml-6 text-right">
              <div className="mb-3">
                <span className="text-2xl font-bold text-gray-900">
                  {formatCurrency(destination.price)}
                </span>
                <span className="text-gray-500 text-sm block">per person</span>
              </div>
              <Link
                href={`/destinations/${destination.id}`}
                className="bg-primary text-primary-foreground hover:bg-primary/90 px-4 py-2 rounded-lg font-medium transition-colors inline-block"
              >
                View Details
              </Link>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default function DestinationsPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    }>
      <DestinationsPageContent />
    </Suspense>
  )
}
