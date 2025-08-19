'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { CreditCard, Plus, Trash2, Edit, Shield, Star, ArrowLeft } from 'lucide-react'
import { useAuth } from '@/contexts/auth-context'
import { PaymentsService, PaymentMethod } from '@/lib/payments'

export default function PaymentMethodsPage() {
  const { user } = useAuth()
  const router = useRouter()
  const [paymentMethods, setPaymentMethods] = useState<PaymentMethod[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showAddForm, setShowAddForm] = useState(false)

  useEffect(() => {
    if (!user) {
      router.push('/auth/login')
      return
    }

    loadPaymentMethods()
  }, [user, router])

  const loadPaymentMethods = async () => {
    try {
      // Mock payment methods for demonstration
      const mockMethods: PaymentMethod[] = [
        {
          id: 'pm_1234567890',
          type: 'card',
          card: {
            brand: 'visa',
            last4: '4242',
            exp_month: 12,
            exp_year: 2025,
          },
        },
        {
          id: 'pm_0987654321',
          type: 'card',
          card: {
            brand: 'mastercard',
            last4: '5555',
            exp_month: 8,
            exp_year: 2026,
          },
        },
      ]
      
      setPaymentMethods(mockMethods)
    } catch (err) {
      setError('Failed to load payment methods')
      console.error('Error loading payment methods:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleDeletePaymentMethod = async (paymentMethodId: string) => {
    if (!confirm('Are you sure you want to delete this payment method?')) {
      return
    }

    try {
      // In real implementation, call API to delete payment method
      setPaymentMethods(prev => prev.filter(pm => pm.id !== paymentMethodId))
    } catch (err) {
      setError('Failed to delete payment method')
      console.error('Error deleting payment method:', err)
    }
  }

  const getCardIcon = (brand: string) => {
    switch (brand.toLowerCase()) {
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

  const getCardBrandName = (brand: string) => {
    switch (brand.toLowerCase()) {
      case 'visa':
        return 'Visa'
      case 'mastercard':
        return 'Mastercard'
      case 'amex':
        return 'American Express'
      case 'discover':
        return 'Discover'
      default:
        return brand.charAt(0).toUpperCase() + brand.slice(1)
    }
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white border-b">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center">
            <button
              onClick={() => router.back()}
              className="flex items-center text-gray-600 hover:text-primary transition-colors mr-4"
            >
              <ArrowLeft className="h-5 w-5 mr-2" />
              Back
            </button>
            <h1 className="text-2xl font-bold text-gray-900">Payment Methods</h1>
          </div>
        </div>
      </div>

      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Security Notice */}
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-8">
          <div className="flex items-center">
            <Shield className="h-5 w-5 text-blue-600 mr-2" />
            <div>
              <p className="text-sm text-blue-700 font-medium">Your payment information is secure</p>
              <p className="text-sm text-blue-600">
                We use industry-standard encryption and never store your full card details.
              </p>
            </div>
          </div>
        </div>

        {/* Error Message */}
        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
            <p className="text-red-700">{error}</p>
          </div>
        )}

        {/* Add Payment Method Button */}
        <div className="flex justify-between items-center mb-6">
          <h2 className="text-xl font-semibold text-gray-900">Saved Payment Methods</h2>
          <button
            onClick={() => setShowAddForm(true)}
            className="flex items-center px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors"
          >
            <Plus className="h-4 w-4 mr-2" />
            Add Payment Method
          </button>
        </div>

        {/* Payment Methods List */}
        {paymentMethods.length === 0 ? (
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-12 text-center">
            <CreditCard className="h-16 w-16 text-gray-300 mx-auto mb-4" />
            <h3 className="text-xl font-semibold text-gray-900 mb-2">No payment methods saved</h3>
            <p className="text-gray-600 mb-6">
              Add a payment method to make booking faster and easier.
            </p>
            <button
              onClick={() => setShowAddForm(true)}
              className="bg-primary text-primary-foreground px-6 py-3 rounded-lg hover:bg-primary/90 transition-colors"
            >
              Add Your First Payment Method
            </button>
          </div>
        ) : (
          <div className="space-y-4">
            {paymentMethods.map((method, index) => (
              <div
                key={method.id}
                className="bg-white rounded-lg shadow-sm border border-gray-200 p-6"
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-4">
                    <div className="text-3xl">
                      {getCardIcon(method.card?.brand || '')}
                    </div>
                    <div>
                      <div className="flex items-center space-x-2">
                        <h3 className="text-lg font-semibold text-gray-900">
                          {getCardBrandName(method.card?.brand || '')} •••• {method.card?.last4}
                        </h3>
                        {index === 0 && (
                          <span className="inline-flex items-center px-2 py-1 text-xs font-medium rounded-full bg-yellow-100 text-yellow-800">
                            <Star className="h-3 w-3 mr-1" />
                            Default
                          </span>
                        )}
                      </div>
                      <p className="text-gray-600">
                        Expires {method.card?.exp_month?.toString().padStart(2, '0')}/{method.card?.exp_year}
                      </p>
                    </div>
                  </div>
                  
                  <div className="flex items-center space-x-2">
                    <button
                      onClick={() => {
                        // Handle edit payment method
                        alert('Edit payment method functionality would be implemented here')
                      }}
                      className="p-2 text-gray-400 hover:text-gray-600 transition-colors"
                    >
                      <Edit className="h-4 w-4" />
                    </button>
                    <button
                      onClick={() => handleDeletePaymentMethod(method.id)}
                      className="p-2 text-gray-400 hover:text-red-600 transition-colors"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Add Payment Method Form Modal */}
        {showAddForm && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
            <div className="bg-white rounded-lg max-w-md w-full p-6">
              <div className="flex justify-between items-center mb-4">
                <h3 className="text-lg font-semibold text-gray-900">Add Payment Method</h3>
                <button
                  onClick={() => setShowAddForm(false)}
                  className="text-gray-400 hover:text-gray-600"
                >
                  ×
                </button>
              </div>
              
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Card Number
                  </label>
                  <input
                    type="text"
                    placeholder="1234 5678 9012 3456"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                  />
                </div>
                
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Expiry Date
                    </label>
                    <input
                      type="text"
                      placeholder="MM/YY"
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      CVC
                    </label>
                    <input
                      type="text"
                      placeholder="123"
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                    />
                  </div>
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Cardholder Name
                  </label>
                  <input
                    type="text"
                    placeholder="John Doe"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                  />
                </div>
                
                <div className="flex items-center">
                  <input
                    type="checkbox"
                    id="setDefault"
                    className="h-4 w-4 text-primary focus:ring-primary border-gray-300 rounded"
                  />
                  <label htmlFor="setDefault" className="ml-2 block text-sm text-gray-700">
                    Set as default payment method
                  </label>
                </div>
              </div>
              
              <div className="flex space-x-3 mt-6">
                <button
                  onClick={() => setShowAddForm(false)}
                  className="flex-1 px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
                >
                  Cancel
                </button>
                <button
                  onClick={() => {
                    // Handle add payment method
                    alert('Add payment method functionality would be implemented here')
                    setShowAddForm(false)
                  }}
                  className="flex-1 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors"
                >
                  Add Card
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Payment Security Info */}
        <div className="mt-8 bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Payment Security</h3>
          <div className="space-y-3 text-sm text-gray-600">
            <div className="flex items-start">
              <Shield className="h-4 w-4 text-green-600 mr-2 mt-0.5 flex-shrink-0" />
              <p>All payment information is encrypted using industry-standard SSL technology.</p>
            </div>
            <div className="flex items-start">
              <Shield className="h-4 w-4 text-green-600 mr-2 mt-0.5 flex-shrink-0" />
              <p>We never store your full credit card number or security code.</p>
            </div>
            <div className="flex items-start">
              <Shield className="h-4 w-4 text-green-600 mr-2 mt-0.5 flex-shrink-0" />
              <p>Payment processing is handled by Stripe, a PCI DSS Level 1 certified provider.</p>
            </div>
            <div className="flex items-start">
              <Shield className="h-4 w-4 text-green-600 mr-2 mt-0.5 flex-shrink-0" />
              <p>You can remove saved payment methods at any time.</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
