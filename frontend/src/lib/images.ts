import api from './api'

export interface ImageUploadOptions {
  maxSize?: number // in bytes
  allowedTypes?: string[]
  quality?: number // 0-1 for compression
  maxWidth?: number
  maxHeight?: number
  generateThumbnail?: boolean
  thumbnailSize?: number
}

export interface ImageMetadata {
  id: string
  filename: string
  originalName: string
  mimeType: string
  size: number
  width: number
  height: number
  url: string
  thumbnailUrl?: string
  uploadedAt: string
  uploadedBy: string
  tags?: string[]
  alt?: string
  caption?: string
}

export interface ImageUploadResult {
  success: boolean
  image?: ImageMetadata
  error?: string
  progress?: number
}

export interface ImageGallery {
  id: string
  name: string
  description?: string
  images: ImageMetadata[]
  coverImage?: string
  createdAt: string
  updatedAt: string
}

export class ImagesService {
  private static readonly DEFAULT_OPTIONS: ImageUploadOptions = {
    maxSize: 10 * 1024 * 1024, // 10MB
    allowedTypes: ['image/jpeg', 'image/png', 'image/webp', 'image/gif'],
    quality: 0.8,
    maxWidth: 2048,
    maxHeight: 2048,
    generateThumbnail: true,
    thumbnailSize: 300,
  }

  // Upload single image
  static async uploadImage(
    file: File,
    options: ImageUploadOptions = {},
    onProgress?: (progress: number) => void
  ): Promise<ImageUploadResult> {
    const opts = { ...this.DEFAULT_OPTIONS, ...options }

    // Validate file
    const validation = this.validateFile(file, opts)
    if (!validation.valid) {
      return { success: false, error: validation.error }
    }

    try {
      // Process image (resize, compress, etc.)
      const processedFile = await this.processImage(file, opts)
      
      // Create form data
      const formData = new FormData()
      formData.append('image', processedFile)
      formData.append('options', JSON.stringify(opts))

      // Upload with progress tracking
      const response = await this.uploadWithProgress(formData, onProgress)
      
      return {
        success: true,
        image: response.data
      }
    } catch (error: any) {
      console.error('Image upload error:', error)
      return {
        success: false,
        error: error.message || 'Upload failed'
      }
    }
  }

  // Upload multiple images
  static async uploadMultipleImages(
    files: File[],
    options: ImageUploadOptions = {},
    onProgress?: (progress: number) => void
  ): Promise<ImageUploadResult[]> {
    const results: ImageUploadResult[] = []
    let totalProgress = 0

    for (let i = 0; i < files.length; i++) {
      const file = files[i]
      
      const result = await this.uploadImage(file, options, (fileProgress) => {
        const overallProgress = ((i + fileProgress / 100) / files.length) * 100
        onProgress?.(overallProgress)
      })
      
      results.push(result)
      totalProgress = ((i + 1) / files.length) * 100
      onProgress?.(totalProgress)
    }

    return results
  }

  // Get image by ID
  static async getImage(imageId: string): Promise<ImageMetadata> {
    const response = await api.get<ImageMetadata>(`/api/images/${imageId}`)
    return response.data
  }

  // Get images with pagination and filtering
  static async getImages(params: {
    page?: number
    limit?: number
    tags?: string[]
    mimeType?: string
    uploadedBy?: string
    sortBy?: 'uploadedAt' | 'filename' | 'size'
    sortOrder?: 'asc' | 'desc'
  } = {}): Promise<{ images: ImageMetadata[]; total: number; page: number; limit: number }> {
    const response = await api.get<{ images: ImageMetadata[]; total: number; page: number; limit: number }>(
      '/api/images',
      { params }
    )
    return response.data
  }

  // Update image metadata
  static async updateImageMetadata(
    imageId: string,
    metadata: Partial<Pick<ImageMetadata, 'alt' | 'caption' | 'tags'>>
  ): Promise<ImageMetadata> {
    const response = await api.patch<ImageMetadata>(`/api/images/${imageId}`, metadata)
    return response.data
  }

  // Delete image
  static async deleteImage(imageId: string): Promise<void> {
    await api.delete(`/api/images/${imageId}`)
  }

  // Create image gallery
  static async createGallery(gallery: Omit<ImageGallery, 'id' | 'createdAt' | 'updatedAt'>): Promise<ImageGallery> {
    const response = await api.post<ImageGallery>('/api/galleries', gallery)
    return response.data
  }

  // Get gallery by ID
  static async getGallery(galleryId: string): Promise<ImageGallery> {
    const response = await api.get<ImageGallery>(`/api/galleries/${galleryId}`)
    return response.data
  }

  // Update gallery
  static async updateGallery(galleryId: string, updates: Partial<ImageGallery>): Promise<ImageGallery> {
    const response = await api.patch<ImageGallery>(`/api/galleries/${galleryId}`, updates)
    return response.data
  }

  // Delete gallery
  static async deleteGallery(galleryId: string): Promise<void> {
    await api.delete(`/api/galleries/${galleryId}`)
  }

  // Generate optimized image URL
  static getOptimizedImageUrl(
    imageUrl: string,
    options: {
      width?: number
      height?: number
      quality?: number
      format?: 'webp' | 'jpeg' | 'png'
      fit?: 'cover' | 'contain' | 'fill'
    } = {}
  ): string {
    if (!imageUrl) return ''
    
    // If it's already an optimized URL or external URL, return as-is
    if (imageUrl.includes('?') || imageUrl.startsWith('http')) {
      return imageUrl
    }

    const params = new URLSearchParams()
    if (options.width) params.append('w', options.width.toString())
    if (options.height) params.append('h', options.height.toString())
    if (options.quality) params.append('q', Math.round(options.quality * 100).toString())
    if (options.format) params.append('f', options.format)
    if (options.fit) params.append('fit', options.fit)

    const queryString = params.toString()
    return queryString ? `${imageUrl}?${queryString}` : imageUrl
  }

  // Validate file before upload
  private static validateFile(file: File, options: ImageUploadOptions): { valid: boolean; error?: string } {
    // Check file size
    if (options.maxSize && file.size > options.maxSize) {
      return {
        valid: false,
        error: `File size (${this.formatFileSize(file.size)}) exceeds maximum allowed size (${this.formatFileSize(options.maxSize)})`
      }
    }

    // Check file type
    if (options.allowedTypes && !options.allowedTypes.includes(file.type)) {
      return {
        valid: false,
        error: `File type ${file.type} is not allowed. Allowed types: ${options.allowedTypes.join(', ')}`
      }
    }

    return { valid: true }
  }

  // Process image (resize, compress, etc.)
  private static async processImage(file: File, options: ImageUploadOptions): Promise<File> {
    return new Promise((resolve, reject) => {
      const canvas = document.createElement('canvas')
      const ctx = canvas.getContext('2d')
      const img = new Image()

      img.onload = () => {
        try {
          // Calculate new dimensions
          let { width, height } = img
          const maxWidth = options.maxWidth || width
          const maxHeight = options.maxHeight || height

          if (width > maxWidth || height > maxHeight) {
            const ratio = Math.min(maxWidth / width, maxHeight / height)
            width *= ratio
            height *= ratio
          }

          // Set canvas dimensions
          canvas.width = width
          canvas.height = height

          // Draw and compress image
          ctx?.drawImage(img, 0, 0, width, height)
          
          canvas.toBlob(
            (blob) => {
              if (blob) {
                const processedFile = new File([blob], file.name, {
                  type: file.type,
                  lastModified: Date.now()
                })
                resolve(processedFile)
              } else {
                reject(new Error('Failed to process image'))
              }
            },
            file.type,
            options.quality || 0.8
          )
        } catch (error) {
          reject(error)
        }
      }

      img.onerror = () => reject(new Error('Failed to load image'))
      img.src = URL.createObjectURL(file)
    })
  }

  // Upload with progress tracking
  private static async uploadWithProgress(
    formData: FormData,
    onProgress?: (progress: number) => void
  ): Promise<any> {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest()

      xhr.upload.addEventListener('progress', (event) => {
        if (event.lengthComputable) {
          const progress = (event.loaded / event.total) * 100
          onProgress?.(progress)
        }
      })

      xhr.addEventListener('load', () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            const response = JSON.parse(xhr.responseText)
            resolve({ data: response })
          } catch (error) {
            reject(new Error('Invalid response format'))
          }
        } else {
          reject(new Error(`Upload failed with status ${xhr.status}`))
        }
      })

      xhr.addEventListener('error', () => {
        reject(new Error('Upload failed'))
      })

      xhr.open('POST', '/api/images/upload')
      xhr.send(formData)
    })
  }

  // Format file size for display
  private static formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 Bytes'
    
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  // Mock implementations for development
  static async uploadMockImage(file: File): Promise<ImageUploadResult> {
    console.log('📸 Mock Image Upload:', file.name, file.size, file.type)
    
    // Simulate upload delay
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    const mockImage: ImageMetadata = {
      id: `img_${Date.now()}`,
      filename: `${Date.now()}_${file.name}`,
      originalName: file.name,
      mimeType: file.type,
      size: file.size,
      width: 1920,
      height: 1080,
      url: URL.createObjectURL(file),
      thumbnailUrl: URL.createObjectURL(file),
      uploadedAt: new Date().toISOString(),
      uploadedBy: 'current_user',
      tags: [],
    }
    
    return {
      success: true,
      image: mockImage
    }
  }

  static getMockImages(): ImageMetadata[] {
    return [
      {
        id: 'img_1',
        filename: 'maldives_beach.jpg',
        originalName: 'Beautiful Maldives Beach.jpg',
        mimeType: 'image/jpeg',
        size: 2048576,
        width: 1920,
        height: 1080,
        url: 'https://images.unsplash.com/photo-1506905925346-21bda4d32df4?w=1920&h=1080&fit=crop',
        thumbnailUrl: 'https://images.unsplash.com/photo-1506905925346-21bda4d32df4?w=300&h=200&fit=crop',
        uploadedAt: '2024-01-15T10:00:00Z',
        uploadedBy: 'admin',
        tags: ['beach', 'maldives', 'tropical'],
        alt: 'Beautiful beach in Maldives with crystal clear water',
        caption: 'Paradise found in the Maldives'
      },
      {
        id: 'img_2',
        filename: 'amazon_rainforest.jpg',
        originalName: 'Amazon Rainforest Canopy.jpg',
        mimeType: 'image/jpeg',
        size: 3145728,
        width: 2048,
        height: 1365,
        url: 'https://images.unsplash.com/photo-1441974231531-c6227db76b6e?w=2048&h=1365&fit=crop',
        thumbnailUrl: 'https://images.unsplash.com/photo-1441974231531-c6227db76b6e?w=300&h=200&fit=crop',
        uploadedAt: '2024-01-16T14:30:00Z',
        uploadedBy: 'admin',
        tags: ['rainforest', 'amazon', 'nature', 'green'],
        alt: 'Lush Amazon rainforest canopy',
        caption: 'The heart of the Amazon rainforest'
      }
    ]
  }
}
