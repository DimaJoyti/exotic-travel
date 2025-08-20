'use client'

import React, { useState } from 'react'
import { useRouter } from 'next/navigation'
import CampaignManager from '@/components/marketing/campaign-manager'
import CampaignBuilder from '@/components/marketing/campaign-builder'
import AnalyticsDashboard from '@/components/marketing/analytics-dashboard'
import { Campaign } from '@/lib/marketing-api'

export default function CampaignsPage() {
  const router = useRouter()
  const [currentView, setCurrentView] = useState<'list' | 'create' | 'edit' | 'analytics'>('list')
  const [selectedCampaign, setSelectedCampaign] = useState<Campaign | null>(null)

  const handleCreateCampaign = () => {
    setSelectedCampaign(null)
    setCurrentView('create')
  }

  const handleEditCampaign = (campaign: Campaign) => {
    setSelectedCampaign(campaign)
    setCurrentView('edit')
  }

  const handleViewAnalytics = (campaign: Campaign) => {
    setSelectedCampaign(campaign)
    setCurrentView('analytics')
  }

  const handleSaveCampaign = (campaign: Campaign) => {
    // In real implementation, save to API
    console.log('Saving campaign:', campaign)
    setCurrentView('list')
    setSelectedCampaign(null)
  }

  const handleCancel = () => {
    setCurrentView('list')
    setSelectedCampaign(null)
  }

  switch (currentView) {
    case 'create':
      return (
        <CampaignBuilder
          onSave={handleSaveCampaign}
          onCancel={handleCancel}
        />
      )
    
    case 'edit':
      return (
        <CampaignBuilder
          campaign={selectedCampaign!}
          onSave={handleSaveCampaign}
          onCancel={handleCancel}
        />
      )
    
    case 'analytics':
      return (
        <AnalyticsDashboard
          campaignId={selectedCampaign?.id}
        />
      )
    
    default:
      return (
        <CampaignManager
          onCreateCampaign={handleCreateCampaign}
          onEditCampaign={handleEditCampaign}
          onViewAnalytics={handleViewAnalytics}
        />
      )
  }
}
