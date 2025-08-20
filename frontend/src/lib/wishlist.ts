/**
 * Wishlist Service for Exotic Travel Booking Platform
 * 
 * Provides wishlist functionality with local storage and sync capabilities
 */

import { Destination } from '@/types'

export interface WishlistItem {
  id: string
  destination_id: string
  destination: Destination
  added_at: string
  notes?: string
  priority?: 'low' | 'medium' | 'high'
  tags?: string[]
}

export interface WishlistCollection {
  id: string
  name: string
  description?: string
  items: WishlistItem[]
  created_at: string
  updated_at: string
  is_public: boolean
  color?: string
}

export interface WishlistStats {
  total_items: number
  total_collections: number
  most_popular_destination: string
  average_price: number
  countries_count: number
}

class WishlistService {
  private static readonly STORAGE_KEY = 'exotic_travel_wishlist'
  private static readonly COLLECTIONS_KEY = 'exotic_travel_wishlist_collections'
  private static listeners: Set<() => void> = new Set()

  // Local Storage Management
  private static getStoredWishlist(): WishlistItem[] {
    if (typeof window === 'undefined') return []

    try {
      const stored = localStorage.getItem(this.STORAGE_KEY)
      return stored ? JSON.parse(stored) : []
    } catch (error) {
      console.error('Error reading wishlist from localStorage:', error)
      return []
    }
  }

  private static setStoredWishlist(items: WishlistItem[]): void {
    if (typeof window === 'undefined') return
    
    try {
      localStorage.setItem(this.STORAGE_KEY, JSON.stringify(items))
      this.notifyListeners()
    } catch (error) {
      console.error('Error saving wishlist to localStorage:', error)
    }
  }

  private static getStoredCollections(): WishlistCollection[] {
    if (typeof window === 'undefined') return []
    
    try {
      const stored = localStorage.getItem(this.COLLECTIONS_KEY)
      return stored ? JSON.parse(stored) : []
    } catch (error) {
      console.error('Error reading collections from localStorage:', error)
      return []
    }
  }

  private static setStoredCollections(collections: WishlistCollection[]): void {
    if (typeof window === 'undefined') return
    
    try {
      localStorage.setItem(this.COLLECTIONS_KEY, JSON.stringify(collections))
      this.notifyListeners()
    } catch (error) {
      console.error('Error saving collections to localStorage:', error)
    }
  }

  // Event Listeners
  static addListener(callback: () => void): () => void {
    this.listeners.add(callback)
    return () => this.listeners.delete(callback)
  }

  private static notifyListeners(): void {
    this.listeners.forEach(callback => callback())
  }

  // Wishlist Item Management
  static async addToWishlist(destination: Destination, notes?: string, priority?: 'low' | 'medium' | 'high'): Promise<WishlistItem> {
    const items = this.getStoredWishlist()
    
    // Check if already in wishlist
    const existingItem = items.find(item => item.destination_id === destination.id.toString())
    if (existingItem) {
      throw new Error('Destination is already in your wishlist')
    }

    const newItem: WishlistItem = {
      id: `wishlist_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      destination_id: destination.id.toString(),
      destination,
      added_at: new Date().toISOString(),
      notes,
      priority: priority || 'medium',
      tags: []
    }

    const updatedItems = [...items, newItem]
    this.setStoredWishlist(updatedItems)

    // Simulate API call for sync
    await this.syncToServer([newItem])

    return newItem
  }

  static async removeFromWishlist(destinationId: string): Promise<void> {
    const items = this.getStoredWishlist()
    const updatedItems = items.filter(item => item.destination_id !== destinationId)
    this.setStoredWishlist(updatedItems)

    // Simulate API call for sync
    await this.syncToServer(updatedItems)
  }

  static async updateWishlistItem(itemId: string, updates: Partial<WishlistItem>): Promise<WishlistItem> {
    const items = this.getStoredWishlist()
    const itemIndex = items.findIndex(item => item.id === itemId)
    
    if (itemIndex === -1) {
      throw new Error('Wishlist item not found')
    }

    const updatedItem = { ...items[itemIndex], ...updates }
    items[itemIndex] = updatedItem
    this.setStoredWishlist(items)

    // Simulate API call for sync
    await this.syncToServer(items)

    return updatedItem
  }

  static getWishlist(): WishlistItem[] {
    return this.getStoredWishlist()
  }

  static isInWishlist(destinationId: string): boolean {
    const items = this.getStoredWishlist()
    return items.some(item => item.destination_id === destinationId)
  }

  static getWishlistItem(destinationId: string): WishlistItem | null {
    const items = this.getStoredWishlist()
    return items.find(item => item.destination_id === destinationId) || null
  }

  // Collection Management
  static async createCollection(name: string, description?: string, color?: string): Promise<WishlistCollection> {
    const collections = this.getStoredCollections()
    
    const newCollection: WishlistCollection = {
      id: `collection_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      name,
      description,
      items: [],
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      is_public: false,
      color: color || '#3B82F6'
    }

    const updatedCollections = [...collections, newCollection]
    this.setStoredCollections(updatedCollections)

    return newCollection
  }

  static async addToCollection(collectionId: string, wishlistItemId: string): Promise<void> {
    const collections = this.getStoredCollections()
    const items = this.getStoredWishlist()
    
    const collection = collections.find(c => c.id === collectionId)
    const item = items.find(i => i.id === wishlistItemId)
    
    if (!collection || !item) {
      throw new Error('Collection or item not found')
    }

    if (!collection.items.find(i => i.id === wishlistItemId)) {
      collection.items.push(item)
      collection.updated_at = new Date().toISOString()
      this.setStoredCollections(collections)
    }
  }

  static getCollections(): WishlistCollection[] {
    return this.getStoredCollections()
  }

  static async deleteCollection(collectionId: string): Promise<void> {
    const collections = this.getStoredCollections()
    const updatedCollections = collections.filter(c => c.id !== collectionId)
    this.setStoredCollections(updatedCollections)
  }

  // Statistics
  static getWishlistStats(): WishlistStats {
    const items = this.getStoredWishlist()
    const collections = this.getStoredCollections()

    if (items.length === 0) {
      return {
        total_items: 0,
        total_collections: collections.length,
        most_popular_destination: '',
        average_price: 0,
        countries_count: 0
      }
    }

    const countries = new Set(items.map(item => item.destination.country))
    const totalPrice = items.reduce((sum, item) => sum + item.destination.price, 0)
    
    // Find most common destination (simplified)
    const destinationCounts = items.reduce((acc, item) => {
      acc[item.destination.name] = (acc[item.destination.name] || 0) + 1
      return acc
    }, {} as Record<string, number>)
    
    const mostPopular = Object.entries(destinationCounts)
      .sort(([,a], [,b]) => b - a)[0]?.[0] || ''

    return {
      total_items: items.length,
      total_collections: collections.length,
      most_popular_destination: mostPopular,
      average_price: Math.round(totalPrice / items.length),
      countries_count: countries.size
    }
  }

  // Search and Filter
  static searchWishlist(query: string): WishlistItem[] {
    const items = this.getStoredWishlist()
    const lowercaseQuery = query.toLowerCase()
    
    return items.filter(item => 
      item.destination.name.toLowerCase().includes(lowercaseQuery) ||
      item.destination.country.toLowerCase().includes(lowercaseQuery) ||
      item.destination.city.toLowerCase().includes(lowercaseQuery) ||
      item.notes?.toLowerCase().includes(lowercaseQuery) ||
      item.tags?.some(tag => tag.toLowerCase().includes(lowercaseQuery))
    )
  }

  static filterWishlist(filters: {
    priority?: 'low' | 'medium' | 'high'
    priceRange?: [number, number]
    countries?: string[]
    tags?: string[]
  }): WishlistItem[] {
    const items = this.getStoredWishlist()
    
    return items.filter(item => {
      if (filters.priority && item.priority !== filters.priority) return false
      
      if (filters.priceRange) {
        const [min, max] = filters.priceRange
        if (item.destination.price < min || item.destination.price > max) return false
      }
      
      if (filters.countries && filters.countries.length > 0) {
        if (!filters.countries.includes(item.destination.country)) return false
      }
      
      if (filters.tags && filters.tags.length > 0) {
        if (!item.tags || !filters.tags.some(tag => item.tags!.includes(tag))) return false
      }
      
      return true
    })
  }

  // Sync with Server (Mock Implementation)
  private static async syncToServer(items: WishlistItem[]): Promise<void> {
    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 100))
    
    console.log('🔄 Wishlist synced to server:', {
      items_count: items.length,
      timestamp: new Date().toISOString()
    })
  }

  // Export/Import
  static exportWishlist(): string {
    const data = {
      wishlist: this.getStoredWishlist(),
      collections: this.getStoredCollections(),
      exported_at: new Date().toISOString()
    }
    
    return JSON.stringify(data, null, 2)
  }

  static async importWishlist(jsonData: string): Promise<void> {
    try {
      const data = JSON.parse(jsonData)
      
      if (data.wishlist) {
        this.setStoredWishlist(data.wishlist)
      }
      
      if (data.collections) {
        this.setStoredCollections(data.collections)
      }
      
      await this.syncToServer(data.wishlist || [])
    } catch (error) {
      throw new Error('Invalid wishlist data format')
    }
  }

  // Clear all data
  static async clearWishlist(): Promise<void> {
    this.setStoredWishlist([])
    this.setStoredCollections([])
    await this.syncToServer([])
  }
}

export { WishlistService }
