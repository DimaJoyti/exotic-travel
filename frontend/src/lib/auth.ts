import api from './api'
import { User, AuthResponse } from '@/types'

export interface LoginCredentials {
  email: string
  password: string
}

export interface RegisterCredentials {
  email: string
  password: string
  first_name: string
  last_name: string
}

export class AuthService {
  private static TOKEN_KEY = 'auth_token'
  private static USER_KEY = 'auth_user'

  static getToken(): string | null {
    if (typeof window === 'undefined') return null
    return localStorage.getItem(this.TOKEN_KEY)
  }

  static setToken(token: string): void {
    if (typeof window === 'undefined') return
    localStorage.setItem(this.TOKEN_KEY, token)
  }

  static removeToken(): void {
    if (typeof window === 'undefined') return
    localStorage.removeItem(this.TOKEN_KEY)
    localStorage.removeItem(this.USER_KEY)
  }

  static getUser(): User | null {
    if (typeof window === 'undefined') return null
    const userStr = localStorage.getItem(this.USER_KEY)
    if (!userStr) return null
    try {
      return JSON.parse(userStr)
    } catch {
      return null
    }
  }

  static setUser(user: User): void {
    if (typeof window === 'undefined') return
    localStorage.setItem(this.USER_KEY, JSON.stringify(user))
  }

  static async login(credentials: LoginCredentials): Promise<AuthResponse> {
    const response = await api.post<AuthResponse>('/api/auth/login', credentials)
    const { token, user } = response.data
    
    this.setToken(token)
    this.setUser(user)
    
    return response.data
  }

  static async register(credentials: RegisterCredentials): Promise<AuthResponse> {
    const response = await api.post<AuthResponse>('/api/auth/register', credentials)
    const { token, user } = response.data
    
    this.setToken(token)
    this.setUser(user)
    
    return response.data
  }

  static async logout(): Promise<void> {
    this.removeToken()
    // Optionally call logout endpoint if needed
    // await api.post('/api/auth/logout')
  }

  static async refreshToken(): Promise<string> {
    const response = await api.post<{ token: string }>('/api/auth/refresh')
    const { token } = response.data
    this.setToken(token)
    return token
  }

  static async getCurrentUser(): Promise<User> {
    const response = await api.get<User>('/api/auth/me')
    const user = response.data
    this.setUser(user)
    return user
  }

  static isAuthenticated(): boolean {
    return !!this.getToken()
  }
}
