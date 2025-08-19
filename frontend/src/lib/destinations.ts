import api from './api'
import { Destination } from '@/types'

export interface DestinationFilter {
  country?: string
  city?: string
  min_price?: number
  max_price?: number
  duration?: number
  max_guests?: number
  search?: string
  limit?: number
  offset?: number
}

export interface DestinationsResponse {
  destinations: Destination[]
  total: number
  page: number
  limit: number
}

export class DestinationsService {
  static async getDestinations(filter: DestinationFilter = {}): Promise<Destination[]> {
    const params = new URLSearchParams()
    
    Object.entries(filter).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        params.append(key, value.toString())
      }
    })

    const response = await api.get<Destination[]>(`/api/destinations?${params.toString()}`)
    return response.data
  }

  static async getDestination(id: number): Promise<Destination> {
    const response = await api.get<Destination>(`/api/destinations/${id}`)
    return response.data
  }

  static async searchDestinations(query: string, limit = 20, offset = 0): Promise<Destination[]> {
    const params = new URLSearchParams({
      q: query,
      limit: limit.toString(),
      offset: offset.toString(),
    })

    const response = await api.get<Destination[]>(`/api/destinations/search?${params.toString()}`)
    return response.data
  }

  static async getPopularDestinations(limit = 6): Promise<Destination[]> {
    const response = await api.get<Destination[]>(`/api/destinations?limit=${limit}`)
    return response.data
  }

  static async getDestinationsByCountry(country: string, limit = 10): Promise<Destination[]> {
    const response = await api.get<Destination[]>(`/api/destinations?country=${encodeURIComponent(country)}&limit=${limit}`)
    return response.data
  }

  // Mock data for development when backend is not available
  static getMockDestinations(): Destination[] {
    return [
      {
        id: 1,
        name: "Maldives Paradise Resort",
        description: "Experience luxury in overwater bungalows surrounded by crystal-clear turquoise waters. This exclusive resort offers world-class diving, spa treatments, and gourmet dining with stunning ocean views.",
        country: "Maldives",
        city: "Malé",
        price: 2500,
        duration: 7,
        max_guests: 4,
        images: [
          "https://images.unsplash.com/photo-1506905925346-21bda4d32df4?w=800&h=600&fit=crop",
          "https://images.unsplash.com/photo-1540979388789-6cee28a1cdc9?w=800&h=600&fit=crop"
        ],
        features: ["Overwater Bungalows", "Private Beach", "Spa & Wellness", "Scuba Diving", "Fine Dining", "Airport Transfer"],
        created_at: "2024-01-15T10:00:00Z",
        updated_at: "2024-01-15T10:00:00Z"
      },
      {
        id: 2,
        name: "Amazon Rainforest Adventure",
        description: "Embark on an unforgettable journey deep into the Amazon rainforest. Stay in eco-lodges, spot exotic wildlife, and learn about indigenous cultures while contributing to conservation efforts.",
        country: "Brazil",
        city: "Manaus",
        price: 1800,
        duration: 10,
        max_guests: 8,
        images: [
          "https://images.unsplash.com/photo-1544735716-392fe2489ffa?w=800&h=600&fit=crop",
          "https://images.unsplash.com/photo-1516026672322-bc52d61a55d5?w=800&h=600&fit=crop"
        ],
        features: ["Eco Lodge", "Wildlife Spotting", "Canoe Expeditions", "Indigenous Culture", "Conservation Program", "Expert Guides"],
        created_at: "2024-01-10T10:00:00Z",
        updated_at: "2024-01-10T10:00:00Z"
      },
      {
        id: 3,
        name: "Sahara Desert Glamping",
        description: "Sleep under the stars in luxury desert camps while exploring the vast Sahara. Enjoy camel trekking, traditional Berber cuisine, and breathtaking sunrises over endless sand dunes.",
        country: "Morocco",
        city: "Merzouga",
        price: 1200,
        duration: 5,
        max_guests: 6,
        images: [
          "https://images.unsplash.com/photo-1509316975850-ff9c5deb0cd9?w=800&h=600&fit=crop",
          "https://images.unsplash.com/photo-1518548419970-58e3b4079ab2?w=800&h=600&fit=crop"
        ],
        features: ["Luxury Tents", "Camel Trekking", "Stargazing", "Traditional Cuisine", "Berber Culture", "Sandboarding"],
        created_at: "2024-01-05T10:00:00Z",
        updated_at: "2024-01-05T10:00:00Z"
      },
      {
        id: 4,
        name: "Antarctic Expedition Cruise",
        description: "Journey to the last frontier on Earth aboard a luxury expedition vessel. Witness massive icebergs, encounter penguins and whales, and explore the pristine Antarctic wilderness.",
        country: "Antarctica",
        city: "Antarctic Peninsula",
        price: 8500,
        duration: 14,
        max_guests: 12,
        images: [
          "https://images.unsplash.com/photo-1518837695005-2083093ee35b?w=800&h=600&fit=crop",
          "https://images.unsplash.com/photo-1544966503-7cc5ac882d5f?w=800&h=600&fit=crop"
        ],
        features: ["Expedition Cruise", "Wildlife Viewing", "Zodiac Landings", "Expert Naturalists", "Photography Workshops", "All Meals Included"],
        created_at: "2024-01-01T10:00:00Z",
        updated_at: "2024-01-01T10:00:00Z"
      },
      {
        id: 5,
        name: "Bali Temple & Rice Terraces",
        description: "Discover the spiritual heart of Bali through ancient temples, emerald rice terraces, and traditional villages. Experience authentic Balinese culture, yoga retreats, and volcanic landscapes.",
        country: "Indonesia",
        city: "Ubud",
        price: 950,
        duration: 8,
        max_guests: 10,
        images: [
          "https://images.unsplash.com/photo-1537953773345-d172ccf13cf1?w=800&h=600&fit=crop",
          "https://images.unsplash.com/photo-1518548419970-58e3b4079ab2?w=800&h=600&fit=crop"
        ],
        features: ["Temple Tours", "Rice Terraces", "Yoga Retreats", "Traditional Villages", "Volcano Hiking", "Cultural Workshops"],
        created_at: "2024-01-20T10:00:00Z",
        updated_at: "2024-01-20T10:00:00Z"
      },
      {
        id: 6,
        name: "Iceland Northern Lights",
        description: "Chase the magical Aurora Borealis across Iceland's dramatic landscapes. Explore ice caves, geysers, and waterfalls while staying in cozy lodges with panoramic views of the night sky.",
        country: "Iceland",
        city: "Reykjavik",
        price: 2200,
        duration: 6,
        max_guests: 8,
        images: [
          "https://images.unsplash.com/photo-1506905925346-21bda4d32df4?w=800&h=600&fit=crop",
          "https://images.unsplash.com/photo-1540979388789-6cee28a1cdc9?w=800&h=600&fit=crop"
        ],
        features: ["Northern Lights", "Ice Caves", "Geysers", "Waterfalls", "Hot Springs", "Photography Tours"],
        created_at: "2024-01-25T10:00:00Z",
        updated_at: "2024-01-25T10:00:00Z"
      }
    ]
  }
}
