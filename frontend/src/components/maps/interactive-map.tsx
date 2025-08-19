'use client'

import React, { useRef, useEffect, useState, useCallback } from 'react'
import mapboxgl from 'mapbox-gl'
import { motion, AnimatePresence } from 'framer-motion'
import { MapPin, Navigation, ZoomIn, ZoomOut, Layers, X } from 'lucide-react'
import { Destination } from '@/types'
import { Button } from '@/components/ui/button'
import { FadeIn, ScaleIn } from '@/components/ui/animated'

// Mapbox CSS
import 'mapbox-gl/dist/mapbox-gl.css'

interface InteractiveMapProps {
  destinations: Destination[]
  selectedDestination?: Destination | null
  onDestinationSelect?: (destination: Destination) => void
  onDestinationHover?: (destination: Destination | null) => void
  center?: [number, number]
  zoom?: number
  height?: string
  showControls?: boolean
  showClustering?: boolean
  className?: string
}

interface MapMarker {
  destination: Destination
  marker: mapboxgl.Marker
  popup: mapboxgl.Popup
}

// Set your Mapbox access token here
mapboxgl.accessToken = process.env.NEXT_PUBLIC_MAPBOX_ACCESS_TOKEN || 'pk.eyJ1IjoiZXhhbXBsZSIsImEiOiJjbGV4YW1wbGUifQ.example'

export default function InteractiveMap({
  destinations,
  selectedDestination,
  onDestinationSelect,
  onDestinationHover,
  center = [0, 20], // Default center
  zoom = 2,
  height = '400px',
  showControls = true,
  showClustering = true,
  className = ''
}: InteractiveMapProps) {
  const mapContainer = useRef<HTMLDivElement>(null)
  const map = useRef<mapboxgl.Map | null>(null)
  const [mapLoaded, setMapLoaded] = useState(false)
  const [markers, setMarkers] = useState<MapMarker[]>([])
  const [hoveredDestination, setHoveredDestination] = useState<Destination | null>(null)
  const [mapStyle, setMapStyle] = useState<'streets' | 'satellite' | 'outdoors'>('streets')

  // Initialize map
  useEffect(() => {
    if (!mapContainer.current || map.current) return

    map.current = new mapboxgl.Map({
      container: mapContainer.current,
      style: 'mapbox://styles/mapbox/streets-v12',
      center: center,
      zoom: zoom,
      attributionControl: false
    })

    map.current.on('load', () => {
      setMapLoaded(true)
      
      if (showClustering) {
        setupClustering()
      }
    })

    // Add navigation controls
    if (showControls) {
      map.current.addControl(new mapboxgl.NavigationControl(), 'top-right')
    }

    return () => {
      if (map.current) {
        map.current.remove()
        map.current = null
      }
    }
  }, [])

  // Setup clustering for destinations
  const setupClustering = useCallback(() => {
    if (!map.current || !mapLoaded) return

    // Add destination data source
    map.current.addSource('destinations', {
      type: 'geojson',
      data: {
        type: 'FeatureCollection',
        features: destinations.map(destination => ({
          type: 'Feature',
          properties: {
            id: destination.id,
            name: destination.name,
            price: destination.price,
            rating: 4.8, // Mock rating
            image: destination.images[0]
          },
          geometry: {
            type: 'Point',
            coordinates: [0, 0] // Default coordinates - would be fetched from geocoding service
          }
        }))
      },
      cluster: true,
      clusterMaxZoom: 14,
      clusterRadius: 50
    })

    // Add cluster circles
    map.current.addLayer({
      id: 'clusters',
      type: 'circle',
      source: 'destinations',
      filter: ['has', 'point_count'],
      paint: {
        'circle-color': [
          'step',
          ['get', 'point_count'],
          '#3B82F6', // brand-500
          100,
          '#1D4ED8', // brand-700
          750,
          '#1E3A8A'  // brand-900
        ],
        'circle-radius': [
          'step',
          ['get', 'point_count'],
          20,
          100,
          30,
          750,
          40
        ]
      }
    })

    // Add cluster count labels
    map.current.addLayer({
      id: 'cluster-count',
      type: 'symbol',
      source: 'destinations',
      filter: ['has', 'point_count'],
      layout: {
        'text-field': '{point_count_abbreviated}',
        'text-font': ['DIN Offc Pro Medium', 'Arial Unicode MS Bold'],
        'text-size': 12
      },
      paint: {
        'text-color': '#ffffff'
      }
    })

    // Add individual destination points
    map.current.addLayer({
      id: 'unclustered-point',
      type: 'circle',
      source: 'destinations',
      filter: ['!', ['has', 'point_count']],
      paint: {
        'circle-color': '#14B8A6', // accent-500
        'circle-radius': 8,
        'circle-stroke-width': 2,
        'circle-stroke-color': '#ffffff'
      }
    })

    // Click handlers
    map.current.on('click', 'clusters', (e) => {
      const features = map.current!.queryRenderedFeatures(e.point, {
        layers: ['clusters']
      })
      const clusterIdValue = features[0].properties!.cluster_id as number

      (map.current!.getSource('destinations') as any)!.getClusterExpansionZoom(
        clusterIdValue,
        (err: any, zoom: number) => {
          if (err) return
          
          map.current!.easeTo({
            center: (features[0].geometry as any).coordinates,
            zoom: zoom
          })
        }
      )
    })

    map.current.on('click', 'unclustered-point', (e) => {
      const feature = e.features![0]
      const destinationId = feature.properties!.id
      const destination = destinations.find(d => d.id === destinationId)
      
      if (destination && onDestinationSelect) {
        onDestinationSelect(destination)
      }
    })

    // Hover effects
    map.current.on('mouseenter', 'unclustered-point', (e) => {
      map.current!.getCanvas().style.cursor = 'pointer'
      const feature = e.features![0]
      const destinationId = feature.properties!.id
      const destination = destinations.find(d => d.id === destinationId)
      
      if (destination) {
        setHoveredDestination(destination)
        onDestinationHover?.(destination)
      }
    })

    map.current.on('mouseleave', 'unclustered-point', () => {
      map.current!.getCanvas().style.cursor = ''
      setHoveredDestination(null)
      onDestinationHover?.(null)
    })

  }, [destinations, mapLoaded, onDestinationSelect, onDestinationHover])

  // Update clustering when destinations change
  useEffect(() => {
    if (mapLoaded && showClustering) {
      setupClustering()
    }
  }, [destinations, mapLoaded, setupClustering, showClustering])

  // Handle selected destination
  useEffect(() => {
    if (!map.current || !selectedDestination) return

    // Fly to selected destination
    map.current.flyTo({
      center: [0, 0], // Default coordinates - would be fetched from geocoding service
      zoom: 12,
      duration: 2000
    })
  }, [selectedDestination])

  // Change map style
  const changeMapStyle = (style: 'streets' | 'satellite' | 'outdoors') => {
    if (!map.current) return

    const styleUrls = {
      streets: 'mapbox://styles/mapbox/streets-v12',
      satellite: 'mapbox://styles/mapbox/satellite-streets-v12',
      outdoors: 'mapbox://styles/mapbox/outdoors-v12'
    }

    map.current.setStyle(styleUrls[style])
    setMapStyle(style)

    // Re-add layers after style change
    map.current.once('styledata', () => {
      if (showClustering) {
        setupClustering()
      }
    })
  }

  // Zoom controls
  const zoomIn = () => map.current?.zoomIn()
  const zoomOut = () => map.current?.zoomOut()

  // Fit bounds to show all destinations
  const fitBounds = () => {
    if (!map.current || destinations.length === 0) return

    const bounds = new mapboxgl.LngLatBounds()
    // In a real implementation, destinations would have coordinates
    // For now, we'll use a default bounds
    bounds.extend([-180, -85])
    bounds.extend([180, 85])

    map.current.fitBounds(bounds, { padding: 50 })
  }

  return (
    <div className={`relative ${className}`} style={{ height }}>
      {/* Map Container */}
      <div ref={mapContainer} className="w-full h-full rounded-lg overflow-hidden" />

      {/* Custom Controls */}
      {showControls && (
        <div className="absolute top-4 left-4 space-y-2">
          {/* Style Switcher */}
          <FadeIn delay={0.2}>
            <div className="bg-white rounded-lg shadow-lg p-2">
              <div className="flex space-x-1">
                {(['streets', 'satellite', 'outdoors'] as const).map((style) => (
                  <Button
                    key={style}
                    variant={mapStyle === style ? 'primary' : 'ghost'}
                    size="sm"
                    onClick={() => changeMapStyle(style)}
                    className="text-xs capitalize"
                  >
                    {style}
                  </Button>
                ))}
              </div>
            </div>
          </FadeIn>

          {/* Zoom Controls */}
          <FadeIn delay={0.3}>
            <div className="bg-white rounded-lg shadow-lg p-1">
              <div className="flex flex-col space-y-1">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={zoomIn}
                  className="p-2"
                >
                  <ZoomIn className="h-4 w-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={zoomOut}
                  className="p-2"
                >
                  <ZoomOut className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </FadeIn>

          {/* Fit Bounds */}
          <FadeIn delay={0.4}>
            <Button
              variant="ghost"
              size="sm"
              onClick={fitBounds}
              className="bg-white shadow-lg p-2"
              title="Show all destinations"
            >
              <Navigation className="h-4 w-4" />
            </Button>
          </FadeIn>
        </div>
      )}

      {/* Hovered Destination Info */}
      <AnimatePresence>
        {hoveredDestination && (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: 20 }}
            className="absolute bottom-4 left-4 bg-white rounded-lg shadow-xl p-4 max-w-sm"
          >
            <div className="flex items-center space-x-3">
              <div className="w-12 h-12 bg-gray-200 rounded-lg overflow-hidden flex-shrink-0">
                {hoveredDestination.images[0] && (
                  <img
                    src={hoveredDestination.images[0]}
                    alt={hoveredDestination.name}
                    className="w-full h-full object-cover"
                  />
                )}
              </div>
              <div>
                <h4 className="font-semibold text-gray-900">{hoveredDestination.name}</h4>
                <p className="text-sm text-gray-500">{hoveredDestination.city}, {hoveredDestination.country}</p>
                <p className="text-sm font-medium text-brand-600">${hoveredDestination.price}</p>
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Loading State */}
      {!mapLoaded && (
        <div className="absolute inset-0 bg-gray-100 rounded-lg flex items-center justify-center">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-brand-500 mx-auto mb-2"></div>
            <p className="text-gray-600">Loading map...</p>
          </div>
        </div>
      )}
    </div>
  )
}
