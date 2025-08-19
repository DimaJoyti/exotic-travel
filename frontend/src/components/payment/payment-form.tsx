'use client'

import React, { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { CreditCard, Lock, AlertCircle, CheckCircle } from 'lucide-react'
import { PaymentsService } from '@/lib/payments'

const paymentSchema = z.object({
  cardNumber: z
    .string()
    .min(1, 'Card number is required')
    .refine((val) => PaymentsService.validateCardNumber(val), 'Invalid card number'),
  expiryMonth: z
    .string()
    .min(1, 'Expiry month is required')
    .regex(/^(0[1-9]|1[0-2])$/, 'Invalid month'),
  expiryYear: z
    .string()
    .min(1, 'Expiry year is required')
    .regex(/^\d{2}$/, 'Invalid year'),
  cvc: z
    .string()
    .min(1, 'CVC is required')
    .refine((val) => PaymentsService.validateCVC(val), 'Invalid CVC'),
  cardholderName: z
    .string()
    .min(1, 'Cardholder name is required')
    .min(2, 'Name must be at least 2 characters'),
  billingAddress: z.object({
    line1: z.string().min(1, 'Address is required'),
    city: z.string().min(1, 'City is required'),
    state: z.string().min(1, 'State is required'),
    postalCode: z.string().min(1, 'Postal code is required'),
    country: z.string().min(1, 'Country is required'),
  }),
  saveCard: z.boolean().optional(),
})

type PaymentFormData = z.infer<typeof paymentSchema>

interface PaymentFormProps {
  amount: number
  currency: string
  onSuccess: (paymentIntentId: string) => void
  onError: (error: string) => void
  loading?: boolean
}

export default function PaymentForm({
  amount,
  currency,
  onSuccess,
  onError,
  loading = false,
}: PaymentFormProps) {
  const [processing, setProcessing] = useState(false)
  const [cardBrand, setCardBrand] = useState('')

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
    setValue,
  } = useForm<PaymentFormData>({
    resolver: zodResolver(paymentSchema),
    defaultValues: {
      billingAddress: {
        country: 'US',
      },
      saveCard: false,
    },
  })

  const cardNumber = watch('cardNumber')
  const expiryMonth = watch('expiryMonth')
  const expiryYear = watch('expiryYear')

  // Update card brand when card number changes
  React.useEffect(() => {
    if (cardNumber) {
      const brand = PaymentsService.getCardBrand(cardNumber)
      setCardBrand(brand)
    }
  }, [cardNumber])

  const handleCardNumberChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const formatted = PaymentsService.formatCardNumber(e.target.value)
    setValue('cardNumber', formatted)
  }

  const onSubmit = async (data: PaymentFormData) => {
    setProcessing(true)

    try {
      // Validate expiry date
      if (!PaymentsService.validateExpiryDate(data.expiryMonth, `20${data.expiryYear}`)) {
        throw new Error('Card has expired or invalid expiry date')
      }

      // Create payment intent
      const paymentIntent = await PaymentsService.createMockPaymentIntent({
        amount: amount * 100, // Convert to cents
        currency,
        booking_id: 1, // This would come from props
        customer_email: 'user@example.com', // This would come from auth context
        metadata: {
          cardholder_name: data.cardholderName,
          card_brand: cardBrand,
        },
      })

      // Simulate payment confirmation
      const confirmation = await PaymentsService.confirmMockPayment(paymentIntent.id)

      if (confirmation.status === 'succeeded') {
        onSuccess(confirmation.payment_intent_id)
      } else {
        throw new Error('Payment failed. Please try again.')
      }
    } catch (error: any) {
      onError(error.message || 'Payment processing failed')
    } finally {
      setProcessing(false)
    }
  }

  const getCardIcon = () => {
    switch (cardBrand) {
      case 'visa':
        return '💳'
      case 'mastercard':
        return '💳'
      case 'amex':
        return '💳'
      case 'discover':
        return '💳'
      default:
        return '💳'
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {/* Security Notice */}
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <div className="flex items-center">
          <Lock className="h-5 w-5 text-blue-600 mr-2" />
          <p className="text-sm text-blue-700">
            Your payment information is encrypted and secure. We use industry-standard SSL encryption.
          </p>
        </div>
      </div>

      {/* Card Information */}
      <div className="space-y-4">
        <h3 className="text-lg font-semibold text-gray-900">Payment Information</h3>
        
        {/* Card Number */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Card Number
          </label>
          <div className="relative">
            <input
              type="text"
              {...register('cardNumber')}
              onChange={handleCardNumberChange}
              placeholder="1234 5678 9012 3456"
              maxLength={19}
              className={`w-full pl-4 pr-12 py-3 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent ${
                errors.cardNumber ? 'border-red-300' : 'border-gray-300'
              }`}
            />
            <div className="absolute inset-y-0 right-0 pr-3 flex items-center">
              <span className="text-2xl">{getCardIcon()}</span>
            </div>
          </div>
          {errors.cardNumber && (
            <p className="mt-1 text-sm text-red-600">{errors.cardNumber.message}</p>
          )}
        </div>

        {/* Expiry and CVC */}
        <div className="grid grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Month
            </label>
            <select
              {...register('expiryMonth')}
              className={`w-full px-3 py-3 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent ${
                errors.expiryMonth ? 'border-red-300' : 'border-gray-300'
              }`}
            >
              <option value="">MM</option>
              {Array.from({ length: 12 }, (_, i) => {
                const month = (i + 1).toString().padStart(2, '0')
                return (
                  <option key={month} value={month}>
                    {month}
                  </option>
                )
              })}
            </select>
            {errors.expiryMonth && (
              <p className="mt-1 text-sm text-red-600">{errors.expiryMonth.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Year
            </label>
            <select
              {...register('expiryYear')}
              className={`w-full px-3 py-3 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent ${
                errors.expiryYear ? 'border-red-300' : 'border-gray-300'
              }`}
            >
              <option value="">YY</option>
              {Array.from({ length: 10 }, (_, i) => {
                const year = (new Date().getFullYear() + i).toString().slice(-2)
                return (
                  <option key={year} value={year}>
                    {year}
                  </option>
                )
              })}
            </select>
            {errors.expiryYear && (
              <p className="mt-1 text-sm text-red-600">{errors.expiryYear.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              CVC
            </label>
            <input
              type="text"
              {...register('cvc')}
              placeholder="123"
              maxLength={4}
              className={`w-full px-3 py-3 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent ${
                errors.cvc ? 'border-red-300' : 'border-gray-300'
              }`}
            />
            {errors.cvc && (
              <p className="mt-1 text-sm text-red-600">{errors.cvc.message}</p>
            )}
          </div>
        </div>

        {/* Cardholder Name */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Cardholder Name
          </label>
          <input
            type="text"
            {...register('cardholderName')}
            placeholder="John Doe"
            className={`w-full px-3 py-3 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent ${
              errors.cardholderName ? 'border-red-300' : 'border-gray-300'
            }`}
          />
          {errors.cardholderName && (
            <p className="mt-1 text-sm text-red-600">{errors.cardholderName.message}</p>
          )}
        </div>
      </div>

      {/* Billing Address */}
      <div className="space-y-4">
        <h3 className="text-lg font-semibold text-gray-900">Billing Address</h3>
        
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Address
          </label>
          <input
            type="text"
            {...register('billingAddress.line1')}
            placeholder="123 Main Street"
            className={`w-full px-3 py-3 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent ${
              errors.billingAddress?.line1 ? 'border-red-300' : 'border-gray-300'
            }`}
          />
          {errors.billingAddress?.line1 && (
            <p className="mt-1 text-sm text-red-600">{errors.billingAddress.line1.message}</p>
          )}
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              City
            </label>
            <input
              type="text"
              {...register('billingAddress.city')}
              placeholder="New York"
              className={`w-full px-3 py-3 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent ${
                errors.billingAddress?.city ? 'border-red-300' : 'border-gray-300'
              }`}
            />
            {errors.billingAddress?.city && (
              <p className="mt-1 text-sm text-red-600">{errors.billingAddress.city.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              State
            </label>
            <input
              type="text"
              {...register('billingAddress.state')}
              placeholder="NY"
              className={`w-full px-3 py-3 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent ${
                errors.billingAddress?.state ? 'border-red-300' : 'border-gray-300'
              }`}
            />
            {errors.billingAddress?.state && (
              <p className="mt-1 text-sm text-red-600">{errors.billingAddress.state.message}</p>
            )}
          </div>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Postal Code
            </label>
            <input
              type="text"
              {...register('billingAddress.postalCode')}
              placeholder="10001"
              className={`w-full px-3 py-3 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent ${
                errors.billingAddress?.postalCode ? 'border-red-300' : 'border-gray-300'
              }`}
            />
            {errors.billingAddress?.postalCode && (
              <p className="mt-1 text-sm text-red-600">{errors.billingAddress.postalCode.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Country
            </label>
            <select
              {...register('billingAddress.country')}
              className={`w-full px-3 py-3 border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent ${
                errors.billingAddress?.country ? 'border-red-300' : 'border-gray-300'
              }`}
            >
              <option value="US">United States</option>
              <option value="CA">Canada</option>
              <option value="GB">United Kingdom</option>
              <option value="AU">Australia</option>
              <option value="DE">Germany</option>
              <option value="FR">France</option>
            </select>
            {errors.billingAddress?.country && (
              <p className="mt-1 text-sm text-red-600">{errors.billingAddress.country.message}</p>
            )}
          </div>
        </div>
      </div>

      {/* Save Card Option */}
      <div className="flex items-center">
        <input
          type="checkbox"
          {...register('saveCard')}
          className="h-4 w-4 text-primary focus:ring-primary border-gray-300 rounded"
        />
        <label className="ml-2 block text-sm text-gray-700">
          Save this card for future purchases
        </label>
      </div>

      {/* Submit Button */}
      <button
        type="submit"
        disabled={processing || loading}
        className="w-full bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed font-semibold py-3 px-6 rounded-lg transition-colors flex items-center justify-center"
      >
        {processing ? (
          <>
            <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
            Processing Payment...
          </>
        ) : (
          <>
            <Lock className="h-4 w-4 mr-2" />
            Pay {PaymentsService.formatAmount(amount * 100, currency)}
          </>
        )}
      </button>

      {/* Security Footer */}
      <div className="text-center text-xs text-gray-500">
        <p>🔒 Secured by 256-bit SSL encryption</p>
      </div>
    </form>
  )
}
