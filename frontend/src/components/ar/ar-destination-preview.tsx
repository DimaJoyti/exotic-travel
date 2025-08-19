'use client'

import React, { useState, useRef, useEffect } from 'react'
import { Camera, Smartphone, Eye, Maximize, X, RotateCcw, Zap, Layers, MapPin, Info } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { Button } from '@/components/ui/button'
import { FadeIn, ScaleIn } from '@/components/ui/animated'
import { Destination } from '@/types'

interface ARDestinationPreviewProps {
  destination: Destination
  onClose?: () => void
  className?: string
}

interface ARMarker {
  id: string
  position: { x: number; y: number; z: number }
  label: string
  description: string
  type: 'poi' | 'hotel' | 'restaurant' | 'activity'
  icon: React.ReactNode
}

interface ARSession {
  isActive: boolean
  isSupported: boolean
  hasPermission: boolean
  error: string | null
}

export default function ARDestinationPreview({ destination, onClose, className = '' }: ARDestinationPreviewProps) {
  const videoRef = useRef<HTMLVideoElement>(null)
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  
  const [arSession, setArSession] = useState<ARSession>({
    isActive: false,
    isSupported: false,
    hasPermission: false,
    error: null
  })
  
  const [markers, setMarkers] = useState<ARMarker[]>([])
  const [selectedMarker, setSelectedMarker] = useState<ARMarker | null>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [deviceOrientation, setDeviceOrientation] = useState({ alpha: 0, beta: 0, gamma: 0 })

  // Check AR support and permissions
  useEffect(() => {
    checkARSupport()
    generateARMarkers()
  }, [destination])

  const checkARSupport = async () => {
    try {
      // Check if WebXR is supported
      if ('xr' in navigator) {
        const isSupported = await (navigator as any).xr.isSessionSupported('immersive-ar')
        setArSession(prev => ({ ...prev, isSupported }))
      } else {
        // Fallback to camera-based AR simulation
        setArSession(prev => ({ ...prev, isSupported: true }))
      }
    } catch (error) {
      console.error('AR support check failed:', error)
      setArSession(prev => ({ ...prev, error: 'AR not supported on this device' }))
    }
  }

  const generateARMarkers = () => {
    // Generate AR markers based on destination features
    const newMarkers: ARMarker[] = [
      {
        id: 'main-attraction',
        position: { x: 0, y: 0, z: -5 },
        label: destination.name,
        description: destination.description,
        type: 'poi',
        icon: <MapPin className="h-4 w-4" />
      },
      {
        id: 'hotel-1',
        position: { x: -2, y: 0, z: -3 },
        label: 'Luxury Resort',
        description: 'Premium accommodation with ocean views',
        type: 'hotel',
        icon: <Layers className="h-4 w-4" />
      },
      {
        id: 'restaurant-1',
        position: { x: 2, y: 0, z: -4 },
        label: 'Local Cuisine',
        description: 'Authentic local dining experience',
        type: 'restaurant',
        icon: <Zap className="h-4 w-4" />
      }
    ]

    // Add markers based on destination features
    destination.features.forEach((feature, index) => {
      newMarkers.push({
        id: `feature-${index}`,
        position: { 
          x: Math.cos(index * 0.5) * 3, 
          y: Math.sin(index * 0.3), 
          z: -2 - index 
        },
        label: feature,
        description: `Experience ${feature} at ${destination.name}`,
        type: 'activity',
        icon: <Eye className="h-4 w-4" />
      })
    })

    setMarkers(newMarkers)
  }

  const startARSession = async () => {
    setIsLoading(true)
    
    try {
      // Request camera permission
      const stream = await navigator.mediaDevices.getUserMedia({ 
        video: { facingMode: 'environment' } 
      })
      
      if (videoRef.current) {
        videoRef.current.srcObject = stream
        videoRef.current.play()
      }

      // Start device orientation tracking
      if (window.DeviceOrientationEvent) {
        window.addEventListener('deviceorientation', handleDeviceOrientation)
      }

      setArSession(prev => ({ 
        ...prev, 
        isActive: true, 
        hasPermission: true,
        error: null 
      }))
      
    } catch (error) {
      console.error('Failed to start AR session:', error)
      setArSession(prev => ({ 
        ...prev, 
        error: 'Camera permission denied or not available' 
      }))
    } finally {
      setIsLoading(false)
    }
  }

  const stopARSession = () => {
    if (videoRef.current?.srcObject) {
      const stream = videoRef.current.srcObject as MediaStream
      stream.getTracks().forEach(track => track.stop())
      videoRef.current.srcObject = null
    }

    window.removeEventListener('deviceorientation', handleDeviceOrientation)
    
    setArSession(prev => ({ ...prev, isActive: false }))
  }

  const handleDeviceOrientation = (event: DeviceOrientationEvent) => {
    setDeviceOrientation({
      alpha: event.alpha || 0,
      beta: event.beta || 0,
      gamma: event.gamma || 0
    })
  }

  const renderARMarker = (marker: ARMarker, index: number) => {
    // Calculate marker position based on device orientation
    const adjustedX = marker.position.x + (deviceOrientation.gamma * 0.01)
    const adjustedY = marker.position.y + (deviceOrientation.beta * 0.01)
    
    // Convert 3D position to 2D screen coordinates (simplified)
    const screenX = 50 + adjustedX * 10 // Center + offset
    const screenY = 50 + adjustedY * 10
    
    return (
      <motion.div
        key={marker.id}
        initial={{ opacity: 0, scale: 0 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ delay: index * 0.2 }}
        className="absolute transform -translate-x-1/2 -translate-y-1/2 cursor-pointer"
        style={{ 
          left: `${Math.max(10, Math.min(90, screenX))}%`,
          top: `${Math.max(10, Math.min(90, screenY))}%`
        }}
        onClick={() => setSelectedMarker(marker)}
      >
        <div className="relative">
          {/* Marker Pin */}
          <div className={`
            w-12 h-12 rounded-full flex items-center justify-center text-white shadow-lg
            ${marker.type === 'poi' ? 'bg-red-500' :
              marker.type === 'hotel' ? 'bg-blue-500' :
              marker.type === 'restaurant' ? 'bg-green-500' :
              'bg-purple-500'
            }
          `}>
            {marker.icon}
          </div>
          
          {/* Pulsing Animation */}
          <div className={`
            absolute inset-0 rounded-full animate-ping opacity-30
            ${marker.type === 'poi' ? 'bg-red-500' :
              marker.type === 'hotel' ? 'bg-blue-500' :
              marker.type === 'restaurant' ? 'bg-green-500' :
              'bg-purple-500'
            }
          `} />
          
          {/* Label */}
          <div className="absolute top-full left-1/2 transform -translate-x-1/2 mt-2 bg-black bg-opacity-75 text-white px-2 py-1 rounded text-xs whitespace-nowrap">
            {marker.label}
          </div>
        </div>
      </motion.div>
    )
  }

  return (
    <div className={`fixed inset-0 z-50 bg-black ${className}`}>
      {/* Header */}
      <div className="absolute top-0 left-0 right-0 z-10 bg-gradient-to-b from-black to-transparent p-4">
        <div className="flex items-center justify-between text-white">
          <div>
            <h2 className="text-xl font-bold">{destination.name}</h2>
            <p className="text-sm opacity-75">AR Preview</p>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={onClose}
            className="text-white hover:bg-white/20"
          >
            <X className="h-5 w-5" />
          </Button>
        </div>
      </div>

      {/* AR View */}
      <div ref={containerRef} className="relative w-full h-full">
        {!arSession.isActive ? (
          /* AR Start Screen */
          <div className="flex items-center justify-center h-full bg-gradient-to-br from-gray-900 to-black text-white">
            <FadeIn>
              <div className="text-center max-w-md px-6">
                <div className="mb-8">
                  <Camera className="h-16 w-16 mx-auto mb-4 text-brand-500" />
                  <h3 className="text-2xl font-bold mb-2">AR Destination Preview</h3>
                  <p className="text-gray-300">
                    Experience {destination.name} in augmented reality. Point your camera around to discover points of interest.
                  </p>
                </div>

                {arSession.error ? (
                  <div className="mb-6 p-4 bg-red-500/20 border border-red-500 rounded-lg">
                    <p className="text-red-300">{arSession.error}</p>
                  </div>
                ) : null}

                <div className="space-y-4">
                  <Button
                    onClick={startARSession}
                    disabled={isLoading || !arSession.isSupported}
                    className="w-full bg-brand-500 hover:bg-brand-600"
                  >
                    <div className="flex items-center space-x-2">
                      {isLoading ? (
                        <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white" />
                      ) : (
                        <Smartphone className="h-4 w-4" />
                      )}
                      <span>
                        {isLoading ? 'Starting AR...' : 'Start AR Experience'}
                      </span>
                    </div>
                  </Button>

                  {!arSession.isSupported && (
                    <p className="text-sm text-gray-400">
                      AR not supported on this device. Try on a mobile device with camera access.
                    </p>
                  )}
                </div>
              </div>
            </FadeIn>
          </div>
        ) : (
          /* Active AR View */
          <>
            {/* Camera Feed */}
            <video
              ref={videoRef}
              className="w-full h-full object-cover"
              playsInline
              muted
            />

            {/* AR Overlay Canvas */}
            <canvas
              ref={canvasRef}
              className="absolute inset-0 pointer-events-none"
            />

            {/* AR Markers */}
            <div className="absolute inset-0">
              {markers.map((marker, index) => renderARMarker(marker, index))}
            </div>

            {/* AR Controls */}
            <div className="absolute bottom-4 left-4 right-4 flex justify-center space-x-4">
              <Button
                variant="secondary"
                size="sm"
                onClick={() => setDeviceOrientation({ alpha: 0, beta: 0, gamma: 0 })}
                className="bg-black/50 text-white border-white/20"
              >
                <RotateCcw className="h-4 w-4 mr-2" />
                Reset View
              </Button>
              
              <Button
                variant="secondary"
                size="sm"
                onClick={stopARSession}
                className="bg-red-500/80 text-white border-red-400/20"
              >
                Stop AR
              </Button>
            </div>
          </>
        )}
      </div>

      {/* Marker Info Modal */}
      <AnimatePresence>
        {selectedMarker && (
          <motion.div
            initial={{ opacity: 0, y: 50 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: 50 }}
            className="absolute bottom-20 left-4 right-4 bg-white rounded-lg p-4 shadow-xl"
          >
            <div className="flex items-start justify-between mb-2">
              <div className="flex items-center space-x-2">
                <div className={`
                  w-8 h-8 rounded-full flex items-center justify-center text-white
                  ${selectedMarker.type === 'poi' ? 'bg-red-500' :
                    selectedMarker.type === 'hotel' ? 'bg-blue-500' :
                    selectedMarker.type === 'restaurant' ? 'bg-green-500' :
                    'bg-purple-500'
                  }
                `}>
                  {selectedMarker.icon}
                </div>
                <h4 className="font-semibold text-gray-900">{selectedMarker.label}</h4>
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setSelectedMarker(null)}
              >
                <X className="h-4 w-4" />
              </Button>
            </div>
            <p className="text-gray-600 text-sm">{selectedMarker.description}</p>
            <div className="mt-3 flex space-x-2">
              <Button size="sm" className="flex-1">
                Learn More
              </Button>
              <Button size="sm" variant="outline" className="flex-1">
                <Info className="h-4 w-4 mr-1" />
                Details
              </Button>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}
