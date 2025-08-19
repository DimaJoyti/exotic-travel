'use client'

import React, { createContext, useContext, useEffect, useState } from 'react'
import { User } from '@/types'
import { AuthService, LoginCredentials, RegisterCredentials } from '@/lib/auth'

interface AuthContextType {
  user: User | null
  loading: boolean
  login: (credentials: LoginCredentials) => Promise<void>
  register: (credentials: RegisterCredentials) => Promise<void>
  logout: () => Promise<void>
  refreshUser: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const initAuth = async () => {
      try {
        const token = AuthService.getToken()
        if (token) {
          const storedUser = AuthService.getUser()
          if (storedUser) {
            setUser(storedUser)
            // Optionally refresh user data from server
            try {
              const currentUser = await AuthService.getCurrentUser()
              setUser(currentUser)
            } catch (error) {
              // If refresh fails, keep stored user or logout
              console.error('Failed to refresh user:', error)
            }
          }
        }
      } catch (error) {
        console.error('Auth initialization error:', error)
        AuthService.removeToken()
      } finally {
        setLoading(false)
      }
    }

    initAuth()
  }, [])

  const login = async (credentials: LoginCredentials) => {
    try {
      const { user: authUser } = await AuthService.login(credentials)
      setUser(authUser)
    } catch (error) {
      throw error
    }
  }

  const register = async (credentials: RegisterCredentials) => {
    try {
      const { user: authUser } = await AuthService.register(credentials)
      setUser(authUser)
    } catch (error) {
      throw error
    }
  }

  const logout = async () => {
    try {
      await AuthService.logout()
      setUser(null)
    } catch (error) {
      console.error('Logout error:', error)
      // Still clear local state even if server call fails
      setUser(null)
    }
  }

  const refreshUser = async () => {
    try {
      const currentUser = await AuthService.getCurrentUser()
      setUser(currentUser)
    } catch (error) {
      console.error('Failed to refresh user:', error)
      throw error
    }
  }

  const value: AuthContextType = {
    user,
    loading,
    login,
    register,
    logout,
    refreshUser,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
