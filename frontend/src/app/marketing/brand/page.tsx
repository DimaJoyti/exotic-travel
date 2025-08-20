'use client'

import React, { useState } from 'react'
import BrandManager from '@/components/marketing/brand-manager'
import { Brand } from '@/lib/marketing-api'

export default function BrandPage() {
  const [selectedBrand, setSelectedBrand] = useState<Brand | null>(null)

  const handleSaveBrand = (brand: Brand) => {
    console.log('Saving brand:', brand)
    // In real implementation, save to API
  }

  const handleCancel = () => {
    setSelectedBrand(null)
  }

  return (
    <BrandManager
      brandId={1} // In real implementation, this would come from route params or selection
      onSave={handleSaveBrand}
      onCancel={handleCancel}
    />
  )
}
