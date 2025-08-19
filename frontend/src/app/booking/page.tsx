'use client'

import { useState, useEffect, Suspense } from 'react'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Calendar, Users, MapPin, Clock, ArrowLeft, CreditCard } from 'lucide-react'
import { useAuth } from '@/contexts/auth-context'
import { DestinationsService } from '@/lib/destinations'
import { BookingsService, CreateBookingRequest } from '@/lib/bookings'
import { bookingSchema, BookingFormData } from '@/lib/validations'
import { Destination } from '@/types'
import { formatCurrency } from '@/lib/utils'
import GuestDetailsForm from '@/components/booking/guest-details-form'
import DateSelector from '@/components/booking/date-selector'
import BookingSummary from '@/components/booking/booking-summary'
import PaymentForm from '@/components/payment/payment-form'
import { useSafeSearchParams } from '@/hooks/use-safe-search-params'

function BookingPageContent() {
  const searchParams = useSafeSearchParams()
  const router = useRouter()
  const { user } = useAuth()

  const [destination, setDestination] = useState<Destination | null>(null)
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [currentStep, setCurrentStep] = useState(1)
  const [availableDates, setAvailableDates] = useState<string[]>([])
  const [destinationId, setDestinationId] = useState<number | null>(null)

  // Initialize destination ID from search params after component mounts
  useEffect(() => {
    if (searchParams) {
      const id = searchParams.get('destination') ? Number(searchParams.get('destination')) : null
      setDestinationId(id)
    }
  }, [searchParams])

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
    trigger,
  } = useForm<BookingFormData>({
    resolver: zodResolver(bookingSchema),
    defaultValues: {
      destination_id: destinationId || 0,
      guests: 2,
      guest_details: [],
      special_requests: '',
    },
  })

  const watchedValues = watch()

  // Load destination data
  useEffect(() => {
    const loadDestination = async () => {
      if (!destinationId) {
        setError('No destination selected')
        setLoading(false)
        return
      }

      try {
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
        setValue('destination_id', data.id)
      } catch (err) {
        setError('Failed to load destination')
        console.error('Error loading destination:', err)
      } finally {
        setLoading(false)
      }
    }

    loadDestination()
  }, [destinationId, setValue])

  // Load available dates when destination changes
  useEffect(() => {
    const loadAvailableDates = async () => {
      if (!destination) return

      try {
        const currentMonth = new Date().toISOString().slice(0, 7)
        let dates: string[]
        try {
          dates = await BookingsService.getAvailableDates(destination.id, currentMonth)
        } catch (apiError) {
          console.warn('API not available, using mock data:', apiError)
          dates = BookingsService.getMockAvailableDates(currentMonth)
        }
        setAvailableDates(dates)
      } catch (err) {
        console.error('Error loading available dates:', err)
      }
    }

    loadAvailableDates()
  }, [destination])

  // Redirect to login if not authenticated
  useEffect(() => {
    if (!loading && !user) {
      router.push('/auth/login')
    }
  }, [user, loading, router])

  const calculateTotalPrice = () => {
    if (!destination || !watchedValues.check_in_date || !watchedValues.check_out_date) {
      return 0
    }

    const checkIn = new Date(watchedValues.check_in_date)
    const checkOut = new Date(watchedValues.check_out_date)
    const nights = Math.ceil((checkOut.getTime() - checkIn.getTime()) / (1000 * 60 * 60 * 24))
    
    return BookingsService.calculateTotalPrice(destination.price, watchedValues.guests, nights)
  }

  const handleNextStep = async () => {
    let fieldsToValidate: (keyof BookingFormData)[] = []

    switch (currentStep) {
      case 1:
        fieldsToValidate = ['check_in_date', 'check_out_date', 'guests']
        break
      case 2:
        fieldsToValidate = ['guest_details']
        break
    }

    const isValid = await trigger(fieldsToValidate)
    if (isValid) {
      setCurrentStep(currentStep + 1)
    }
  }

  const handlePreviousStep = () => {
    setCurrentStep(currentStep - 1)
  }

  const onSubmit = async (data: BookingFormData) => {
    setSubmitting(true)
    setError('')

    try {
      const totalPrice = calculateTotalPrice()
      
      const bookingRequest: CreateBookingRequest = {
        destination_id: data.destination_id,
        check_in_date: data.check_in_date,
        check_out_date: data.check_out_date,
        guests: data.guests,
        guest_details: data.guest_details,
        special_requests: data.special_requests,
        total_price: totalPrice,
      }

      let confirmation
      try {
        confirmation = await BookingsService.createBooking(bookingRequest)
      } catch (apiError) {
        console.warn('API not available, using mock booking:', apiError)
        confirmation = await BookingsService.createMockBooking(bookingRequest)
      }

      // Redirect to confirmation page
      router.push(`/booking/confirmation?booking=${confirmation.booking.id}&confirmation=${confirmation.confirmation_number}`)
    } catch (err: any) {
      setError(err.message || 'Failed to create booking. Please try again.')
    } finally {
      setSubmitting(false)
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
          <h1 className="text-2xl font-bold text-gray-900 mb-4">Booking Error</h1>
          <p className="text-gray-600 mb-6">{error || 'Unable to load booking information'}</p>
          <button
            onClick={() => router.push('/destinations')}
            className="bg-primary text-primary-foreground px-6 py-3 rounded-lg hover:bg-primary/90 transition-colors"
          >
            Browse Destinations
          </button>
        </div>
      </div>
    )
  }

  const totalPrice = calculateTotalPrice()

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <button
              onClick={() => router.back()}
              className="flex items-center text-gray-600 hover:text-primary transition-colors"
            >
              <ArrowLeft className="h-5 w-5 mr-2" />
              Back to destination
            </button>
            <div className="text-sm text-gray-500">
              Step {currentStep} of 3
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Main Content */}
          <div className="lg:col-span-2">
            <div className="bg-white rounded-2xl shadow-sm p-6 mb-6">
              {/* Destination Info */}
              <div className="flex items-center space-x-4 mb-6 pb-6 border-b">
                <div className="w-16 h-16 bg-gray-200 rounded-lg flex-shrink-0">
                  {/* Placeholder for destination image */}
                </div>
                <div>
                  <h1 className="text-xl font-bold text-gray-900">{destination.name}</h1>
                  <div className="flex items-center text-gray-500 mt-1">
                    <MapPin className="h-4 w-4 mr-1" />
                    {destination.city}, {destination.country}
                  </div>
                  <div className="flex items-center text-gray-500 mt-1">
                    <Clock className="h-4 w-4 mr-1" />
                    {destination.duration} days
                  </div>
                </div>
              </div>

              {/* Step Content */}
              <form onSubmit={handleSubmit(onSubmit)}>
                {currentStep === 1 && (
                  <DateSelector
                    destination={destination}
                    availableDates={availableDates}
                    register={register}
                    setValue={setValue}
                    watch={watch}
                    errors={errors}
                  />
                )}

                {currentStep === 2 && (
                  <GuestDetailsForm
                    guests={watchedValues.guests}
                    register={register}
                    setValue={setValue}
                    watch={watch}
                    errors={errors}
                  />
                )}

                {currentStep === 3 && (
                  <div>
                    <h2 className="text-2xl font-bold text-gray-900 mb-6">Review & Payment</h2>

                    {/* Special Requests */}
                    <div className="mb-6">
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Special Requests (Optional)
                      </label>
                      <textarea
                        {...register('special_requests')}
                        rows={4}
                        className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                        placeholder="Any special dietary requirements, accessibility needs, or other requests..."
                      />
                    </div>

                    {/* Payment Form */}
                    <div className="mb-6">
                      <PaymentForm
                        amount={totalPrice}
                        currency="USD"
                        onSuccess={(paymentIntentId) => {
                          console.log('Payment successful:', paymentIntentId)
                          // Continue with booking creation
                          handleSubmit(onSubmit)()
                        }}
                        onError={(error) => {
                          setError(error)
                        }}
                        loading={submitting}
                      />
                    </div>

                    {error && (
                      <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
                        <p className="text-red-700">{error}</p>
                      </div>
                    )}
                  </div>
                )}

                {/* Navigation Buttons */}
                <div className="flex justify-between pt-6 border-t">
                  {currentStep > 1 && (
                    <button
                      type="button"
                      onClick={handlePreviousStep}
                      className="px-6 py-3 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
                    >
                      Previous
                    </button>
                  )}

                  {currentStep < 3 && (
                    <button
                      type="button"
                      onClick={handleNextStep}
                      className="ml-auto px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors"
                    >
                      Next
                    </button>
                  )}

                  {/* Payment is handled by the PaymentForm component in step 3 */}
                </div>
              </form>
            </div>
          </div>

          {/* Booking Summary Sidebar */}
          <div className="lg:col-span-1">
            <BookingSummary
              destination={destination}
              bookingData={watchedValues}
              totalPrice={totalPrice}
            />
          </div>
        </div>
      </div>
    </div>
  )
}

export default function BookingPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    }>
      <BookingPageContent />
    </Suspense>
  )
}
