import api from './api'
import { User } from '@/types'

export interface UpdateProfileRequest {
  first_name?: string
  last_name?: string
  email?: string
  phone?: string
  date_of_birth?: string
  address?: string
  city?: string
  country?: string
  postal_code?: string
  emergency_contact_name?: string
  emergency_contact_phone?: string
  dietary_preferences?: string
  travel_preferences?: string
}

export interface ChangePasswordRequest {
  current_password: string
  new_password: string
  confirm_password: string
}

export interface UserPreferences {
  email_notifications: boolean
  sms_notifications: boolean
  marketing_emails: boolean
  newsletter: boolean
  currency: string
  language: string
  timezone: string
}

export interface UserStats {
  total_bookings: number
  total_spent: number
  destinations_visited: number
  countries_visited: number
  favorite_destination_type: string
  member_since: string
  loyalty_points: number
  next_trip?: {
    destination: string
    date: string
    days_until: number
  }
}

export class UsersService {
  static async updateProfile(userId: number, data: UpdateProfileRequest): Promise<User> {
    const response = await api.patch<User>(`/api/users/${userId}`, data)
    return response.data
  }

  static async changePassword(userId: number, data: ChangePasswordRequest): Promise<void> {
    await api.patch(`/api/users/${userId}/password`, data)
  }

  static async getUserPreferences(userId: number): Promise<UserPreferences> {
    const response = await api.get<UserPreferences>(`/api/users/${userId}/preferences`)
    return response.data
  }

  static async updateUserPreferences(userId: number, preferences: Partial<UserPreferences>): Promise<UserPreferences> {
    const response = await api.patch<UserPreferences>(`/api/users/${userId}/preferences`, preferences)
    return response.data
  }

  static async getUserStats(userId: number): Promise<UserStats> {
    const response = await api.get<UserStats>(`/api/users/${userId}/stats`)
    return response.data
  }

  static async deleteAccount(userId: number, password: string): Promise<void> {
    await api.delete(`/api/users/${userId}`, { data: { password } })
  }

  static async uploadAvatar(userId: number, file: File): Promise<string> {
    const formData = new FormData()
    formData.append('avatar', file)
    
    const response = await api.post<{ avatar_url: string }>(`/api/users/${userId}/avatar`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    
    return response.data.avatar_url
  }

  // Mock data for development
  static getMockUserStats(): UserStats {
    return {
      total_bookings: 8,
      total_spent: 24750,
      destinations_visited: 12,
      countries_visited: 8,
      favorite_destination_type: 'Beach & Islands',
      member_since: '2023-03-15',
      loyalty_points: 2475,
      next_trip: {
        destination: 'Maldives Paradise Resort',
        date: '2024-04-15',
        days_until: 45
      }
    }
  }

  static getMockUserPreferences(): UserPreferences {
    return {
      email_notifications: true,
      sms_notifications: false,
      marketing_emails: true,
      newsletter: true,
      currency: 'USD',
      language: 'en',
      timezone: 'America/New_York'
    }
  }

  static async updateMockProfile(data: UpdateProfileRequest): Promise<User> {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    // Return updated user data (in real app, this would come from server)
    return {
      id: 1,
      email: data.email || 'user@example.com',
      first_name: data.first_name || 'John',
      last_name: data.last_name || 'Doe',
      role: 'user',
      created_at: '2023-03-15T10:00:00Z',
      updated_at: new Date().toISOString()
    }
  }

  static async changeMockPassword(data: ChangePasswordRequest): Promise<void> {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    // In real app, validate current password and update
    if (data.current_password !== 'currentpass') {
      throw new Error('Current password is incorrect')
    }
    
    if (data.new_password !== data.confirm_password) {
      throw new Error('New passwords do not match')
    }
    
    if (data.new_password.length < 8) {
      throw new Error('New password must be at least 8 characters')
    }
  }

  static async updateMockPreferences(preferences: Partial<UserPreferences>): Promise<UserPreferences> {
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 500))
    
    const currentPrefs = this.getMockUserPreferences()
    return { ...currentPrefs, ...preferences }
  }
}
