import api from './api'
import { Booking, BookingStatus } from '@/types'

export interface CreateBookingRequest {
  destination_id: number
  check_in_date: string
  check_out_date: string
  guests: number
  guest_details: GuestDetail[]
  special_requests?: string
  total_price: number
}

export interface GuestDetail {
  first_name: string
  last_name: string
  email: string
  phone?: string
  date_of_birth?: string
  passport_number?: string
  dietary_requirements?: string
}

export interface BookingConfirmation {
  booking: Booking
  confirmation_number: string
  payment_intent_id?: string
}

export class BookingsService {
  static async createBooking(bookingData: CreateBookingRequest): Promise<BookingConfirmation> {
    const response = await api.post<BookingConfirmation>('/api/bookings', bookingData)
    return response.data
  }

  static async getBooking(id: number): Promise<Booking> {
    const response = await api.get<Booking>(`/api/bookings/${id}`)
    return response.data
  }

  static async getUserBookings(userId: number): Promise<Booking[]> {
    const response = await api.get<Booking[]>(`/api/users/${userId}/bookings`)
    return response.data
  }

  static async updateBooking(id: number, updates: Partial<Booking>): Promise<Booking> {
    const response = await api.patch<Booking>(`/api/bookings/${id}`, updates)
    return response.data
  }

  static async cancelBooking(id: number, reason?: string): Promise<Booking> {
    const response = await api.patch<Booking>(`/api/bookings/${id}/cancel`, { reason })
    return response.data
  }

  static async getAvailableDates(destinationId: number, month: string): Promise<string[]> {
    const response = await api.get<string[]>(`/api/destinations/${destinationId}/availability?month=${month}`)
    return response.data
  }

  static calculateTotalPrice(basePrice: number, guests: number, duration: number): number {
    return basePrice * guests * duration
  }

  static generateConfirmationNumber(): string {
    const prefix = 'ET'
    const timestamp = Date.now().toString().slice(-6)
    const random = Math.random().toString(36).substring(2, 6).toUpperCase()
    return `${prefix}${timestamp}${random}`
  }

  // Mock booking creation for development
  static async createMockBooking(bookingData: CreateBookingRequest): Promise<BookingConfirmation> {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 2000))

    const booking: Booking = {
      id: Math.floor(Math.random() * 10000),
      user_id: 1, // Mock user ID
      destination_id: bookingData.destination_id,
      check_in_date: bookingData.check_in_date,
      check_out_date: bookingData.check_out_date,
      guests: bookingData.guests,
      total_price: bookingData.total_price,
      status: BookingStatus.CONFIRMED,
      special_requests: bookingData.special_requests,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    }

    const confirmation: BookingConfirmation = {
      booking,
      confirmation_number: this.generateConfirmationNumber(),
      payment_intent_id: `pi_mock_${Math.random().toString(36).substring(7)}`
    }

    return confirmation
  }

  // Mock available dates (exclude some random dates to simulate bookings)
  static getMockAvailableDates(month: string): string[] {
    const year = new Date().getFullYear()
    const monthNum = parseInt(month.split('-')[1]) - 1
    const daysInMonth = new Date(year, monthNum + 1, 0).getDate()
    
    const availableDates: string[] = []
    const today = new Date()
    
    for (let day = 1; day <= daysInMonth; day++) {
      const date = new Date(year, monthNum, day)
      
      // Only include future dates
      if (date >= today) {
        // Randomly exclude some dates to simulate bookings (20% chance)
        if (Math.random() > 0.2) {
          availableDates.push(date.toISOString().split('T')[0])
        }
      }
    }
    
    return availableDates
  }
}
