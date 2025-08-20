'use client'

import React, { useState } from 'react'
import IntegrationManager from '@/components/marketing/integration-manager'

interface Integration {
  id: number
  platform: string
  accountId: string
  accountName?: string
  status: 'active' | 'inactive' | 'error' | 'expired'
  lastSync?: string
  connectedAt: string
  features: string[]
  metrics?: {
    campaigns: number
    spend: number
    impressions: number
  }
}

export default function IntegrationsPage() {
  const [integrations, setIntegrations] = useState<Integration[]>([])

  const handleIntegrationChange = (updatedIntegrations: Integration[]) => {
    setIntegrations(updatedIntegrations)
    // In real implementation, you might want to sync this with global state
    console.log('Integrations updated:', updatedIntegrations)
  }

  return (
    <IntegrationManager
      onIntegrationChange={handleIntegrationChange}
    />
  )
}
