'use client'

import { useState, useRef, useEffect } from 'react'
import { ImagesService } from '@/lib/images'

interface OptimizedImageProps {
  src: string
  alt: string
  width?: number
  height?: number
  quality?: number
  format?: 'webp' | 'jpeg' | 'png'
  fit?: 'cover' | 'contain' | 'fill'
  lazy?: boolean
  placeholder?: string
  blurDataURL?: string
  sizes?: string
  priority?: boolean
  className?: string
  style?: React.CSSProperties
  onLoad?: () => void
  onError?: () => void
  onClick?: () => void
}

export default function OptimizedImage({
  src,
  alt,
  width,
  height,
  quality = 0.8,
  format = 'webp',
  fit = 'cover',
  lazy = true,
  placeholder,
  blurDataURL,
  sizes,
  priority = false,
  className = '',
  style,
  onLoad,
  onError,
  onClick
}: OptimizedImageProps) {
  const [isLoaded, setIsLoaded] = useState(false)
  const [isError, setIsError] = useState(false)
  const [isInView, setIsInView] = useState(!lazy || priority)
  const imgRef = useRef<HTMLImageElement>(null)
  const observerRef = useRef<IntersectionObserver | null>(null)

  // Set up intersection observer for lazy loading
  useEffect(() => {
    if (!lazy || priority || isInView) return

    observerRef.current = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            setIsInView(true)
            observerRef.current?.disconnect()
          }
        })
      },
      {
        rootMargin: '50px'
      }
    )

    if (imgRef.current) {
      observerRef.current.observe(imgRef.current)
    }

    return () => {
      observerRef.current?.disconnect()
    }
  }, [lazy, priority, isInView])

  const handleLoad = () => {
    setIsLoaded(true)
    onLoad?.()
  }

  const handleError = () => {
    setIsError(true)
    onError?.()
  }

  // Generate optimized image URL
  const optimizedSrc = isInView ? ImagesService.getOptimizedImageUrl(src, {
    width,
    height,
    quality,
    format,
    fit
  }) : ''

  // Generate srcSet for responsive images
  const generateSrcSet = () => {
    if (!isInView) return ''
    
    const breakpoints = [480, 768, 1024, 1280, 1920]
    const srcSet = breakpoints
      .filter(bp => !width || bp <= width * 2) // Only include relevant breakpoints
      .map(bp => {
        const url = ImagesService.getOptimizedImageUrl(src, {
          width: bp,
          height: height ? Math.round((height / (width || bp)) * bp) : undefined,
          quality,
          format,
          fit
        })
        return `${url} ${bp}w`
      })
      .join(', ')
    
    return srcSet
  }

  const srcSet = generateSrcSet()

  // Placeholder component
  const PlaceholderComponent = () => (
    <div 
      className={`bg-gray-200 animate-pulse ${className}`}
      style={{
        width: width ? `${width}px` : '100%',
        height: height ? `${height}px` : '100%',
        aspectRatio: width && height ? `${width}/${height}` : undefined,
        ...style
      }}
    >
      {placeholder && (
        <div className="flex items-center justify-center h-full text-gray-400 text-sm">
          {placeholder}
        </div>
      )}
    </div>
  )

  // Error component
  const ErrorComponent = () => (
    <div 
      className={`bg-gray-100 border border-gray-200 flex items-center justify-center ${className}`}
      style={{
        width: width ? `${width}px` : '100%',
        height: height ? `${height}px` : '100%',
        aspectRatio: width && height ? `${width}/${height}` : undefined,
        ...style
      }}
    >
      <div className="text-center text-gray-400">
        <svg className="w-8 h-8 mx-auto mb-2" fill="currentColor" viewBox="0 0 20 20">
          <path fillRule="evenodd" d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z" clipRule="evenodd" />
        </svg>
        <p className="text-xs">Failed to load</p>
      </div>
    </div>
  )

  if (isError) {
    return <ErrorComponent />
  }

  if (!isInView) {
    return (
      <div ref={imgRef}>
        <PlaceholderComponent />
      </div>
    )
  }

  return (
    <div className="relative">
      {/* Blur placeholder */}
      {blurDataURL && !isLoaded && (
        <img
          src={blurDataURL}
          alt=""
          className={`absolute inset-0 w-full h-full object-cover filter blur-sm ${className}`}
          style={style}
        />
      )}
      
      {/* Main image */}
      <img
        ref={imgRef}
        src={optimizedSrc}
        srcSet={srcSet}
        sizes={sizes}
        alt={alt}
        width={width}
        height={height}
        loading={lazy && !priority ? 'lazy' : 'eager'}
        className={`transition-opacity duration-300 ${
          isLoaded ? 'opacity-100' : 'opacity-0'
        } ${className}`}
        style={style}
        onLoad={handleLoad}
        onError={handleError}
        onClick={onClick}
      />
      
      {/* Loading placeholder */}
      {!isLoaded && !blurDataURL && (
        <div className="absolute inset-0">
          <PlaceholderComponent />
        </div>
      )}
    </div>
  )
}

// Hero image component with advanced features
interface HeroImageProps {
  src: string
  alt: string
  overlay?: boolean
  overlayOpacity?: number
  children?: React.ReactNode
  className?: string
  imageClassName?: string
  priority?: boolean
}

export function HeroImage({
  src,
  alt,
  overlay = false,
  overlayOpacity = 0.4,
  children,
  className = '',
  imageClassName = '',
  priority = true
}: HeroImageProps) {
  return (
    <div className={`relative overflow-hidden ${className}`}>
      <OptimizedImage
        src={src}
        alt={alt}
        width={1920}
        height={1080}
        quality={0.9}
        format="webp"
        fit="cover"
        lazy={!priority}
        priority={priority}
        sizes="100vw"
        className={`w-full h-full object-cover ${imageClassName}`}
      />
      
      {overlay && (
        <div 
          className="absolute inset-0 bg-black"
          style={{ opacity: overlayOpacity }}
        />
      )}
      
      {children && (
        <div className="absolute inset-0 flex items-center justify-center">
          {children}
        </div>
      )}
    </div>
  )
}

// Avatar component with fallback
interface AvatarImageProps {
  src?: string
  alt: string
  size?: number
  fallback?: string
  className?: string
  onClick?: () => void
}

export function AvatarImage({
  src,
  alt,
  size = 40,
  fallback,
  className = '',
  onClick
}: AvatarImageProps) {
  const [hasError, setHasError] = useState(false)

  const initials = fallback || alt
    .split(' ')
    .map(word => word.charAt(0))
    .join('')
    .toUpperCase()
    .slice(0, 2)

  if (!src || hasError) {
    return (
      <div
        className={`flex items-center justify-center bg-primary text-primary-foreground font-medium rounded-full ${className}`}
        style={{ width: size, height: size, fontSize: size * 0.4 }}
        onClick={onClick}
      >
        {initials}
      </div>
    )
  }

  return (
    <OptimizedImage
      src={src}
      alt={alt}
      width={size}
      height={size}
      quality={0.9}
      format="webp"
      fit="cover"
      className={`rounded-full ${className}`}
      style={{ width: size, height: size }}
      onError={() => setHasError(true)}
      onClick={onClick}
    />
  )
}

// Thumbnail component
interface ThumbnailProps {
  src: string
  alt: string
  size?: number
  className?: string
  onClick?: () => void
}

export function Thumbnail({
  src,
  alt,
  size = 100,
  className = '',
  onClick
}: ThumbnailProps) {
  return (
    <div 
      className={`relative overflow-hidden rounded-lg cursor-pointer ${className}`}
      style={{ width: size, height: size }}
      onClick={onClick}
    >
      <OptimizedImage
        src={src}
        alt={alt}
        width={size}
        height={size}
        quality={0.8}
        format="webp"
        fit="cover"
        className="w-full h-full object-cover hover:scale-105 transition-transform duration-200"
      />
    </div>
  )
}

// Gallery thumbnail with selection
interface GalleryThumbnailProps {
  src: string
  alt: string
  selected?: boolean
  size?: number
  className?: string
  onClick?: () => void
}

export function GalleryThumbnail({
  src,
  alt,
  selected = false,
  size = 100,
  className = '',
  onClick
}: GalleryThumbnailProps) {
  return (
    <div 
      className={`relative overflow-hidden rounded-lg cursor-pointer border-2 transition-colors ${
        selected ? 'border-primary' : 'border-gray-200 hover:border-gray-300'
      } ${className}`}
      style={{ width: size, height: size }}
      onClick={onClick}
    >
      <OptimizedImage
        src={src}
        alt={alt}
        width={size}
        height={size}
        quality={0.8}
        format="webp"
        fit="cover"
        className="w-full h-full object-cover"
      />
      
      {selected && (
        <div className="absolute inset-0 bg-primary bg-opacity-20 flex items-center justify-center">
          <div className="w-6 h-6 bg-primary rounded-full flex items-center justify-center">
            <svg className="w-4 h-4 text-white" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
            </svg>
          </div>
        </div>
      )}
    </div>
  )
}
