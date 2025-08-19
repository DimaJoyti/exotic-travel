'use client'

import Link from 'next/link'
import { Search, MapPin, Users } from 'lucide-react'
import { useState } from 'react'
import Button from '@/components/ui/button'

export default function HeroSection() {
  const [searchQuery, setSearchQuery] = useState('')
  const [destination, setDestination] = useState('')
  const [guests, setGuests] = useState('2')

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    // Redirect to destinations page with search parameters
    const params = new URLSearchParams()
    if (searchQuery) params.set('search', searchQuery)
    if (destination) params.set('country', destination)
    if (guests) params.set('max_guests', guests)
    
    window.location.href = `/destinations?${params.toString()}`
  }

  return (
    <section className="relative bg-gradient-to-br from-blue-600 via-purple-600 to-teal-600 text-white">
      {/* Background Image Overlay */}
      <div className="absolute inset-0 bg-black/30"></div>
      
      {/* Background Pattern */}
      <div className="absolute inset-0 opacity-10">
        <div className="absolute inset-0" style={{
          backgroundImage: `url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23ffffff' fill-opacity='0.1'%3E%3Ccircle cx='30' cy='30' r='2'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E")`
        }}></div>
      </div>

      <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24 lg:py-32">
        <div className="text-center">
          {/* Main Heading */}
          <h1 className="text-4xl md:text-6xl lg:text-7xl font-bold mb-6">
            Discover
            <span className="block text-transparent bg-clip-text bg-gradient-to-r from-yellow-400 to-orange-500">
              Exotic Adventures
            </span>
          </h1>
          
          {/* Subtitle */}
          <p className="text-xl md:text-2xl text-gray-200 mb-8 max-w-3xl mx-auto">
            Embark on extraordinary journeys to the world's most breathtaking destinations. 
            From pristine beaches to ancient wonders, your next adventure awaits.
          </p>

          {/* Search Form */}
          <div className="max-w-4xl mx-auto mb-12">
            <form onSubmit={handleSearch} className="bg-white/10 backdrop-blur-md rounded-2xl p-6 shadow-2xl">
              <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                {/* Search Input */}
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
                  <input
                    type="text"
                    placeholder="Search destinations..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="w-full pl-10 pr-4 py-3 bg-white/20 backdrop-blur-sm border border-white/30 rounded-lg text-white placeholder-gray-300 focus:outline-none focus:ring-2 focus:ring-white/50 focus:border-transparent transition-all duration-200"
                  />
                </div>

                {/* Destination Select */}
                <div className="relative">
                  <MapPin className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
                  <select
                    value={destination}
                    onChange={(e) => setDestination(e.target.value)}
                    className="w-full pl-10 pr-4 py-3 bg-white/20 backdrop-blur-sm border border-white/30 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/50 focus:border-transparent appearance-none transition-all duration-200"
                  >
                    <option value="" className="text-gray-900">Any Country</option>
                    <option value="Maldives" className="text-gray-900">Maldives</option>
                    <option value="Brazil" className="text-gray-900">Brazil</option>
                    <option value="Morocco" className="text-gray-900">Morocco</option>
                    <option value="Antarctica" className="text-gray-900">Antarctica</option>
                    <option value="Thailand" className="text-gray-900">Thailand</option>
                    <option value="Iceland" className="text-gray-900">Iceland</option>
                  </select>
                </div>

                {/* Guests Select */}
                <div className="relative">
                  <Users className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
                  <select
                    value={guests}
                    onChange={(e) => setGuests(e.target.value)}
                    className="w-full pl-10 pr-4 py-3 bg-white/20 backdrop-blur-sm border border-white/30 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-white/50 focus:border-transparent appearance-none transition-all duration-200"
                  >
                    <option value="1" className="text-gray-900">1 Guest</option>
                    <option value="2" className="text-gray-900">2 Guests</option>
                    <option value="3" className="text-gray-900">3 Guests</option>
                    <option value="4" className="text-gray-900">4 Guests</option>
                    <option value="5" className="text-gray-900">5+ Guests</option>
                  </select>
                </div>

                {/* Search Button */}
                <Button
                  type="submit"
                  className="bg-gradient-to-r from-orange-500 to-red-500 hover:from-orange-600 hover:to-red-600 text-white font-semibold py-3 px-6 rounded-lg transition-all duration-200 transform hover:scale-105 shadow-lg"
                >
                  <Search className="h-5 w-5 mr-2" />
                  Search
                </Button>
              </div>
            </form>
          </div>

          {/* CTA Buttons */}
          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
            <Link href="/destinations">
              <Button
                size="lg"
                className="bg-white text-gray-900 hover:bg-gray-100 shadow-lg hover:shadow-xl transform hover:scale-105 transition-all duration-200 rounded-full px-8 py-4"
              >
                Explore Destinations
              </Button>
            </Link>
            <Link href="/about">
              <Button
                variant="outline"
                size="lg"
                className="border-2 border-white text-white hover:bg-white hover:text-gray-900 bg-transparent rounded-full px-8 py-4 transform hover:scale-105 transition-all duration-200"
              >
                Learn More
              </Button>
            </Link>
          </div>
        </div>
      </div>

      {/* Scroll Indicator */}
      <div className="absolute bottom-8 left-1/2 transform -translate-x-1/2 animate-bounce">
        <div className="w-6 h-10 border-2 border-white rounded-full flex justify-center">
          <div className="w-1 h-3 bg-white rounded-full mt-2 animate-pulse"></div>
        </div>
      </div>
    </section>
  )
}
