'use client'

import { useState } from 'react'
import { Star } from 'lucide-react'

interface StarRatingProps {
  rating: number
  maxRating?: number
  size?: 'sm' | 'md' | 'lg'
  interactive?: boolean
  showValue?: boolean
  onRatingChange?: (rating: number) => void
  className?: string
}

export default function StarRating({
  rating,
  maxRating = 5,
  size = 'md',
  interactive = false,
  showValue = false,
  onRatingChange,
  className = ''
}: StarRatingProps) {
  const [hoverRating, setHoverRating] = useState(0)

  const getSizeClasses = () => {
    switch (size) {
      case 'sm':
        return 'h-4 w-4'
      case 'lg':
        return 'h-6 w-6'
      default:
        return 'h-5 w-5'
    }
  }

  const getTextSize = () => {
    switch (size) {
      case 'sm':
        return 'text-sm'
      case 'lg':
        return 'text-lg'
      default:
        return 'text-base'
    }
  }

  const handleStarClick = (starRating: number) => {
    if (interactive && onRatingChange) {
      onRatingChange(starRating)
    }
  }

  const handleStarHover = (starRating: number) => {
    if (interactive) {
      setHoverRating(starRating)
    }
  }

  const handleMouseLeave = () => {
    if (interactive) {
      setHoverRating(0)
    }
  }

  const displayRating = hoverRating || rating
  const stars = []

  for (let i = 1; i <= maxRating; i++) {
    const isFilled = i <= displayRating
    const isPartiallyFilled = i === Math.ceil(displayRating) && displayRating % 1 !== 0
    
    stars.push(
      <button
        key={i}
        type="button"
        onClick={() => handleStarClick(i)}
        onMouseEnter={() => handleStarHover(i)}
        disabled={!interactive}
        className={`relative ${interactive ? 'cursor-pointer hover:scale-110 transition-transform' : 'cursor-default'} ${getSizeClasses()}`}
      >
        {/* Background star */}
        <Star
          className={`absolute inset-0 ${getSizeClasses()} text-gray-300`}
          fill="currentColor"
        />
        
        {/* Filled star */}
        <Star
          className={`absolute inset-0 ${getSizeClasses()} text-yellow-400 transition-opacity ${
            isFilled ? 'opacity-100' : 'opacity-0'
          }`}
          fill="currentColor"
        />
        
        {/* Partially filled star */}
        {isPartiallyFilled && (
          <div
            className="absolute inset-0 overflow-hidden"
            style={{ width: `${(displayRating % 1) * 100}%` }}
          >
            <Star
              className={`${getSizeClasses()} text-yellow-400`}
              fill="currentColor"
            />
          </div>
        )}
      </button>
    )
  }

  return (
    <div 
      className={`flex items-center space-x-1 ${className}`}
      onMouseLeave={handleMouseLeave}
    >
      <div className="flex items-center space-x-0.5">
        {stars}
      </div>
      
      {showValue && (
        <span className={`ml-2 font-medium text-gray-700 ${getTextSize()}`}>
          {rating.toFixed(1)}
        </span>
      )}
      
      {interactive && hoverRating > 0 && (
        <span className={`ml-2 text-gray-500 ${getTextSize()}`}>
          {hoverRating} star{hoverRating !== 1 ? 's' : ''}
        </span>
      )}
    </div>
  )
}

// Utility component for displaying rating summary
interface RatingSummaryProps {
  rating: number
  totalReviews: number
  size?: 'sm' | 'md' | 'lg'
  className?: string
}

export function RatingSummary({ 
  rating, 
  totalReviews, 
  size = 'md',
  className = '' 
}: RatingSummaryProps) {
  const getTextSize = () => {
    switch (size) {
      case 'sm':
        return 'text-sm'
      case 'lg':
        return 'text-lg'
      default:
        return 'text-base'
    }
  }

  const getRatingText = (rating: number): string => {
    if (rating >= 4.5) return 'Excellent'
    if (rating >= 4.0) return 'Very Good'
    if (rating >= 3.5) return 'Good'
    if (rating >= 3.0) return 'Average'
    if (rating >= 2.0) return 'Poor'
    return 'Terrible'
  }

  return (
    <div className={`flex items-center space-x-2 ${className}`}>
      <StarRating rating={rating} size={size} />
      <span className={`font-semibold text-gray-900 ${getTextSize()}`}>
        {rating.toFixed(1)}
      </span>
      <span className={`text-gray-600 ${getTextSize()}`}>
        ({totalReviews} review{totalReviews !== 1 ? 's' : ''})
      </span>
      <span className={`text-gray-500 ${getTextSize()}`}>
        • {getRatingText(rating)}
      </span>
    </div>
  )
}

// Component for rating distribution bars
interface RatingDistributionProps {
  distribution: {
    5: number
    4: number
    3: number
    2: number
    1: number
  }
  totalReviews: number
  className?: string
}

export function RatingDistribution({ 
  distribution, 
  totalReviews,
  className = '' 
}: RatingDistributionProps) {
  return (
    <div className={`space-y-2 ${className}`}>
      {[5, 4, 3, 2, 1].map((rating) => {
        const count = distribution[rating as keyof typeof distribution]
        const percentage = totalReviews > 0 ? (count / totalReviews) * 100 : 0
        
        return (
          <div key={rating} className="flex items-center space-x-3">
            <div className="flex items-center space-x-1 w-12">
              <span className="text-sm text-gray-600">{rating}</span>
              <Star className="h-3 w-3 text-yellow-400" fill="currentColor" />
            </div>
            
            <div className="flex-1 bg-gray-200 rounded-full h-2">
              <div
                className="bg-yellow-400 h-2 rounded-full transition-all duration-300"
                style={{ width: `${percentage}%` }}
              />
            </div>
            
            <span className="text-sm text-gray-600 w-8 text-right">
              {count}
            </span>
          </div>
        )
      })}
    </div>
  )
}
