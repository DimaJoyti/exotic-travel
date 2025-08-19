'use client'

import React from 'react'
import { Metadata } from 'next'
import InnovationDashboard from '@/components/advanced/innovation-dashboard'
import { FadeIn } from '@/components/ui/animated'

export default function InnovationPage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-white">
      <div className="container mx-auto px-4 py-8">
        <FadeIn>
          <InnovationDashboard />
        </FadeIn>
      </div>
    </div>
  )
}
