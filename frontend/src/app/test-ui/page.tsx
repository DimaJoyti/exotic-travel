'use client'

import { WishlistButton } from '@/components/wishlist/wishlist-button'
import { Destination } from '@/types'

export default function TestUIPage() {
  const testDestination: Destination = {
    id: 1,
    name: "Test Destination",
    description: "A test destination for UI testing",
    country: "Test Country",
    city: "Test City",
    price: 1000,
    duration: 7,
    max_guests: 4,
    images: ["https://images.unsplash.com/photo-1506905925346-21bda4d32df4?w=800&h=600&fit=crop"],
    features: ["Test Feature 1", "Test Feature 2"],
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z"
  }

  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-3xl font-bold mb-8">UI Test Page</h1>
        
        <div className="bg-white rounded-lg p-6 shadow-sm">
          <h2 className="text-xl font-semibold mb-4">Wishlist Button Test</h2>
          
          <div className="space-y-6">
            <div>
              <h3 className="text-lg font-medium mb-2">Icon Variants</h3>
              <div className="flex gap-4 items-center">
                <WishlistButton destination={testDestination} variant="icon" size="sm" />
                <WishlistButton destination={testDestination} variant="icon" size="md" />
                <WishlistButton destination={testDestination} variant="icon" size="lg" />
              </div>
            </div>
            
            <div>
              <h3 className="text-lg font-medium mb-2">Button Variants</h3>
              <div className="flex gap-4 items-center">
                <WishlistButton destination={testDestination} variant="button" size="sm" showLabel />
                <WishlistButton destination={testDestination} variant="button" size="md" showLabel />
                <WishlistButton destination={testDestination} variant="button" size="lg" showLabel />
              </div>
            </div>
          </div>
        </div>
        
        <div className="mt-8 bg-green-50 border border-green-200 rounded-lg p-4">
          <p className="text-green-800">
            ✅ If you can see this page without errors, the UI components are working correctly!
          </p>
        </div>
      </div>
    </div>
  )
}
