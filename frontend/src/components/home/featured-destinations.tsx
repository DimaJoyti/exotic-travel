'use client'

import Link from 'next/link'
import Image from 'next/image'
import { MapPin, Clock, Users, Star, Heart, ArrowRight } from 'lucide-react'
import { motion } from 'framer-motion'
import { formatCurrency } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import WishlistButton from '@/components/wishlist/wishlist-button'
import { Destination } from '@/types'

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
      features: ["Overwater Bungalows", "Private Beach", "Spa & Wellness"],
      created_at: "2024-01-01T00:00:00Z",
      updated_at: "2024-01-01T00:00:00Z"
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
      features: ["Eco Lodge", "Wildlife Spotting", "Canoe Expeditions"],
      created_at: "2024-01-01T00:00:00Z",
      updated_at: "2024-01-01T00:00:00Z"
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
      features: ["Luxury Tents", "Camel Trekking", "Stargazing"],
      created_at: "2024-01-01T00:00:00Z",
      updated_at: "2024-01-01T00:00:00Z"
    }
  ]

  const displayDestinations = destinations.length > 0 ? destinations.slice(0, 3) : mockDestinations

  return (
    <section className="py-20 bg-gradient-to-b from-gray-50 to-white">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Section Header */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6 }}
          className="text-center mb-16"
        >
          <h2 className="text-4xl md:text-5xl font-display font-bold text-gray-900 mb-6">
            Featured Destinations
          </h2>
          <p className="text-xl text-gray-600 max-w-4xl mx-auto leading-relaxed">
            Discover our most popular exotic destinations, carefully curated for unforgettable experiences
          </p>
        </motion.div>

        {/* Destinations Grid */}
        <motion.div
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ staggerChildren: 0.2, delayChildren: 0.1 }}
          className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8 mb-16"
        >
          {displayDestinations.map((destination, index) => (
            <motion.div
              key={destination.id}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: index * 0.2 }}
              whileHover={{ y: -8, scale: 1.02 }}
            >
              <div className="bg-white rounded-3xl shadow-xl overflow-hidden group border border-gray-100">
                {/* Image */}
                <div className="relative h-72 overflow-hidden">
                  <Image
                    src={destination.images[0] || '/placeholder-destination.jpg'}
                    alt={destination.name}
                    fill
                    className="object-cover group-hover:scale-110 transition-transform duration-700"
                  />

                  {/* Gradient Overlay */}
                  <div className="absolute inset-0 bg-gradient-to-t from-black/60 via-transparent to-transparent" />

                  {/* Rating Badge */}
                  <div className="absolute top-4 right-4 bg-white/95 backdrop-blur-sm rounded-full px-3 py-1.5 shadow-lg">
                    <div className="flex items-center space-x-1">
                      <Star className="h-4 w-4 text-yellow-400 fill-current" />
                      <span className="text-sm font-semibold text-gray-900">4.9</span>
                    </div>
                  </div>

                  {/* Wishlist Button */}
                  <div className="absolute top-4 left-4">
                    <WishlistButton
                      destination={destination}
                      size="md"
                      variant="icon"
                    />
                  </div>

                  {/* Price Badge */}
                  <div className="absolute bottom-4 left-4 bg-blue-500 text-white px-4 py-2 rounded-full font-bold shadow-lg">
                    {formatCurrency(destination.price)}
                  </div>
                </div>
                
                {/* Content */}
                <div className="p-8">
                  <div className="flex items-center text-sm text-gray-500 mb-3">
                    <MapPin className="h-4 w-4 mr-2" />
                    {destination.city}, {destination.country}
                  </div>

                  <h3 className="text-2xl font-bold text-gray-900 mb-3 group-hover:text-blue-600 transition-colors">
                    {destination.name}
                  </h3>

                  <p className="text-gray-600 mb-6 line-clamp-2 leading-relaxed">
                    {destination.description}
                  </p>

                  {/* Features */}
                  <div className="flex flex-wrap gap-2 mb-6">
                    {destination.features.slice(0, 2).map((feature, featureIndex) => (
                      <span
                        key={featureIndex}
                        className="bg-blue-50 text-blue-700 text-sm px-3 py-1.5 rounded-full font-medium"
                      >
                        {feature}
                      </span>
                    ))}
                    {destination.features.length > 2 && (
                      <span className="text-sm text-gray-500 px-3 py-1.5">
                        +{destination.features.length - 2} more
                      </span>
                    )}
                  </div>

                  {/* Details */}
                  <div className="flex items-center justify-between text-sm text-gray-500 mb-6">
                    <div className="flex items-center">
                      <Clock className="h-4 w-4 mr-1" />
                      {destination.duration} days
                    </div>
                    <div className="flex items-center">
                      <Users className="h-4 w-4 mr-1" />
                      Up to {destination.max_guests} guests
                    </div>
                  </div>

                  {/* CTA */}
                  <Link href={`/destinations/${destination.id}`}>
                    <Button
                      variant="primary"
                      size="lg"
                      className="w-full font-semibold shadow-lg hover:shadow-xl"
                      rightIcon={<ArrowRight className="h-4 w-4" />}
                    >
                      View Details
                    </Button>
                  </Link>
                </div>
              </div>
            </motion.div>
          ))}
        </motion.div>


        {/* View All Button */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.6, delay: 0.4 }}
          className="text-center"
        >
          <Link href="/destinations">
            <Button
              variant="outline"
              size="xl"
              className="border-2 border-blue-500 text-blue-600 hover:bg-blue-500 hover:text-white px-12 py-4 font-semibold shadow-lg hover:shadow-xl rounded-full"
              rightIcon={<ArrowRight className="h-5 w-5" />}
            >
              View All Destinations
            </Button>
          </Link>
        </motion.div>
      </div>
    </section>
  )
}
