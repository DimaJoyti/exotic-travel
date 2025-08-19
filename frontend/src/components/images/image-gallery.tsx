'use client'

import { useState, useEffect } from 'react'
import { ChevronLeft, ChevronRight, X, ZoomIn, Download, Share2, Heart, MoreHorizontal } from 'lucide-react'
import { ImageMetadata, ImagesService } from '@/lib/images'

interface ImageGalleryProps {
  images: ImageMetadata[]
  initialIndex?: number
  showThumbnails?: boolean
  showControls?: boolean
  showMetadata?: boolean
  className?: string
  onImageChange?: (index: number, image: ImageMetadata) => void
}

export default function ImageGallery({
  images,
  initialIndex = 0,
  showThumbnails = true,
  showControls = true,
  showMetadata = false,
  className = '',
  onImageChange
}: ImageGalleryProps) {
  const [currentIndex, setCurrentIndex] = useState(initialIndex)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [isLoading, setIsLoading] = useState(false)

  const currentImage = images[currentIndex]

  useEffect(() => {
    if (onImageChange && currentImage) {
      onImageChange(currentIndex, currentImage)
    }
  }, [currentIndex, currentImage, onImageChange])

  const goToPrevious = () => {
    setCurrentIndex(prev => (prev > 0 ? prev - 1 : images.length - 1))
  }

  const goToNext = () => {
    setCurrentIndex(prev => (prev < images.length - 1 ? prev + 1 : 0))
  }

  const goToImage = (index: number) => {
    setCurrentIndex(index)
  }

  const toggleFullscreen = () => {
    setIsFullscreen(!isFullscreen)
  }

  const handleDownload = async () => {
    if (!currentImage) return
    
    try {
      const response = await fetch(currentImage.url)
      const blob = await response.blob()
      const url = URL.createObjectURL(blob)
      
      const a = document.createElement('a')
      a.href = url
      a.download = currentImage.originalName || currentImage.filename
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
    } catch (error) {
      console.error('Download failed:', error)
    }
  }

  const handleShare = async () => {
    if (!currentImage) return
    
    if (navigator.share) {
      try {
        await navigator.share({
          title: currentImage.alt || currentImage.filename,
          text: currentImage.caption || 'Check out this image',
          url: currentImage.url
        })
      } catch (error) {
        console.error('Share failed:', error)
      }
    } else {
      // Fallback: copy to clipboard
      try {
        await navigator.clipboard.writeText(currentImage.url)
        alert('Image URL copied to clipboard!')
      } catch (error) {
        console.error('Copy failed:', error)
      }
    }
  }

  if (!images.length) {
    return (
      <div className={`bg-gray-100 rounded-lg p-8 text-center ${className}`}>
        <p className="text-gray-500">No images to display</p>
      </div>
    )
  }

  const galleryContent = (
    <div className="relative">
      {/* Main Image */}
      <div className="relative bg-black rounded-lg overflow-hidden">
        <img
          src={ImagesService.getOptimizedImageUrl(currentImage.url, {
            width: isFullscreen ? 1920 : 800,
            height: isFullscreen ? 1080 : 600,
            quality: 0.9,
            fit: 'contain'
          })}
          alt={currentImage.alt || currentImage.filename}
          className="w-full h-auto max-h-[60vh] object-contain"
          onLoad={() => setIsLoading(false)}
          onLoadStart={() => setIsLoading(true)}
        />
        
        {/* Loading Overlay */}
        {isLoading && (
          <div className="absolute inset-0 bg-black bg-opacity-50 flex items-center justify-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-white"></div>
          </div>
        )}

        {/* Navigation Arrows */}
        {showControls && images.length > 1 && (
          <>
            <button
              onClick={goToPrevious}
              className="absolute left-4 top-1/2 transform -translate-y-1/2 bg-black bg-opacity-50 text-white p-2 rounded-full hover:bg-opacity-75 transition-opacity"
            >
              <ChevronLeft className="h-6 w-6" />
            </button>
            <button
              onClick={goToNext}
              className="absolute right-4 top-1/2 transform -translate-y-1/2 bg-black bg-opacity-50 text-white p-2 rounded-full hover:bg-opacity-75 transition-opacity"
            >
              <ChevronRight className="h-6 w-6" />
            </button>
          </>
        )}

        {/* Top Controls */}
        {showControls && (
          <div className="absolute top-4 right-4 flex space-x-2">
            <button
              onClick={toggleFullscreen}
              className="bg-black bg-opacity-50 text-white p-2 rounded-full hover:bg-opacity-75 transition-opacity"
            >
              <ZoomIn className="h-5 w-5" />
            </button>
            <button
              onClick={handleDownload}
              className="bg-black bg-opacity-50 text-white p-2 rounded-full hover:bg-opacity-75 transition-opacity"
            >
              <Download className="h-5 w-5" />
            </button>
            <button
              onClick={handleShare}
              className="bg-black bg-opacity-50 text-white p-2 rounded-full hover:bg-opacity-75 transition-opacity"
            >
              <Share2 className="h-5 w-5" />
            </button>
          </div>
        )}

        {/* Image Counter */}
        {images.length > 1 && (
          <div className="absolute bottom-4 left-4 bg-black bg-opacity-50 text-white px-3 py-1 rounded-full text-sm">
            {currentIndex + 1} / {images.length}
          </div>
        )}
      </div>

      {/* Image Metadata */}
      {showMetadata && currentImage && (
        <div className="mt-4 p-4 bg-gray-50 rounded-lg">
          {currentImage.caption && (
            <h3 className="text-lg font-semibold text-gray-900 mb-2">
              {currentImage.caption}
            </h3>
          )}
          {currentImage.alt && (
            <p className="text-gray-700 mb-3">{currentImage.alt}</p>
          )}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm text-gray-600">
            <div>
              <span className="font-medium">Dimensions:</span>
              <br />
              {currentImage.width} × {currentImage.height}
            </div>
            <div>
              <span className="font-medium">Size:</span>
              <br />
              {(currentImage.size / (1024 * 1024)).toFixed(2)} MB
            </div>
            <div>
              <span className="font-medium">Format:</span>
              <br />
              {currentImage.mimeType.split('/')[1].toUpperCase()}
            </div>
            <div>
              <span className="font-medium">Uploaded:</span>
              <br />
              {new Date(currentImage.uploadedAt).toLocaleDateString()}
            </div>
          </div>
          {currentImage.tags && currentImage.tags.length > 0 && (
            <div className="mt-3">
              <span className="font-medium text-gray-700">Tags:</span>
              <div className="flex flex-wrap gap-2 mt-1">
                {currentImage.tags.map((tag, index) => (
                  <span
                    key={index}
                    className="px-2 py-1 bg-primary/10 text-primary text-xs rounded-full"
                  >
                    {tag}
                  </span>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Thumbnails */}
      {showThumbnails && images.length > 1 && (
        <div className="mt-4">
          <div className="flex space-x-2 overflow-x-auto pb-2">
            {images.map((image, index) => (
              <button
                key={image.id}
                onClick={() => goToImage(index)}
                className={`flex-shrink-0 w-20 h-20 rounded-lg overflow-hidden border-2 transition-colors ${
                  index === currentIndex
                    ? 'border-primary'
                    : 'border-gray-200 hover:border-gray-300'
                }`}
              >
                <img
                  src={ImagesService.getOptimizedImageUrl(image.thumbnailUrl || image.url, {
                    width: 80,
                    height: 80,
                    fit: 'cover'
                  })}
                  alt={image.alt || image.filename}
                  className="w-full h-full object-cover"
                />
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  )

  // Fullscreen Modal
  if (isFullscreen) {
    return (
      <div className="fixed inset-0 bg-black bg-opacity-95 z-50 flex items-center justify-center p-4">
        <button
          onClick={toggleFullscreen}
          className="absolute top-4 right-4 text-white p-2 rounded-full hover:bg-white hover:bg-opacity-20 transition-colors z-10"
        >
          <X className="h-6 w-6" />
        </button>
        <div className="w-full max-w-7xl">
          {galleryContent}
        </div>
      </div>
    )
  }

  return <div className={className}>{galleryContent}</div>
}

// Grid Gallery Component
interface ImageGridProps {
  images: ImageMetadata[]
  columns?: number
  gap?: number
  onImageClick?: (index: number, image: ImageMetadata) => void
  showOverlay?: boolean
  className?: string
}

export function ImageGrid({
  images,
  columns = 3,
  gap = 4,
  onImageClick,
  showOverlay = true,
  className = ''
}: ImageGridProps) {
  const [selectedIndex, setSelectedIndex] = useState<number | null>(null)

  const handleImageClick = (index: number, image: ImageMetadata) => {
    setSelectedIndex(index)
    onImageClick?.(index, image)
  }

  const closeGallery = () => {
    setSelectedIndex(null)
  }

  return (
    <>
      <div 
        className={`grid gap-${gap} ${className}`}
        style={{ gridTemplateColumns: `repeat(${columns}, 1fr)` }}
      >
        {images.map((image, index) => (
          <div
            key={image.id}
            className="relative aspect-square bg-gray-100 rounded-lg overflow-hidden cursor-pointer group"
            onClick={() => handleImageClick(index, image)}
          >
            <img
              src={ImagesService.getOptimizedImageUrl(image.url, {
                width: 400,
                height: 400,
                fit: 'cover'
              })}
              alt={image.alt || image.filename}
              className="w-full h-full object-cover transition-transform group-hover:scale-105"
            />
            
            {showOverlay && (
              <div className="absolute inset-0 bg-black bg-opacity-0 group-hover:bg-opacity-30 transition-opacity flex items-center justify-center">
                <ZoomIn className="h-8 w-8 text-white opacity-0 group-hover:opacity-100 transition-opacity" />
              </div>
            )}
            
            {image.caption && (
              <div className="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black to-transparent p-4">
                <p className="text-white text-sm font-medium truncate">
                  {image.caption}
                </p>
              </div>
            )}
          </div>
        ))}
      </div>

      {/* Fullscreen Gallery Modal */}
      {selectedIndex !== null && (
        <div className="fixed inset-0 bg-black bg-opacity-95 z-50 flex items-center justify-center p-4">
          <button
            onClick={closeGallery}
            className="absolute top-4 right-4 text-white p-2 rounded-full hover:bg-white hover:bg-opacity-20 transition-colors z-10"
          >
            <X className="h-6 w-6" />
          </button>
          <div className="w-full max-w-7xl">
            <ImageGallery
              images={images}
              initialIndex={selectedIndex}
              showThumbnails={true}
              showControls={true}
              showMetadata={true}
            />
          </div>
        </div>
      )}
    </>
  )
}
