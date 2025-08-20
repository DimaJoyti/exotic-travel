'use client'

import React from 'react'
import ContentGenerator from '@/components/marketing/content-generator'

export default function ContentPage() {
  const handleContentGenerated = (content: any) => {
    console.log('Content generated:', content)
    // Handle the generated content (save to database, show success message, etc.)
  }

  return (
    <ContentGenerator
      campaignId={1} // In real implementation, this would come from route params or selection
      onContentGenerated={handleContentGenerated}
    />
  )
}
