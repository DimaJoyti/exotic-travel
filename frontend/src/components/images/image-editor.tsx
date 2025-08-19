'use client'

import { useState, useRef, useEffect } from 'react'
import { 
  RotateCw, 
  RotateCcw, 
  Crop, 
  Sliders, 
  Download, 
  Undo, 
  Redo,
  ZoomIn,
  ZoomOut,
  Move,
  Square,
  Circle
} from 'lucide-react'
import { ImageOptimizer, ImageOptimizationOptions } from '@/lib/image-optimization'

interface ImageEditorProps {
  src: string
  alt: string
  onSave?: (editedImage: Blob) => void
  onCancel?: () => void
  className?: string
}

interface EditState {
  rotation: number
  scale: number
  brightness: number
  contrast: number
  saturation: number
  cropArea?: {
    x: number
    y: number
    width: number
    height: number
  }
}

export default function ImageEditor({
  src,
  alt,
  onSave,
  onCancel,
  className = ''
}: ImageEditorProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const imageRef = useRef<HTMLImageElement>(null)
  const [editState, setEditState] = useState<EditState>({
    rotation: 0,
    scale: 1,
    brightness: 100,
    contrast: 100,
    saturation: 100,
  })
  const [history, setHistory] = useState<EditState[]>([])
  const [historyIndex, setHistoryIndex] = useState(-1)
  const [activeTab, setActiveTab] = useState<'adjust' | 'crop' | 'rotate'>('adjust')
  const [isLoading, setIsLoading] = useState(false)

  useEffect(() => {
    if (imageRef.current) {
      drawImage()
    }
  }, [editState])

  const drawImage = () => {
    const canvas = canvasRef.current
    const img = imageRef.current
    if (!canvas || !img) return

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    // Set canvas size
    canvas.width = img.naturalWidth
    canvas.height = img.naturalHeight

    // Clear canvas
    ctx.clearRect(0, 0, canvas.width, canvas.height)

    // Save context
    ctx.save()

    // Apply transformations
    ctx.translate(canvas.width / 2, canvas.height / 2)
    ctx.rotate((editState.rotation * Math.PI) / 180)
    ctx.scale(editState.scale, editState.scale)

    // Apply filters
    ctx.filter = `brightness(${editState.brightness}%) contrast(${editState.contrast}%) saturate(${editState.saturation}%)`

    // Draw image
    ctx.drawImage(img, -img.naturalWidth / 2, -img.naturalHeight / 2)

    // Restore context
    ctx.restore()

    // Draw crop area if active
    if (editState.cropArea && activeTab === 'crop') {
      ctx.strokeStyle = '#3b82f6'
      ctx.lineWidth = 2
      ctx.setLineDash([5, 5])
      ctx.strokeRect(
        editState.cropArea.x,
        editState.cropArea.y,
        editState.cropArea.width,
        editState.cropArea.height
      )
    }
  }

  const addToHistory = (state: EditState) => {
    const newHistory = history.slice(0, historyIndex + 1)
    newHistory.push({ ...state })
    setHistory(newHistory)
    setHistoryIndex(newHistory.length - 1)
  }

  const updateEditState = (updates: Partial<EditState>) => {
    const newState = { ...editState, ...updates }
    setEditState(newState)
    addToHistory(newState)
  }

  const undo = () => {
    if (historyIndex > 0) {
      setHistoryIndex(historyIndex - 1)
      setEditState(history[historyIndex - 1])
    }
  }

  const redo = () => {
    if (historyIndex < history.length - 1) {
      setHistoryIndex(historyIndex + 1)
      setEditState(history[historyIndex + 1])
    }
  }

  const rotate = (degrees: number) => {
    updateEditState({ rotation: editState.rotation + degrees })
  }

  const handleSliderChange = (property: keyof EditState, value: number) => {
    updateEditState({ [property]: value })
  }

  const resetEdits = () => {
    const initialState: EditState = {
      rotation: 0,
      scale: 1,
      brightness: 100,
      contrast: 100,
      saturation: 100,
    }
    setEditState(initialState)
    setHistory([initialState])
    setHistoryIndex(0)
  }

  const handleSave = async () => {
    if (!canvasRef.current) return

    setIsLoading(true)
    try {
      const blob = await new Promise<Blob>((resolve, reject) => {
        canvasRef.current!.toBlob((blob) => {
          if (blob) resolve(blob)
          else reject(new Error('Failed to create blob'))
        }, 'image/jpeg', 0.9)
      })

      onSave?.(blob)
    } catch (error) {
      console.error('Failed to save image:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const handleDownload = async () => {
    if (!canvasRef.current) return

    const link = document.createElement('a')
    link.download = `edited_${alt || 'image'}.jpg`
    link.href = canvasRef.current.toDataURL('image/jpeg', 0.9)
    link.click()
  }

  return (
    <div className={`bg-white rounded-lg shadow-lg ${className}`}>
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-gray-200">
        <h3 className="text-lg font-semibold text-gray-900">Image Editor</h3>
        <div className="flex space-x-2">
          <button
            onClick={undo}
            disabled={historyIndex <= 0}
            className="p-2 text-gray-600 hover:text-gray-900 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Undo className="h-4 w-4" />
          </button>
          <button
            onClick={redo}
            disabled={historyIndex >= history.length - 1}
            className="p-2 text-gray-600 hover:text-gray-900 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Redo className="h-4 w-4" />
          </button>
        </div>
      </div>

      <div className="flex">
        {/* Sidebar */}
        <div className="w-64 border-r border-gray-200 p-4">
          {/* Tabs */}
          <div className="flex space-x-1 mb-6">
            {[
              { id: 'adjust', label: 'Adjust', icon: Sliders },
              { id: 'rotate', label: 'Rotate', icon: RotateCw },
              { id: 'crop', label: 'Crop', icon: Crop },
            ].map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as any)}
                className={`flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-colors ${
                  activeTab === tab.id
                    ? 'bg-primary text-primary-foreground'
                    : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
                }`}
              >
                <tab.icon className="h-4 w-4 mr-2" />
                {tab.label}
              </button>
            ))}
          </div>

          {/* Controls */}
          <div className="space-y-6">
            {activeTab === 'adjust' && (
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Brightness: {editState.brightness}%
                  </label>
                  <input
                    type="range"
                    min="0"
                    max="200"
                    value={editState.brightness}
                    onChange={(e) => handleSliderChange('brightness', parseInt(e.target.value))}
                    className="w-full"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Contrast: {editState.contrast}%
                  </label>
                  <input
                    type="range"
                    min="0"
                    max="200"
                    value={editState.contrast}
                    onChange={(e) => handleSliderChange('contrast', parseInt(e.target.value))}
                    className="w-full"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Saturation: {editState.saturation}%
                  </label>
                  <input
                    type="range"
                    min="0"
                    max="200"
                    value={editState.saturation}
                    onChange={(e) => handleSliderChange('saturation', parseInt(e.target.value))}
                    className="w-full"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Scale: {(editState.scale * 100).toFixed(0)}%
                  </label>
                  <input
                    type="range"
                    min="0.1"
                    max="3"
                    step="0.1"
                    value={editState.scale}
                    onChange={(e) => handleSliderChange('scale', parseFloat(e.target.value))}
                    className="w-full"
                  />
                </div>
              </div>
            )}

            {activeTab === 'rotate' && (
              <div className="space-y-4">
                <div className="flex space-x-2">
                  <button
                    onClick={() => rotate(-90)}
                    className="flex-1 flex items-center justify-center px-3 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
                  >
                    <RotateCcw className="h-4 w-4 mr-2" />
                    90° Left
                  </button>
                  <button
                    onClick={() => rotate(90)}
                    className="flex-1 flex items-center justify-center px-3 py-2 border border-gray-300 rounded-lg hover:bg-gray-50"
                  >
                    <RotateCw className="h-4 w-4 mr-2" />
                    90° Right
                  </button>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Custom Rotation: {editState.rotation}°
                  </label>
                  <input
                    type="range"
                    min="-180"
                    max="180"
                    value={editState.rotation}
                    onChange={(e) => handleSliderChange('rotation', parseInt(e.target.value))}
                    className="w-full"
                  />
                </div>
              </div>
            )}

            {activeTab === 'crop' && (
              <div className="space-y-4">
                <p className="text-sm text-gray-600">
                  Click and drag on the image to select crop area
                </p>
                <div className="flex space-x-2">
                  <button className="flex-1 flex items-center justify-center px-3 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">
                    <Square className="h-4 w-4 mr-2" />
                    Square
                  </button>
                  <button className="flex-1 flex items-center justify-center px-3 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">
                    <Circle className="h-4 w-4 mr-2" />
                    Circle
                  </button>
                </div>
              </div>
            )}
          </div>

          {/* Reset Button */}
          <button
            onClick={resetEdits}
            className="w-full mt-6 px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
          >
            Reset All
          </button>
        </div>

        {/* Canvas Area */}
        <div className="flex-1 p-4">
          <div className="relative bg-gray-100 rounded-lg overflow-hidden" style={{ height: '500px' }}>
            <img
              ref={imageRef}
              src={src}
              alt={alt}
              className="hidden"
              onLoad={drawImage}
            />
            <canvas
              ref={canvasRef}
              className="max-w-full max-h-full object-contain mx-auto"
              style={{ display: 'block' }}
            />
          </div>

          {/* Action Buttons */}
          <div className="flex justify-between mt-4">
            <div className="flex space-x-2">
              <button
                onClick={handleDownload}
                className="flex items-center px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
              >
                <Download className="h-4 w-4 mr-2" />
                Download
              </button>
            </div>

            <div className="flex space-x-2">
              <button
                onClick={onCancel}
                className="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleSave}
                disabled={isLoading}
                className="flex items-center px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {isLoading ? (
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                ) : null}
                {isLoading ? 'Saving...' : 'Save Changes'}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
