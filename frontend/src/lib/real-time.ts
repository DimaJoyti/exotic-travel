/**
 * Real-time Updates Service for Exotic Travel Booking Platform
 * 
 * Provides real-time availability and pricing updates using WebSocket simulation
 */

import { Destination } from '@/types'

export interface RealTimeUpdate {
  type: 'availability' | 'price' | 'booking' | 'view_count'
  destination_id: string
  data: any
  timestamp: string
}

export interface AvailabilityUpdate {
  destination_id: string
  available_dates: string[]
  unavailable_dates: string[]
  last_updated: string
}

export interface PriceUpdate {
  destination_id: string
  old_price: number
  new_price: number
  change_percentage: number
  reason: 'demand' | 'seasonal' | 'promotion' | 'last_minute'
  valid_until: string
}

export interface BookingUpdate {
  destination_id: string
  recent_bookings: number
  total_bookings_today: number
  trending_score: number
}

export interface ViewCountUpdate {
  destination_id: string
  current_viewers: number
  total_views_today: number
  popularity_rank: number
}

type UpdateListener = (update: RealTimeUpdate) => void

class RealTimeService {
  private static listeners: Set<UpdateListener> = new Set()
  private static isConnected = false
  private static reconnectAttempts = 0
  private static maxReconnectAttempts = 5
  private static reconnectDelay = 1000
  private static updateInterval: NodeJS.Timeout | null = null
  private static destinations: Destination[] = []

  // Connection Management
  static connect(destinations: Destination[] = []): Promise<void> {
    return new Promise((resolve) => {
      this.destinations = destinations
      this.isConnected = true
      this.reconnectAttempts = 0
      
      console.log('🔌 Real-time service connected')
      
      // Start simulated updates
      this.startSimulatedUpdates()
      
      // Simulate connection delay
      setTimeout(() => {
        this.notifyListeners({
          type: 'booking',
          destination_id: 'system',
          data: { message: 'Real-time updates connected' },
          timestamp: new Date().toISOString()
        })
        resolve()
      }, 500)
    })
  }

  static disconnect(): void {
    this.isConnected = false
    if (this.updateInterval) {
      clearInterval(this.updateInterval)
      this.updateInterval = null
    }
    console.log('🔌 Real-time service disconnected')
  }

  static isConnectedStatus(): boolean {
    return this.isConnected
  }

  // Event Listeners
  static addListener(callback: UpdateListener): () => void {
    this.listeners.add(callback)
    return () => this.listeners.delete(callback)
  }

  private static notifyListeners(update: RealTimeUpdate): void {
    this.listeners.forEach(callback => {
      try {
        callback(update)
      } catch (error) {
        console.error('Error in real-time listener:', error)
      }
    })
  }

  // Simulated Updates
  private static startSimulatedUpdates(): void {
    if (this.updateInterval) {
      clearInterval(this.updateInterval)
    }

    this.updateInterval = setInterval(() => {
      if (!this.isConnected || this.destinations.length === 0) return

      // Generate random updates
      const updateType = Math.random()
      
      if (updateType < 0.3) {
        this.simulatePriceUpdate()
      } else if (updateType < 0.6) {
        this.simulateAvailabilityUpdate()
      } else if (updateType < 0.8) {
        this.simulateBookingUpdate()
      } else {
        this.simulateViewCountUpdate()
      }
    }, 3000 + Math.random() * 7000) // Random interval between 3-10 seconds
  }

  private static simulatePriceUpdate(): void {
    const destination = this.getRandomDestination()
    if (!destination) return

    const changePercentage = (Math.random() - 0.5) * 20 // -10% to +10%
    const oldPrice = destination.price
    const newPrice = Math.round(oldPrice * (1 + changePercentage / 100))
    
    const reasons: PriceUpdate['reason'][] = ['demand', 'seasonal', 'promotion', 'last_minute']
    const reason = reasons[Math.floor(Math.random() * reasons.length)]

    const priceUpdate: PriceUpdate = {
      destination_id: destination.id.toString(),
      old_price: oldPrice,
      new_price: newPrice,
      change_percentage: Math.round(changePercentage * 100) / 100,
      reason,
      valid_until: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString() // 24 hours
    }

    this.notifyListeners({
      type: 'price',
      destination_id: destination.id.toString(),
      data: priceUpdate,
      timestamp: new Date().toISOString()
    })
  }

  private static simulateAvailabilityUpdate(): void {
    const destination = this.getRandomDestination()
    if (!destination) return

    // Generate some random available/unavailable dates
    const availableDates: string[] = []
    const unavailableDates: string[] = []
    
    for (let i = 0; i < 30; i++) {
      const date = new Date()
      date.setDate(date.getDate() + i)
      const dateString = date.toISOString().split('T')[0]
      
      if (Math.random() > 0.3) {
        availableDates.push(dateString)
      } else {
        unavailableDates.push(dateString)
      }
    }

    const availabilityUpdate: AvailabilityUpdate = {
      destination_id: destination.id.toString(),
      available_dates: availableDates,
      unavailable_dates: unavailableDates,
      last_updated: new Date().toISOString()
    }

    this.notifyListeners({
      type: 'availability',
      destination_id: destination.id.toString(),
      data: availabilityUpdate,
      timestamp: new Date().toISOString()
    })
  }

  private static simulateBookingUpdate(): void {
    const destination = this.getRandomDestination()
    if (!destination) return

    const bookingUpdate: BookingUpdate = {
      destination_id: destination.id.toString(),
      recent_bookings: Math.floor(Math.random() * 5) + 1,
      total_bookings_today: Math.floor(Math.random() * 50) + 10,
      trending_score: Math.random() * 100
    }

    this.notifyListeners({
      type: 'booking',
      destination_id: destination.id.toString(),
      data: bookingUpdate,
      timestamp: new Date().toISOString()
    })
  }

  private static simulateViewCountUpdate(): void {
    const destination = this.getRandomDestination()
    if (!destination) return

    const viewUpdate: ViewCountUpdate = {
      destination_id: destination.id.toString(),
      current_viewers: Math.floor(Math.random() * 20) + 1,
      total_views_today: Math.floor(Math.random() * 500) + 100,
      popularity_rank: Math.floor(Math.random() * 100) + 1
    }

    this.notifyListeners({
      type: 'view_count',
      destination_id: destination.id.toString(),
      data: viewUpdate,
      timestamp: new Date().toISOString()
    })
  }

  private static getRandomDestination(): Destination | null {
    if (this.destinations.length === 0) return null
    return this.destinations[Math.floor(Math.random() * this.destinations.length)]
  }

  // Manual Update Triggers
  static triggerPriceUpdate(destinationId: string, newPrice: number, reason: PriceUpdate['reason']): void {
    const destination = this.destinations.find(d => d.id.toString() === destinationId)
    if (!destination) return

    const changePercentage = ((newPrice - destination.price) / destination.price) * 100

    const priceUpdate: PriceUpdate = {
      destination_id: destinationId,
      old_price: destination.price,
      new_price: newPrice,
      change_percentage: Math.round(changePercentage * 100) / 100,
      reason,
      valid_until: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString()
    }

    this.notifyListeners({
      type: 'price',
      destination_id: destinationId,
      data: priceUpdate,
      timestamp: new Date().toISOString()
    })
  }

  static triggerBookingUpdate(destinationId: string, bookingCount: number): void {
    const bookingUpdate: BookingUpdate = {
      destination_id: destinationId,
      recent_bookings: bookingCount,
      total_bookings_today: Math.floor(Math.random() * 50) + bookingCount,
      trending_score: Math.min(100, bookingCount * 10)
    }

    this.notifyListeners({
      type: 'booking',
      destination_id: destinationId,
      data: bookingUpdate,
      timestamp: new Date().toISOString()
    })
  }

  // Connection Recovery
  private static async reconnect(): Promise<void> {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('❌ Max reconnection attempts reached')
      return
    }

    this.reconnectAttempts++
    console.log(`🔄 Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})`)

    await new Promise(resolve => setTimeout(resolve, this.reconnectDelay * this.reconnectAttempts))

    try {
      await this.connect(this.destinations)
      console.log('✅ Reconnected successfully')
    } catch (error) {
      console.error('❌ Reconnection failed:', error)
      this.reconnect()
    }
  }

  // Utility Methods
  static getConnectionStatus(): {
    connected: boolean
    reconnectAttempts: number
    lastUpdate: string | null
  } {
    return {
      connected: this.isConnected,
      reconnectAttempts: this.reconnectAttempts,
      lastUpdate: new Date().toISOString()
    }
  }

  static simulateConnectionLoss(): void {
    console.log('🔌 Simulating connection loss')
    this.disconnect()
    
    // Attempt to reconnect after a delay
    setTimeout(() => {
      this.reconnect()
    }, 2000)
  }

  // Analytics
  static getUpdateStats(): {
    totalUpdates: number
    updatesByType: Record<string, number>
    averageUpdateInterval: number
  } {
    // This would be implemented with actual tracking in a real application
    return {
      totalUpdates: Math.floor(Math.random() * 1000) + 100,
      updatesByType: {
        price: Math.floor(Math.random() * 100) + 20,
        availability: Math.floor(Math.random() * 100) + 30,
        booking: Math.floor(Math.random() * 100) + 40,
        view_count: Math.floor(Math.random() * 100) + 50
      },
      averageUpdateInterval: 5000
    }
  }
}

export { RealTimeService }
