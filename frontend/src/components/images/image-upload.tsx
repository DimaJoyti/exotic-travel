'use client'

import { useState, useRef, useCallback } from 'react'
import { useDropzone } from 'react-dropzone'
import { Upload, X, Image as ImageIcon, AlertCircle, Check } from 'lucide-react'
import { ImagesService, ImageUploadOptions, ImageUploadResult, ImageMetadata } from '@/lib/images'

interface ImageUploadProps {
  multiple?: boolean
  maxFiles?: number
  options?: ImageUploadOptions
  onUploadComplete?: (results: ImageUploadResult[]) => void
  onUploadProgress?: (progress: number) => void
  className?: string
  accept?: string
  disabled?: boolean
}

interface UploadingFile {
  file: File
  progress: number
  result?: ImageUploadResult
  preview: string
}

export default function ImageUpload({
  multiple = false,
  maxFiles = 10,
  options = {},
  onUploadComplete,
  onUploadProgress,
  className = '',
  accept = 'image/*',
  disabled = false
}: ImageUploadProps) {
  const [uploadingFiles, setUploadingFiles] = useState<UploadingFile[]>([])
  const [isUploading, setIsUploading] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const onDrop = useCallback(async (acceptedFiles: File[]) => {
    if (disabled || isUploading) return

    // Limit number of files
    const filesToUpload = acceptedFiles.slice(0, maxFiles)
    
    // Create preview objects
    const uploadingFiles: UploadingFile[] = filesToUpload.map(file => ({
      file,
      progress: 0,
      preview: URL.createObjectURL(file)
    }))

    setUploadingFiles(uploadingFiles)
    setIsUploading(true)

    try {
      const results: ImageUploadResult[] = []
      
      for (let i = 0; i < filesToUpload.length; i++) {
        const file = filesToUpload[i]
        
        // Update progress for current file
        setUploadingFiles(prev => prev.map((uf, index) => 
          index === i ? { ...uf, progress: 0 } : uf
        ))

        const result = await ImagesService.uploadMockImage(file)
        
        // Update with result
        setUploadingFiles(prev => prev.map((uf, index) => 
          index === i ? { ...uf, progress: 100, result } : uf
        ))

        results.push(result)
        
        // Update overall progress
        const overallProgress = ((i + 1) / filesToUpload.length) * 100
        onUploadProgress?.(overallProgress)
      }

      onUploadComplete?.(results)
      
      // Clear uploading files after a delay
      setTimeout(() => {
        setUploadingFiles([])
      }, 2000)
      
    } catch (error) {
      console.error('Upload error:', error)
    } finally {
      setIsUploading(false)
    }
  }, [disabled, isUploading, maxFiles, onUploadComplete, onUploadProgress])

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept: { [accept]: [] },
    multiple,
    maxFiles,
    disabled: disabled || isUploading,
    maxSize: options.maxSize || 10 * 1024 * 1024, // 10MB default
  })

  const removeUploadingFile = (index: number) => {
    setUploadingFiles(prev => {
      const newFiles = [...prev]
      URL.revokeObjectURL(newFiles[index].preview)
      newFiles.splice(index, 1)
      return newFiles
    })
  }

  const triggerFileInput = () => {
    fileInputRef.current?.click()
  }

  return (
    <div className={`space-y-4 ${className}`}>
      {/* Drop Zone */}
      <div
        {...getRootProps()}
        className={`
          border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors
          ${isDragActive 
            ? 'border-primary bg-primary/5' 
            : 'border-gray-300 hover:border-gray-400'
          }
          ${disabled || isUploading ? 'opacity-50 cursor-not-allowed' : ''}
        `}
      >
        <input {...getInputProps()} ref={fileInputRef} />
        
        <div className="space-y-4">
          <div className="mx-auto w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center">
            <Upload className={`h-8 w-8 ${isDragActive ? 'text-primary' : 'text-gray-400'}`} />
          </div>
          
          <div>
            <p className="text-lg font-medium text-gray-900">
              {isDragActive ? 'Drop images here' : 'Upload images'}
            </p>
            <p className="text-sm text-gray-600 mt-1">
              Drag and drop {multiple ? 'images' : 'an image'} here, or{' '}
              <button
                type="button"
                onClick={triggerFileInput}
                className="text-primary hover:text-primary/80 font-medium"
                disabled={disabled || isUploading}
              >
                browse files
              </button>
            </p>
          </div>
          
          <div className="text-xs text-gray-500">
            <p>Supported formats: JPEG, PNG, WebP, GIF</p>
            <p>Maximum size: {Math.round((options.maxSize || 10 * 1024 * 1024) / (1024 * 1024))}MB per file</p>
            {multiple && <p>Maximum {maxFiles} files</p>}
          </div>
        </div>
      </div>

      {/* Uploading Files */}
      {uploadingFiles.length > 0 && (
        <div className="space-y-3">
          <h4 className="text-sm font-medium text-gray-900">
            {isUploading ? 'Uploading...' : 'Upload Complete'}
          </h4>
          
          <div className="space-y-2">
            {uploadingFiles.map((uploadingFile, index) => (
              <div key={index} className="bg-white border border-gray-200 rounded-lg p-4">
                <div className="flex items-center space-x-4">
                  {/* Preview */}
                  <div className="w-16 h-16 bg-gray-100 rounded-lg overflow-hidden flex-shrink-0">
                    <img
                      src={uploadingFile.preview}
                      alt={uploadingFile.file.name}
                      className="w-full h-full object-cover"
                    />
                  </div>
                  
                  {/* File Info */}
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-gray-900 truncate">
                      {uploadingFile.file.name}
                    </p>
                    <p className="text-xs text-gray-500">
                      {(uploadingFile.file.size / (1024 * 1024)).toFixed(2)} MB
                    </p>
                    
                    {/* Progress Bar */}
                    <div className="mt-2">
                      <div className="bg-gray-200 rounded-full h-2">
                        <div
                          className={`h-2 rounded-full transition-all duration-300 ${
                            uploadingFile.result?.success 
                              ? 'bg-green-500' 
                              : uploadingFile.result?.error 
                              ? 'bg-red-500' 
                              : 'bg-primary'
                          }`}
                          style={{ width: `${uploadingFile.progress}%` }}
                        />
                      </div>
                      <div className="flex justify-between items-center mt-1">
                        <span className="text-xs text-gray-500">
                          {uploadingFile.progress}%
                        </span>
                        {uploadingFile.result && (
                          <span className="text-xs">
                            {uploadingFile.result.success ? (
                              <span className="text-green-600 flex items-center">
                                <Check className="h-3 w-3 mr-1" />
                                Complete
                              </span>
                            ) : (
                              <span className="text-red-600 flex items-center">
                                <AlertCircle className="h-3 w-3 mr-1" />
                                Failed
                              </span>
                            )}
                          </span>
                        )}
                      </div>
                    </div>
                    
                    {/* Error Message */}
                    {uploadingFile.result?.error && (
                      <p className="text-xs text-red-600 mt-1">
                        {uploadingFile.result.error}
                      </p>
                    )}
                  </div>
                  
                  {/* Remove Button */}
                  {!isUploading && (
                    <button
                      onClick={() => removeUploadingFile(index)}
                      className="p-1 text-gray-400 hover:text-red-600 transition-colors"
                    >
                      <X className="h-4 w-4" />
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

// Simple image upload button component
interface ImageUploadButtonProps {
  onUpload: (result: ImageUploadResult) => void
  options?: ImageUploadOptions
  disabled?: boolean
  className?: string
  children?: React.ReactNode
}

export function ImageUploadButton({
  onUpload,
  options = {},
  disabled = false,
  className = '',
  children
}: ImageUploadButtonProps) {
  const [uploading, setUploading] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleFileSelect = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) return

    setUploading(true)
    try {
      const result = await ImagesService.uploadMockImage(file)
      onUpload(result)
    } catch (error) {
      console.error('Upload error:', error)
      onUpload({ success: false, error: 'Upload failed' })
    } finally {
      setUploading(false)
      // Reset input
      if (fileInputRef.current) {
        fileInputRef.current.value = ''
      }
    }
  }

  return (
    <>
      <input
        ref={fileInputRef}
        type="file"
        accept="image/*"
        onChange={handleFileSelect}
        className="hidden"
        disabled={disabled || uploading}
      />
      <button
        type="button"
        onClick={() => fileInputRef.current?.click()}
        disabled={disabled || uploading}
        className={`flex items-center justify-center ${className}`}
      >
        {uploading ? (
          <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-current mr-2"></div>
        ) : (
          <ImageIcon className="h-4 w-4 mr-2" />
        )}
        {children || (uploading ? 'Uploading...' : 'Upload Image')}
      </button>
    </>
  )
}
