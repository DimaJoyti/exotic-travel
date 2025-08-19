'use client'

import React, { useRef, useEffect, useState } from 'react'
import { RotateCcw, Maximize, Minimize, Move, ZoomIn, ZoomOut, Play, Pause, RotateCw } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { Button } from '@/components/ui/button'

interface PanoramicViewerProps {
  imageUrl: string
  title?: string
  autoRotate?: boolean
  rotationSpeed?: number
  initialYaw?: number
  initialPitch?: number
  minPitch?: number
  maxPitch?: number
  className?: string
  onLoad?: () => void
  onError?: (error: string) => void
}

interface ViewerState {
  yaw: number
  pitch: number
  zoom: number
  isAutoRotating: boolean
  isDragging: boolean
  isFullscreen: boolean
}

export default function PanoramicViewer({
  imageUrl,
  title,
  autoRotate = false,
  rotationSpeed = 0.5,
  initialYaw = 0,
  initialPitch = 0,
  minPitch = -90,
  maxPitch = 90,
  className = '',
  onLoad,
  onError
}: PanoramicViewerProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const imageRef = useRef<HTMLImageElement | null>(null)
  const animationFrameRef = useRef<number>()
  const lastMousePos = useRef<{ x: number; y: number } | null>(null)

  const [viewerState, setViewerState] = useState<ViewerState>({
    yaw: initialYaw,
    pitch: initialPitch,
    zoom: 1,
    isAutoRotating: autoRotate,
    isDragging: false,
    isFullscreen: false
  })

  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showControls, setShowControls] = useState(true)

  // Initialize panoramic viewer
  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    // Load panoramic image
    const img = new Image()
    img.crossOrigin = 'anonymous'
    
    img.onload = () => {
      imageRef.current = img
      setIsLoading(false)
      onLoad?.()
      startRenderLoop()
    }

    img.onerror = () => {
      const errorMsg = 'Failed to load panoramic image'
      setError(errorMsg)
      setIsLoading(false)
      onError?.(errorMsg)
    }

    img.src = imageUrl

    return () => {
      if (animationFrameRef.current) {
        cancelAnimationFrame(animationFrameRef.current)
      }
    }
  }, [imageUrl, onLoad, onError])

  // Render loop for panoramic view
  const startRenderLoop = () => {
    const render = () => {
      renderPanorama()
      
      // Auto rotation
      if (viewerState.isAutoRotating && !viewerState.isDragging) {
        setViewerState(prev => ({
          ...prev,
          yaw: (prev.yaw + rotationSpeed) % 360
        }))
      }

      animationFrameRef.current = requestAnimationFrame(render)
    }
    
    render()
  }

  // Render panoramic image on canvas
  const renderPanorama = () => {
    const canvas = canvasRef.current
    const image = imageRef.current
    if (!canvas || !image) return

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const { width, height } = canvas
    const { yaw, pitch, zoom } = viewerState

    // Clear canvas
    ctx.clearRect(0, 0, width, height)

    // Calculate field of view
    const fov = 90 / zoom
    const fovRad = (fov * Math.PI) / 180

    // Calculate visible portion of panorama
    const yawRad = (yaw * Math.PI) / 180
    const pitchRad = (pitch * Math.PI) / 180

    // Simple equirectangular projection
    const sourceX = ((yaw % 360) / 360) * image.width
    const sourceY = ((pitch + 90) / 180) * image.height
    const sourceWidth = (fov / 360) * image.width
    const sourceHeight = (fov / 180) * image.height

    // Draw the visible portion
    try {
      ctx.drawImage(
        image,
        Math.max(0, sourceX - sourceWidth / 2),
        Math.max(0, sourceY - sourceHeight / 2),
        Math.min(sourceWidth, image.width),
        Math.min(sourceHeight, image.height),
        0,
        0,
        width,
        height
      )
    } catch (error) {
      console.error('Error rendering panorama:', error)
    }
  }

  // Mouse/touch event handlers
  const handleMouseDown = (e: React.MouseEvent) => {
    setViewerState(prev => ({ ...prev, isDragging: true }))
    lastMousePos.current = { x: e.clientX, y: e.clientY }
  }

  const handleMouseMove = (e: React.MouseEvent) => {
    if (!viewerState.isDragging || !lastMousePos.current) return

    const deltaX = e.clientX - lastMousePos.current.x
    const deltaY = e.clientY - lastMousePos.current.y

    setViewerState(prev => ({
      ...prev,
      yaw: (prev.yaw - deltaX * 0.5) % 360,
      pitch: Math.max(minPitch, Math.min(maxPitch, prev.pitch + deltaY * 0.5))
    }))

    lastMousePos.current = { x: e.clientX, y: e.clientY }
  }

  const handleMouseUp = () => {
    setViewerState(prev => ({ ...prev, isDragging: false }))
    lastMousePos.current = null
  }

  // Control functions
  const toggleAutoRotate = () => {
    setViewerState(prev => ({ ...prev, isAutoRotating: !prev.isAutoRotating }))
  }

  const resetView = () => {
    setViewerState(prev => ({
      ...prev,
      yaw: initialYaw,
      pitch: initialPitch,
      zoom: 1
    }))
  }

  const zoomIn = () => {
    setViewerState(prev => ({ ...prev, zoom: Math.min(prev.zoom * 1.2, 3) }))
  }

  const zoomOut = () => {
    setViewerState(prev => ({ ...prev, zoom: Math.max(prev.zoom / 1.2, 0.5) }))
  }

  const toggleFullscreen = () => {
    if (!document.fullscreenElement) {
      containerRef.current?.requestFullscreen()
      setViewerState(prev => ({ ...prev, isFullscreen: true }))
    } else {
      document.exitFullscreen()
      setViewerState(prev => ({ ...prev, isFullscreen: false }))
    }
  }

  // Resize canvas to container
  useEffect(() => {
    const resizeCanvas = () => {
      const canvas = canvasRef.current
      const container = containerRef.current
      if (!canvas || !container) return

      const rect = container.getBoundingClientRect()
      canvas.width = rect.width
      canvas.height = rect.height
    }

    resizeCanvas()
    window.addEventListener('resize', resizeCanvas)
    return () => window.removeEventListener('resize', resizeCanvas)
  }, [viewerState.isFullscreen])

  return (
    <div 
      ref={containerRef}
      className={`relative bg-black rounded-lg overflow-hidden ${className}`}
      style={{ height: viewerState.isFullscreen ? '100vh' : '400px' }}
      onMouseEnter={() => setShowControls(true)}
      onMouseLeave={() => setShowControls(false)}
    >
      {/* Canvas */}
      <canvas
        ref={canvasRef}
        className="w-full h-full cursor-move"
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseUp}
      />

      {/* Loading State */}
      {isLoading && (
        <div className="absolute inset-0 flex items-center justify-center bg-black bg-opacity-75">
          <div className="text-center text-white">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-white mx-auto mb-4"></div>
            <p>Loading 360° view...</p>
          </div>
        </div>
      )}

      {/* Error State */}
      {error && (
        <div className="absolute inset-0 flex items-center justify-center bg-black bg-opacity-75">
          <div className="text-center text-white">
            <p className="text-red-400 mb-2">Failed to load panoramic view</p>
            <p className="text-sm opacity-75">{error}</p>
          </div>
        </div>
      )}

      {/* Controls */}
      <AnimatePresence>
        {showControls && !isLoading && !error && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="absolute inset-0 pointer-events-none"
          >
            {/* Title */}
            {title && (
              <div className="absolute top-4 left-4 bg-black bg-opacity-50 text-white px-3 py-2 rounded-lg pointer-events-auto">
                <h3 className="font-medium">{title}</h3>
                <p className="text-xs opacity-75">360° Panoramic View</p>
              </div>
            )}

            {/* Top Controls */}
            <div className="absolute top-4 right-4 flex space-x-2 pointer-events-auto">
              <Button
                variant="ghost"
                size="sm"
                onClick={toggleAutoRotate}
                className="bg-black bg-opacity-50 text-white hover:bg-opacity-75"
              >
                {viewerState.isAutoRotating ? (
                  <Pause className="h-4 w-4" />
                ) : (
                  <Play className="h-4 w-4" />
                )}
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={toggleFullscreen}
                className="bg-black bg-opacity-50 text-white hover:bg-opacity-75"
              >
                {viewerState.isFullscreen ? (
                  <Minimize className="h-4 w-4" />
                ) : (
                  <Maximize className="h-4 w-4" />
                )}
              </Button>
            </div>

            {/* Bottom Controls */}
            <div className="absolute bottom-4 left-1/2 transform -translate-x-1/2 flex items-center space-x-2 bg-black bg-opacity-50 rounded-lg p-2 pointer-events-auto">
              <Button
                variant="ghost"
                size="sm"
                onClick={zoomOut}
                className="text-white hover:bg-white hover:bg-opacity-20"
              >
                <ZoomOut className="h-4 w-4" />
              </Button>
              
              <div className="text-white text-sm px-2">
                {Math.round(viewerState.zoom * 100)}%
              </div>
              
              <Button
                variant="ghost"
                size="sm"
                onClick={zoomIn}
                className="text-white hover:bg-white hover:bg-opacity-20"
              >
                <ZoomIn className="h-4 w-4" />
              </Button>
              
              <div className="w-px h-6 bg-white bg-opacity-30 mx-2" />
              
              <Button
                variant="ghost"
                size="sm"
                onClick={resetView}
                className="text-white hover:bg-white hover:bg-opacity-20"
              >
                <RotateCcw className="h-4 w-4" />
              </Button>
            </div>

            {/* Instructions */}
            <div className="absolute bottom-4 right-4 bg-black bg-opacity-50 text-white text-xs px-3 py-2 rounded-lg pointer-events-auto">
              <div className="flex items-center space-x-1">
                <Move className="h-3 w-3" />
                <span>Drag to look around</span>
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Rotation Indicator */}
      {viewerState.isAutoRotating && (
        <motion.div
          animate={{ rotate: 360 }}
          transition={{ duration: 2, repeat: Infinity, ease: 'linear' }}
          className="absolute top-4 left-1/2 transform -translate-x-1/2 text-white"
        >
          <RotateCw className="h-4 w-4" />
        </motion.div>
      )}
    </div>
  )
}

// 360° Gallery Component
interface PanoramicGalleryProps {
  images: Array<{
    id: string
    url: string
    title: string
    description?: string
  }>
  className?: string
}

export function PanoramicGallery({ images, className = '' }: PanoramicGalleryProps) {
  const [currentIndex, setCurrentIndex] = useState(0)

  if (images.length === 0) {
    return (
      <div className={`bg-gray-100 rounded-lg p-8 text-center ${className}`}>
        <p className="text-gray-500">No 360° images available</p>
      </div>
    )
  }

  const currentImage = images[currentIndex]

  return (
    <div className={`space-y-4 ${className}`}>
      <PanoramicViewer
        imageUrl={currentImage.url}
        title={currentImage.title}
        autoRotate={true}
        className="w-full"
      />

      {/* Image Selector */}
      {images.length > 1 && (
        <div className="flex space-x-2 overflow-x-auto pb-2">
          {images.map((image, index) => (
            <button
              key={image.id}
              onClick={() => setCurrentIndex(index)}
              className={`
                flex-shrink-0 w-20 h-12 rounded-lg overflow-hidden border-2 transition-colors
                ${index === currentIndex 
                  ? 'border-brand-500' 
                  : 'border-gray-200 hover:border-gray-300'
                }
              `}
            >
              <img
                src={image.url}
                alt={image.title}
                className="w-full h-full object-cover"
              />
            </button>
          ))}
        </div>
      )}

      {/* Image Info */}
      <div className="text-center">
        <h3 className="font-semibold text-gray-900">{currentImage.title}</h3>
        {currentImage.description && (
          <p className="text-sm text-gray-600 mt-1">{currentImage.description}</p>
        )}
      </div>
    </div>
  )
}
