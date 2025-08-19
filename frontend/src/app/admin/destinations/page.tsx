'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import Link from 'next/link'
import { 
  Plus, 
  Search, 
  Filter, 
  Eye, 
  Edit, 
  Trash2, 
  MapPin, 
  DollarSign,
  Users,
  Star,
  MoreHorizontal
} from 'lucide-react'
import { useAuth } from '@/contexts/auth-context'
import { DestinationsService } from '@/lib/destinations'
import { Destination } from '@/types'
import { formatCurrency } from '@/lib/utils'
import AdminLayout from '@/components/admin/admin-layout'

export default function AdminDestinationsPage() {
  const { user } = useAuth()
  const router = useRouter()
  const [destinations, setDestinations] = useState<Destination[]>([])
  const [loading, setLoading] = useState(true)
  const [searchQuery, setSearchQuery] = useState('')
  const [filterStatus, setFilterStatus] = useState('all')
  const [selectedDestinations, setSelectedDestinations] = useState<number[]>([])

  useEffect(() => {
    if (!user) {
      router.push('/auth/login')
      return
    }

    if (user.role !== 'admin') {
      router.push('/dashboard')
      return
    }

    loadDestinations()
  }, [user, router])

  const loadDestinations = async () => {
    try {
      const data = DestinationsService.getMockDestinations()
      setDestinations(data)
    } catch (error) {
      console.error('Error loading destinations:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteDestination = async (id: number) => {
    if (!confirm('Are you sure you want to delete this destination?')) {
      return
    }

    try {
      // In real implementation, call API to delete destination
      setDestinations(prev => prev.filter(d => d.id !== id))
    } catch (error) {
      console.error('Error deleting destination:', error)
    }
  }

  const handleBulkDelete = async () => {
    if (selectedDestinations.length === 0) return
    
    if (!confirm(`Are you sure you want to delete ${selectedDestinations.length} destinations?`)) {
      return
    }

    try {
      setDestinations(prev => prev.filter(d => !selectedDestinations.includes(d.id)))
      setSelectedDestinations([])
    } catch (error) {
      console.error('Error deleting destinations:', error)
    }
  }

  const toggleDestinationSelection = (id: number) => {
    setSelectedDestinations(prev => 
      prev.includes(id) 
        ? prev.filter(destId => destId !== id)
        : [...prev, id]
    )
  }

  const toggleSelectAll = () => {
    if (selectedDestinations.length === filteredDestinations.length) {
      setSelectedDestinations([])
    } else {
      setSelectedDestinations(filteredDestinations.map(d => d.id))
    }
  }

  const filteredDestinations = destinations.filter(destination => {
    const matchesSearch = destination.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         destination.country.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         destination.city.toLowerCase().includes(searchQuery.toLowerCase())
    
    // Add status filtering logic here if needed
    return matchesSearch
  })

  if (loading) {
    return (
      <AdminLayout>
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      </AdminLayout>
    )
  }

  return (
    <AdminLayout>
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Destinations</h1>
            <p className="text-gray-600 mt-1">Manage your travel destinations and packages</p>
          </div>
          <Link
            href="/admin/destinations/new"
            className="flex items-center px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors"
          >
            <Plus className="h-4 w-4 mr-2" />
            Add Destination
          </Link>
        </div>
      </div>

      {/* Filters and Search */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 mb-6">
        <div className="flex flex-col md:flex-row gap-4">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
            <input
              type="text"
              placeholder="Search destinations..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
            />
          </div>
          <select
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          >
            <option value="all">All Status</option>
            <option value="active">Active</option>
            <option value="inactive">Inactive</option>
            <option value="draft">Draft</option>
          </select>
          <button className="flex items-center px-3 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors">
            <Filter className="h-4 w-4 mr-2" />
            More Filters
          </button>
        </div>

        {/* Bulk Actions */}
        {selectedDestinations.length > 0 && (
          <div className="mt-4 flex items-center justify-between bg-blue-50 border border-blue-200 rounded-lg p-3">
            <span className="text-sm text-blue-700">
              {selectedDestinations.length} destination{selectedDestinations.length !== 1 ? 's' : ''} selected
            </span>
            <div className="flex space-x-2">
              <button
                onClick={handleBulkDelete}
                className="text-sm text-red-600 hover:text-red-800 font-medium"
              >
                Delete Selected
              </button>
              <button
                onClick={() => setSelectedDestinations([])}
                className="text-sm text-gray-600 hover:text-gray-800"
              >
                Clear Selection
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Destinations Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredDestinations.map((destination) => (
          <div key={destination.id} className="bg-white rounded-lg shadow-sm border border-gray-200 overflow-hidden">
            {/* Image */}
            <div className="relative h-48 bg-gray-200">
              <img
                src={destination.images[0] || '/placeholder-destination.jpg'}
                alt={destination.name}
                className="w-full h-full object-cover"
              />
              <div className="absolute top-3 left-3">
                <input
                  type="checkbox"
                  checked={selectedDestinations.includes(destination.id)}
                  onChange={() => toggleDestinationSelection(destination.id)}
                  className="h-4 w-4 text-primary focus:ring-primary border-gray-300 rounded"
                />
              </div>
              <div className="absolute top-3 right-3">
                <div className="relative">
                  <button className="p-1 bg-white rounded-full shadow-sm hover:bg-gray-50">
                    <MoreHorizontal className="h-4 w-4 text-gray-600" />
                  </button>
                </div>
              </div>
            </div>

            {/* Content */}
            <div className="p-4">
              <div className="flex items-start justify-between mb-2">
                <h3 className="text-lg font-semibold text-gray-900 line-clamp-1">
                  {destination.name}
                </h3>
                <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800">
                  Active
                </span>
              </div>
              
              <div className="flex items-center text-sm text-gray-600 mb-3">
                <MapPin className="h-4 w-4 mr-1" />
                {destination.city}, {destination.country}
              </div>

              <p className="text-sm text-gray-600 mb-4 line-clamp-2">
                {destination.description}
              </p>

              {/* Stats */}
              <div className="grid grid-cols-3 gap-4 mb-4 text-center">
                <div>
                  <p className="text-lg font-semibold text-gray-900">{formatCurrency(destination.price)}</p>
                  <p className="text-xs text-gray-500">Price</p>
                </div>
                <div>
                  <p className="text-lg font-semibold text-gray-900">{destination.duration}</p>
                  <p className="text-xs text-gray-500">Days</p>
                </div>
                <div>
                  <p className="text-lg font-semibold text-gray-900">{destination.max_guests}</p>
                  <p className="text-xs text-gray-500">Max Guests</p>
                </div>
              </div>

              {/* Actions */}
              <div className="flex space-x-2">
                <Link
                  href={`/destinations/${destination.id}`}
                  className="flex-1 flex items-center justify-center px-3 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
                >
                  <Eye className="h-4 w-4 mr-1" />
                  View
                </Link>
                <Link
                  href={`/admin/destinations/${destination.id}/edit`}
                  className="flex-1 flex items-center justify-center px-3 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors"
                >
                  <Edit className="h-4 w-4 mr-1" />
                  Edit
                </Link>
                <button
                  onClick={() => handleDeleteDestination(destination.id)}
                  className="px-3 py-2 border border-red-300 text-red-600 rounded-lg hover:bg-red-50 transition-colors"
                >
                  <Trash2 className="h-4 w-4" />
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Empty State */}
      {filteredDestinations.length === 0 && (
        <div className="text-center py-12">
          <MapPin className="h-16 w-16 text-gray-300 mx-auto mb-4" />
          <h3 className="text-xl font-semibold text-gray-900 mb-2">No destinations found</h3>
          <p className="text-gray-600 mb-6">
            {searchQuery ? 'Try adjusting your search criteria.' : 'Get started by adding your first destination.'}
          </p>
          {!searchQuery && (
            <Link
              href="/admin/destinations/new"
              className="inline-flex items-center px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors"
            >
              <Plus className="h-4 w-4 mr-2" />
              Add Destination
            </Link>
          )}
        </div>
      )}

      {/* Pagination */}
      {filteredDestinations.length > 0 && (
        <div className="mt-8 flex items-center justify-between">
          <p className="text-sm text-gray-700">
            Showing {filteredDestinations.length} of {destinations.length} destinations
          </p>
          <div className="flex space-x-2">
            <button className="px-3 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors">
              Previous
            </button>
            <button className="px-3 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors">
              1
            </button>
            <button className="px-3 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors">
              Next
            </button>
          </div>
        </div>
      )}
    </AdminLayout>
  )
}
