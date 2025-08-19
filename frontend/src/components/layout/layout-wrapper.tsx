'use client'

import { usePathname } from 'next/navigation'
import Header from '@/components/navigation/header'
import Footer from '@/components/navigation/footer'
import { useAuth } from '@/contexts/auth-context'

interface LayoutWrapperProps {
  children: React.ReactNode
}

export default function LayoutWrapper({ children }: LayoutWrapperProps) {
  const pathname = usePathname()
  const { user, logout } = useAuth()
  
  // Don't show header/footer on auth pages
  const isAuthPage = pathname?.startsWith('/auth')

  if (isAuthPage) {
    return <>{children}</>
  }

  return (
    <>
      <Header user={user} onLogout={logout} />
      <main className="flex-1">
        {children}
      </main>
      <Footer />
    </>
  )
}
