'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { 
  Upload, 
  Image as ImageIcon, 
  HardDrive, 
  Eye, 
  Download,
  Trash2,
  Settings,
  BarChart3
} from 'lucide-react'
import { useAuth } from '@/contexts/auth-context'
import { ImagesService, ImageMetadata } from '@/lib/images'
import AdminLayout from '@/components/admin/admin-layout'
import ImageManager from '@/components/images/image-manager'

export default function AdminImagesPage() {
  const { user } = useAuth()
  const router = useRouter()
  const [images, setImages] = useState<ImageMetadata[]>([])
  const [loading, setLoading] = useState(true)
  const [stats, setStats] = useState({
    totalImages: 0,
    totalSize: 0,
    storageUsed: 0,
    storageLimit: 10 * 1024 * 1024 * 1024, // 10GB
  })

  useEffect(() => {
    if (!user) {
      router.push('/auth/login')
      return
    }

    if (user.role !== 'admin') {
      router.push('/dashboard')
      return
    }

    loadImages()
  }, [user, router])

  const loadImages = async () => {
    try {
      const mockImages = ImagesService.getMockImages()
      setImages(mockImages)
      
      // Calculate stats
      const totalSize = mockImages.reduce((sum, img) => sum + img.size, 0)
      setStats({
        totalImages: mockImages.length,
        totalSize,
        storageUsed: totalSize,
        storageLimit: 10 * 1024 * 1024 * 1024, // 10GB
      })
    } catch (error) {
      console.error('Error loading images:', error)
    } finally {
      setLoading(false)
    }
  }

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes'
    
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  const getStoragePercentage = (): number => {
    return (stats.storageUsed / stats.storageLimit) * 100
  }

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
            <h1 className="text-3xl font-bold text-gray-900">Image Management</h1>
            <p className="text-gray-600 mt-1">Manage your platform's images and media assets</p>
          </div>
          <div className="flex space-x-3">
            <button className="flex items-center px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors">
              <Settings className="h-4 w-4 mr-2" />
              Settings
            </button>
            <button className="flex items-center px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors">
              <Download className="h-4 w-4 mr-2" />
              Export
            </button>
          </div>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center">
            <div className="p-2 bg-blue-100 rounded-lg">
              <ImageIcon className="h-6 w-6 text-blue-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600">Total Images</p>
              <p className="text-2xl font-bold text-gray-900">{stats.totalImages.toLocaleString()}</p>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center">
            <div className="p-2 bg-green-100 rounded-lg">
              <HardDrive className="h-6 w-6 text-green-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600">Storage Used</p>
              <p className="text-2xl font-bold text-gray-900">{formatFileSize(stats.storageUsed)}</p>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center">
            <div className="p-2 bg-purple-100 rounded-lg">
              <BarChart3 className="h-6 w-6 text-purple-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600">Storage Usage</p>
              <p className="text-2xl font-bold text-gray-900">{getStoragePercentage().toFixed(1)}%</p>
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center">
            <div className="p-2 bg-yellow-100 rounded-lg">
              <Upload className="h-6 w-6 text-yellow-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600">Avg. File Size</p>
              <p className="text-2xl font-bold text-gray-900">
                {stats.totalImages > 0 ? formatFileSize(stats.totalSize / stats.totalImages) : '0 Bytes'}
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Storage Usage Bar */}
      <div className="bg-white rounded-lg shadow p-6 mb-8">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900">Storage Usage</h3>
          <span className="text-sm text-gray-600">
            {formatFileSize(stats.storageUsed)} of {formatFileSize(stats.storageLimit)} used
          </span>
        </div>
        
        <div className="w-full bg-gray-200 rounded-full h-3">
          <div
            className={`h-3 rounded-full transition-all duration-300 ${
              getStoragePercentage() > 90 
                ? 'bg-red-500' 
                : getStoragePercentage() > 75 
                ? 'bg-yellow-500' 
                : 'bg-green-500'
            }`}
            style={{ width: `${Math.min(getStoragePercentage(), 100)}%` }}
          />
        </div>
        
        {getStoragePercentage() > 80 && (
          <div className="mt-3 p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
            <p className="text-sm text-yellow-800">
              <strong>Warning:</strong> You're approaching your storage limit. Consider upgrading your plan or cleaning up unused images.
            </p>
          </div>
        )}
      </div>

      {/* Recent Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 mb-8">
        <div className="lg:col-span-2">
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Recent Uploads</h3>
            
            <div className="space-y-3">
              {images.slice(0, 5).map((image) => (
                <div key={image.id} className="flex items-center space-x-4">
                  <img
                    src={ImagesService.getOptimizedImageUrl(image.thumbnailUrl || image.url, {
                      width: 60,
                      height: 60,
                      fit: 'cover'
                    })}
                    alt={image.alt || image.filename}
                    className="w-12 h-12 object-cover rounded-lg"
                  />
                  
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-gray-900 truncate">
                      {image.originalName}
                    </p>
                    <p className="text-xs text-gray-500">
                      {formatFileSize(image.size)} • {new Date(image.uploadedAt).toLocaleDateString()}
                    </p>
                  </div>
                  
                  <div className="flex space-x-2">
                    <button className="p-1 text-gray-400 hover:text-blue-600">
                      <Eye className="h-4 w-4" />
                    </button>
                    <button className="p-1 text-gray-400 hover:text-red-600">
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">File Types</h3>
          
          <div className="space-y-3">
            {[
              { type: 'JPEG', count: 45, percentage: 60 },
              { type: 'PNG', count: 20, percentage: 27 },
              { type: 'WebP', count: 8, percentage: 11 },
              { type: 'GIF', count: 2, percentage: 2 },
            ].map((fileType) => (
              <div key={fileType.type} className="flex items-center justify-between">
                <span className="text-sm text-gray-600">{fileType.type}</span>
                <div className="flex items-center space-x-3">
                  <div className="w-20 bg-gray-200 rounded-full h-2">
                    <div
                      className="bg-primary h-2 rounded-full"
                      style={{ width: `${fileType.percentage}%` }}
                    />
                  </div>
                  <span className="text-sm font-medium text-gray-900 w-8 text-right">
                    {fileType.count}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Image Manager */}
      <div className="bg-white rounded-lg shadow">
        <ImageManager
          selectionMode="multiple"
          showUpload={true}
          onSelectionChange={(selectedIds) => {
            console.log('Selected images:', selectedIds)
          }}
        />
      </div>

      {/* Optimization Tips */}
      <div className="mt-8 bg-blue-50 border border-blue-200 rounded-lg p-6">
        <h3 className="text-lg font-semibold text-blue-900 mb-4">💡 Optimization Tips</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm text-blue-800">
          <div>
            <h4 className="font-medium mb-2">Image Formats</h4>
            <ul className="space-y-1">
              <li>• Use WebP for better compression</li>
              <li>• JPEG for photos, PNG for graphics</li>
              <li>• Avoid large GIF files</li>
            </ul>
          </div>
          <div>
            <h4 className="font-medium mb-2">File Sizes</h4>
            <ul className="space-y-1">
              <li>• Keep images under 1MB when possible</li>
              <li>• Use appropriate dimensions</li>
              <li>• Compress images before upload</li>
            </ul>
          </div>
          <div>
            <h4 className="font-medium mb-2">SEO & Accessibility</h4>
            <ul className="space-y-1">
              <li>• Add descriptive alt text</li>
              <li>• Use meaningful filenames</li>
              <li>• Include relevant tags</li>
            </ul>
          </div>
          <div>
            <h4 className="font-medium mb-2">Performance</h4>
            <ul className="space-y-1">
              <li>• Enable lazy loading</li>
              <li>• Use responsive images</li>
              <li>• Optimize for mobile</li>
            </ul>
          </div>
        </div>
      </div>
    </AdminLayout>
  )
}
