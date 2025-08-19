'use client'

import Link from 'next/link'
import Image from 'next/image'
import { MapPin, Clock, Users, Star } from 'lucide-react'
import { formatCurrency } from '@/lib/utils'

interface Destination {
  id: number
  name: string
  description: string
  country: string
  city: string
  price: number
  duration: number
  max_guests: number
  images: string[]
  features: string[]
}

interface FeaturedDestinationsProps {
  destinations?: Destination[]
}

export default function FeaturedDestinations({ destinations = [] }: FeaturedDestinationsProps) {
  // Mock data if no destinations provided
  const mockDestinations: Destination[] = [
    {
      id: 1,
      name: "Maldives Paradise Resort",
      description: "Experience luxury in overwater bungalows surrounded by crystal-clear turquoise waters.",
      country: "Maldives",
      city: "Malé",
      price: 2500,
      duration: 7,
      max_guests: 4,
      images: ["https://images.unsplash.com/photo-1506905925346-21bda4d32df4?w=800&h=600&fit=crop"],
      features: ["Overwater Bungalows", "Private Beach", "Spa & Wellness"]
    },
    {
      id: 2,
      name: "Amazon Rainforest Adventure",
      description: "Embark on an unforgettable journey deep into the Amazon rainforest.",
      country: "Brazil",
      city: "Manaus",
      price: 1800,
      duration: 10,
      max_guests: 8,
      images: ["https://images.unsplash.com/photo-1544735716-392fe2489ffa?w=800&h=600&fit=crop"],
      features: ["Eco Lodge", "Wildlife Spotting", "Canoe Expeditions"]
    },
    {
      id: 3,
      name: "Sahara Desert Glamping",
      description: "Sleep under the stars in luxury desert camps while exploring the vast Sahara.",
      country: "Morocco",
      city: "Merzouga",
      price: 1200,
      duration: 5,
      max_guests: 6,
      images: ["https://images.unsplash.com/photo-1509316975850-ff9c5deb0cd9?w=800&h=600&fit=crop"],
      features: ["Luxury Tents", "Camel Trekking", "Stargazing"]
    }
  ]

  const displayDestinations = destinations.length > 0 ? destinations.slice(0, 3) : mockDestinations

  return (
    <section className="py-16 bg-gray-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Section Header */}
        <div className="text-center mb-12">
          <h2 className="text-3xl md:text-4xl font-bold text-gray-900 mb-4">
            Featured Destinations
          </h2>
          <p className="text-xl text-gray-600 max-w-3xl mx-auto">
            Discover our most popular exotic destinations, carefully curated for unforgettable experiences
          </p>
        </div>

        {/* Destinations Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8 mb-12">
          {displayDestinations.map((destination) => (
            <div
              key={destination.id}
              className="bg-white rounded-2xl shadow-lg overflow-hidden hover:shadow-xl transition-shadow duration-300 group"
            >
              {/* Image */}
              <div className="relative h-64 overflow-hidden">
                <Image
                  src={destination.images[0] || '/placeholder-destination.jpg'}
                  alt={destination.name}
                  fill
                  className="object-cover group-hover:scale-110 transition-transform duration-300"
                />
                <div className="absolute top-4 right-4 bg-white/90 backdrop-blur-sm rounded-full px-3 py-1">
                  <div className="flex items-center space-x-1">
                    <Star className="h-4 w-4 text-yellow-400 fill-current" />
                    <span className="text-sm font-medium">4.8</span>
                  </div>
                </div>
              </div>

              {/* Content */}
              <div className="p-6">
                <div className="flex items-center text-sm text-gray-500 mb-2">
                  <MapPin className="h-4 w-4 mr-1" />
                  {destination.city}, {destination.country}
                </div>
                
                <h3 className="text-xl font-bold text-gray-900 mb-2 group-hover:text-primary transition-colors">
                  {destination.name}
                </h3>
                
                <p className="text-gray-600 mb-4 line-clamp-2">
                  {destination.description}
                </p>

                {/* Features */}
                <div className="flex flex-wrap gap-2 mb-4">
                  {destination.features.slice(0, 2).map((feature, index) => (
                    <span
                      key={index}
                      className="bg-primary/10 text-primary text-xs px-2 py-1 rounded-full"
                    >
                      {feature}
                    </span>
                  ))}
                  {destination.features.length > 2 && (
                    <span className="text-xs text-gray-500">
                      +{destination.features.length - 2} more
                    </span>
                  )}
                </div>

                {/* Details */}
                <div className="flex items-center justify-between text-sm text-gray-500 mb-4">
                  <div className="flex items-center">
                    <Clock className="h-4 w-4 mr-1" />
                    {destination.duration} days
                  </div>
                  <div className="flex items-center">
                    <Users className="h-4 w-4 mr-1" />
                    Up to {destination.max_guests} guests
                  </div>
                </div>

                {/* Price and CTA */}
                <div className="flex items-center justify-between">
                  <div>
                    <span className="text-2xl font-bold text-gray-900">
                      {formatCurrency(destination.price)}
                    </span>
                    <span className="text-gray-500 text-sm ml-1">per person</span>
                  </div>
                  <Link
                    href={`/destinations/${destination.id}`}
                    className="bg-primary text-primary-foreground hover:bg-primary/90 px-4 py-2 rounded-lg font-medium transition-colors"
                  >
                    View Details
                  </Link>
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* View All Button */}
        <div className="text-center">
          <Link
            href="/destinations"
            className="inline-flex items-center bg-primary text-primary-foreground hover:bg-primary/90 px-8 py-3 rounded-full font-semibold transition-colors"
          >
            View All Destinations
            <svg
              className="ml-2 h-5 w-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M17 8l4 4m0 0l-4 4m4-4H3"
              />
            </svg>
          </Link>
        </div>
      </div>
    </section>
  )
}
