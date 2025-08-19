'use client'

import { useState, useEffect } from 'react'
import { 
  Search, 
  Filter, 
  Grid, 
  List, 
  Upload, 
  Trash2, 
  Edit, 
  Download,
  Tag,
  Calendar,
  FileImage,
  MoreHorizontal
} from 'lucide-react'
import { ImageMetadata, ImagesService, ImageUploadResult } from '@/lib/images'
import ImageUpload from './image-upload'
import { ImageGrid } from './image-gallery'

interface ImageManagerProps {
  onImageSelect?: (image: ImageMetadata) => void
  selectionMode?: 'single' | 'multiple' | 'none'
  selectedImages?: string[]
  onSelectionChange?: (selectedIds: string[]) => void
  showUpload?: boolean
  className?: string
}

export default function ImageManager({
  onImageSelect,
  selectionMode = 'none',
  selectedImages = [],
  onSelectionChange,
  showUpload = true,
  className = ''
}: ImageManagerProps) {
  const [images, setImages] = useState<ImageMetadata[]>([])
  const [loading, setLoading] = useState(true)
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedTags, setSelectedTags] = useState<string[]>([])
  const [sortBy, setSortBy] = useState<'uploadedAt' | 'filename' | 'size'>('uploadedAt')
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc')
  const [showUploadModal, setShowUploadModal] = useState(false)
  const [selectedImageIds, setSelectedImageIds] = useState<string[]>(selectedImages)

  useEffect(() => {
    loadImages()
  }, [searchQuery, selectedTags, sortBy, sortOrder])

  useEffect(() => {
    setSelectedImageIds(selectedImages)
  }, [selectedImages])

  const loadImages = async () => {
    try {
      // In real implementation, this would call the API
      const mockImages = ImagesService.getMockImages()
      setImages(mockImages)
    } catch (error) {
      console.error('Error loading images:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleUploadComplete = (results: ImageUploadResult[]) => {
    const successfulUploads = results
      .filter(result => result.success && result.image)
      .map(result => result.image!)
    
    setImages(prev => [...successfulUploads, ...prev])
    setShowUploadModal(false)
  }

  const handleImageSelect = (image: ImageMetadata) => {
    if (selectionMode === 'none') {
      onImageSelect?.(image)
      return
    }

    const imageId = image.id
    let newSelection: string[]

    if (selectionMode === 'single') {
      newSelection = [imageId]
    } else {
      // Multiple selection
      if (selectedImageIds.includes(imageId)) {
        newSelection = selectedImageIds.filter(id => id !== imageId)
      } else {
        newSelection = [...selectedImageIds, imageId]
      }
    }

    setSelectedImageIds(newSelection)
    onSelectionChange?.(newSelection)
    
    if (selectionMode === 'single') {
      onImageSelect?.(image)
    }
  }

  const handleDeleteImages = async (imageIds: string[]) => {
    if (!confirm(`Are you sure you want to delete ${imageIds.length} image(s)?`)) {
      return
    }

    try {
      // In real implementation, call API to delete images
      setImages(prev => prev.filter(img => !imageIds.includes(img.id)))
      setSelectedImageIds(prev => prev.filter(id => !imageIds.includes(id)))
      onSelectionChange?.(selectedImageIds.filter(id => !imageIds.includes(id)))
    } catch (error) {
      console.error('Error deleting images:', error)
    }
  }

  const filteredImages = images.filter(image => {
    const matchesSearch = searchQuery === '' || 
      image.filename.toLowerCase().includes(searchQuery.toLowerCase()) ||
      image.alt?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      image.caption?.toLowerCase().includes(searchQuery.toLowerCase())

    const matchesTags = selectedTags.length === 0 ||
      selectedTags.some(tag => image.tags?.includes(tag))

    return matchesSearch && matchesTags
  })

  const sortedImages = [...filteredImages].sort((a, b) => {
    let comparison = 0
    
    switch (sortBy) {
      case 'filename':
        comparison = a.filename.localeCompare(b.filename)
        break
      case 'size':
        comparison = a.size - b.size
        break
      case 'uploadedAt':
      default:
        comparison = new Date(a.uploadedAt).getTime() - new Date(b.uploadedAt).getTime()
        break
    }

    return sortOrder === 'desc' ? -comparison : comparison
  })

  const allTags = Array.from(new Set(images.flatMap(img => img.tags || [])))

  if (loading) {
    return (
      <div className={`bg-white rounded-lg border border-gray-200 p-8 ${className}`}>
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-gray-200 rounded w-1/4"></div>
          <div className="grid grid-cols-3 gap-4">
            {[1, 2, 3, 4, 5, 6].map(i => (
              <div key={i} className="aspect-square bg-gray-200 rounded"></div>
            ))}
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className={`bg-white rounded-lg border border-gray-200 ${className}`}>
      {/* Header */}
      <div className="p-6 border-b border-gray-200">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900">Image Manager</h3>
          <div className="flex items-center space-x-2">
            {showUpload && (
              <button
                onClick={() => setShowUploadModal(true)}
                className="flex items-center px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors"
              >
                <Upload className="h-4 w-4 mr-2" />
                Upload
              </button>
            )}
            
            {selectedImageIds.length > 0 && (
              <button
                onClick={() => handleDeleteImages(selectedImageIds)}
                className="flex items-center px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
              >
                <Trash2 className="h-4 w-4 mr-2" />
                Delete ({selectedImageIds.length})
              </button>
            )}
          </div>
        </div>

        {/* Search and Filters */}
        <div className="flex flex-col md:flex-row gap-4">
          <div className="flex-1 relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
            <input
              type="text"
              placeholder="Search images..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
            />
          </div>
          
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as any)}
            className="px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          >
            <option value="uploadedAt">Upload Date</option>
            <option value="filename">Filename</option>
            <option value="size">File Size</option>
          </select>
          
          <select
            value={sortOrder}
            onChange={(e) => setSortOrder(e.target.value as any)}
            className="px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          >
            <option value="desc">Descending</option>
            <option value="asc">Ascending</option>
          </select>
          
          <div className="flex border border-gray-300 rounded-lg">
            <button
              onClick={() => setViewMode('grid')}
              className={`p-2 ${viewMode === 'grid' ? 'bg-primary text-primary-foreground' : 'text-gray-600 hover:bg-gray-50'}`}
            >
              <Grid className="h-4 w-4" />
            </button>
            <button
              onClick={() => setViewMode('list')}
              className={`p-2 ${viewMode === 'list' ? 'bg-primary text-primary-foreground' : 'text-gray-600 hover:bg-gray-50'}`}
            >
              <List className="h-4 w-4" />
            </button>
          </div>
        </div>

        {/* Tags Filter */}
        {allTags.length > 0 && (
          <div className="mt-4">
            <div className="flex flex-wrap gap-2">
              {allTags.map(tag => (
                <button
                  key={tag}
                  onClick={() => {
                    if (selectedTags.includes(tag)) {
                      setSelectedTags(prev => prev.filter(t => t !== tag))
                    } else {
                      setSelectedTags(prev => [...prev, tag])
                    }
                  }}
                  className={`px-3 py-1 text-sm rounded-full transition-colors ${
                    selectedTags.includes(tag)
                      ? 'bg-primary text-primary-foreground'
                      : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                  }`}
                >
                  <Tag className="h-3 w-3 mr-1 inline" />
                  {tag}
                </button>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Content */}
      <div className="p-6">
        {sortedImages.length === 0 ? (
          <div className="text-center py-12">
            <FileImage className="h-16 w-16 text-gray-300 mx-auto mb-4" />
            <h3 className="text-xl font-semibold text-gray-900 mb-2">No images found</h3>
            <p className="text-gray-600 mb-6">
              {searchQuery || selectedTags.length > 0 
                ? 'Try adjusting your search criteria.' 
                : 'Upload some images to get started.'
              }
            </p>
            {showUpload && !searchQuery && selectedTags.length === 0 && (
              <button
                onClick={() => setShowUploadModal(true)}
                className="bg-primary text-primary-foreground px-6 py-3 rounded-lg hover:bg-primary/90 transition-colors"
              >
                Upload Images
              </button>
            )}
          </div>
        ) : viewMode === 'grid' ? (
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
            {sortedImages.map(image => (
              <div
                key={image.id}
                className={`relative aspect-square bg-gray-100 rounded-lg overflow-hidden cursor-pointer group border-2 transition-colors ${
                  selectedImageIds.includes(image.id)
                    ? 'border-primary'
                    : 'border-transparent hover:border-gray-300'
                }`}
                onClick={() => handleImageSelect(image)}
              >
                <img
                  src={ImagesService.getOptimizedImageUrl(image.url, {
                    width: 300,
                    height: 300,
                    fit: 'cover'
                  })}
                  alt={image.alt || image.filename}
                  className="w-full h-full object-cover"
                />
                
                {selectionMode !== 'none' && (
                  <div className="absolute top-2 left-2">
                    <input
                      type="checkbox"
                      checked={selectedImageIds.includes(image.id)}
                      onChange={() => handleImageSelect(image)}
                      className="h-4 w-4 text-primary focus:ring-primary border-gray-300 rounded"
                    />
                  </div>
                )}
                
                <div className="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black to-transparent p-3">
                  <p className="text-white text-sm font-medium truncate">
                    {image.caption || image.filename}
                  </p>
                  <p className="text-white text-xs opacity-75">
                    {(image.size / (1024 * 1024)).toFixed(2)} MB
                  </p>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="space-y-2">
            {sortedImages.map(image => (
              <div
                key={image.id}
                className={`flex items-center p-4 border rounded-lg cursor-pointer transition-colors ${
                  selectedImageIds.includes(image.id)
                    ? 'border-primary bg-primary/5'
                    : 'border-gray-200 hover:border-gray-300'
                }`}
                onClick={() => handleImageSelect(image)}
              >
                {selectionMode !== 'none' && (
                  <input
                    type="checkbox"
                    checked={selectedImageIds.includes(image.id)}
                    onChange={() => handleImageSelect(image)}
                    className="h-4 w-4 text-primary focus:ring-primary border-gray-300 rounded mr-4"
                  />
                )}
                
                <img
                  src={ImagesService.getOptimizedImageUrl(image.thumbnailUrl || image.url, {
                    width: 80,
                    height: 80,
                    fit: 'cover'
                  })}
                  alt={image.alt || image.filename}
                  className="w-16 h-16 object-cover rounded-lg mr-4"
                />
                
                <div className="flex-1 min-w-0">
                  <h4 className="text-sm font-medium text-gray-900 truncate">
                    {image.caption || image.filename}
                  </h4>
                  <p className="text-xs text-gray-500">
                    {image.width} × {image.height} • {(image.size / (1024 * 1024)).toFixed(2)} MB
                  </p>
                  <p className="text-xs text-gray-500">
                    {new Date(image.uploadedAt).toLocaleDateString()}
                  </p>
                </div>
                
                <div className="flex items-center space-x-2">
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      // Handle edit
                    }}
                    className="p-1 text-gray-400 hover:text-gray-600"
                  >
                    <Edit className="h-4 w-4" />
                  </button>
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      handleDeleteImages([image.id])
                    }}
                    className="p-1 text-gray-400 hover:text-red-600"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Upload Modal */}
      {showUploadModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="p-6 border-b border-gray-200">
              <div className="flex justify-between items-center">
                <h3 className="text-lg font-semibold text-gray-900">Upload Images</h3>
                <button
                  onClick={() => setShowUploadModal(false)}
                  className="text-gray-400 hover:text-gray-600"
                >
                  ×
                </button>
              </div>
            </div>
            <div className="p-6">
              <ImageUpload
                multiple={true}
                maxFiles={10}
                onUploadComplete={handleUploadComplete}
              />
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
