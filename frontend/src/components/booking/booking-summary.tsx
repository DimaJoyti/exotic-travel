'use client'

import Image from 'next/image'
import { MapPin, Calendar, Users, Clock, Shield, CreditCard } from 'lucide-react'
import { Destination } from '@/types'
import { BookingFormData } from '@/lib/validations'
import { formatCurrency } from '@/lib/utils'

interface BookingSummaryProps {
  destination: Destination
  bookingData: Partial<BookingFormData>
  totalPrice: number
}

export default function BookingSummary({
  destination,
  bookingData,
  totalPrice,
}: BookingSummaryProps) {
  const calculateNights = () => {
    if (!bookingData.check_in_date || !bookingData.check_out_date) return 0
    
    const checkIn = new Date(bookingData.check_in_date)
    const checkOut = new Date(bookingData.check_out_date)
    const diffTime = checkOut.getTime() - checkIn.getTime()
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))
    
    return Math.max(0, diffDays)
  }

  const nights = calculateNights()
  const guests = bookingData.guests || 0

  return (
    <div className="bg-white rounded-2xl shadow-sm border border-gray-200 p-6 sticky top-8">
      <h3 className="text-lg font-semibold text-gray-900 mb-6">Booking Summary</h3>
      
      {/* Destination Info */}
      <div className="mb-6">
        <div className="relative h-32 rounded-lg overflow-hidden mb-4">
          <Image
            src={destination.images[0] || '/placeholder-destination.jpg'}
            alt={destination.name}
            fill
            className="object-cover"
          />
        </div>
        
        <h4 className="font-semibold text-gray-900 mb-2">{destination.name}</h4>
        <div className="flex items-center text-sm text-gray-500 mb-1">
          <MapPin className="h-4 w-4 mr-1" />
          {destination.city}, {destination.country}
        </div>
        <div className="flex items-center text-sm text-gray-500">
          <Clock className="h-4 w-4 mr-1" />
          {destination.duration} days experience
        </div>
      </div>

      {/* Booking Details */}
      <div className="space-y-4 mb-6">
        {bookingData.check_in_date && bookingData.check_out_date && (
          <div className="flex items-center justify-between py-3 border-b border-gray-100">
            <div className="flex items-center">
              <Calendar className="h-5 w-5 text-gray-400 mr-3" />
              <div>
                <p className="text-sm font-medium text-gray-900">Dates</p>
                <p className="text-sm text-gray-500">
                  {new Date(bookingData.check_in_date).toLocaleDateString()} - {new Date(bookingData.check_out_date).toLocaleDateString()}
                </p>
              </div>
            </div>
            <div className="text-right">
              <p className="text-sm font-medium text-gray-900">{nights} night{nights !== 1 ? 's' : ''}</p>
            </div>
          </div>
        )}

        {guests > 0 && (
          <div className="flex items-center justify-between py-3 border-b border-gray-100">
            <div className="flex items-center">
              <Users className="h-5 w-5 text-gray-400 mr-3" />
              <div>
                <p className="text-sm font-medium text-gray-900">Guests</p>
                <p className="text-sm text-gray-500">{guests} traveler{guests !== 1 ? 's' : ''}</p>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Price Breakdown */}
      {totalPrice > 0 && (
        <div className="space-y-3 mb-6">
          <h4 className="font-semibold text-gray-900">Price Breakdown</h4>
          
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-gray-600">
                {formatCurrency(destination.price)} × {guests} guest{guests !== 1 ? 's' : ''} × {nights} night{nights !== 1 ? 's' : ''}
              </span>
              <span className="text-gray-900">{formatCurrency(totalPrice)}</span>
            </div>
            
            <div className="flex justify-between">
              <span className="text-gray-600">Service fee</span>
              <span className="text-gray-900">Included</span>
            </div>
            
            <div className="flex justify-between">
              <span className="text-gray-600">Taxes</span>
              <span className="text-gray-900">Included</span>
            </div>
          </div>
          
          <div className="border-t border-gray-200 pt-3">
            <div className="flex justify-between items-center">
              <span className="text-lg font-semibold text-gray-900">Total</span>
              <span className="text-lg font-bold text-primary">{formatCurrency(totalPrice)}</span>
            </div>
          </div>
        </div>
      )}

      {/* What's Included */}
      <div className="mb-6">
        <h4 className="font-semibold text-gray-900 mb-3">What's Included</h4>
        <div className="space-y-2">
          {destination.features.slice(0, 4).map((feature, index) => (
            <div key={index} className="flex items-center text-sm text-gray-600">
              <div className="w-2 h-2 bg-green-500 rounded-full mr-3 flex-shrink-0"></div>
              {feature}
            </div>
          ))}
          {destination.features.length > 4 && (
            <p className="text-sm text-gray-500 ml-5">
              +{destination.features.length - 4} more included
            </p>
          )}
        </div>
      </div>

      {/* Security & Trust */}
      <div className="space-y-3 pt-6 border-t border-gray-200">
        <div className="flex items-center text-sm text-gray-600">
          <Shield className="h-4 w-4 mr-2 text-green-500" />
          Free cancellation up to 48 hours
        </div>
        <div className="flex items-center text-sm text-gray-600">
          <CreditCard className="h-4 w-4 mr-2 text-green-500" />
          Secure payment processing
        </div>
        <div className="flex items-center text-sm text-gray-600">
          <Shield className="h-4 w-4 mr-2 text-green-500" />
          24/7 customer support
        </div>
      </div>

      {/* Contact Support */}
      <div className="mt-6 pt-6 border-t border-gray-200">
        <p className="text-sm text-gray-600 mb-2">Need help with your booking?</p>
        <button className="text-sm text-primary hover:text-primary/80 font-medium">
          Contact our travel experts
        </button>
      </div>
    </div>
  )
}
