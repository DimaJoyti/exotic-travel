import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'
import ClientProviders from '@/components/providers/client-providers'
import LayoutWrapper from '@/components/layout/layout-wrapper'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'ExoticTravel - Discover Extraordinary Adventures',
  description: 'Book your next exotic adventure with us. Discover unique destinations, luxury accommodations, and unforgettable experiences around the world.',
  keywords: 'exotic travel, adventure booking, luxury travel, unique destinations, travel experiences',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" suppressHydrationWarning={true}>
      <body className={inter.className}>
        <ClientProviders>
          <div className="min-h-screen bg-background flex flex-col">
            <LayoutWrapper>
              {children}
            </LayoutWrapper>
          </div>
        </ClientProviders>
      </body>
    </html>
  )
}
