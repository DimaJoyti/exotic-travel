'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import Image from 'next/image'
import { Calendar, MapPin, Users, Clock, Filter, Search, Download, Eye, X } from 'lucide-react'
import { useAuth } from '@/contexts/auth-context'
import { BookingsService } from '@/lib/bookings'
import { DestinationsService } from '@/lib/destinations'
import { Booking, Destination, BookingStatus } from '@/types'
import { formatCurrency } from '@/lib/utils'

export default function BookingsPage() {
  const { user } = useAuth()
  const router = useRouter()
  const [bookings, setBookings] = useState<Booking[]>([])
  const [destinations, setDestinations] = useState<{ [key: number]: Destination }>({})
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState<string>('all')
  const [searchQuery, setSearchQuery] = useState('')

  useEffect(() => {
    if (!user) {
      router.push('/auth/login')
      return
    }

    const loadBookings = async () => {
      try {
        // For demo purposes, create mock bookings
        const mockBookings: Booking[] = [
          {
            id: 1,
            user_id: user.id,
            destination_id: 1,
            check_in_date: '2024-03-15',
            check_out_date: '2024-03-22',
            guests: 2,
            total_price: 5000,
            status: BookingStatus.CONFIRMED,
            special_requests: 'Honeymoon suite requested',
            created_at: '2024-01-15T10:00:00Z',
            updated_at: '2024-01-15T10:00:00Z'
          },
          {
            id: 2,
            user_id: user.id,
            destination_id: 2,
            check_in_date: '2024-04-10',
            check_out_date: '2024-04-20',
            guests: 4,
            total_price: 7200,
            status: BookingStatus.PENDING,
            special_requests: '',
            created_at: '2024-02-01T10:00:00Z',
            updated_at: '2024-02-01T10:00:00Z'
          },
          {
            id: 3,
            user_id: user.id,
            destination_id: 3,
            check_in_date: '2023-12-01',
            check_out_date: '2023-12-06',
            guests: 2,
            total_price: 2400,
            status: BookingStatus.COMPLETED,
            special_requests: 'Vegetarian meals',
            created_at: '2023-10-15T10:00:00Z',
            updated_at: '2023-12-06T10:00:00Z'
          }
        ]

        setBookings(mockBookings)

        // Load destination data for each booking
        const destinationData: { [key: number]: Destination } = {}
        const mockDestinations = DestinationsService.getMockDestinations()
        
        for (const booking of mockBookings) {
          const destination = mockDestinations.find(d => d.id === booking.destination_id)
          if (destination) {
            destinationData[booking.destination_id] = destination
          }
        }
        
        setDestinations(destinationData)
      } catch (error) {
        console.error('Error loading bookings:', error)
      } finally {
        setLoading(false)
      }
    }

    loadBookings()
  }, [user, router])

  const filteredBookings = bookings.filter(booking => {
    const matchesFilter = filter === 'all' || booking.status.toLowerCase() === filter.toLowerCase()
    const destination = destinations[booking.destination_id]
    const matchesSearch = !searchQuery || 
      destination?.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      destination?.country.toLowerCase().includes(searchQuery.toLowerCase())
    
    return matchesFilter && matchesSearch
  })

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'confirmed':
        return 'bg-green-100 text-green-800'
      case 'pending':
        return 'bg-yellow-100 text-yellow-800'
      case 'cancelled':
        return 'bg-red-100 text-red-800'
      case 'completed':
        return 'bg-blue-100 text-blue-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const calculateNights = (checkIn: string, checkOut: string) => {
    const start = new Date(checkIn)
    const end = new Date(checkOut)
    const diffTime = end.getTime() - start.getTime()
    return Math.ceil(diffTime / (1000 * 60 * 60 * 24))
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">My Bookings</h1>
          <p className="text-gray-600">Manage your travel bookings and view trip history</p>
        </div>

        {/* Filters and Search */}
        <div className="bg-white rounded-lg shadow-sm p-6 mb-8">
          <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
            <div className="flex items-center space-x-4">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
                <input
                  type="text"
                  placeholder="Search bookings..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                />
              </div>
              
              <div className="relative">
                <Filter className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
                <select
                  value={filter}
                  onChange={(e) => setFilter(e.target.value)}
                  className="pl-10 pr-8 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent appearance-none bg-white"
                >
                  <option value="all">All Bookings</option>
                  <option value="confirmed">Confirmed</option>
                  <option value="pending">Pending</option>
                  <option value="completed">Completed</option>
                  <option value="cancelled">Cancelled</option>
                </select>
              </div>
            </div>

            <div className="text-sm text-gray-500">
              {filteredBookings.length} booking{filteredBookings.length !== 1 ? 's' : ''} found
            </div>
          </div>
        </div>

        {/* Bookings List */}
        {filteredBookings.length === 0 ? (
          <div className="bg-white rounded-lg shadow-sm p-12 text-center">
            <Calendar className="h-16 w-16 text-gray-300 mx-auto mb-4" />
            <h3 className="text-xl font-semibold text-gray-900 mb-2">No bookings found</h3>
            <p className="text-gray-600 mb-6">
              {searchQuery || filter !== 'all' 
                ? 'Try adjusting your search or filter criteria'
                : "You haven't made any bookings yet. Start planning your next adventure!"
              }
            </p>
            <Link
              href="/destinations"
              className="bg-primary text-primary-foreground px-6 py-3 rounded-lg hover:bg-primary/90 transition-colors inline-block"
            >
              Browse Destinations
            </Link>
          </div>
        ) : (
          <div className="space-y-6">
            {filteredBookings.map((booking) => {
              const destination = destinations[booking.destination_id]
              if (!destination) return null

              return (
                <div key={booking.id} className="bg-white rounded-lg shadow-sm overflow-hidden">
                  <div className="p-6">
                    <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between">
                      <div className="flex items-start space-x-4 mb-4 lg:mb-0">
                        <div className="relative w-20 h-20 rounded-lg overflow-hidden flex-shrink-0">
                          <Image
                            src={destination.images[0] || '/placeholder-destination.jpg'}
                            alt={destination.name}
                            fill
                            className="object-cover"
                          />
                        </div>
                        
                        <div className="flex-1">
                          <div className="flex items-center space-x-2 mb-2">
                            <h3 className="text-lg font-semibold text-gray-900">
                              {destination.name}
                            </h3>
                            <span className={`px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(booking.status)}`}>
                              {booking.status.charAt(0).toUpperCase() + booking.status.slice(1)}
                            </span>
                          </div>
                          
                          <div className="flex items-center text-sm text-gray-500 mb-1">
                            <MapPin className="h-4 w-4 mr-1" />
                            {destination.city}, {destination.country}
                          </div>
                          
                          <div className="flex items-center space-x-4 text-sm text-gray-500">
                            <div className="flex items-center">
                              <Calendar className="h-4 w-4 mr-1" />
                              {new Date(booking.check_in_date).toLocaleDateString()} - {new Date(booking.check_out_date).toLocaleDateString()}
                            </div>
                            <div className="flex items-center">
                              <Clock className="h-4 w-4 mr-1" />
                              {calculateNights(booking.check_in_date, booking.check_out_date)} nights
                            </div>
                            <div className="flex items-center">
                              <Users className="h-4 w-4 mr-1" />
                              {booking.guests} guest{booking.guests !== 1 ? 's' : ''}
                            </div>
                          </div>
                        </div>
                      </div>

                      <div className="flex flex-col lg:items-end space-y-3">
                        <div className="text-right">
                          <p className="text-2xl font-bold text-gray-900">
                            {formatCurrency(booking.total_price)}
                          </p>
                          <p className="text-sm text-gray-500">Total paid</p>
                        </div>
                        
                        <div className="flex space-x-2">
                          <Link
                            href={`/bookings/${booking.id}`}
                            className="flex items-center px-3 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
                          >
                            <Eye className="h-4 w-4 mr-1" />
                            View
                          </Link>
                          <button className="flex items-center px-3 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors">
                            <Download className="h-4 w-4 mr-1" />
                            Download
                          </button>
                          {booking.status === BookingStatus.CONFIRMED && (
                            <button className="flex items-center px-3 py-2 border border-red-300 rounded-lg text-red-700 hover:bg-red-50 transition-colors">
                              <X className="h-4 w-4 mr-1" />
                              Cancel
                            </button>
                          )}
                        </div>
                      </div>
                    </div>

                    {booking.special_requests && (
                      <div className="mt-4 pt-4 border-t border-gray-100">
                        <p className="text-sm text-gray-600">
                          <span className="font-medium">Special requests:</span> {booking.special_requests}
                        </p>
                      </div>
                    )}
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </div>
    </div>
  )
}
