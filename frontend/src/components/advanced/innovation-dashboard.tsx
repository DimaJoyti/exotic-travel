'use client'

import React, { useState, useEffect } from 'react'
import { 
  Zap, Brain, Mic, Camera, Coins, Smartphone, Cloud, TrendingUp, 
  Shield, Award, MapPin, Thermometer, Wind, Users, Clock, Star,
  ChevronRight, Play, Pause, Settings, RefreshCw, Eye
} from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { Button } from '@/components/ui/button'
import { FadeIn, StaggerContainer, StaggerItem, HoverAnimation, ScaleIn } from '@/components/ui/animated'
import ARDestinationPreview from '@/components/ar/ar-destination-preview'
import VoiceSearch from '@/components/search/voice-search'
import { BlockchainLoyaltyService, LoyaltyToken, NFTReward } from '@/lib/blockchain-loyalty'
import { IoTIntegrationService, SmartRecommendation, WeatherData, AirQualityData } from '@/lib/iot-integration'
import { AdvancedBookingService, DynamicPricing } from '@/lib/advanced-booking'
import { useAuth } from '@/contexts/auth-context'
import { Destination } from '@/types'

interface InnovationDashboardProps {
  className?: string
}

interface FeatureCard {
  id: string
  title: string
  description: string
  icon: React.ReactNode
  status: 'active' | 'demo' | 'coming_soon'
  category: 'ar_vr' | 'voice' | 'blockchain' | 'iot' | 'ai'
  component?: React.ReactNode
}

export default function InnovationDashboard({ className = '' }: InnovationDashboardProps) {
  const { user } = useAuth()
  const [activeFeature, setActiveFeature] = useState<string | null>(null)
  const [loyaltyTokens, setLoyaltyTokens] = useState<LoyaltyToken[]>([])
  const [nftRewards, setNftRewards] = useState<NFTReward[]>([])
  const [smartRecommendations, setSmartRecommendations] = useState<SmartRecommendation[]>([])
  const [weatherData, setWeatherData] = useState<WeatherData | null>(null)
  const [airQualityData, setAirQualityData] = useState<AirQualityData | null>(null)
  const [dynamicPricing, setDynamicPricing] = useState<DynamicPricing | null>(null)
  const [isInitializing, setIsInitializing] = useState(true)

  // Mock destination for demos
  const mockDestination: Destination = {
    id: 1,
    name: 'Santorini, Greece',
    description: 'Beautiful Greek island with stunning sunsets and white-washed buildings',
    country: 'Greece',
    city: 'Santorini',
    price: 2500,
    duration: 7,
    max_guests: 4,
    images: ['/images/santorini.jpg'],
    features: ['Beach Access', 'Cultural Sites', 'Romantic Setting', 'Photography Tours'],
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z'
  }

  useEffect(() => {
    initializeAdvancedFeatures()
  }, [user])

  const initializeAdvancedFeatures = async () => {
    setIsInitializing(true)
    
    try {
      // Initialize IoT service
      await IoTIntegrationService.initialize()
      
      // Load user's blockchain assets
      if (user) {
        const tokens = await BlockchainLoyaltyService.getUserTokens?.(user.id.toString()) || []
        const nfts = await BlockchainLoyaltyService.getUserNFTs(user.id.toString())
        setLoyaltyTokens(tokens)
        setNftRewards(nfts)
      }
      
      // Get smart recommendations
      const recommendations = await IoTIntegrationService.generateSmartRecommendations('1')
      setSmartRecommendations(recommendations)
      
      // Get weather data
      const weather = await IoTIntegrationService.getWeatherData('1')
      setWeatherData(weather)
      
      // Get air quality data
      const airQuality = await IoTIntegrationService.getAirQualityData(36.3932, 25.4615)
      setAirQualityData(airQuality)
      
      // Get dynamic pricing
      const pricing = await AdvancedBookingService.getDynamicPricing(
        '1',
        new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
        new Date(Date.now() + 37 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
        2
      )
      setDynamicPricing(pricing)
      
    } catch (error) {
      console.error('Error initializing advanced features:', error)
    } finally {
      setIsInitializing(false)
    }
  }

  const features: FeatureCard[] = [
    {
      id: 'ar_preview',
      title: 'AR Destination Preview',
      description: 'Experience destinations in augmented reality before you book',
      icon: <Camera className="h-6 w-6" />,
      status: 'demo',
      category: 'ar_vr',
      component: (
        <ARDestinationPreview
          destination={mockDestination}
          onClose={() => setActiveFeature(null)}
        />
      )
    },
    {
      id: 'voice_search',
      title: 'Voice-Powered Search',
      description: 'Search and book destinations using natural voice commands',
      icon: <Mic className="h-6 w-6" />,
      status: 'active',
      category: 'voice',
      component: (
        <div className="p-6 bg-white rounded-lg">
          <VoiceSearch
            onSearch={(query) => console.log('Voice search:', query)}
            onVoiceCommand={(command) => console.log('Voice command:', command)}
          />
        </div>
      )
    },
    {
      id: 'blockchain_loyalty',
      title: 'Blockchain Loyalty Program',
      description: 'Earn crypto tokens and NFT rewards for your travels',
      icon: <Coins className="h-6 w-6" />,
      status: 'active',
      category: 'blockchain'
    },
    {
      id: 'iot_integration',
      title: 'Smart Travel Insights',
      description: 'Real-time data from IoT sensors for optimal travel decisions',
      icon: <Smartphone className="h-6 w-6" />,
      status: 'active',
      category: 'iot'
    },
    {
      id: 'dynamic_pricing',
      title: 'AI Dynamic Pricing',
      description: 'Smart pricing that adapts to demand, weather, and market conditions',
      icon: <TrendingUp className="h-6 w-6" />,
      status: 'active',
      category: 'ai'
    },
    {
      id: 'smart_contracts',
      title: 'Smart Contract Booking',
      description: 'Automated booking protection with blockchain smart contracts',
      icon: <Shield className="h-6 w-6" />,
      status: 'demo',
      category: 'blockchain'
    }
  ]

  const handleFeatureClick = (featureId: string) => {
    setActiveFeature(activeFeature === featureId ? null : featureId)
  }

  const handleVoiceSearch = (query: string) => {
    console.log('Voice search query:', query)
    // Implement voice search logic
  }

  const handleVoiceCommand = (command: any) => {
    console.log('Voice command:', command)
    // Implement voice command logic
  }

  const formatCurrency = (amount: number): string => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD'
    }).format(amount)
  }

  if (isInitializing) {
    return (
      <div className={`flex items-center justify-center py-12 ${className}`}>
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-brand-500 mx-auto mb-4"></div>
          <p className="text-gray-600">Initializing advanced features...</p>
        </div>
      </div>
    )
  }

  return (
    <div className={`space-y-8 ${className}`}>
      {/* Header */}
      <FadeIn>
        <div className="text-center">
          <div className="flex items-center justify-center space-x-2 mb-4">
            <Zap className="h-8 w-8 text-brand-500" />
            <h1 className="text-4xl font-bold text-gray-900">Innovation Hub</h1>
          </div>
          <p className="text-xl text-gray-600 max-w-3xl mx-auto">
            Experience the future of travel with cutting-edge AR/VR, voice AI, blockchain technology, 
            and IoT integration for the ultimate smart travel experience.
          </p>
        </div>
      </FadeIn>

      {/* Feature Categories */}
      <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-8">
        {['ar_vr', 'voice', 'blockchain', 'iot', 'ai'].map((category) => {
          const categoryFeatures = features.filter(f => f.category === category)
          const activeCount = categoryFeatures.filter(f => f.status === 'active').length
          
          return (
            <div key={category} className="text-center p-4 bg-white rounded-lg border border-gray-200">
              <div className="text-2xl font-bold text-brand-500">{activeCount}</div>
              <div className="text-sm text-gray-600 capitalize">{category.replace('_', ' & ')}</div>
            </div>
          )
        })}
      </div>

      {/* Advanced Features Grid */}
      <StaggerContainer staggerDelay={0.1}>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {features.map((feature) => (
            <StaggerItem key={feature.id}>
              <HoverAnimation hoverY={-8} hoverScale={1.02}>
                <motion.div
                  className={`
                    bg-white rounded-xl p-6 border-2 cursor-pointer transition-all duration-300
                    ${activeFeature === feature.id 
                      ? 'border-brand-500 shadow-lg' 
                      : 'border-gray-200 hover:border-brand-300 hover:shadow-md'
                    }
                  `}
                  onClick={() => handleFeatureClick(feature.id)}
                >
                  <div className="flex items-center justify-between mb-4">
                    <div className={`
                      p-3 rounded-lg
                      ${feature.status === 'active' ? 'bg-green-100 text-green-600' :
                        feature.status === 'demo' ? 'bg-blue-100 text-blue-600' :
                        'bg-gray-100 text-gray-600'
                      }
                    `}>
                      {feature.icon}
                    </div>
                    <span className={`
                      text-xs px-2 py-1 rounded-full font-medium
                      ${feature.status === 'active' ? 'bg-green-100 text-green-700' :
                        feature.status === 'demo' ? 'bg-blue-100 text-blue-700' :
                        'bg-gray-100 text-gray-700'
                      }
                    `}>
                      {feature.status.replace('_', ' ')}
                    </span>
                  </div>
                  
                  <h3 className="text-lg font-semibold text-gray-900 mb-2">{feature.title}</h3>
                  <p className="text-gray-600 text-sm mb-4">{feature.description}</p>
                  
                  <div className="flex items-center justify-between">
                    <span className="text-xs text-gray-500 capitalize">
                      {feature.category.replace('_', ' & ')}
                    </span>
                    <ChevronRight className={`
                      h-4 w-4 transition-transform
                      ${activeFeature === feature.id ? 'rotate-90' : ''}
                    `} />
                  </div>
                </motion.div>
              </HoverAnimation>
            </StaggerItem>
          ))}
        </div>
      </StaggerContainer>

      {/* Active Feature Display */}
      <AnimatePresence>
        {activeFeature && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            className="overflow-hidden"
          >
            <div className="bg-gray-50 rounded-xl p-6">
              {features.find(f => f.id === activeFeature)?.component || (
                <div className="text-center py-8">
                  <div className="text-gray-500 mb-4">
                    {features.find(f => f.id === activeFeature)?.icon}
                  </div>
                  <h3 className="text-lg font-semibold text-gray-900 mb-2">
                    {features.find(f => f.id === activeFeature)?.title}
                  </h3>
                  <p className="text-gray-600">
                    {features.find(f => f.id === activeFeature)?.description}
                  </p>
                </div>
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Real-time Data Dashboard */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* IoT Data */}
        <div className="bg-white rounded-xl p-6 border border-gray-200">
          <div className="flex items-center space-x-2 mb-4">
            <Cloud className="h-5 w-5 text-blue-500" />
            <h3 className="text-lg font-semibold text-gray-900">Live IoT Data</h3>
          </div>
          
          <div className="space-y-4">
            {weatherData && (
              <div className="flex items-center justify-between p-3 bg-blue-50 rounded-lg">
                <div className="flex items-center space-x-2">
                  <Thermometer className="h-4 w-4 text-blue-600" />
                  <span className="text-sm font-medium text-blue-900">Temperature</span>
                </div>
                <span className="text-blue-700 font-semibold">
                  {Math.round(weatherData.data.temperature)}°C
                </span>
              </div>
            )}
            
            {airQualityData && (
              <div className="flex items-center justify-between p-3 bg-green-50 rounded-lg">
                <div className="flex items-center space-x-2">
                  <Wind className="h-4 w-4 text-green-600" />
                  <span className="text-sm font-medium text-green-900">Air Quality</span>
                </div>
                <span className="text-green-700 font-semibold">
                  AQI {airQualityData.data.aqi}
                </span>
              </div>
            )}
          </div>
        </div>

        {/* Dynamic Pricing */}
        <div className="bg-white rounded-xl p-6 border border-gray-200">
          <div className="flex items-center space-x-2 mb-4">
            <TrendingUp className="h-5 w-5 text-green-500" />
            <h3 className="text-lg font-semibold text-gray-900">Smart Pricing</h3>
          </div>
          
          {dynamicPricing && (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-gray-600">Current Price</span>
                <span className="text-2xl font-bold text-gray-900">
                  {formatCurrency(dynamicPricing.current_price)}
                </span>
              </div>
              
              <div className="flex items-center justify-between text-sm">
                <span className="text-gray-600">Base Price</span>
                <span className="text-gray-500">
                  {formatCurrency(dynamicPricing.base_price)}
                </span>
              </div>
              
              <div className="flex items-center justify-between text-sm">
                <span className="text-gray-600">Price Trend</span>
                <span className={`
                  font-medium capitalize
                  ${dynamicPricing.predicted_price_trend === 'increasing' ? 'text-red-600' :
                    dynamicPricing.predicted_price_trend === 'decreasing' ? 'text-green-600' :
                    'text-gray-600'
                  }
                `}>
                  {dynamicPricing.predicted_price_trend}
                </span>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Blockchain Assets */}
      {user && (
        <div className="bg-white rounded-xl p-6 border border-gray-200">
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center space-x-2">
              <Coins className="h-5 w-5 text-yellow-500" />
              <h3 className="text-lg font-semibold text-gray-900">Your Blockchain Assets</h3>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={() => BlockchainLoyaltyService.connectWallet()}
            >
              Connect Wallet
            </Button>
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Loyalty Tokens */}
            <div>
              <h4 className="font-medium text-gray-900 mb-3">Loyalty Tokens</h4>
              <div className="space-y-2">
                {loyaltyTokens.length > 0 ? loyaltyTokens.map((token) => (
                  <div key={token.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                    <span className="text-sm font-medium text-gray-900">{token.metadata.name}</span>
                    <span className="text-sm text-gray-600">{token.amount}</span>
                  </div>
                )) : (
                  <p className="text-gray-500 text-sm">No tokens yet. Start traveling to earn rewards!</p>
                )}
              </div>
            </div>
            
            {/* NFT Rewards */}
            <div>
              <h4 className="font-medium text-gray-900 mb-3">NFT Achievements</h4>
              <div className="space-y-2">
                {nftRewards.length > 0 ? nftRewards.map((nft) => (
                  <div key={nft.token_id} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                    <span className="text-sm font-medium text-gray-900">{nft.name}</span>
                    <Award className="h-4 w-4 text-yellow-500" />
                  </div>
                )) : (
                  <p className="text-gray-500 text-sm">Complete bookings to unlock NFT achievements!</p>
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Smart Recommendations */}
      {smartRecommendations.length > 0 && (
        <div className="bg-white rounded-xl p-6 border border-gray-200">
          <div className="flex items-center space-x-2 mb-4">
            <Brain className="h-5 w-5 text-purple-500" />
            <h3 className="text-lg font-semibold text-gray-900">Smart Recommendations</h3>
          </div>
          
          <div className="space-y-3">
            {smartRecommendations.slice(0, 3).map((recommendation) => (
              <div
                key={recommendation.id}
                className={`
                  p-4 rounded-lg border-l-4
                  ${recommendation.priority === 'critical' ? 'border-red-500 bg-red-50' :
                    recommendation.priority === 'high' ? 'border-orange-500 bg-orange-50' :
                    recommendation.priority === 'medium' ? 'border-yellow-500 bg-yellow-50' :
                    'border-blue-500 bg-blue-50'
                  }
                `}
              >
                <h4 className="font-medium text-gray-900 mb-1">{recommendation.title}</h4>
                <p className="text-sm text-gray-700 mb-2">{recommendation.message}</p>
                {recommendation.suggested_actions.length > 0 && (
                  <div className="flex flex-wrap gap-2">
                    {recommendation.suggested_actions.map((action, index) => (
                      <span
                        key={index}
                        className="text-xs bg-white px-2 py-1 rounded-full text-gray-600"
                      >
                        {action}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
