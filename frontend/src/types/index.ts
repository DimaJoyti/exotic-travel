export interface User {
  id: number
  email: string
  first_name: string
  last_name: string
  role: string
  created_at: string
  updated_at: string
}

export interface Destination {
  id: number
  name: string
  description: string
  country: string
  city: string
  price: number
  duration: number
  max_guests: number
  images: string[]
  features: string[]
  created_at: string
  updated_at: string
}

export enum BookingStatus {
  PENDING = 'pending',
  CONFIRMED = 'confirmed',
  CANCELLED = 'cancelled',
  COMPLETED = 'completed'
}

export interface Booking {
  id: number
  user_id: number
  destination_id: number
  check_in_date: string
  check_out_date: string
  guests: number
  total_price: number
  status: BookingStatus | string
  payment_status?: string
  payment_intent_id?: string
  special_requests?: string
  created_at: string
  updated_at: string
  user?: User
  destination?: Destination
}

export interface Review {
  id: number
  user_id: number
  destination_id: number
  booking_id?: number
  rating: number
  title: string
  comment: string
  pros?: string[]
  cons?: string[]
  travel_date: string
  verified_booking: boolean
  helpful_count: number
  created_at: string
  updated_at: string
  user?: User
  destination?: Destination
}

export interface ReviewStats {
  average_rating: number
  total_reviews: number
  rating_distribution: {
    5: number
    4: number
    3: number
    2: number
    1: number
  }
}

export interface Review {
  id: number
  user_id: number
  destination_id: number
  rating: number
  comment: string
  created_at: string
  user?: User
}

export interface AuthResponse {
  token: string
  user: User
}

export interface ApiError {
  message: string
  code?: string
  details?: any
}
