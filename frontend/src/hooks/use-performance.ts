import { useEffect, useRef, useCallback, useMemo } from 'react'
import { usePerformanceMonitor } from '@/lib/performance'

// Hook for measuring component render performance
export function useRenderPerformance(componentName: string) {
  const { markStart, markEnd, recordMetric } = usePerformanceMonitor()
  const renderCount = useRef(0)
  const mountTime = useRef<number>(0)

  useEffect(() => {
    mountTime.current = performance.now()
    markStart(`${componentName}-mount`)
    
    return () => {
      markEnd(`${componentName}-mount`)
      const totalRenderTime = performance.now() - mountTime.current
      recordMetric(`${componentName}-total-time`, totalRenderTime)
      recordMetric(`${componentName}-render-count`, renderCount.current)
    }
  }, [componentName, markStart, markEnd, recordMetric])

  useEffect(() => {
    renderCount.current++
    markStart(`${componentName}-render-${renderCount.current}`)
    
    return () => {
      markEnd(`${componentName}-render-${renderCount.current}`)
    }
  })

  return {
    renderCount: renderCount.current,
    recordCustomMetric: recordMetric,
  }
}

// Hook for debouncing expensive operations
export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value)

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value)
    }, delay)

    return () => {
      clearTimeout(handler)
    }
  }, [value, delay])

  return debouncedValue
}

// Hook for throttling function calls
export function useThrottle<T extends (...args: any[]) => any>(
  func: T,
  delay: number
): T {
  const lastCall = useRef<number>(0)
  const timeoutRef = useRef<NodeJS.Timeout | null>(null)

  return useCallback(
    ((...args: Parameters<T>) => {
      const now = Date.now()
      
      if (now - lastCall.current >= delay) {
        lastCall.current = now
        return func(...args)
      } else {
        if (timeoutRef.current) {
          clearTimeout(timeoutRef.current)
        }
        
        timeoutRef.current = setTimeout(() => {
          lastCall.current = Date.now()
          func(...args)
        }, delay - (now - lastCall.current))
      }
    }) as T,
    [func, delay]
  )
}

// Hook for lazy loading with Intersection Observer
export function useLazyLoad(
  threshold: number = 0.1,
  rootMargin: string = '50px'
) {
  const [isVisible, setIsVisible] = useState(false)
  const [isLoaded, setIsLoaded] = useState(false)
  const elementRef = useRef<HTMLElement>(null)

  useEffect(() => {
    const element = elementRef.current
    if (!element) return

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting && !isLoaded) {
          setIsVisible(true)
          setIsLoaded(true)
          observer.unobserve(element)
        }
      },
      { threshold, rootMargin }
    )

    observer.observe(element)

    return () => {
      observer.unobserve(element)
    }
  }, [threshold, rootMargin, isLoaded])

  return { ref: elementRef, isVisible, isLoaded }
}

// Hook for measuring API call performance
export function useApiPerformance() {
  const { markStart, markEnd, recordMetric } = usePerformanceMonitor()

  const measureApiCall = useCallback(
    async <T>(
      apiName: string,
      apiCall: () => Promise<T>
    ): Promise<T> => {
      const startTime = performance.now()
      markStart(`api-${apiName}`)

      try {
        const result = await apiCall()
        const endTime = performance.now()
        const duration = endTime - startTime
        
        markEnd(`api-${apiName}`)
        recordMetric(`api-${apiName}-success`, duration)
        recordMetric(`api-${apiName}-status`, 200)
        
        return result
      } catch (error) {
        const endTime = performance.now()
        const duration = endTime - startTime
        
        markEnd(`api-${apiName}`)
        recordMetric(`api-${apiName}-error`, duration)
        recordMetric(`api-${apiName}-status`, 500)
        
        throw error
      }
    },
    [markStart, markEnd, recordMetric]
  )

  return { measureApiCall }
}

// Hook for optimizing expensive calculations
export function useExpensiveCalculation<T>(
  calculation: () => T,
  dependencies: React.DependencyList,
  componentName?: string
): T {
  const { recordMetric } = usePerformanceMonitor()

  return useMemo(() => {
    const startTime = performance.now()
    const result = calculation()
    const endTime = performance.now()
    const duration = endTime - startTime

    if (componentName) {
      recordMetric(`${componentName}-calculation`, duration)
    }

    // Warn if calculation takes too long
    if (duration > 16) { // More than one frame at 60fps
      console.warn(
        `Expensive calculation detected: ${componentName || 'unknown'} took ${duration.toFixed(2)}ms`
      )
    }

    return result
  }, dependencies)
}

// Hook for virtual scrolling performance
export function useVirtualScrolling<T>(
  items: T[],
  itemHeight: number,
  containerHeight: number
) {
  const [scrollTop, setScrollTop] = useState(0)

  const visibleItems = useMemo(() => {
    const startIndex = Math.floor(scrollTop / itemHeight)
    const endIndex = Math.min(
      startIndex + Math.ceil(containerHeight / itemHeight) + 1,
      items.length
    )

    return {
      startIndex,
      endIndex,
      items: items.slice(startIndex, endIndex),
      totalHeight: items.length * itemHeight,
      offsetY: startIndex * itemHeight,
    }
  }, [items, itemHeight, containerHeight, scrollTop])

  const handleScroll = useThrottle((event: React.UIEvent<HTMLDivElement>) => {
    setScrollTop(event.currentTarget.scrollTop)
  }, 16) // Throttle to 60fps

  return {
    visibleItems,
    handleScroll,
    scrollTop,
  }
}

// Hook for image optimization
export function useImageOptimization() {
  const [isWebPSupported, setIsWebPSupported] = useState<boolean | null>(null)
  const [isAvifSupported, setIsAvifSupported] = useState<boolean | null>(null)

  useEffect(() => {
    // Check WebP support
    const webpTest = new Image()
    webpTest.onload = webpTest.onerror = () => {
      setIsWebPSupported(webpTest.height === 2)
    }
    webpTest.src = 'data:image/webp;base64,UklGRjoAAABXRUJQVlA4IC4AAACyAgCdASoCAAIALmk0mk0iIiIiIgBoSygABc6WWgAA/veff/0PP8bA//LwYAAA'

    // Check AVIF support
    const avifTest = new Image()
    avifTest.onload = avifTest.onerror = () => {
      setIsAvifSupported(avifTest.height === 2)
    }
    avifTest.src = 'data:image/avif;base64,AAAAIGZ0eXBhdmlmAAAAAGF2aWZtaWYxbWlhZk1BMUIAAADybWV0YQAAAAAAAAAoaGRscgAAAAAAAAAAcGljdAAAAAAAAAAAAAAAAGxpYmF2aWYAAAAADnBpdG0AAAAAAAEAAAAeaWxvYwAAAABEAAABAAEAAAABAAABGgAAAB0AAAAoaWluZgAAAAAAAQAAABppbmZlAgAAAAABAABhdjAxQ29sb3IAAAAAamlwcnAAAABLaXBjbwAAABRpc3BlAAAAAAAAAAIAAAACAAAAEHBpeGkAAAAAAwgICAAAAAxhdjFDgQ0MAAAAABNjb2xybmNseAACAAIAAYAAAAAXaXBtYQAAAAAAAAABAAEEAQKDBAAAACVtZGF0EgAKCBgABogQEAwgMg8f8D///8WfhwB8+ErK42A='
  }, [])

  const getOptimizedImageUrl = useCallback(
    (originalUrl: string, width?: number, height?: number) => {
      if (!originalUrl) return originalUrl

      // If it's already a Next.js optimized image, return as is
      if (originalUrl.includes('/_next/image')) {
        return originalUrl
      }

      // Build optimized URL
      const params = new URLSearchParams()
      params.set('url', originalUrl)
      
      if (width) params.set('w', width.toString())
      if (height) params.set('h', height.toString())
      
      // Use best supported format
      if (isAvifSupported) {
        params.set('f', 'avif')
      } else if (isWebPSupported) {
        params.set('f', 'webp')
      }

      return `/_next/image?${params.toString()}`
    },
    [isWebPSupported, isAvifSupported]
  )

  return {
    isWebPSupported,
    isAvifSupported,
    getOptimizedImageUrl,
  }
}

// Hook for bundle size monitoring
export function useBundleMonitoring() {
  const { recordMetric } = usePerformanceMonitor()

  useEffect(() => {
    if (typeof window !== 'undefined' && 'performance' in window) {
      // Monitor JavaScript bundle sizes
      const resources = performance.getEntriesByType('resource') as PerformanceResourceTiming[]
      
      let totalJSSize = 0
      let totalCSSSize = 0
      
      resources.forEach(resource => {
        if (resource.name.includes('.js')) {
          totalJSSize += resource.transferSize || 0
        } else if (resource.name.includes('.css')) {
          totalCSSSize += resource.transferSize || 0
        }
      })

      recordMetric('bundle-js-size', totalJSSize)
      recordMetric('bundle-css-size', totalCSSSize)
      recordMetric('bundle-total-size', totalJSSize + totalCSSSize)
    }
  }, [recordMetric])
}

// Import useState for hooks that need it
import { useState } from 'react'
