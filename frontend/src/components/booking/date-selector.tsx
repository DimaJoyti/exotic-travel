'use client'

import { useState, useEffect } from 'react'
import { Calendar, Users, Plus, Minus } from 'lucide-react'
import { UseFormRegister, UseFormSetValue, UseFormWatch, FieldErrors } from 'react-hook-form'
import { BookingFormData } from '@/lib/validations'
import { Destination } from '@/types'

interface DateSelectorProps {
  destination: Destination
  availableDates: string[]
  register: UseFormRegister<BookingFormData>
  setValue: UseFormSetValue<BookingFormData>
  watch: UseFormWatch<BookingFormData>
  errors: FieldErrors<BookingFormData>
}

export default function DateSelector({
  destination,
  availableDates,
  register,
  setValue,
  watch,
  errors,
}: DateSelectorProps) {
  const [selectedCheckIn, setSelectedCheckIn] = useState('')
  const [selectedCheckOut, setSelectedCheckOut] = useState('')
  const [guests, setGuests] = useState(2)

  const watchedCheckIn = watch('check_in_date')
  const watchedCheckOut = watch('check_out_date')
  const watchedGuests = watch('guests')

  useEffect(() => {
    if (watchedCheckIn) setSelectedCheckIn(watchedCheckIn)
    if (watchedCheckOut) setSelectedCheckOut(watchedCheckOut)
    if (watchedGuests) setGuests(watchedGuests)
  }, [watchedCheckIn, watchedCheckOut, watchedGuests])

  const handleCheckInChange = (date: string) => {
    setSelectedCheckIn(date)
    setValue('check_in_date', date)
    
    // Auto-set checkout date based on destination duration
    if (date) {
      const checkInDate = new Date(date)
      const checkOutDate = new Date(checkInDate)
      checkOutDate.setDate(checkInDate.getDate() + destination.duration)
      
      const checkOutString = checkOutDate.toISOString().split('T')[0]
      setSelectedCheckOut(checkOutString)
      setValue('check_out_date', checkOutString)
    }
  }

  const handleCheckOutChange = (date: string) => {
    setSelectedCheckOut(date)
    setValue('check_out_date', date)
  }

  const handleGuestsChange = (newGuests: number) => {
    if (newGuests >= 1 && newGuests <= destination.max_guests) {
      setGuests(newGuests)
      setValue('guests', newGuests)
    }
  }

  const isDateAvailable = (date: string) => {
    return availableDates.includes(date)
  }

  const getMinDate = () => {
    const today = new Date()
    return today.toISOString().split('T')[0]
  }

  const getMinCheckOutDate = () => {
    if (!selectedCheckIn) return getMinDate()
    
    const checkInDate = new Date(selectedCheckIn)
    checkInDate.setDate(checkInDate.getDate() + 1)
    return checkInDate.toISOString().split('T')[0]
  }

  const calculateNights = () => {
    if (!selectedCheckIn || !selectedCheckOut) return 0
    
    const checkIn = new Date(selectedCheckIn)
    const checkOut = new Date(selectedCheckOut)
    const diffTime = checkOut.getTime() - checkIn.getTime()
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))
    
    return Math.max(0, diffDays)
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-6">Select Dates & Guests</h2>
      
      <div className="space-y-6">
        {/* Date Selection */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Check-in Date
            </label>
            <div className="relative">
              <Calendar className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
              <input
                type="date"
                {...register('check_in_date')}
                value={selectedCheckIn}
                onChange={(e) => handleCheckInChange(e.target.value)}
                min={getMinDate()}
                className={`w-full pl-10 pr-3 py-3 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent ${
                  errors.check_in_date ? 'border-red-300' : 'border-gray-300'
                }`}
              />
            </div>
            {errors.check_in_date && (
              <p className="mt-1 text-sm text-red-600">{errors.check_in_date.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Check-out Date
            </label>
            <div className="relative">
              <Calendar className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
              <input
                type="date"
                {...register('check_out_date')}
                value={selectedCheckOut}
                onChange={(e) => handleCheckOutChange(e.target.value)}
                min={getMinCheckOutDate()}
                className={`w-full pl-10 pr-3 py-3 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent ${
                  errors.check_out_date ? 'border-red-300' : 'border-gray-300'
                }`}
              />
            </div>
            {errors.check_out_date && (
              <p className="mt-1 text-sm text-red-600">{errors.check_out_date.message}</p>
            )}
          </div>
        </div>

        {/* Duration Display */}
        {selectedCheckIn && selectedCheckOut && (
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-blue-900">Trip Duration</p>
                <p className="text-lg font-bold text-blue-900">
                  {calculateNights()} night{calculateNights() !== 1 ? 's' : ''}
                </p>
              </div>
              <div className="text-right">
                <p className="text-sm text-blue-700">
                  {new Date(selectedCheckIn).toLocaleDateString()} - {new Date(selectedCheckOut).toLocaleDateString()}
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Guest Selection */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Number of Guests
          </label>
          <div className="flex items-center space-x-4">
            <div className="flex items-center border border-gray-300 rounded-lg">
              <button
                type="button"
                onClick={() => handleGuestsChange(guests - 1)}
                disabled={guests <= 1}
                className="p-3 text-gray-500 hover:text-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <Minus className="h-4 w-4" />
              </button>
              
              <div className="px-4 py-3 border-x border-gray-300">
                <div className="flex items-center space-x-2">
                  <Users className="h-5 w-5 text-gray-400" />
                  <span className="font-medium">{guests}</span>
                </div>
              </div>
              
              <button
                type="button"
                onClick={() => handleGuestsChange(guests + 1)}
                disabled={guests >= destination.max_guests}
                className="p-3 text-gray-500 hover:text-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <Plus className="h-4 w-4" />
              </button>
            </div>
            
            <div className="text-sm text-gray-500">
              Maximum {destination.max_guests} guests
            </div>
          </div>
          {errors.guests && (
            <p className="mt-1 text-sm text-red-600">{errors.guests.message}</p>
          )}
        </div>

        {/* Availability Notice */}
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
          <h3 className="text-sm font-medium text-yellow-900 mb-2">Availability Notice</h3>
          <p className="text-sm text-yellow-700">
            Dates are subject to availability. We'll confirm your booking within 24 hours.
            {destination.duration > 1 && (
              <span className="block mt-1">
                This is a {destination.duration}-day experience. Your check-out date has been automatically set.
              </span>
            )}
          </p>
        </div>

        {/* Pricing Preview */}
        {selectedCheckIn && selectedCheckOut && guests > 0 && (
          <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
            <h3 className="text-sm font-medium text-gray-900 mb-3">Pricing Breakdown</h3>
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-gray-600">
                  ${destination.price.toLocaleString()} × {guests} guest{guests !== 1 ? 's' : ''} × {calculateNights()} night{calculateNights() !== 1 ? 's' : ''}
                </span>
                <span className="font-medium">
                  ${(destination.price * guests * calculateNights()).toLocaleString()}
                </span>
              </div>
              <div className="flex justify-between text-base font-semibold pt-2 border-t">
                <span>Total</span>
                <span>${(destination.price * guests * calculateNights()).toLocaleString()}</span>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
