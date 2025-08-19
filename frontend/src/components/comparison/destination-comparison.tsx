'use client'

import React, { useState, useEffect } from 'react'
import { X, Plus, ArrowRight, Star, MapPin, Clock, Users, DollarSign, Check, Minus } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { Destination } from '@/types'
import { Button } from '@/components/ui/button'
import { FadeIn, StaggerContainer, StaggerItem } from '@/components/ui/animated'

interface DestinationComparisonProps {
  destinations: Destination[]
  onClose?: () => void
  onDestinationSelect?: (destination: Destination) => void
  className?: string
}

interface ComparisonFeature {
  label: string
  key: keyof Destination | 'rating' | 'availability'
  type: 'text' | 'number' | 'currency' | 'array' | 'boolean' | 'rating'
  icon?: React.ReactNode
}

const comparisonFeatures: ComparisonFeature[] = [
  { label: 'Price', key: 'price', type: 'currency', icon: <DollarSign className="h-4 w-4" /> },
  { label: 'Duration', key: 'duration', type: 'number', icon: <Clock className="h-4 w-4" /> },
  { label: 'Max Guests', key: 'max_guests', type: 'number', icon: <Users className="h-4 w-4" /> },
  { label: 'Location', key: 'city', type: 'text', icon: <MapPin className="h-4 w-4" /> },
  { label: 'Country', key: 'country', type: 'text' },
  { label: 'Rating', key: 'rating', type: 'rating', icon: <Star className="h-4 w-4" /> },
  { label: 'Features', key: 'features', type: 'array' },
  { label: 'Availability', key: 'availability', type: 'boolean' }
]

export default function DestinationComparison({
  destinations,
  onClose,
  onDestinationSelect,
  className = ''
}: DestinationComparisonProps) {
  const [selectedDestinations, setSelectedDestinations] = useState<Destination[]>(destinations.slice(0, 3))
  const [highlightedFeature, setHighlightedFeature] = useState<string | null>(null)

  // Limit to 3 destinations for comparison
  const maxDestinations = 3

  const addDestination = (destination: Destination) => {
    if (selectedDestinations.length < maxDestinations && 
        !selectedDestinations.find(d => d.id === destination.id)) {
      setSelectedDestinations([...selectedDestinations, destination])
    }
  }

  const removeDestination = (destinationId: string) => {
    setSelectedDestinations(selectedDestinations.filter(d => d.id.toString() !== destinationId))
  }

  const getFeatureValue = (destination: Destination, feature: ComparisonFeature): any => {
    switch (feature.key) {
      case 'rating':
        return 4.8 // Mock rating
      case 'availability':
        return true // Mock availability
      default:
        return destination[feature.key as keyof Destination]
    }
  }

  const formatFeatureValue = (value: any, type: ComparisonFeature['type']): React.ReactNode => {
    switch (type) {
      case 'currency':
        return `$${value?.toLocaleString() || 0}`
      case 'number':
        return value?.toString() || '0'
      case 'rating':
        return (
          <div className="flex items-center space-x-1">
            <Star className="h-4 w-4 fill-yellow-400 text-yellow-400" />
            <span>{value || '0'}</span>
          </div>
        )
      case 'array':
        return Array.isArray(value) ? value.length : 0
      case 'boolean':
        return value ? (
          <Check className="h-4 w-4 text-green-500" />
        ) : (
          <Minus className="h-4 w-4 text-red-500" />
        )
      default:
        return value?.toString() || '-'
    }
  }

  const getBestValue = (feature: ComparisonFeature): string | null => {
    if (selectedDestinations.length < 2) return null

    const values = selectedDestinations.map(d => getFeatureValue(d, feature))
    
    switch (feature.type) {
      case 'currency':
        const minPrice = Math.min(...values)
        return selectedDestinations.find(d => getFeatureValue(d, feature) === minPrice)?.id.toString() || null
      case 'number':
        if (feature.key === 'duration' || feature.key === 'max_guests') {
          const maxValue = Math.max(...values)
          return selectedDestinations.find(d => getFeatureValue(d, feature) === maxValue)?.id.toString() || null
        }
        break
      case 'rating':
        const maxRating = Math.max(...values)
        return selectedDestinations.find(d => getFeatureValue(d, feature) === maxRating)?.id.toString() || null
      case 'array':
        const maxFeatures = Math.max(...values.map(v => Array.isArray(v) ? v.length : 0))
        return selectedDestinations.find(d => {
          const val = getFeatureValue(d, feature)
          return Array.isArray(val) ? val.length === maxFeatures : false
        })?.id.toString() || null
    }
    
    return null
  }

  return (
    <div className={`bg-white rounded-2xl shadow-2xl overflow-hidden ${className}`}>
      {/* Header */}
      <div className="bg-gradient-to-r from-brand-500 to-accent-500 p-6 text-white">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold">Compare Destinations</h2>
            <p className="text-white/90 mt-1">
              Compare up to {maxDestinations} destinations side by side
            </p>
          </div>
          {onClose && (
            <Button
              variant="ghost"
              size="sm"
              onClick={onClose}
              className="text-white hover:bg-white/20"
            >
              <X className="h-5 w-5" />
            </Button>
          )}
        </div>
      </div>

      {/* Destination Selection */}
      {selectedDestinations.length < maxDestinations && (
        <div className="p-6 border-b border-gray-200">
          <h3 className="text-lg font-semibold mb-4">Add Destinations to Compare</h3>
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
            {destinations
              .filter(d => !selectedDestinations.find(selected => selected.id === d.id))
              .slice(0, 8)
              .map(destination => (
                <motion.div
                  key={destination.id}
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                  onClick={() => addDestination(destination)}
                  className="cursor-pointer bg-gray-50 rounded-lg p-3 hover:bg-gray-100 transition-colors"
                >
                  <div className="aspect-video bg-gray-200 rounded mb-2 overflow-hidden">
                    {destination.images[0] && (
                      <img
                        src={destination.images[0]}
                        alt={destination.name}
                        className="w-full h-full object-cover"
                      />
                    )}
                  </div>
                  <h4 className="font-medium text-sm truncate">{destination.name}</h4>
                  <p className="text-xs text-gray-500">{destination.country}</p>
                </motion.div>
              ))}
          </div>
        </div>
      )}

      {/* Comparison Table */}
      {selectedDestinations.length > 0 && (
        <div className="overflow-x-auto">
          <table className="w-full">
            {/* Destination Headers */}
            <thead>
              <tr className="border-b border-gray-200">
                <th className="text-left p-4 w-48">
                  <div className="font-semibold text-gray-900">Features</div>
                </th>
                {selectedDestinations.map((destination, index) => (
                  <th key={destination.id} className="text-left p-4 min-w-64">
                    <StaggerItem>
                      <div className="relative">
                        {/* Remove Button */}
                        <button
                          onClick={() => removeDestination(destination.id.toString())}
                          className="absolute -top-2 -right-2 w-6 h-6 bg-red-500 text-white rounded-full flex items-center justify-center hover:bg-red-600 transition-colors z-10"
                        >
                          <X className="h-3 w-3" />
                        </button>

                        {/* Destination Image */}
                        <div className="aspect-video bg-gray-200 rounded-lg mb-3 overflow-hidden">
                          {destination.images[0] && (
                            <img
                              src={destination.images[0]}
                              alt={destination.name}
                              className="w-full h-full object-cover"
                            />
                          )}
                        </div>

                        {/* Destination Info */}
                        <h3 className="font-bold text-lg text-gray-900 mb-1">
                          {destination.name}
                        </h3>
                        <p className="text-sm text-gray-500 mb-2">
                          {destination.city}, {destination.country}
                        </p>

                        {/* Select Button */}
                        {onDestinationSelect && (
                          <Button
                            size="sm"
                            onClick={() => onDestinationSelect(destination)}
                            className="w-full"
                            rightIcon={<ArrowRight className="h-4 w-4" />}
                          >
                            View Details
                          </Button>
                        )}
                      </div>
                    </StaggerItem>
                  </th>
                ))}
              </tr>
            </thead>

            {/* Feature Comparison */}
            <tbody>
              {comparisonFeatures.map((feature, featureIndex) => {
                const bestValueId = getBestValue(feature)
                
                return (
                  <motion.tr
                    key={feature.key}
                    initial={{ opacity: 0, x: -20 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: featureIndex * 0.05 }}
                    className={`
                      border-b border-gray-100 hover:bg-gray-50 transition-colors
                      ${highlightedFeature === feature.key ? 'bg-blue-50' : ''}
                    `}
                    onMouseEnter={() => setHighlightedFeature(feature.key)}
                    onMouseLeave={() => setHighlightedFeature(null)}
                  >
                    <td className="p-4 font-medium text-gray-900">
                      <div className="flex items-center space-x-2">
                        {feature.icon}
                        <span>{feature.label}</span>
                      </div>
                    </td>
                    {selectedDestinations.map((destination) => {
                      const value = getFeatureValue(destination, feature)
                      const isBest = bestValueId === destination.id.toString()
                      
                      return (
                        <td
                          key={destination.id}
                          className={`
                            p-4 transition-all
                            ${isBest ? 'bg-green-50 border-l-4 border-green-500' : ''}
                          `}
                        >
                          <div className={`
                            ${isBest ? 'font-semibold text-green-700' : 'text-gray-700'}
                          `}>
                            {feature.type === 'array' && Array.isArray(value) ? (
                              <div className="space-y-1">
                                <div className="font-medium">{value.length} features</div>
                                <div className="text-xs text-gray-500">
                                  {value.slice(0, 3).join(', ')}
                                  {value.length > 3 && ` +${value.length - 3} more`}
                                </div>
                              </div>
                            ) : (
                              formatFeatureValue(value, feature.type)
                            )}
                            {isBest && (
                              <motion.div
                                initial={{ scale: 0 }}
                                animate={{ scale: 1 }}
                                className="inline-flex items-center ml-2 text-green-600"
                              >
                                <Check className="h-3 w-3" />
                              </motion.div>
                            )}
                          </div>
                        </td>
                      )
                    })}
                  </motion.tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}

      {/* Empty State */}
      {selectedDestinations.length === 0 && (
        <div className="p-12 text-center">
          <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <Plus className="h-8 w-8 text-gray-400" />
          </div>
          <h3 className="text-lg font-semibold text-gray-900 mb-2">
            No destinations selected
          </h3>
          <p className="text-gray-500">
            Add destinations above to start comparing their features
          </p>
        </div>
      )}

      {/* Footer */}
      {selectedDestinations.length > 0 && (
        <div className="p-6 bg-gray-50 border-t border-gray-200">
          <div className="flex items-center justify-between">
            <div className="text-sm text-gray-600">
              Comparing {selectedDestinations.length} of {maxDestinations} destinations
            </div>
            <div className="flex items-center space-x-2">
              <div className="flex items-center space-x-1 text-xs text-gray-500">
                <Check className="h-3 w-3 text-green-500" />
                <span>Best value highlighted</span>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
