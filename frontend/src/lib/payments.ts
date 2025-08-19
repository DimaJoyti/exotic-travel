import api from './api'

export interface PaymentIntent {
  id: string
  client_secret: string
  amount: number
  currency: string
  status: string
  metadata?: Record<string, string>
}

export interface CreatePaymentIntentRequest {
  amount: number
  currency: string
  booking_id: number
  customer_email: string
  metadata?: Record<string, string>
}

export interface PaymentMethod {
  id: string
  type: string
  card?: {
    brand: string
    last4: string
    exp_month: number
    exp_year: number
  }
}

export interface PaymentConfirmation {
  payment_intent_id: string
  status: string
  amount_received: number
  receipt_url?: string
  created: number
}

export class PaymentsService {
  static async createPaymentIntent(data: CreatePaymentIntentRequest): Promise<PaymentIntent> {
    const response = await api.post<PaymentIntent>('/api/payments/create-intent', data)
    return response.data
  }

  static async confirmPayment(paymentIntentId: string): Promise<PaymentConfirmation> {
    const response = await api.post<PaymentConfirmation>(`/api/payments/${paymentIntentId}/confirm`)
    return response.data
  }

  static async getPaymentMethods(customerId: string): Promise<PaymentMethod[]> {
    const response = await api.get<PaymentMethod[]>(`/api/payments/customers/${customerId}/payment-methods`)
    return response.data
  }

  static async savePaymentMethod(customerId: string, paymentMethodId: string): Promise<PaymentMethod> {
    const response = await api.post<PaymentMethod>(`/api/payments/customers/${customerId}/payment-methods`, {
      payment_method_id: paymentMethodId
    })
    return response.data
  }

  static async refundPayment(paymentIntentId: string, amount?: number): Promise<any> {
    const response = await api.post(`/api/payments/${paymentIntentId}/refund`, { amount })
    return response.data
  }

  // Mock payment processing for development
  static async createMockPaymentIntent(data: CreatePaymentIntentRequest): Promise<PaymentIntent> {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 1000))

    return {
      id: `pi_mock_${Math.random().toString(36).substring(7)}`,
      client_secret: `pi_mock_${Math.random().toString(36).substring(7)}_secret_${Math.random().toString(36).substring(7)}`,
      amount: data.amount,
      currency: data.currency,
      status: 'requires_payment_method',
      metadata: data.metadata
    }
  }

  static async confirmMockPayment(paymentIntentId: string): Promise<PaymentConfirmation> {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 2000))

    // Simulate random success/failure (90% success rate)
    const isSuccess = Math.random() > 0.1

    if (!isSuccess) {
      throw new Error('Your card was declined. Please try a different payment method.')
    }

    return {
      payment_intent_id: paymentIntentId,
      status: 'succeeded',
      amount_received: 5000, // Mock amount
      receipt_url: `https://pay.stripe.com/receipts/mock_${Math.random().toString(36).substring(7)}`,
      created: Date.now()
    }
  }

  static formatAmount(amount: number, currency: string = 'USD'): string {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency.toUpperCase(),
    }).format(amount / 100) // Stripe amounts are in cents
  }

  static validateCardNumber(cardNumber: string): boolean {
    // Remove spaces and non-digits
    const cleaned = cardNumber.replace(/\D/g, '')
    
    // Check length
    if (cleaned.length < 13 || cleaned.length > 19) {
      return false
    }

    // Luhn algorithm
    let sum = 0
    let isEven = false
    
    for (let i = cleaned.length - 1; i >= 0; i--) {
      let digit = parseInt(cleaned[i])
      
      if (isEven) {
        digit *= 2
        if (digit > 9) {
          digit -= 9
        }
      }
      
      sum += digit
      isEven = !isEven
    }
    
    return sum % 10 === 0
  }

  static validateExpiryDate(month: string, year: string): boolean {
    const currentDate = new Date()
    const currentYear = currentDate.getFullYear()
    const currentMonth = currentDate.getMonth() + 1

    const expMonth = parseInt(month)
    const expYear = parseInt(year)

    if (expMonth < 1 || expMonth > 12) {
      return false
    }

    if (expYear < currentYear) {
      return false
    }

    if (expYear === currentYear && expMonth < currentMonth) {
      return false
    }

    return true
  }

  static validateCVC(cvc: string): boolean {
    const cleaned = cvc.replace(/\D/g, '')
    return cleaned.length >= 3 && cleaned.length <= 4
  }

  static getCardBrand(cardNumber: string): string {
    const cleaned = cardNumber.replace(/\D/g, '')
    
    if (/^4/.test(cleaned)) return 'visa'
    if (/^5[1-5]/.test(cleaned)) return 'mastercard'
    if (/^3[47]/.test(cleaned)) return 'amex'
    if (/^6(?:011|5)/.test(cleaned)) return 'discover'
    
    return 'unknown'
  }

  static formatCardNumber(cardNumber: string): string {
    const cleaned = cardNumber.replace(/\D/g, '')
    const brand = this.getCardBrand(cleaned)
    
    if (brand === 'amex') {
      return cleaned.replace(/(\d{4})(\d{6})(\d{5})/, '$1 $2 $3')
    } else {
      return cleaned.replace(/(\d{4})/g, '$1 ').trim()
    }
  }
}
