'use client'

import { useState, useEffect, Suspense } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { CheckCircle, Download, Share2, Calendar, MapPin, Users, CreditCard } from 'lucide-react'
import { PaymentsService } from '@/lib/payments'
import { BookingsService } from '@/lib/bookings'
import { DestinationsService } from '@/lib/destinations'
import { NotificationsService } from '@/lib/notifications'
import { useAuth } from '@/contexts/auth-context'
import { formatCurrency } from '@/lib/utils'
import { useSafeSearchParams } from '@/hooks/use-safe-search-params'

function PaymentSuccessContent() {
  const searchParams = useSafeSearchParams()
  const router = useRouter()
  const { user } = useAuth()

  const [loading, setLoading] = useState(true)
  const [paymentData, setPaymentData] = useState<any>(null)
  const [bookingData, setBookingData] = useState<any>(null)
  const [error, setError] = useState('')
  const [paymentIntentId, setPaymentIntentId] = useState<string | null>(null)
  const [bookingId, setBookingId] = useState<string | null>(null)

  // Initialize from search params after component mounts
  useEffect(() => {
    if (searchParams) {
      setPaymentIntentId(searchParams.get('payment_intent'))
      setBookingId(searchParams.get('booking_id'))
    }
  }, [searchParams])

  useEffect(() => {
    const verifyPaymentAndBooking = async () => {
      if (!paymentIntentId) {
        setError('No payment information found')
        setLoading(false)
        return
      }

      try {
        // Verify payment status
        let paymentConfirmation
        try {
          paymentConfirmation = await PaymentsService.confirmPayment(paymentIntentId)
        } catch (apiError) {
          console.warn('API not available, using mock confirmation:', apiError)
          paymentConfirmation = await PaymentsService.confirmMockPayment(paymentIntentId)
        }

        if (paymentConfirmation.status !== 'succeeded') {
          throw new Error('Payment was not successful')
        }

        setPaymentData(paymentConfirmation)

        // Load booking data if available
        if (bookingId) {
          try {
            const booking = await BookingsService.getBooking(Number(bookingId))
            setBookingData(booking)

            // Load destination data
            const destination = await DestinationsService.getDestination(booking.destination_id)
            setBookingData((prev: any) => ({ ...prev, destination }))

            // Send confirmation email
            if (user?.email) {
              await NotificationsService.sendMockBookingConfirmation(user.email, {
                confirmation_number: `ET${booking.id}${Date.now().toString().slice(-4)}`,
                destination_name: destination.name,
                check_in_date: booking.check_in_date,
                check_out_date: booking.check_out_date,
                guests: booking.guests,
                total_price: booking.total_price,
              })
            }
          } catch (bookingError) {
            console.warn('Could not load booking data:', bookingError)
          }
        }
      } catch (err: any) {
        setError(err.message || 'Failed to verify payment')
      } finally {
        setLoading(false)
      }
    }

    verifyPaymentAndBooking()
  }, [paymentIntentId, bookingId, user])

  const handleDownloadReceipt = () => {
    if (paymentData?.receipt_url) {
      window.open(paymentData.receipt_url, '_blank')
    } else {
      // Generate mock receipt
      alert('Receipt download would start here')
    }
  }

  const handleShareSuccess = async () => {
    const shareData = {
      title: 'Payment Successful - ExoticTravel',
      text: bookingData?.destination 
        ? `I just booked an amazing trip to ${bookingData.destination.name}!`
        : 'I just completed my booking with ExoticTravel!',
      url: window.location.href,
    }

    if (navigator.share) {
      try {
        await navigator.share(shareData)
      } catch (err) {
        console.log('Error sharing:', err)
      }
    } else {
      navigator.clipboard.writeText(shareData.url)
      alert('Link copied to clipboard!')
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
          <p className="text-gray-600">Verifying your payment...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="mx-auto flex items-center justify-center h-16 w-16 rounded-full bg-red-100 mb-4">
            <CreditCard className="h-8 w-8 text-red-600" />
          </div>
          <h1 className="text-2xl font-bold text-gray-900 mb-4">Payment Verification Failed</h1>
          <p className="text-gray-600 mb-6">{error}</p>
          <div className="space-x-4">
            <Link
              href="/destinations"
              className="bg-primary text-primary-foreground px-6 py-3 rounded-lg hover:bg-primary/90 transition-colors"
            >
              Browse Destinations
            </Link>
            <Link
              href="/dashboard"
              className="border border-gray-300 text-gray-700 px-6 py-3 rounded-lg hover:bg-gray-50 transition-colors"
            >
              Go to Dashboard
            </Link>
          </div>
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
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Payment Successful!</h1>
          <p className="text-lg text-gray-600">
            Your payment has been processed successfully.
          </p>
        </div>

        {/* Payment Details */}
        <div className="bg-white rounded-2xl shadow-sm border border-gray-200 p-6 mb-6">
          <div className="flex flex-col md:flex-row md:items-center md:justify-between mb-6">
            <div>
              <h2 className="text-xl font-semibold text-gray-900 mb-1">
                Payment Confirmation
              </h2>
              <p className="text-gray-600">Transaction ID: {paymentData?.payment_intent_id}</p>
            </div>
            <div className="mt-4 md:mt-0 flex space-x-3">
              <button
                onClick={handleDownloadReceipt}
                className="flex items-center px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
              >
                <Download className="h-4 w-4 mr-2" />
                Download Receipt
              </button>
              <button
                onClick={handleShareSuccess}
                className="flex items-center px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
              >
                <Share2 className="h-4 w-4 mr-2" />
                Share
              </button>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            {/* Payment Information */}
            <div>
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Payment Details</h3>
              <div className="space-y-3">
                <div className="flex justify-between">
                  <span className="text-gray-600">Amount Paid</span>
                  <span className="font-semibold text-gray-900">
                    {formatCurrency(paymentData?.amount_received || 0)}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Payment Method</span>
                  <span className="text-gray-900">Credit Card</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Transaction Date</span>
                  <span className="text-gray-900">
                    {new Date(paymentData?.created || Date.now()).toLocaleDateString()}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Status</span>
                  <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800">
                    Completed
                  </span>
                </div>
              </div>
            </div>

            {/* Booking Information */}
            {bookingData && (
              <div>
                <h3 className="text-lg font-semibold text-gray-900 mb-4">Booking Details</h3>
                <div className="space-y-3">
                  {bookingData.destination && (
                    <div>
                      <span className="text-gray-600 block">Destination</span>
                      <span className="font-semibold text-gray-900">{bookingData.destination.name}</span>
                      <div className="flex items-center text-sm text-gray-500 mt-1">
                        <MapPin className="h-4 w-4 mr-1" />
                        {bookingData.destination.city}, {bookingData.destination.country}
                      </div>
                    </div>
                  )}
                  <div className="flex justify-between">
                    <span className="text-gray-600">Check-in</span>
                    <span className="text-gray-900">
                      {new Date(bookingData.check_in_date).toLocaleDateString()}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-600">Check-out</span>
                    <span className="text-gray-900">
                      {new Date(bookingData.check_out_date).toLocaleDateString()}
                    </span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-600">Guests</span>
                    <span className="text-gray-900">{bookingData.guests}</span>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Next Steps */}
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-6 mb-6">
          <h3 className="text-lg font-semibold text-blue-900 mb-4">What's Next?</h3>
          <div className="space-y-3 text-sm text-blue-700">
            <div className="flex items-start">
              <div className="flex-shrink-0 w-6 h-6 bg-blue-200 rounded-full flex items-center justify-center mr-3 mt-0.5">
                <span className="text-xs font-semibold text-blue-900">1</span>
              </div>
              <div>
                <p className="font-medium">Confirmation Email</p>
                <p>You'll receive a detailed confirmation email with your booking details.</p>
              </div>
            </div>
            <div className="flex items-start">
              <div className="flex-shrink-0 w-6 h-6 bg-blue-200 rounded-full flex items-center justify-center mr-3 mt-0.5">
                <span className="text-xs font-semibold text-blue-900">2</span>
              </div>
              <div>
                <p className="font-medium">Travel Documents</p>
                <p>We'll send you detailed travel information 7 days before departure.</p>
              </div>
            </div>
            <div className="flex items-start">
              <div className="flex-shrink-0 w-6 h-6 bg-blue-200 rounded-full flex items-center justify-center mr-3 mt-0.5">
                <span className="text-xs font-semibold text-blue-900">3</span>
              </div>
              <div>
                <p className="font-medium">Customer Support</p>
                <p>Our team is available 24/7 for any questions or assistance you need.</p>
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

        {/* Support */}
        <div className="text-center mt-8 pt-8 border-t border-gray-200">
          <p className="text-gray-600 mb-2">Need help with your booking?</p>
          <Link
            href="/contact"
            className="text-primary hover:text-primary/80 font-medium"
          >
            Contact our support team
          </Link>
        </div>
      </div>
    </div>
  )
}

export default function PaymentSuccessPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    }>
      <PaymentSuccessContent />
    </Suspense>
  )
}
