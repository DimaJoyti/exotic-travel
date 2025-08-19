'use client'

import { useState, useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import Image from 'next/image'
import Link from 'next/link'
import { ArrowLeft, MapPin, Clock, Users, Star, Calendar, Heart, Share2, Check } from 'lucide-react'
import { Destination } from '@/types'
import { DestinationsService } from '@/lib/destinations'
import { formatCurrency } from '@/lib/utils'
import { useAuth } from '@/contexts/auth-context'
import ReviewsSection from '@/components/reviews/reviews-section'
import { ImageGrid } from '@/components/images/image-gallery'
import OptimizedImage, { HeroImage } from '@/components/images/optimized-image'
import { ImagesService } from '@/lib/images'

export default function DestinationDetailPage() {
  const params = useParams()
  const router = useRouter()
  const { user } = useAuth()
  const [destination, setDestination] = useState<Destination | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [selectedImageIndex, setSelectedImageIndex] = useState(0)
  const [isFavorite, setIsFavorite] = useState(false)

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
            <div className="flex items-center space-x-2">
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
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Main Content */}
          <div className="lg:col-span-2">
            {/* Image Gallery */}
            <div className="mb-8">
              <ImageGrid
                images={destination.images.map((image, index) => ({
                  id: `img_${index}`,
                  filename: `${destination.name}_${index + 1}.jpg`,
                  originalName: `${destination.name} Image ${index + 1}`,
                  mimeType: 'image/jpeg',
                  size: 2048576,
                  width: 1920,
                  height: 1080,
                  url: image,
                  thumbnailUrl: image,
                  uploadedAt: new Date().toISOString(),
                  uploadedBy: 'admin',
                  alt: `${destination.name} - Image ${index + 1}`,
                  caption: `Beautiful view of ${destination.name}`,
                  tags: ['destination', destination.country.toLowerCase()]
                }))}
                columns={1}
                showOverlay={true}
                className="mb-4"
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

          {/* Booking Sidebar */}
          <div className="lg:col-span-1">
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
    </div>
  )
}
