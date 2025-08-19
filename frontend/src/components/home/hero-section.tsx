import Link from 'next/link'
import { Search, ArrowRight } from 'lucide-react'
import { Button } from '@/components/ui/button'

export default function HeroSection() {
  return (
    <section className="relative bg-gradient-to-br from-blue-600 via-purple-600 to-teal-600 text-white">
      {/* Background overlay */}
      <div className="absolute inset-0 bg-black/30"></div>

      {/* Content */}
      <div className="relative z-20 py-32 md:py-40 lg:py-48">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="w-full text-center">
            <div className="mb-12">
              {/* Main Heading */}
              <h1 className="text-5xl md:text-7xl lg:text-8xl font-bold mb-6 leading-tight">
                <span className="block">Discover Your Next</span>
                <span className="block text-transparent bg-clip-text bg-gradient-to-r from-yellow-400 via-red-400 to-pink-500">
                  Epic Adventure
                </span>
              </h1>

              {/* Subtitle */}
              <p className="text-xl md:text-2xl text-gray-200 mb-8 max-w-4xl mx-auto leading-relaxed">
                Embark on extraordinary journeys to the world's most breathtaking destinations.
                From pristine beaches to ancient wonders, your next adventure awaits.
              </p>
            </div>

            {/* Simplified Search Section */}
            <div className="max-w-5xl mx-auto mb-12">
              <div className="bg-white/10 backdrop-blur-xl rounded-3xl p-8 shadow-2xl border border-white/20">
                <div className="flex flex-col sm:flex-row gap-4 items-center justify-center">
                  <div className="flex items-center space-x-3 text-white">
                    <Search className="h-6 w-6" />
                    <span className="text-lg font-medium">Search Amazing Destinations</span>
                  </div>
                  <Link href="/destinations">
                    <Button
                      variant="secondary"
                      size="lg"
                      className="bg-white text-gray-900 hover:bg-gray-100 shadow-xl hover:shadow-2xl px-8 py-3 font-semibold"
                      rightIcon={<ArrowRight className="h-4 w-4" />}
                    >
                      Start Exploring
                    </Button>
                  </Link>
                </div>
              </div>
            </div>

            {/* CTA Buttons */}
            <div className="flex flex-col sm:flex-row gap-6 justify-center items-center">
              <Link href="/destinations">
                <Button
                  size="xl"
                  variant="secondary"
                  className="bg-white text-gray-900 hover:bg-gray-100 shadow-xl hover:shadow-2xl rounded-full px-10 py-4 font-semibold text-lg"
                  rightIcon={<ArrowRight className="h-5 w-5" />}
                >
                  Explore Destinations
                </Button>
              </Link>
              <Link href="/about">
                <Button
                  variant="outline"
                  size="xl"
                  className="border-2 border-white text-white hover:bg-white hover:text-gray-900 bg-transparent rounded-full px-10 py-4 font-semibold text-lg"
                >
                  Learn More
                </Button>
              </Link>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
