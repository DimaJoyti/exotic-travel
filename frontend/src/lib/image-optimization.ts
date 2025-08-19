// Image optimization utilities for client-side processing

export interface ImageOptimizationOptions {
  maxWidth?: number
  maxHeight?: number
  quality?: number
  format?: 'jpeg' | 'png' | 'webp'
  maintainAspectRatio?: boolean
}

export interface ImageDimensions {
  width: number
  height: number
}

export interface OptimizedImageResult {
  blob: Blob
  dataUrl: string
  dimensions: ImageDimensions
  originalSize: number
  optimizedSize: number
  compressionRatio: number
}

export class ImageOptimizer {
  private static readonly DEFAULT_OPTIONS: ImageOptimizationOptions = {
    maxWidth: 1920,
    maxHeight: 1080,
    quality: 0.8,
    format: 'jpeg',
    maintainAspectRatio: true,
  }

  /**
   * Optimize an image file
   */
  static async optimizeImage(
    file: File,
    options: ImageOptimizationOptions = {}
  ): Promise<OptimizedImageResult> {
    const opts = { ...this.DEFAULT_OPTIONS, ...options }
    
    return new Promise((resolve, reject) => {
      const img = new Image()
      
      img.onload = () => {
        try {
          const result = this.processImage(img, file, opts)
          resolve(result)
        } catch (error) {
          reject(error)
        }
      }
      
      img.onerror = () => {
        reject(new Error('Failed to load image'))
      }
      
      img.src = URL.createObjectURL(file)
    })
  }

  /**
   * Process the loaded image
   */
  private static processImage(
    img: HTMLImageElement,
    originalFile: File,
    options: ImageOptimizationOptions
  ): OptimizedImageResult {
    const canvas = document.createElement('canvas')
    const ctx = canvas.getContext('2d')
    
    if (!ctx) {
      throw new Error('Canvas context not available')
    }

    // Calculate new dimensions
    const newDimensions = this.calculateDimensions(
      { width: img.width, height: img.height },
      options
    )

    // Set canvas size
    canvas.width = newDimensions.width
    canvas.height = newDimensions.height

    // Draw image on canvas
    ctx.drawImage(img, 0, 0, newDimensions.width, newDimensions.height)

    // Convert to blob
    return new Promise((resolve, reject) => {
      canvas.toBlob(
        (blob) => {
          if (!blob) {
            reject(new Error('Failed to create blob'))
            return
          }

          const dataUrl = canvas.toDataURL(`image/${options.format}`, options.quality)
          const compressionRatio = originalFile.size / blob.size

          resolve({
            blob,
            dataUrl,
            dimensions: newDimensions,
            originalSize: originalFile.size,
            optimizedSize: blob.size,
            compressionRatio,
          })
        },
        `image/${options.format}`,
        options.quality
      )
    }) as any // Type assertion to match return type
  }

  /**
   * Calculate optimal dimensions based on constraints
   */
  private static calculateDimensions(
    original: ImageDimensions,
    options: ImageOptimizationOptions
  ): ImageDimensions {
    const { maxWidth = Infinity, maxHeight = Infinity, maintainAspectRatio = true } = options

    if (!maintainAspectRatio) {
      return {
        width: Math.min(original.width, maxWidth),
        height: Math.min(original.height, maxHeight),
      }
    }

    // Calculate aspect ratio
    const aspectRatio = original.width / original.height

    // Determine limiting dimension
    let newWidth = Math.min(original.width, maxWidth)
    let newHeight = Math.min(original.height, maxHeight)

    // Adjust to maintain aspect ratio
    if (newWidth / aspectRatio > newHeight) {
      newWidth = newHeight * aspectRatio
    } else {
      newHeight = newWidth / aspectRatio
    }

    return {
      width: Math.round(newWidth),
      height: Math.round(newHeight),
    }
  }

  /**
   * Generate multiple sizes for responsive images
   */
  static async generateResponsiveSizes(
    file: File,
    sizes: number[] = [480, 768, 1024, 1280, 1920],
    options: Omit<ImageOptimizationOptions, 'maxWidth' | 'maxHeight'> = {}
  ): Promise<{ size: number; result: OptimizedImageResult }[]> {
    const results: { size: number; result: OptimizedImageResult }[] = []

    for (const size of sizes) {
      try {
        const result = await this.optimizeImage(file, {
          ...options,
          maxWidth: size,
          maxHeight: size,
        })
        results.push({ size, result })
      } catch (error) {
        console.warn(`Failed to generate ${size}px version:`, error)
      }
    }

    return results
  }

  /**
   * Create thumbnail from image
   */
  static async createThumbnail(
    file: File,
    size: number = 300,
    options: Omit<ImageOptimizationOptions, 'maxWidth' | 'maxHeight'> = {}
  ): Promise<OptimizedImageResult> {
    return this.optimizeImage(file, {
      ...options,
      maxWidth: size,
      maxHeight: size,
      quality: options.quality || 0.7,
    })
  }

  /**
   * Validate image file
   */
  static validateImageFile(file: File): { valid: boolean; error?: string } {
    // Check if it's an image
    if (!file.type.startsWith('image/')) {
      return { valid: false, error: 'File is not an image' }
    }

    // Check supported formats
    const supportedFormats = ['image/jpeg', 'image/png', 'image/webp', 'image/gif']
    if (!supportedFormats.includes(file.type)) {
      return { valid: false, error: `Unsupported format: ${file.type}` }
    }

    return { valid: true }
  }

  /**
   * Get image metadata without processing
   */
  static async getImageMetadata(file: File): Promise<{
    dimensions: ImageDimensions
    size: number
    type: string
    name: string
  }> {
    return new Promise((resolve, reject) => {
      const img = new Image()
      
      img.onload = () => {
        resolve({
          dimensions: { width: img.width, height: img.height },
          size: file.size,
          type: file.type,
          name: file.name,
        })
        URL.revokeObjectURL(img.src)
      }
      
      img.onerror = () => {
        reject(new Error('Failed to load image for metadata'))
        URL.revokeObjectURL(img.src)
      }
      
      img.src = URL.createObjectURL(file)
    })
  }

  /**
   * Convert image to different format
   */
  static async convertFormat(
    file: File,
    targetFormat: 'jpeg' | 'png' | 'webp',
    quality: number = 0.8
  ): Promise<Blob> {
    const result = await this.optimizeImage(file, {
      format: targetFormat,
      quality,
      maxWidth: Infinity,
      maxHeight: Infinity,
    })
    
    return result.blob
  }

  /**
   * Compress image while maintaining dimensions
   */
  static async compressImage(
    file: File,
    quality: number = 0.8
  ): Promise<OptimizedImageResult> {
    return this.optimizeImage(file, {
      quality,
      maxWidth: Infinity,
      maxHeight: Infinity,
    })
  }

  /**
   * Calculate file size reduction
   */
  static calculateSizeReduction(originalSize: number, newSize: number): {
    reduction: number
    percentage: string
    saved: string
  } {
    const reduction = originalSize - newSize
    const percentage = ((reduction / originalSize) * 100).toFixed(1)
    const saved = this.formatFileSize(reduction)
    
    return { reduction, percentage: `${percentage}%`, saved }
  }

  /**
   * Format file size for display
   */
  static formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 Bytes'
    
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  /**
   * Check if WebP is supported
   */
  static isWebPSupported(): boolean {
    const canvas = document.createElement('canvas')
    canvas.width = 1
    canvas.height = 1
    return canvas.toDataURL('image/webp').indexOf('data:image/webp') === 0
  }

  /**
   * Get optimal format for the browser
   */
  static getOptimalFormat(): 'webp' | 'jpeg' {
    return this.isWebPSupported() ? 'webp' : 'jpeg'
  }

  /**
   * Batch optimize multiple images
   */
  static async batchOptimize(
    files: File[],
    options: ImageOptimizationOptions = {},
    onProgress?: (completed: number, total: number) => void
  ): Promise<OptimizedImageResult[]> {
    const results: OptimizedImageResult[] = []
    
    for (let i = 0; i < files.length; i++) {
      try {
        const result = await this.optimizeImage(files[i], options)
        results.push(result)
        onProgress?.(i + 1, files.length)
      } catch (error) {
        console.error(`Failed to optimize ${files[i].name}:`, error)
        // Continue with other files
      }
    }
    
    return results
  }
}
