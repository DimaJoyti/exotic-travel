import React from 'react'
import Link from 'next/link'
import { Brain, Sparkles, Target, Zap } from 'lucide-react'
import { FadeIn } from '@/components/ui/animated'

// interface RecommendationCategory {
//   id: string
//   name: string
//   description: string
//   icon: React.ReactNode
//   recommendations: RecommendationResult[]
// }

export default function RecommendationsPage() {

  // Placeholder implementation

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Hero Section */}
      <div className="bg-gradient-to-r from-brand-500 to-accent-500 text-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
          <FadeIn>
            <div className="text-center">
              <div className="flex items-center justify-center space-x-2 mb-4">
                <Sparkles className="h-8 w-8" />
                <h1 className="text-4xl md:text-5xl font-bold">AI-Powered Recommendations</h1>
              </div>
              <p className="text-xl text-white/90 max-w-3xl mx-auto mb-8">
                Discover your perfect destinations with our intelligent recommendation engine,
                personalized just for you based on your preferences and travel behavior.
              </p>
              <div className="flex items-center justify-center space-x-6 text-sm">
                <div className="flex items-center space-x-2">
                  <Target className="h-5 w-5" />
                  <span>95% Accuracy</span>
                </div>
                <div className="flex items-center space-x-2">
                  <Zap className="h-5 w-5" />
                  <span>Real-time Updates</span>
                </div>
                <div className="flex items-center space-x-2">
                  <Brain className="h-5 w-5" />
                  <span>Machine Learning</span>
                </div>
              </div>
            </div>
          </FadeIn>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <div className="text-center">
          <Brain className="h-16 w-16 text-gray-400 mx-auto mb-4" />
          <h2 className="text-2xl font-bold text-gray-900 mb-2">AI Recommendations Coming Soon</h2>
          <p className="text-gray-600 mb-6">
            We're building an intelligent recommendation system that will provide personalized destination suggestions.
          </p>
          <Link
            href="/destinations"
            className="inline-flex items-center px-6 py-3 bg-brand-500 text-white font-medium rounded-lg hover:bg-brand-600 transition-colors"
          >
            Browse Destinations
          </Link>
        </div>
      </div>
    </div>
  )
}
