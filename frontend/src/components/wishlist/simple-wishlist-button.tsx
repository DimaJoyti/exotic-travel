'use client'

import React, { useState, useEffect } from 'react'
import { Heart, Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'

interface SimpleWishlistButtonProps {
  destinationId: string | number
  size?: 'sm' | 'md' | 'lg'
  className?: string
  onToggle?: (isInWishlist: boolean) => void
}

export function SimpleWishlistButton({
  destinationId,
  size = 'md',
  className = '',
  onToggle
}: SimpleWishlistButtonProps) {
  const [isInWishlist, setIsInWishlist] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [mounted, setMounted] = useState(false)

  const sizeClasses = {
    sm: 'w-8 h-8',
    md: 'w-10 h-10',
    lg: 'w-12 h-12'
  }

  const iconSizes = {
    sm: 'h-4 w-4',
    md: 'h-5 w-5',
    lg: 'h-6 w-6'
  }

  useEffect(() => {
    setMounted(true)
    
    // Check localStorage for wishlist status
    if (typeof window !== 'undefined') {
      try {
        const wishlist = localStorage.getItem('exotic_travel_wishlist')
        if (wishlist) {
          const items = JSON.parse(wishlist)
          const isInList = items.some((item: any) => 
            item.destination_id === destinationId.toString()
          )
          setIsInWishlist(isInList)
        }
      } catch (error) {
        console.error('Error checking wishlist:', error)
      }
    }
  }, [destinationId])

  const handleToggle = async () => {
    if (isLoading || !mounted) return

    setIsLoading(true)

    try {
      if (typeof window !== 'undefined') {
        const wishlist = localStorage.getItem('exotic_travel_wishlist')
        let items = wishlist ? JSON.parse(wishlist) : []

        if (isInWishlist) {
          // Remove from wishlist
          items = items.filter((item: any) => 
            item.destination_id !== destinationId.toString()
          )
          setIsInWishlist(false)
        } else {
          // Add to wishlist
          const newItem = {
            id: `wishlist_${Date.now()}`,
            destination_id: destinationId.toString(),
            added_at: new Date().toISOString()
          }
          items.push(newItem)
          setIsInWishlist(true)
        }

        localStorage.setItem('exotic_travel_wishlist', JSON.stringify(items))
        onToggle?.(!isInWishlist)
      }
    } catch (error) {
      console.error('Error toggling wishlist:', error)
    } finally {
      setIsLoading(false)
    }
  }

  if (!mounted) {
    return (
      <div className={cn(
        sizeClasses[size],
        "bg-white/90 backdrop-blur-sm border border-white/30 rounded-full flex items-center justify-center shadow-lg",
        className
      )}>
        <Heart className={cn(iconSizes[size], "text-gray-400")} />
      </div>
    )
  }

  return (
    <button
      onClick={handleToggle}
      disabled={isLoading}
      aria-label={isInWishlist ? "Remove from wishlist" : "Add to wishlist"}
      className={cn(
        sizeClasses[size],
        "relative flex items-center justify-center rounded-full",
        "bg-white/90 backdrop-blur-sm border border-white/30",
        "hover:bg-white transition-all duration-200 shadow-lg",
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-red-500 focus-visible:ring-offset-2",
        isInWishlist ? "text-red-500" : "text-gray-600 hover:text-red-500",
        className
      )}
    >
      {isLoading ? (
        <Loader2 className={cn(iconSizes[size], "animate-spin")} />
      ) : (
        <Heart 
          className={cn(
            iconSizes[size],
            isInWishlist ? "fill-current" : "",
            "transition-colors duration-200"
          )} 
        />
      )}
    </button>
  )
}

export default SimpleWishlistButton
