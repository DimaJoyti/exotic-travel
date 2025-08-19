'use client'

import { useState, useEffect, Suspense } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { CheckCircle, Calendar, Users, MapPin, Download, Share2, Mail } from 'lucide-react'
import { useAuth } from '@/contexts/auth-context'
import { BookingsService } from '@/lib/bookings'
import { DestinationsService } from '@/lib/destinations'
import { Booking, Destination } from '@/types'
import { formatCurrency } from '@/lib/utils'
import { useSafeSearchParams } from '@/hooks/use-safe-search-params'

function BookingConfirmationContent() {
  const searchParams = useSafeSearchParams()
  const router = useRouter()
  const { user } = useAuth()

  const [booking, setBooking] = useState<Booking | null>(null)
  const [destination, setDestination] = useState<Destination | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [bookingId, setBookingId] = useState<number | null>(null)
  const [confirmationNumber, setConfirmationNumber] = useState('')

  // Initialize from search params after component mounts
  useEffect(() => {
    if (searchParams) {
      try {
        const id = searchParams.get('booking') ? Number(searchParams.get('booking')) : null
        const confirmation = searchParams.get('confirmation') || ''
        setBookingId(id)
        setConfirmationNumber(confirmation)
      } catch (error) {
        console.error('Error reading search params:', error)
        setError('Invalid booking parameters')
      }
    }
  }, [searchParams])

  useEffect(() => {
    const loadBookingData = async () => {
      if (!bookingId || !confirmationNumber) {
        setError('Invalid booking information')
        setLoading(false)
        return
      }

      try {
        // For demo purposes, create mock booking data
        const mockBooking: Booking = {
          id: bookingId,
          user_id: user?.id || 1,
          destination_id: 1, // Default to first destination
          check_in_date: new Date().toISOString().split('T')[0],
          check_out_date: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
          guests: 2,
          total_price: 5000,
          status: 'confirmed' as any,
          special_requests: '',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        }

        setBooking(mockBooking)

        // Load destination data
        try {
          const destinationData = await DestinationsService.getDestination(mockBooking.destination_id)
          setDestination(destinationData)
        } catch (apiError) {
          console.warn('API not available, using mock data:', apiError)
          try {
            const mockDestinations = DestinationsService.getMockDestinations()
            const mockDestination = mockDestinations.find(d => d.id === mockBooking.destination_id)
            if (mockDestination) {
              setDestination(mockDestination)
            }
          } catch (mockError) {
            console.error('Error loading mock destinations:', mockError)
          }
        }
      } catch (err) {
        setError('Failed to load booking information')
        console.error('Error loading booking:', err)
      } finally {
        setLoading(false)
      }
    }

    loadBookingData()
  }, [bookingId, confirmationNumber, user])

  const handleDownloadConfirmation = () => {
    // In a real app, this would generate and download a PDF
    alert('Confirmation PDF download would start here')
  }

  const handleShareBooking = async () => {
    if (navigator.share) {
      try {
        await navigator.share({
          title: 'My Exotic Travel Booking',
          text: `I just booked an amazing trip to ${destination?.name}!`,
          url: window.location.href,
        })
      } catch (err) {
        console.log('Error sharing:', err)
      }
    } else {
      navigator.clipboard.writeText(window.location.href)
      alert('Booking link copied to clipboard!')
    }
  }

  const calculateNights = () => {
    if (!booking) return 0
    
    const checkIn = new Date(booking.check_in_date)
    const checkOut = new Date(booking.check_out_date)
    const diffTime = checkOut.getTime() - checkIn.getTime()
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))
    
    return Math.max(0, diffDays)
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  if (error || !booking || !destination) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-gray-900 mb-4">Booking Not Found</h1>
          <p className="text-gray-600 mb-6">{error || 'Unable to load booking information'}</p>
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
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Success Header */}
        <div className="text-center mb-8">
          <div className="mx-auto flex items-center justify-center h-16 w-16 rounded-full bg-green-100 mb-4">
            <CheckCircle className="h-8 w-8 text-green-600" />
          </div>
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Booking Confirmed!</h1>
          <p className="text-lg text-gray-600">
            Your exotic adventure awaits. We've sent a confirmation email with all the details.
          </p>
        </div>

        {/* Confirmation Details */}
        <div className="bg-white rounded-2xl shadow-sm border border-gray-200 p-6 mb-6">
          <div className="flex flex-col md:flex-row md:items-center md:justify-between mb-6">
            <div>
              <h2 className="text-xl font-semibold text-gray-900 mb-1">
                Confirmation Number
              </h2>
              <p className="text-2xl font-bold text-primary">{confirmationNumber}</p>
            </div>
            <div className="mt-4 md:mt-0 flex space-x-3">
              <button
                onClick={handleDownloadConfirmation}
                className="flex items-center px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
              >
                <Download className="h-4 w-4 mr-2" />
                Download PDF
              </button>
              <button
                onClick={handleShareBooking}
                className="flex items-center px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
              >
                <Share2 className="h-4 w-4 mr-2" />
                Share
              </button>
            </div>
          </div>

          {/* Booking Summary */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
            {/* Trip Details */}
            <div>
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Trip Details</h3>
              
              <div className="space-y-4">
                <div>
                  <h4 className="font-medium text-gray-900 mb-1">{destination.name}</h4>
                  <div className="flex items-center text-gray-500 text-sm">
                    <MapPin className="h-4 w-4 mr-1" />
                    {destination.city}, {destination.country}
                  </div>
                </div>

                <div className="flex items-center justify-between py-3 border-t border-gray-100">
                  <div className="flex items-center">
                    <Calendar className="h-5 w-5 text-gray-400 mr-3" />
                    <div>
                      <p className="text-sm font-medium text-gray-900">Check-in</p>
                      <p className="text-sm text-gray-500">
                        {new Date(booking.check_in_date).toLocaleDateString('en-US', {
                          weekday: 'long',
                          year: 'numeric',
                          month: 'long',
                          day: 'numeric'
                        })}
                      </p>
                    </div>
                  </div>
                </div>

                <div className="flex items-center justify-between py-3 border-t border-gray-100">
                  <div className="flex items-center">
                    <Calendar className="h-5 w-5 text-gray-400 mr-3" />
                    <div>
                      <p className="text-sm font-medium text-gray-900">Check-out</p>
                      <p className="text-sm text-gray-500">
                        {new Date(booking.check_out_date).toLocaleDateString('en-US', {
                          weekday: 'long',
                          year: 'numeric',
                          month: 'long',
                          day: 'numeric'
                        })}
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="text-sm font-medium text-gray-900">
                      {calculateNights()} night{calculateNights() !== 1 ? 's' : ''}
                    </p>
                  </div>
                </div>

                <div className="flex items-center justify-between py-3 border-t border-gray-100">
                  <div className="flex items-center">
                    <Users className="h-5 w-5 text-gray-400 mr-3" />
                    <div>
                      <p className="text-sm font-medium text-gray-900">Guests</p>
                      <p className="text-sm text-gray-500">
                        {booking.guests} traveler{booking.guests !== 1 ? 's' : ''}
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* Payment Summary */}
            <div>
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Payment Summary</h3>
              
              <div className="space-y-3">
                <div className="flex justify-between text-sm">
                  <span className="text-gray-600">
                    {formatCurrency(destination.price)} × {booking.guests} guest{booking.guests !== 1 ? 's' : ''} × {calculateNights()} night{calculateNights() !== 1 ? 's' : ''}
                  </span>
                  <span className="text-gray-900">{formatCurrency(booking.total_price)}</span>
                </div>
                
                <div className="flex justify-between text-sm">
                  <span className="text-gray-600">Service fee</span>
                  <span className="text-gray-900">Included</span>
                </div>
                
                <div className="flex justify-between text-sm">
                  <span className="text-gray-600">Taxes</span>
                  <span className="text-gray-900">Included</span>
                </div>
                
                <div className="border-t border-gray-200 pt-3">
                  <div className="flex justify-between items-center">
                    <span className="text-lg font-semibold text-gray-900">Total Paid</span>
                    <span className="text-lg font-bold text-primary">{formatCurrency(booking.total_price)}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Next Steps */}
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-6 mb-6">
          <h3 className="text-lg font-semibold text-blue-900 mb-4">What happens next?</h3>
          <div className="space-y-3 text-sm text-blue-700">
            <div className="flex items-start">
              <div className="flex-shrink-0 w-6 h-6 bg-blue-200 rounded-full flex items-center justify-center mr-3 mt-0.5">
                <span className="text-xs font-semibold text-blue-900">1</span>
              </div>
              <div>
                <p className="font-medium">Confirmation Email</p>
                <p>You'll receive a detailed confirmation email within the next few minutes.</p>
              </div>
            </div>
            <div className="flex items-start">
              <div className="flex-shrink-0 w-6 h-6 bg-blue-200 rounded-full flex items-center justify-center mr-3 mt-0.5">
                <span className="text-xs font-semibold text-blue-900">2</span>
              </div>
              <div>
                <p className="font-medium">Travel Documents</p>
                <p>We'll send you detailed travel information and itinerary 7 days before departure.</p>
              </div>
            </div>
            <div className="flex items-start">
              <div className="flex-shrink-0 w-6 h-6 bg-blue-200 rounded-full flex items-center justify-center mr-3 mt-0.5">
                <span className="text-xs font-semibold text-blue-900">3</span>
              </div>
              <div>
                <p className="font-medium">24/7 Support</p>
                <p>Our travel experts are available anytime if you have questions or need assistance.</p>
              </div>
            </div>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <Link
            href="/dashboard"
            className="bg-primary text-primary-foreground hover:bg-primary/90 px-6 py-3 rounded-lg font-medium transition-colors text-center"
          >
            View My Bookings
          </Link>
          <Link
            href="/destinations"
            className="border border-gray-300 text-gray-700 hover:bg-gray-50 px-6 py-3 rounded-lg font-medium transition-colors text-center"
          >
            Book Another Trip
          </Link>
        </div>

        {/* Contact Support */}
        <div className="text-center mt-8 pt-8 border-t border-gray-200">
          <p className="text-gray-600 mb-2">Questions about your booking?</p>
          <div className="flex items-center justify-center space-x-4">
            <Link
              href="/contact"
              className="text-primary hover:text-primary/80 font-medium"
            >
              Contact Support
            </Link>
            <span className="text-gray-300">|</span>
            <a
              href="mailto:support@exotictravel.com"
              className="text-primary hover:text-primary/80 font-medium flex items-center"
            >
              <Mail className="h-4 w-4 mr-1" />
              Email Us
            </a>
          </div>
        </div>
      </div>
    </div>
  )
}

export default function BookingConfirmationPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    }>
      <BookingConfirmationContent />
    </Suspense>
  )
}
