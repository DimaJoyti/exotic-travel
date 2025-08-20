'use client'

import React, { useState, useEffect } from 'react'
import { Heart, Plus, Check, Loader2 } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { Destination } from '@/types'
import { WishlistService, WishlistItem } from '@/lib/wishlist'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface WishlistButtonProps {
  destination: Destination
  size?: 'sm' | 'md' | 'lg'
  variant?: 'icon' | 'button'
  showLabel?: boolean
  className?: string
  onWishlistChange?: (isInWishlist: boolean, item?: WishlistItem) => void
}

export function WishlistButton({
  destination,
  size = 'md',
  variant = 'icon',
  showLabel = false,
  className = '',
  onWishlistChange
}: WishlistButtonProps) {
  const [isInWishlist, setIsInWishlist] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [showSuccess, setShowSuccess] = useState(false)
  const [wishlistItem, setWishlistItem] = useState<WishlistItem | null>(null)

  // Check if destination is in wishlist
  useEffect(() => {
    if (!destination?.id) return

    const checkWishlistStatus = () => {
      if (typeof window !== 'undefined') {
        try {
          const inWishlist = WishlistService.isInWishlist(destination.id.toString())
          const item = WishlistService.getWishlistItem(destination.id.toString())
          setIsInWishlist(inWishlist)
          setWishlistItem(item)
        } catch (error) {
          console.error('Error checking wishlist status:', error)
        }
      }
    }

    checkWishlistStatus()

    // Listen for wishlist changes
    const unsubscribe = WishlistService.addListener(checkWishlistStatus)
    return unsubscribe
  }, [destination?.id])

  const handleToggleWishlist = async () => {
    if (isLoading) return

    setIsLoading(true)

    try {
      if (isInWishlist) {
        await WishlistService.removeFromWishlist(destination.id.toString())
        setIsInWishlist(false)
        setWishlistItem(null)
        onWishlistChange?.(false)
      } else {
        const newItem = await WishlistService.addToWishlist(destination)
        setIsInWishlist(true)
        setWishlistItem(newItem)
        setShowSuccess(true)
        onWishlistChange?.(true, newItem)

        // Hide success animation after 2 seconds
        setTimeout(() => setShowSuccess(false), 2000)
      }
    } catch (error) {
      console.error('Error toggling wishlist:', error)
      // You could show a toast notification here
    } finally {
      setIsLoading(false)
    }
  }

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

  if (variant === 'button') {
    return (
      <Button
        variant={isInWishlist ? 'primary' : 'outline'}
        size={size}
        onClick={handleToggleWishlist}
        disabled={isLoading}
        className={cn("relative overflow-hidden", className)}
        leftIcon={
          isLoading ? (
            <Loader2 className={cn(iconSizes[size], "animate-spin")} />
          ) : isInWishlist ? (
            <Heart className={cn(iconSizes[size], "fill-current")} />
          ) : (
            <Heart className={iconSizes[size]} />
          )
        }
        aria-label={isInWishlist ? "Remove from wishlist" : "Add to wishlist"}
      >
        {showLabel && (
          <span className="ml-2">
            {isInWishlist ? 'Saved' : 'Save'}
          </span>
        )}

        {/* Success Animation */}
        <AnimatePresence>
          {showSuccess && (
            <motion.div
              initial={{ scale: 0, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              exit={{ scale: 0, opacity: 0 }}
              className="absolute inset-0 bg-green-500 flex items-center justify-center z-10"
            >
              <Check className={cn(iconSizes[size], "text-white")} />
            </motion.div>
          )}
        </AnimatePresence>
      </Button>
    )
  }

  return (
    <motion.button
      whileHover={{ scale: 1.05 }}
      whileTap={{ scale: 0.95 }}
      onClick={handleToggleWishlist}
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
      {/* Loading State */}
      {isLoading && (
        <Loader2 className={cn(iconSizes[size], "animate-spin")} />
      )}

      {/* Heart Icon with Animation */}
      {!isLoading && (
        <motion.div
          key={isInWishlist ? 'filled' : 'empty'}
          initial={{ scale: 0.5, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          transition={{
            type: 'spring',
            stiffness: 500,
            damping: 15
          }}
        >
          <Heart
            className={cn(
              iconSizes[size],
              isInWishlist ? "fill-current" : "",
              "transition-colors duration-200"
            )}
          />
        </motion.div>
      )}

      {/* Success Animation Overlay */}
      <AnimatePresence>
        {showSuccess && (
          <motion.div
            initial={{ scale: 0, opacity: 0 }}
            animate={{ scale: 1.2, opacity: 1 }}
            exit={{ scale: 0, opacity: 0 }}
            transition={{ duration: 0.3 }}
            className="absolute inset-0 flex items-center justify-center"
          >
            <motion.div
              animate={{ 
                scale: [1, 1.5, 1],
                rotate: [0, 10, -10, 0]
              }}
              transition={{ duration: 0.6 }}
              className="text-red-500"
            >
              <Heart className={`${iconSizes[size]} fill-current`} />
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Floating Hearts Animation */}
      <AnimatePresence>
        {showSuccess && (
          <>
            {[...Array(3)].map((_, i) => (
              <motion.div
                key={i}
                initial={{ 
                  scale: 0, 
                  x: 0, 
                  y: 0, 
                  opacity: 1 
                }}
                animate={{ 
                  scale: [0, 1, 0], 
                  x: (i - 1) * 20, 
                  y: -30 - i * 10, 
                  opacity: [0, 1, 0] 
                }}
                transition={{ 
                  duration: 1, 
                  delay: i * 0.1,
                  ease: 'easeOut'
                }}
                className="absolute pointer-events-none"
              >
                <Heart className="h-3 w-3 text-red-400 fill-current" />
              </motion.div>
            ))}
          </>
        )}
      </AnimatePresence>

      {/* Ripple Effect */}
      <AnimatePresence>
        {showSuccess && (
          <motion.div
            initial={{ scale: 0, opacity: 0.5 }}
            animate={{ scale: 3, opacity: 0 }}
            exit={{ scale: 0, opacity: 0 }}
            transition={{ duration: 0.6 }}
            className="absolute inset-0 rounded-full border-2 border-red-400"
          />
        )}
      </AnimatePresence>
    </motion.button>
  )
}

// Wishlist Counter Component
interface WishlistCounterProps {
  className?: string
}

export function WishlistCounter({ className = '' }: WishlistCounterProps) {
  const [count, setCount] = useState(0)

  useEffect(() => {
    const updateCount = () => {
      if (typeof window !== 'undefined') {
        const wishlist = WishlistService.getWishlist()
        setCount(wishlist.length)
      }
    }

    updateCount()
    const unsubscribe = WishlistService.addListener(updateCount)
    return unsubscribe
  }, [])

  if (count === 0) return null

  return (
    <motion.div
      initial={{ scale: 0 }}
      animate={{ scale: 1 }}
      className={`
        inline-flex items-center justify-center
        min-w-[20px] h-5 px-1.5 
        bg-red-500 text-white text-xs font-bold rounded-full
        ${className}
      `}
    >
      {count > 99 ? '99+' : count}
    </motion.div>
  )
}

// Quick Add to Wishlist Component
interface QuickWishlistProps {
  destinations: Destination[]
  onComplete?: () => void
  className?: string
}

export function QuickWishlist({ destinations, onComplete, className = '' }: QuickWishlistProps) {
  const [selectedDestinations, setSelectedDestinations] = useState<Set<string>>(new Set())
  const [isLoading, setIsLoading] = useState(false)

  const handleToggleDestination = (destinationId: string) => {
    const newSelected = new Set(selectedDestinations)
    if (newSelected.has(destinationId)) {
      newSelected.delete(destinationId)
    } else {
      newSelected.add(destinationId)
    }
    setSelectedDestinations(newSelected)
  }

  const handleAddSelected = async () => {
    if (selectedDestinations.size === 0) return

    setIsLoading(true)
    try {
      const promises = Array.from(selectedDestinations).map(id => {
        const destination = destinations.find(d => d.id.toString() === id)
        return destination ? WishlistService.addToWishlist(destination) : null
      }).filter(Boolean)

      await Promise.all(promises)
      setSelectedDestinations(new Set())
      onComplete?.()
    } catch (error) {
      console.error('Error adding to wishlist:', error)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className={`space-y-4 ${className}`}>
      <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
        {destinations.map(destination => (
          <motion.div
            key={destination.id}
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
            onClick={() => handleToggleDestination(destination.id.toString())}
            className={`
              relative cursor-pointer rounded-lg overflow-hidden border-2 transition-all
              ${selectedDestinations.has(destination.id.toString())
                ? 'border-red-500 ring-2 ring-red-200' 
                : 'border-gray-200 hover:border-gray-300'
              }
            `}
          >
            <div className="aspect-video bg-gray-200">
              {destination.images[0] && (
                <img
                  src={destination.images[0]}
                  alt={destination.name}
                  className="w-full h-full object-cover"
                />
              )}
            </div>
            <div className="p-3">
              <h4 className="font-medium text-sm truncate">{destination.name}</h4>
              <p className="text-xs text-gray-500">{destination.country}</p>
            </div>
            
            {selectedDestinations.has(destination.id.toString()) && (
              <motion.div
                initial={{ scale: 0 }}
                animate={{ scale: 1 }}
                className="absolute top-2 right-2 w-6 h-6 bg-red-500 rounded-full flex items-center justify-center"
              >
                <Check className="h-4 w-4 text-white" />
              </motion.div>
            )}
          </motion.div>
        ))}
      </div>

      {selectedDestinations.size > 0 && (
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="flex items-center justify-between p-4 bg-gray-50 rounded-lg"
        >
          <span className="text-sm text-gray-600">
            {selectedDestinations.size} destination{selectedDestinations.size !== 1 ? 's' : ''} selected
          </span>
          <Button
            onClick={handleAddSelected}
            disabled={isLoading}
            leftIcon={isLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Plus className="h-4 w-4" />}
          >
            Add to Wishlist
          </Button>
        </motion.div>
      )}
    </div>
  )
}

// Default export for backward compatibility
export default WishlistButton
