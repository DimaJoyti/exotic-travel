'use client'

import { useState, useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Image from 'next/image'
import Link from 'next/link'
import { ArrowLeft, MapPin, Clock, Users, Star, Calendar, Heart, Share2, Check, Map, Eye, TrendingUp, Layers, Image as ImageIcon } from 'lucide-react'
import { Destination } from '@/types'
import { DestinationsService } from '@/lib/destinations'
import { formatCurrency } from '@/lib/utils'
import { useAuth } from '@/contexts/auth-context'
import ReviewsSection from '@/components/reviews/reviews-section'
import { ImageGrid } from '@/components/images/image-gallery'
import OptimizedImage, { HeroImage } from '@/components/images/optimized-image'
import { ImagesService } from '@/lib/images'
// import InteractiveMap from '@/components/maps/interactive-map'
// import WishlistButton from '@/components/wishlist/wishlist-button'
// import RealTimeIndicator from '@/components/real-time/real-time-indicators'
// import PanoramicViewer from '@/components/images/panoramic-viewer'
// import DestinationComparison from '@/components/comparison/destination-comparison'

export default function DestinationDetailPage() {
  const params = useParams()
  const router = useRouter()
  const { user } = useAuth()
  const [destination, setDestination] = useState<Destination | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [selectedImageIndex, setSelectedImageIndex] = useState(0)
  const [isFavorite, setIsFavorite] = useState(false)
  const [showMap, setShowMap] = useState(false)
  const [showComparison, setShowComparison] = useState(false)
  const [show360View, setShow360View] = useState(false)
  const [relatedDestinations, setRelatedDestinations] = useState<Destination[]>([])
  const [activeTab, setActiveTab] = useState<'overview' | 'gallery' | 'map' | '360' | 'reviews'>('overview')

  const destinationId = params?.id ? Number(params.id) : null

  useEffect(() => {
    const loadDestination = async () => {
      if (!destinationId) {
        setError('Invalid destination ID')
        setLoading(false)
        return
      }

      setLoading(true)
      setError('')

      try {
        // Try to load from API, fallback to mock data
        let data: Destination
        try {
          data = await DestinationsService.getDestination(destinationId)
        } catch (apiError) {
          console.warn('API not available, using mock data:', apiError)
          const mockDestinations = DestinationsService.getMockDestinations()
          const mockDestination = mockDestinations.find(d => d.id === destinationId)
          if (!mockDestination) {
            throw new Error('Destination not found')
          }
          data = mockDestination
        }

        setDestination(data)

        // Load related destinations for comparison
        try {
          const mockDestinations = DestinationsService.getMockDestinations()
          const related = mockDestinations
            .filter(d => d.id !== data.id && d.country === data.country)
            .slice(0, 3)
          setRelatedDestinations(related)
        } catch (relatedError) {
          console.warn('Failed to load related destinations:', relatedError)
        }
      } catch (err) {
        setError('Failed to load destination. Please try again.')
        console.error('Error loading destination:', err)
      } finally {
        setLoading(false)
      }
    }

    loadDestination()
  }, [destinationId])

  const handleBookNow = () => {
    if (!user) {
      router.push('/auth/login')
      return
    }
    router.push(`/booking?destination=${destinationId}`)
  }

  const handleShare = async () => {
    if (navigator.share) {
      try {
        await navigator.share({
          title: destination?.name,
          text: destination?.description,
          url: window.location.href,
        })
      } catch (err) {
        console.log('Error sharing:', err)
      }
    } else {
      // Fallback: copy to clipboard
      navigator.clipboard.writeText(window.location.href)
      alert('Link copied to clipboard!')
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  if (error || !destination) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-gray-900 mb-4">Destination Not Found</h1>
          <p className="text-gray-600 mb-6">{error || 'The destination you are looking for does not exist.'}</p>
          <Link
            href="/destinations"
            className="bg-primary text-primary-foreground px-6 py-3 rounded-lg hover:bg-primary/90 transition-colors"
          >
            Browse Destinations
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-white">
      {/* Header */}
      <div className="bg-gray-50 border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <button
              onClick={() => router.back()}
              className="flex items-center text-gray-600 hover:text-primary transition-colors"
            >
              <ArrowLeft className="h-5 w-5 mr-2" />
              Back to destinations
            </button>
            <div className="flex items-center space-x-3">
              <button
                onClick={() => setIsFavorite(!isFavorite)}
                className={`p-2 rounded-full transition-colors ${
                  isFavorite ? 'bg-red-100 text-red-600' : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                }`}
              >
                <Heart className={`h-5 w-5 ${isFavorite ? 'fill-current' : ''}`} />
              </button>
              <button
                onClick={handleShare}
                className="p-2 rounded-full bg-gray-100 text-gray-600 hover:bg-gray-200 transition-colors"
              >
                <Share2 className="h-5 w-5" />
              </button>
              <button
                onClick={() => setShowComparison(true)}
                className="p-2 rounded-full bg-gray-100 text-gray-600 hover:bg-gray-200 transition-colors"
                title="Compare destinations"
              >
                <Layers className="h-5 w-5" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Tab Navigation */}
        <div className="mb-8">
          <div className="border-b border-gray-200">
            <nav className="-mb-px flex space-x-8">
              {[
                { id: 'overview', label: 'Overview', icon: Eye },
                { id: 'gallery', label: 'Gallery', icon: ImageIcon },
                { id: 'map', label: 'Map', icon: Map },
                { id: '360', label: '360° View', icon: TrendingUp },
                { id: 'reviews', label: 'Reviews', icon: Star }
              ].map((tab) => {
                const Icon = tab.icon
                return (
                  <button
                    key={tab.id}
                    onClick={() => setActiveTab(tab.id as 'overview' | 'gallery' | 'map' | '360' | 'reviews')}
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
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Main Content */}
          <div className="lg:col-span-2">
            {/* Tab Content */}
            {activeTab === 'overview' && (
              <div className="space-y-8">
                {/* Hero Image */}
                <div className="aspect-video bg-gray-200 rounded-2xl overflow-hidden">
                  <HeroImage
                    src={destination.images[0] || '/placeholder-destination.jpg'}
                    alt={destination.name}
                    className="w-full h-full object-cover"
                  />
                </div>

            {/* Destination Info */}
            <div className="mb-8">
              <div className="flex items-center text-sm text-gray-500 mb-2">
                <MapPin className="h-4 w-4 mr-1" />
                {destination.city}, {destination.country}
              </div>

              <h1 className="text-3xl md:text-4xl font-bold text-gray-900 mb-4">
                {destination.name}
              </h1>

              <div className="flex items-center space-x-6 mb-6">
                <div className="flex items-center">
                  <Star className="h-5 w-5 text-yellow-400 fill-current mr-1" />
                  <span className="font-medium">4.8</span>
                  <span className="text-gray-500 ml-1">(124 reviews)</span>
                </div>
                <div className="flex items-center text-gray-500">
                  <Clock className="h-5 w-5 mr-1" />
                  {destination.duration} days
                </div>
                <div className="flex items-center text-gray-500">
                  <Users className="h-5 w-5 mr-1" />
                  Up to {destination.max_guests} guests
                </div>
              </div>

              <p className="text-gray-700 text-lg leading-relaxed">
                {destination.description}
              </p>
            </div>

            {/* Features */}
            <div className="mb-8">
              <h2 className="text-2xl font-bold text-gray-900 mb-4">What's Included</h2>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {destination.features.map((feature, index) => (
                  <div key={index} className="flex items-center">
                    <Check className="h-5 w-5 text-green-600 mr-3 flex-shrink-0" />
                    <span className="text-gray-700">{feature}</span>
                  </div>
                ))}
              </div>
            </div>

                {/* Itinerary Preview */}
                <div className="mb-8">
                  <h2 className="text-2xl font-bold text-gray-900 mb-4">Sample Itinerary</h2>
                  <div className="space-y-4">
                    {[
                      { day: 1, title: "Arrival & Welcome", description: "Airport transfer and welcome dinner" },
                      { day: 2, title: "Exploration Begins", description: "Guided tour of main attractions" },
                      { day: 3, title: "Adventure Day", description: "Outdoor activities and cultural experiences" },
                      { day: 4, title: "Relaxation", description: "Spa day and leisure activities" },
                    ].slice(0, Math.min(4, destination.duration)).map((item, index) => (
                      <div key={index} className="flex">
                        <div className="flex-shrink-0 w-12 h-12 bg-primary/10 rounded-full flex items-center justify-center mr-4">
                          <span className="text-primary font-semibold">{item.day}</span>
                        </div>
                        <div>
                          <h3 className="font-semibold text-gray-900">{item.title}</h3>
                          <p className="text-gray-600">{item.description}</p>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            )}

            {/* Interactive Map */}
            {activeTab === 'map' && (
              <div className="mb-8">
                <h2 className="text-2xl font-bold text-gray-900 mb-4">Location</h2>
                <div className="bg-gray-200 rounded-lg h-96 flex items-center justify-center">
                  <div className="text-center text-gray-600">
                    <div className="text-lg font-medium mb-2">Interactive Map</div>
                    <div className="text-sm">Map view coming soon</div>
                  </div>
                </div>
              </div>
            )}

            {/* 360° View */}
            {activeTab === '360' && (
              <div className="mb-8">
                <h2 className="text-2xl font-bold text-gray-900 mb-4">360° Experience</h2>
                <div className="bg-gray-200 rounded-lg h-96 flex items-center justify-center">
                  <div className="text-center text-gray-600">
                    <div className="text-lg font-medium mb-2">360° Panoramic View</div>
                    <div className="text-sm">Immersive view coming soon</div>
                  </div>
                </div>
              </div>
            )}

            {/* Reviews */}
            {activeTab === 'reviews' && (
              <div className="mb-8">
                <ReviewsSection destinationId={destination.id} />
              </div>
            )}
          </div>

          {/* Booking Sidebar */}
          <div className="lg:col-span-1 space-y-6">
            {/* Real-time Indicators */}
            <div className="bg-white border border-gray-200 rounded-2xl p-6">
              <h3 className="font-semibold text-gray-900 mb-4">Live Updates</h3>
              <div className="space-y-3">
                <div className="flex items-center space-x-2">
                  <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse"></div>
                  <span className="text-xs text-gray-500">Live updates</span>
                </div>
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 text-blue-700">
                  <div className="font-medium text-sm">Price Stable</div>
                  <div className="text-xs opacity-75">No recent price changes</div>
                </div>
              </div>
            </div>

            {/* Booking Card */}
            <div className="bg-white border border-gray-200 rounded-2xl p-6 sticky top-8">
              <div className="mb-6">
                <div className="flex items-baseline">
                  <span className="text-3xl font-bold text-gray-900">
                    {formatCurrency(destination.price)}
                  </span>
                  <span className="text-gray-500 ml-2">per person</span>
                </div>
                <p className="text-sm text-gray-500 mt-1">
                  {destination.duration} days • Up to {destination.max_guests} guests
                </p>
              </div>

              <div className="space-y-4 mb-6">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Check-in Date
                  </label>
                  <div className="relative">
                    <Calendar className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
                    <input
                      type="date"
                      className="w-full pl-10 pr-3 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                      min={new Date().toISOString().split('T')[0]}
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Number of Guests
                  </label>
                  <select className="w-full px-3 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent">
                    {Array.from({ length: destination.max_guests }, (_, i) => i + 1).map(num => (
                      <option key={num} value={num}>
                        {num} guest{num !== 1 ? 's' : ''}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <button
                onClick={handleBookNow}
                className="w-full bg-primary text-primary-foreground hover:bg-primary/90 font-semibold py-3 px-6 rounded-lg transition-colors mb-4"
              >
                {user ? 'Book Now' : 'Sign In to Book'}
              </button>

              <p className="text-xs text-gray-500 text-center">
                Free cancellation up to 48 hours before departure
              </p>

              <div className="mt-6 pt-6 border-t border-gray-200">
                <h3 className="font-semibold text-gray-900 mb-3">Need help?</h3>
                <div className="space-y-2 text-sm">
                  <Link href="/contact" className="block text-primary hover:text-primary/80">
                    Contact our travel experts
                  </Link>
                  <Link href="/faq" className="block text-primary hover:text-primary/80">
                    View frequently asked questions
                  </Link>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Reviews Section */}
        <div className="mt-16">
          <ReviewsSection destinationId={destination.id} />
        </div>

        {/* Related Destinations */}
        <div className="mt-16">
          <h2 className="text-2xl font-bold text-gray-900 mb-8">Similar Destinations</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {DestinationsService.getMockDestinations()
              .filter(d => d.id !== destination.id && d.country === destination.country)
              .slice(0, 3)
              .map(dest => (
                <Link key={dest.id} href={`/destinations/${dest.id}`} className="group">
                  <div className="bg-white rounded-lg shadow-sm overflow-hidden hover:shadow-md transition-shadow">
                    <div className="relative h-48">
                      <Image
                        src={dest.images[0]}
                        alt={dest.name}
                        fill
                        className="object-cover group-hover:scale-105 transition-transform duration-300"
                      />
                    </div>
                    <div className="p-4">
                      <h3 className="font-semibold text-gray-900 mb-1">{dest.name}</h3>
                      <p className="text-sm text-gray-500 mb-2">{dest.city}, {dest.country}</p>
                      <p className="text-lg font-bold text-primary">{formatCurrency(dest.price)}</p>
                    </div>
                  </div>
                </Link>
              ))}
          </div>
        </div>
      </div>

      {/* Comparison Modal */}
      {showComparison && relatedDestinations.length > 0 && (
        <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
          <div className="max-w-6xl w-full max-h-[90vh] overflow-y-auto bg-white rounded-lg p-8">
            <div className="text-center">
              <h2 className="text-2xl font-bold text-gray-900 mb-4">Destination Comparison</h2>
              <p className="text-gray-600 mb-6">Compare {destination.name} with similar destinations</p>
              <button
                onClick={() => setShowComparison(false)}
                className="bg-brand-500 text-white px-6 py-2 rounded-lg hover:bg-brand-600 transition-colors"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
