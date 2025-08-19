'use client'

import { useState, useEffect } from 'react'
import { useSearchParams } from 'next/navigation'

/**
 * Safe search params hook that handles Next.js 15 issues
 *
 * This hook provides a fallback mechanism for accessing URL search parameters
 * when the built-in useSearchParams hook fails due to runtime errors like:
 * "Cannot read properties of undefined (reading 'call')"
 *
 * The hook tries to use Next.js useSearchParams first, but falls back to
 * manual URL parsing if that fails, ensuring the app continues to work.
 *
 * @returns URLSearchParams object or null if not available
 */
export function useSafeSearchParams(): URLSearchParams | null {
  const [params, setParams] = useState<URLSearchParams | null>(null)
  
  useEffect(() => {
    // Fallback to manual URL parsing if useSearchParams fails
    try {
      if (typeof window !== 'undefined') {
        const urlParams = new URLSearchParams(window.location.search)
        setParams(urlParams)
      }
    } catch (error) {
      console.warn('Error parsing search params:', error)
      setParams(new URLSearchParams())
    }
  }, [])

  // Try to use Next.js useSearchParams, but fall back to manual parsing
  let nextSearchParams: URLSearchParams | null = null
  try {
    nextSearchParams = useSearchParams()
  } catch (error) {
    console.warn('useSearchParams failed, using fallback:', error)
  }

  return nextSearchParams || params
}

/**
 * Get a specific search parameter value safely
 * 
 * @param key - The parameter key to retrieve
 * @param defaultValue - Default value if parameter is not found
 * @returns The parameter value or default value
 */
export function useSafeSearchParam(key: string, defaultValue: string = ''): string {
  const searchParams = useSafeSearchParams()
  
  try {
    return searchParams?.get(key) || defaultValue
  } catch (error) {
    console.warn(`Error getting search param "${key}":`, error)
    return defaultValue
  }
}

/**
 * Get multiple search parameters safely
 *
 * @param keys - Array of parameter keys to retrieve
 * @returns Object with parameter values
 */
export function useSafeSearchParamsMultiple(keys: string[]): Record<string, string> {
  const searchParams = useSafeSearchParams()
  const result: Record<string, string> = {}

  keys.forEach(key => {
    try {
      result[key] = searchParams?.get(key) || ''
    } catch (error) {
      console.warn(`Error getting search param "${key}":`, error)
      result[key] = ''
    }
  })

  return result
}
