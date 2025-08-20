'use client'

import { ReactNode } from 'react'
import { AuthProvider } from '@/contexts/auth-context'
import ErrorBoundary from '@/components/error-boundary'

interface ClientProvidersProps {
  children: ReactNode
}

export function ClientProviders({ children }: ClientProvidersProps) {
  return (
    <ErrorBoundary>
      <AuthProvider>
        {children}
      </AuthProvider>
    </ErrorBoundary>
  )
}

export default ClientProviders
