'use client'

import { useEffect } from 'react'
import { User, Mail, Phone, Calendar, CreditCard, Utensils } from 'lucide-react'
import { UseFormRegister, UseFormSetValue, UseFormWatch, FieldErrors } from 'react-hook-form'
import { BookingFormData, GuestDetailFormData } from '@/lib/validations'

interface GuestDetailsFormProps {
  guests: number
  register: UseFormRegister<BookingFormData>
  setValue: UseFormSetValue<BookingFormData>
  watch: UseFormWatch<BookingFormData>
  errors: FieldErrors<BookingFormData>
}

export default function GuestDetailsForm({
  guests,
  register,
  setValue,
  watch,
  errors,
}: GuestDetailsFormProps) {
  const watchedGuestDetails = watch('guest_details') || []

  // Initialize guest details array when guests number changes
  useEffect(() => {
    const currentDetails = watchedGuestDetails
    const newDetails = Array.from({ length: guests }, (_, index) => {
      return currentDetails[index] || {
        first_name: '',
        last_name: '',
        email: '',
        phone: '',
        date_of_birth: '',
        passport_number: '',
        dietary_requirements: '',
      }
    })
    
    setValue('guest_details', newDetails)
  }, [guests, setValue, watchedGuestDetails])

  const updateGuestDetail = (index: number, field: keyof GuestDetailFormData, value: string) => {
    const currentDetails = [...watchedGuestDetails]
    if (!currentDetails[index]) {
      currentDetails[index] = {
        first_name: '',
        last_name: '',
        email: '',
        phone: '',
        date_of_birth: '',
        passport_number: '',
        dietary_requirements: '',
      }
    }
    currentDetails[index] = { ...currentDetails[index], [field]: value }
    setValue('guest_details', currentDetails)
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-6">Guest Details</h2>
      
      <div className="space-y-8">
        {Array.from({ length: guests }, (_, index) => (
          <div key={index} className="border border-gray-200 rounded-lg p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">
              Guest {index + 1} {index === 0 && '(Primary Contact)'}
            </h3>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* First Name */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  First Name *
                </label>
                <div className="relative">
                  <User className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
                  <input
                    type="text"
                    value={watchedGuestDetails[index]?.first_name || ''}
                    onChange={(e) => updateGuestDetail(index, 'first_name', e.target.value)}
                    className={`w-full pl-10 pr-3 py-3 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent ${
                      errors.guest_details?.[index]?.first_name ? 'border-red-300' : 'border-gray-300'
                    }`}
                    placeholder="Enter first name"
                  />
                </div>
                {errors.guest_details?.[index]?.first_name && (
                  <p className="mt-1 text-sm text-red-600">
                    {errors.guest_details[index]?.first_name?.message}
                  </p>
                )}
              </div>

              {/* Last Name */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Last Name *
                </label>
                <div className="relative">
                  <User className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
                  <input
                    type="text"
                    value={watchedGuestDetails[index]?.last_name || ''}
                    onChange={(e) => updateGuestDetail(index, 'last_name', e.target.value)}
                    className={`w-full pl-10 pr-3 py-3 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent ${
                      errors.guest_details?.[index]?.last_name ? 'border-red-300' : 'border-gray-300'
                    }`}
                    placeholder="Enter last name"
                  />
                </div>
                {errors.guest_details?.[index]?.last_name && (
                  <p className="mt-1 text-sm text-red-600">
                    {errors.guest_details[index]?.last_name?.message}
                  </p>
                )}
              </div>

              {/* Email */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Email Address *
                </label>
                <div className="relative">
                  <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
                  <input
                    type="email"
                    value={watchedGuestDetails[index]?.email || ''}
                    onChange={(e) => updateGuestDetail(index, 'email', e.target.value)}
                    className={`w-full pl-10 pr-3 py-3 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent ${
                      errors.guest_details?.[index]?.email ? 'border-red-300' : 'border-gray-300'
                    }`}
                    placeholder="Enter email address"
                  />
                </div>
                {errors.guest_details?.[index]?.email && (
                  <p className="mt-1 text-sm text-red-600">
                    {errors.guest_details[index]?.email?.message}
                  </p>
                )}
              </div>

              {/* Phone */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Phone Number
                </label>
                <div className="relative">
                  <Phone className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
                  <input
                    type="tel"
                    value={watchedGuestDetails[index]?.phone || ''}
                    onChange={(e) => updateGuestDetail(index, 'phone', e.target.value)}
                    className="w-full pl-10 pr-3 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                    placeholder="Enter phone number"
                  />
                </div>
              </div>

              {/* Date of Birth */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Date of Birth
                </label>
                <div className="relative">
                  <Calendar className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
                  <input
                    type="date"
                    value={watchedGuestDetails[index]?.date_of_birth || ''}
                    onChange={(e) => updateGuestDetail(index, 'date_of_birth', e.target.value)}
                    max={new Date().toISOString().split('T')[0]}
                    className="w-full pl-10 pr-3 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                  />
                </div>
              </div>

              {/* Passport Number */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Passport Number
                </label>
                <div className="relative">
                  <CreditCard className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
                  <input
                    type="text"
                    value={watchedGuestDetails[index]?.passport_number || ''}
                    onChange={(e) => updateGuestDetail(index, 'passport_number', e.target.value)}
                    className="w-full pl-10 pr-3 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                    placeholder="Enter passport number"
                  />
                </div>
              </div>
            </div>

            {/* Dietary Requirements */}
            <div className="mt-4">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Dietary Requirements & Allergies
              </label>
              <div className="relative">
                <Utensils className="absolute left-3 top-3 text-gray-400 h-5 w-5" />
                <textarea
                  value={watchedGuestDetails[index]?.dietary_requirements || ''}
                  onChange={(e) => updateGuestDetail(index, 'dietary_requirements', e.target.value)}
                  rows={3}
                  className="w-full pl-10 pr-3 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                  placeholder="Please specify any dietary requirements, food allergies, or special meal preferences..."
                />
              </div>
            </div>
          </div>
        ))}

        {/* Important Information */}
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <h3 className="text-sm font-medium text-blue-900 mb-2">Important Information</h3>
          <ul className="text-sm text-blue-700 space-y-1">
            <li>• All guest names must match government-issued ID for travel</li>
            <li>• Passport information may be required for international destinations</li>
            <li>• Dietary requirements will be shared with accommodation and tour providers</li>
            <li>• Primary contact will receive all booking confirmations and updates</li>
          </ul>
        </div>

        {/* Validation Errors */}
        {errors.guest_details && typeof errors.guest_details.message === 'string' && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4">
            <p className="text-sm text-red-700">{errors.guest_details.message}</p>
          </div>
        )}
      </div>
    </div>
  )
}
