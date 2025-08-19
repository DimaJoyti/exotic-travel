'use client'

import { useState, useEffect, Suspense } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { XCircle, RefreshCw, CreditCard, AlertTriangle, ArrowLeft } from 'lucide-react'
import { useSafeSearchParams } from '@/hooks/use-safe-search-params'

function PaymentFailedContent() {
  const searchParams = useSafeSearchParams()
  const router = useRouter()

  const [error, setError] = useState('')
  const [paymentIntentId, setPaymentIntentId] = useState('')

  useEffect(() => {
    if (searchParams) {
      const errorMessage = searchParams.get('error') || 'Payment was declined'
      const intentId = searchParams.get('payment_intent') || ''

      setError(errorMessage)
      setPaymentIntentId(intentId)
    }
  }, [searchParams])

  const handleRetryPayment = () => {
    // Go back to the booking page to retry payment
    router.back()
  }

  const getErrorMessage = (error: string) => {
    switch (error.toLowerCase()) {
      case 'card_declined':
        return {
          title: 'Card Declined',
          message: 'Your card was declined. Please try a different payment method or contact your bank.',
          suggestion: 'Check your card details and try again, or use a different card.'
        }
      case 'insufficient_funds':
        return {
          title: 'Insufficient Funds',
          message: 'Your card does not have sufficient funds for this transaction.',
          suggestion: 'Please use a different card or add funds to your account.'
        }
      case 'expired_card':
        return {
          title: 'Expired Card',
          message: 'Your card has expired.',
          suggestion: 'Please use a different card or update your card information.'
        }
      case 'incorrect_cvc':
        return {
          title: 'Incorrect Security Code',
          message: 'The security code (CVC) you entered is incorrect.',
          suggestion: 'Please check the 3-digit code on the back of your card and try again.'
        }
      case 'processing_error':
        return {
          title: 'Processing Error',
          message: 'There was an error processing your payment.',
          suggestion: 'Please try again in a few minutes or contact support if the problem persists.'
        }
      default:
        return {
          title: 'Payment Failed',
          message: error || 'Your payment could not be processed.',
          suggestion: 'Please check your payment details and try again.'
        }
    }
  }

  const errorInfo = getErrorMessage(error)

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Error Header */}
        <div className="text-center mb-8">
          <div className="mx-auto flex items-center justify-center h-16 w-16 rounded-full bg-red-100 mb-4">
            <XCircle className="h-8 w-8 text-red-600" />
          </div>
          <h1 className="text-3xl font-bold text-gray-900 mb-2">{errorInfo.title}</h1>
          <p className="text-lg text-gray-600">
            {errorInfo.message}
          </p>
        </div>

        {/* Error Details */}
        <div className="bg-white rounded-2xl shadow-sm border border-gray-200 p-6 mb-6">
          <div className="flex items-start space-x-4">
            <div className="flex-shrink-0">
              <AlertTriangle className="h-6 w-6 text-yellow-500" />
            </div>
            <div className="flex-1">
              <h3 className="text-lg font-semibold text-gray-900 mb-2">What happened?</h3>
              <p className="text-gray-600 mb-4">{errorInfo.suggestion}</p>
              
              {paymentIntentId && (
                <div className="bg-gray-50 rounded-lg p-4">
                  <p className="text-sm text-gray-600">
                    <span className="font-medium">Transaction ID:</span> {paymentIntentId}
                  </p>
                  <p className="text-xs text-gray-500 mt-1">
                    Please reference this ID when contacting support.
                  </p>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Common Solutions */}
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-6 mb-6">
          <h3 className="text-lg font-semibold text-blue-900 mb-4">Common Solutions</h3>
          <div className="space-y-3 text-sm text-blue-700">
            <div className="flex items-start">
              <div className="flex-shrink-0 w-6 h-6 bg-blue-200 rounded-full flex items-center justify-center mr-3 mt-0.5">
                <span className="text-xs font-semibold text-blue-900">1</span>
              </div>
              <div>
                <p className="font-medium">Check your card details</p>
                <p>Verify that your card number, expiry date, and security code are correct.</p>
              </div>
            </div>
            <div className="flex items-start">
              <div className="flex-shrink-0 w-6 h-6 bg-blue-200 rounded-full flex items-center justify-center mr-3 mt-0.5">
                <span className="text-xs font-semibold text-blue-900">2</span>
              </div>
              <div>
                <p className="font-medium">Try a different card</p>
                <p>Use another credit or debit card if available.</p>
              </div>
            </div>
            <div className="flex items-start">
              <div className="flex-shrink-0 w-6 h-6 bg-blue-200 rounded-full flex items-center justify-center mr-3 mt-0.5">
                <span className="text-xs font-semibold text-blue-900">3</span>
              </div>
              <div>
                <p className="font-medium">Contact your bank</p>
                <p>Your bank may have blocked the transaction for security reasons.</p>
              </div>
            </div>
            <div className="flex items-start">
              <div className="flex-shrink-0 w-6 h-6 bg-blue-200 rounded-full flex items-center justify-center mr-3 mt-0.5">
                <span className="text-xs font-semibold text-blue-900">4</span>
              </div>
              <div>
                <p className="font-medium">Check your internet connection</p>
                <p>Ensure you have a stable internet connection and try again.</p>
              </div>
            </div>
          </div>
        </div>

        {/* Alternative Payment Methods */}
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Alternative Payment Options</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="border border-gray-200 rounded-lg p-4">
              <div className="flex items-center mb-2">
                <CreditCard className="h-5 w-5 text-gray-400 mr-2" />
                <span className="font-medium text-gray-900">Different Card</span>
              </div>
              <p className="text-sm text-gray-600">Try using a different credit or debit card.</p>
            </div>
            <div className="border border-gray-200 rounded-lg p-4">
              <div className="flex items-center mb-2">
                <CreditCard className="h-5 w-5 text-gray-400 mr-2" />
                <span className="font-medium text-gray-900">Bank Transfer</span>
              </div>
              <p className="text-sm text-gray-600">Contact us to arrange a bank transfer payment.</p>
            </div>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex flex-col sm:flex-row gap-4 justify-center mb-8">
          <button
            onClick={handleRetryPayment}
            className="bg-primary text-primary-foreground hover:bg-primary/90 px-6 py-3 rounded-lg font-medium transition-colors flex items-center justify-center"
          >
            <RefreshCw className="h-4 w-4 mr-2" />
            Try Again
          </button>
          <Link
            href="/destinations"
            className="border border-gray-300 text-gray-700 hover:bg-gray-50 px-6 py-3 rounded-lg font-medium transition-colors text-center flex items-center justify-center"
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back to Destinations
          </Link>
        </div>

        {/* Support */}
        <div className="text-center pt-8 border-t border-gray-200">
          <h3 className="text-lg font-semibold text-gray-900 mb-2">Still having trouble?</h3>
          <p className="text-gray-600 mb-4">
            Our support team is here to help you complete your booking.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link
              href="/contact"
              className="text-primary hover:text-primary/80 font-medium"
            >
              Contact Support
            </Link>
            <span className="hidden sm:inline text-gray-300">|</span>
            <a
              href="mailto:support@exotictravel.com"
              className="text-primary hover:text-primary/80 font-medium"
            >
              Email Us
            </a>
            <span className="hidden sm:inline text-gray-300">|</span>
            <a
              href="tel:+1-800-EXOTIC"
              className="text-primary hover:text-primary/80 font-medium"
            >
              Call Us
            </a>
          </div>
        </div>
      </div>
    </div>
  )
}

export default function PaymentFailedPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    }>
      <PaymentFailedContent />
    </Suspense>
  )
}
