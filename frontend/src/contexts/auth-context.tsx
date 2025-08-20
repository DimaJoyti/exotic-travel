'use client'

import React, { createContext, useContext, useEffect, useState } from 'react'
import { User } from '@/types'
import { AuthService, LoginCredentials, RegisterCredentials } from '@/lib/auth'

interface AuthContextType {
  user: User | null
  loading: boolean
  isAuthenticated: boolean
  login: (credentials: LoginCredentials) => Promise<void>
  register: (credentials: RegisterCredentials) => Promise<void>
  logout: () => Promise<void>
  refreshUser: () => Promise<void>
  refreshToken: () => Promise<boolean>
  hasPermission: (permission: string) => boolean
  hasRole: (role: string | string[]) => boolean
  updateUser: (userData: Partial<User>) => void
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const [mounted, setMounted] = useState(false)

  const isAuthenticated = mounted && !!user && !!AuthService.getToken()

  useEffect(() => {
    setMounted(true)
    
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

  const refreshToken = async (): Promise<boolean> => {
    try {
      const token = await AuthService.refreshToken()
      if (token) {
        await refreshUser()
        return true
      }
      return false
    } catch (error) {
      console.error('Failed to refresh token:', error)
      return false
    }
  }

  const hasPermission = (permission: string): boolean => {
    if (!user || !user.permissions) {
      return false
    }

    // Check for wildcard permissions
    const hasWildcard = user.permissions.some(perm => {
      const [resource] = perm.split(':')
      return perm === `${resource}:*` && permission.startsWith(`${resource}:`)
    })

    return hasWildcard || user.permissions.includes(permission)
  }

  const hasRole = (role: string | string[]): boolean => {
    if (!user) {
      return false
    }

    const roles = Array.isArray(role) ? role : [role]
    return roles.includes(user.role)
  }

  const updateUser = (userData: Partial<User>) => {
    if (!user) return

    const updatedUser = { ...user, ...userData }
    setUser(updatedUser)
    AuthService.setUser(updatedUser)
  }

  const value: AuthContextType = {
    user,
    loading,
    isAuthenticated,
    login,
    register,
    logout,
    refreshUser,
    refreshToken,
    hasPermission,
    hasRole,
    updateUser,
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
