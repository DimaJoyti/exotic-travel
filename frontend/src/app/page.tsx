'use client'

import { Suspense, useEffect, useState } from 'react'
import dynamic from 'next/dynamic'

// Dynamically import components to avoid potential webpack issues
const HeroSection = dynamic(() => import('@/components/home/hero-section'), {
  loading: () => <div className="h-96 bg-gray-100 animate-pulse"></div>
})

const FeaturedDestinations = dynamic(() => import('@/components/home/featured-destinations'), {
  loading: () => <div className="h-96 bg-gray-50 animate-pulse"></div>
})

const FeaturesSection = dynamic(() => import('@/components/home/features-section'), {
  loading: () => <div className="h-96 bg-gray-100 animate-pulse"></div>
})

function LoadingSpinner() {
  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-blue-500"></div>
    </div>
  )
}

function HomeContent() {
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) {
    return <LoadingSpinner />
  }

  return (
    <>
      <HeroSection />
      <FeaturedDestinations />
      <FeaturesSection />
    </>
  )
}

export default function Home() {
  return (
    <Suspense fallback={<LoadingSpinner />}>
      <HomeContent />
    </Suspense>
  )
}
