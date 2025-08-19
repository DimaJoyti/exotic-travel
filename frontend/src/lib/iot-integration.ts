/**
 * IoT Integration Service for Smart Travel Features
 * 
 * Integrates with various IoT devices and sensors to provide
 * smart travel experiences and real-time environmental data
 */

export interface IoTDevice {
  id: string
  name: string
  type: 'weather_station' | 'air_quality' | 'crowd_sensor' | 'smart_lock' | 'beacon' | 'camera' | 'environmental'
  location: {
    latitude: number
    longitude: number
    address: string
    destination_id?: string
  }
  status: 'online' | 'offline' | 'maintenance'
  last_update: string
  battery_level?: number
  signal_strength?: number
}

export interface SensorData {
  device_id: string
  timestamp: string
  data: Record<string, any>
  quality_score: number
  confidence: number
}

export interface WeatherData extends SensorData {
  data: {
    temperature: number
    humidity: number
    pressure: number
    wind_speed: number
    wind_direction: number
    precipitation: number
    uv_index: number
    visibility: number
    weather_condition: string
  }
}

export interface AirQualityData extends SensorData {
  data: {
    aqi: number
    pm25: number
    pm10: number
    co2: number
    ozone: number
    no2: number
    so2: number
    health_recommendation: string
  }
}

export interface CrowdData extends SensorData {
  data: {
    crowd_density: number
    estimated_people: number
    noise_level: number
    wait_time_minutes: number
    congestion_level: 'low' | 'medium' | 'high' | 'very_high'
  }
}

export interface SmartRecommendation {
  id: string
  type: 'weather_alert' | 'crowd_avoidance' | 'air_quality_warning' | 'optimal_timing' | 'route_suggestion'
  title: string
  message: string
  priority: 'low' | 'medium' | 'high' | 'critical'
  destination_id: string
  valid_until: string
  action_required: boolean
  suggested_actions: string[]
}

export interface GeofenceEvent {
  id: string
  user_id: string
  device_id: string
  event_type: 'enter' | 'exit' | 'dwell'
  location: {
    latitude: number
    longitude: number
    radius: number
    name: string
  }
  timestamp: string
  duration?: number
  metadata: Record<string, any>
}

class IoTIntegrationService {
  private static devices: Map<string, IoTDevice> = new Map()
  private static sensorData: Map<string, SensorData[]> = new Map()
  private static eventListeners: Map<string, Function[]> = new Map()
  private static geofences: Map<string, any> = new Map()
  private static isInitialized = false

  // Initialize IoT service
  static async initialize(): Promise<void> {
    if (this.isInitialized) return

    try {
      await this.discoverDevices()
      await this.setupGeofencing()
      this.startDataCollection()
      this.isInitialized = true
      
      console.log('IoT Integration Service initialized')
    } catch (error) {
      console.error('Failed to initialize IoT service:', error)
    }
  }

  // Device Discovery and Management
  static async discoverDevices(): Promise<IoTDevice[]> {
    // Simulate device discovery
    const mockDevices: IoTDevice[] = [
      {
        id: 'weather_001',
        name: 'Santorini Weather Station',
        type: 'weather_station',
        location: {
          latitude: 36.3932,
          longitude: 25.4615,
          address: 'Santorini, Greece',
          destination_id: '1'
        },
        status: 'online',
        last_update: new Date().toISOString(),
        battery_level: 85,
        signal_strength: 92
      },
      {
        id: 'air_quality_001',
        name: 'Athens Air Quality Monitor',
        type: 'air_quality',
        location: {
          latitude: 37.9838,
          longitude: 23.7275,
          address: 'Athens, Greece'
        },
        status: 'online',
        last_update: new Date().toISOString(),
        battery_level: 78,
        signal_strength: 88
      },
      {
        id: 'crowd_001',
        name: 'Acropolis Crowd Sensor',
        type: 'crowd_sensor',
        location: {
          latitude: 37.9715,
          longitude: 23.7267,
          address: 'Acropolis, Athens, Greece'
        },
        status: 'online',
        last_update: new Date().toISOString(),
        battery_level: 92,
        signal_strength: 95
      }
    ]

    mockDevices.forEach(device => {
      this.devices.set(device.id, device)
    })

    return mockDevices
  }

  static getDevice(deviceId: string): IoTDevice | null {
    return this.devices.get(deviceId) || null
  }

  static getDevicesByLocation(latitude: number, longitude: number, radiusKm: number = 10): IoTDevice[] {
    const devices: IoTDevice[] = []
    
    this.devices.forEach(device => {
      const distance = this.calculateDistance(
        latitude, longitude,
        device.location.latitude, device.location.longitude
      )
      
      if (distance <= radiusKm) {
        devices.push(device)
      }
    })
    
    return devices
  }

  // Sensor Data Collection
  static async getLatestSensorData(deviceId: string): Promise<SensorData | null> {
    const deviceData = this.sensorData.get(deviceId)
    if (!deviceData || deviceData.length === 0) {
      return null
    }
    
    return deviceData[deviceData.length - 1]
  }

  static async getWeatherData(destinationId: string): Promise<WeatherData | null> {
    // Find weather devices for destination
    const devices = Array.from(this.devices.values()).filter(
      device => device.type === 'weather_station' && device.location.destination_id === destinationId
    )
    
    if (devices.length === 0) return null
    
    // Get latest data from first available device
    const latestData = await this.getLatestSensorData(devices[0].id)
    if (!latestData) {
      // Generate mock weather data
      return this.generateMockWeatherData(devices[0].id)
    }
    
    return latestData as WeatherData
  }

  static async getAirQualityData(latitude: number, longitude: number): Promise<AirQualityData | null> {
    const nearbyDevices = this.getDevicesByLocation(latitude, longitude, 50)
      .filter(device => device.type === 'air_quality')
    
    if (nearbyDevices.length === 0) return null
    
    const latestData = await this.getLatestSensorData(nearbyDevices[0].id)
    if (!latestData) {
      return this.generateMockAirQualityData(nearbyDevices[0].id)
    }
    
    return latestData as AirQualityData
  }

  static async getCrowdData(destinationId: string): Promise<CrowdData | null> {
    const devices = Array.from(this.devices.values()).filter(
      device => device.type === 'crowd_sensor' && device.location.destination_id === destinationId
    )
    
    if (devices.length === 0) return null
    
    const latestData = await this.getLatestSensorData(devices[0].id)
    if (!latestData) {
      return this.generateMockCrowdData(devices[0].id)
    }
    
    return latestData as CrowdData
  }

  // Smart Recommendations
  static async generateSmartRecommendations(destinationId: string): Promise<SmartRecommendation[]> {
    const recommendations: SmartRecommendation[] = []
    
    try {
      // Weather-based recommendations
      const weatherData = await this.getWeatherData(destinationId)
      if (weatherData) {
        const weatherRecs = this.analyzeWeatherData(weatherData, destinationId)
        recommendations.push(...weatherRecs)
      }
      
      // Air quality recommendations
      const destination = this.devices.get(destinationId)
      if (destination) {
        const airQualityData = await this.getAirQualityData(
          destination.location.latitude,
          destination.location.longitude
        )
        if (airQualityData) {
          const airRecs = this.analyzeAirQualityData(airQualityData, destinationId)
          recommendations.push(...airRecs)
        }
      }
      
      // Crowd-based recommendations
      const crowdData = await this.getCrowdData(destinationId)
      if (crowdData) {
        const crowdRecs = this.analyzeCrowdData(crowdData, destinationId)
        recommendations.push(...crowdRecs)
      }
      
    } catch (error) {
      console.error('Error generating smart recommendations:', error)
    }
    
    return recommendations.sort((a, b) => {
      const priorityOrder = { critical: 4, high: 3, medium: 2, low: 1 }
      return priorityOrder[b.priority] - priorityOrder[a.priority]
    })
  }

  // Geofencing
  static async setupGeofencing(): Promise<void> {
    if (!navigator.geolocation) {
      console.warn('Geolocation not supported')
      return
    }

    // Set up geofences around popular destinations
    const geofences = [
      {
        id: 'santorini_center',
        name: 'Santorini Center',
        latitude: 36.3932,
        longitude: 25.4615,
        radius: 1000, // 1km
        destination_id: '1'
      },
      {
        id: 'acropolis',
        name: 'Acropolis',
        latitude: 37.9715,
        longitude: 23.7267,
        radius: 500, // 500m
        destination_id: '2'
      }
    ]

    geofences.forEach(geofence => {
      this.geofences.set(geofence.id, geofence)
    })

    // Start location monitoring
    this.startLocationMonitoring()
  }

  static startLocationMonitoring(): void {
    if (!navigator.geolocation) return

    navigator.geolocation.watchPosition(
      (position) => {
        this.checkGeofences(position.coords.latitude, position.coords.longitude)
      },
      (error) => {
        console.error('Geolocation error:', error)
      },
      {
        enableHighAccuracy: true,
        timeout: 10000,
        maximumAge: 60000
      }
    )
  }

  private static checkGeofences(latitude: number, longitude: number): void {
    this.geofences.forEach((geofence, id) => {
      const distance = this.calculateDistance(
        latitude, longitude,
        geofence.latitude, geofence.longitude
      )

      const isInside = distance <= (geofence.radius / 1000) // Convert to km

      // Trigger geofence events
      if (isInside) {
        this.triggerGeofenceEvent('enter', geofence, latitude, longitude)
      }
    })
  }

  private static triggerGeofenceEvent(
    eventType: 'enter' | 'exit' | 'dwell',
    geofence: any,
    latitude: number,
    longitude: number
  ): void {
    const event: GeofenceEvent = {
      id: `event_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      user_id: 'current_user', // Would get from auth context
      device_id: geofence.id,
      event_type: eventType,
      location: {
        latitude,
        longitude,
        radius: geofence.radius,
        name: geofence.name
      },
      timestamp: new Date().toISOString(),
      metadata: {
        destination_id: geofence.destination_id
      }
    }

    // Emit event to listeners
    this.emitEvent('geofence', event)
  }

  // Event System
  static addEventListener(eventType: string, callback: Function): void {
    if (!this.eventListeners.has(eventType)) {
      this.eventListeners.set(eventType, [])
    }
    this.eventListeners.get(eventType)!.push(callback)
  }

  static removeEventListener(eventType: string, callback: Function): void {
    const listeners = this.eventListeners.get(eventType)
    if (listeners) {
      const index = listeners.indexOf(callback)
      if (index > -1) {
        listeners.splice(index, 1)
      }
    }
  }

  private static emitEvent(eventType: string, data: any): void {
    const listeners = this.eventListeners.get(eventType)
    if (listeners) {
      listeners.forEach(callback => callback(data))
    }
  }

  // Data Analysis Methods
  private static analyzeWeatherData(data: WeatherData, destinationId: string): SmartRecommendation[] {
    const recommendations: SmartRecommendation[] = []
    
    if (data.data.temperature > 35) {
      recommendations.push({
        id: `weather_hot_${Date.now()}`,
        type: 'weather_alert',
        title: 'High Temperature Alert',
        message: `Temperature is ${data.data.temperature}°C. Stay hydrated and seek shade during peak hours.`,
        priority: 'high',
        destination_id: destinationId,
        valid_until: new Date(Date.now() + 6 * 60 * 60 * 1000).toISOString(),
        action_required: true,
        suggested_actions: ['Carry water', 'Wear sunscreen', 'Plan indoor activities during 12-4 PM']
      })
    }
    
    if (data.data.precipitation > 5) {
      recommendations.push({
        id: `weather_rain_${Date.now()}`,
        type: 'weather_alert',
        title: 'Rain Expected',
        message: `${data.data.precipitation}mm of rain expected. Plan indoor activities or carry an umbrella.`,
        priority: 'medium',
        destination_id: destinationId,
        valid_until: new Date(Date.now() + 12 * 60 * 60 * 1000).toISOString(),
        action_required: false,
        suggested_actions: ['Carry umbrella', 'Plan indoor activities', 'Check covered attractions']
      })
    }
    
    return recommendations
  }

  private static analyzeAirQualityData(data: AirQualityData, destinationId: string): SmartRecommendation[] {
    const recommendations: SmartRecommendation[] = []
    
    if (data.data.aqi > 150) {
      recommendations.push({
        id: `air_quality_${Date.now()}`,
        type: 'air_quality_warning',
        title: 'Poor Air Quality',
        message: `Air Quality Index is ${data.data.aqi}. Consider limiting outdoor activities.`,
        priority: 'high',
        destination_id: destinationId,
        valid_until: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
        action_required: true,
        suggested_actions: ['Wear a mask', 'Limit outdoor exercise', 'Stay indoors when possible']
      })
    }
    
    return recommendations
  }

  private static analyzeCrowdData(data: CrowdData, destinationId: string): SmartRecommendation[] {
    const recommendations: SmartRecommendation[] = []
    
    if (data.data.congestion_level === 'very_high') {
      recommendations.push({
        id: `crowd_${Date.now()}`,
        type: 'crowd_avoidance',
        title: 'High Crowd Density',
        message: `Very crowded with ${data.data.wait_time_minutes} min wait time. Consider visiting later.`,
        priority: 'medium',
        destination_id: destinationId,
        valid_until: new Date(Date.now() + 2 * 60 * 60 * 1000).toISOString(),
        action_required: false,
        suggested_actions: ['Visit early morning', 'Try alternative attractions', 'Book skip-the-line tickets']
      })
    }
    
    return recommendations
  }

  // Mock Data Generators
  private static generateMockWeatherData(deviceId: string): WeatherData {
    return {
      device_id: deviceId,
      timestamp: new Date().toISOString(),
      quality_score: 0.95,
      confidence: 0.9,
      data: {
        temperature: 20 + Math.random() * 15,
        humidity: 40 + Math.random() * 40,
        pressure: 1000 + Math.random() * 50,
        wind_speed: Math.random() * 20,
        wind_direction: Math.random() * 360,
        precipitation: Math.random() * 10,
        uv_index: Math.floor(Math.random() * 11),
        visibility: 5 + Math.random() * 15,
        weather_condition: ['sunny', 'cloudy', 'partly_cloudy', 'rainy'][Math.floor(Math.random() * 4)]
      }
    }
  }

  private static generateMockAirQualityData(deviceId: string): AirQualityData {
    const aqi = Math.floor(Math.random() * 200)
    return {
      device_id: deviceId,
      timestamp: new Date().toISOString(),
      quality_score: 0.9,
      confidence: 0.85,
      data: {
        aqi,
        pm25: Math.random() * 50,
        pm10: Math.random() * 100,
        co2: 400 + Math.random() * 200,
        ozone: Math.random() * 0.1,
        no2: Math.random() * 0.05,
        so2: Math.random() * 0.02,
        health_recommendation: aqi < 50 ? 'Good' : aqi < 100 ? 'Moderate' : aqi < 150 ? 'Unhealthy for Sensitive Groups' : 'Unhealthy'
      }
    }
  }

  private static generateMockCrowdData(deviceId: string): CrowdData {
    const density = Math.random()
    const estimatedPeople = Math.floor(density * 1000)
    
    return {
      device_id: deviceId,
      timestamp: new Date().toISOString(),
      quality_score: 0.88,
      confidence: 0.8,
      data: {
        crowd_density: density,
        estimated_people: estimatedPeople,
        noise_level: 40 + Math.random() * 40,
        wait_time_minutes: Math.floor(density * 60),
        congestion_level: density < 0.25 ? 'low' : density < 0.5 ? 'medium' : density < 0.75 ? 'high' : 'very_high'
      }
    }
  }

  // Utility Methods
  private static calculateDistance(lat1: number, lon1: number, lat2: number, lon2: number): number {
    const R = 6371 // Earth's radius in km
    const dLat = (lat2 - lat1) * Math.PI / 180
    const dLon = (lon2 - lon1) * Math.PI / 180
    const a = Math.sin(dLat/2) * Math.sin(dLat/2) +
              Math.cos(lat1 * Math.PI / 180) * Math.cos(lat2 * Math.PI / 180) *
              Math.sin(dLon/2) * Math.sin(dLon/2)
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1-a))
    return R * c
  }

  private static startDataCollection(): void {
    // Simulate periodic data collection
    setInterval(() => {
      this.devices.forEach((device, deviceId) => {
        if (device.status === 'online') {
          let mockData: SensorData
          
          switch (device.type) {
            case 'weather_station':
              mockData = this.generateMockWeatherData(deviceId)
              break
            case 'air_quality':
              mockData = this.generateMockAirQualityData(deviceId)
              break
            case 'crowd_sensor':
              mockData = this.generateMockCrowdData(deviceId)
              break
            default:
              return
          }
          
          // Store data
          if (!this.sensorData.has(deviceId)) {
            this.sensorData.set(deviceId, [])
          }
          
          const deviceData = this.sensorData.get(deviceId)!
          deviceData.push(mockData)
          
          // Keep only last 100 readings
          if (deviceData.length > 100) {
            deviceData.splice(0, deviceData.length - 100)
          }
          
          // Emit data update event
          this.emitEvent('sensor_data', { deviceId, data: mockData })
        }
      })
    }, 60000) // Update every minute
  }
}

export { IoTIntegrationService }
