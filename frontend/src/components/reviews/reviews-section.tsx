'use client'

import { useState, useEffect } from 'react'
import { ChevronDown, Filter, Plus, MessageSquare } from 'lucide-react'
import { Review, ReviewStats } from '@/types'
import { ReviewsService } from '@/lib/reviews'
import { useAuth } from '@/contexts/auth-context'
import StarRating, { RatingSummary, RatingDistribution } from './star-rating'
import ReviewCard from './review-card'
import ReviewForm from './review-form'

interface ReviewsSectionProps {
  destinationId: number
  className?: string
}

export default function ReviewsSection({ destinationId, className = '' }: ReviewsSectionProps) {
  const { user } = useAuth()
  const [reviews, setReviews] = useState<Review[]>([])
  const [stats, setStats] = useState<ReviewStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [showReviewForm, setShowReviewForm] = useState(false)
  const [sortBy, setSortBy] = useState('newest')
  const [currentPage, setCurrentPage] = useState(1)
  const [totalReviews, setTotalReviews] = useState(0)
  const reviewsPerPage = 5

  useEffect(() => {
    loadReviews()
  }, [destinationId, sortBy, currentPage])

  const loadReviews = async () => {
    try {
      const data = await ReviewsService.getMockDestinationReviews(
        destinationId,
        currentPage,
        reviewsPerPage,
        sortBy
      )
      setReviews(data.reviews)
      setStats(data.stats)
      setTotalReviews(data.total)
    } catch (error) {
      console.error('Error loading reviews:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleReviewSubmit = (newReview: Review) => {
    setReviews(prev => [newReview, ...prev])
    setShowReviewForm(false)
    
    // Update stats
    if (stats) {
      const newTotal = stats.total_reviews + 1
      const newAverage = ((stats.average_rating * stats.total_reviews) + newReview.rating) / newTotal
      
      setStats({
        ...stats,
        average_rating: Math.round(newAverage * 10) / 10,
        total_reviews: newTotal,
        rating_distribution: {
          ...stats.rating_distribution,
          [newReview.rating]: stats.rating_distribution[newReview.rating as keyof typeof stats.rating_distribution] + 1
        }
      })
    }
  }

  const handleReviewDelete = (reviewId: number) => {
    setReviews(prev => prev.filter(review => review.id !== reviewId))
    
    // Update stats
    const deletedReview = reviews.find(r => r.id === reviewId)
    if (stats && deletedReview) {
      const newTotal = stats.total_reviews - 1
      let newAverage = 0
      
      if (newTotal > 0) {
        newAverage = ((stats.average_rating * stats.total_reviews) - deletedReview.rating) / newTotal
      }
      
      setStats({
        ...stats,
        average_rating: Math.round(newAverage * 10) / 10,
        total_reviews: newTotal,
        rating_distribution: {
          ...stats.rating_distribution,
          [deletedReview.rating]: Math.max(0, stats.rating_distribution[deletedReview.rating as keyof typeof stats.rating_distribution] - 1)
        }
      })
    }
  }

  const totalPages = Math.ceil(totalReviews / reviewsPerPage)

  if (loading) {
    return (
      <div className={`${className}`}>
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-gray-200 rounded w-1/4"></div>
          <div className="h-32 bg-gray-200 rounded"></div>
          <div className="space-y-3">
            {[1, 2, 3].map(i => (
              <div key={i} className="h-24 bg-gray-200 rounded"></div>
            ))}
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className={`${className}`}>
      {/* Header */}
      <div className="flex items-center justify-between mb-8">
        <h2 className="text-2xl font-bold text-gray-900">
          Reviews & Ratings
        </h2>
        
        {user && (
          <button
            onClick={() => setShowReviewForm(!showReviewForm)}
            className="flex items-center px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors"
          >
            <Plus className="h-4 w-4 mr-2" />
            Write Review
          </button>
        )}
      </div>

      {/* Review Form */}
      {showReviewForm && (
        <div className="mb-8">
          <ReviewForm
            destinationId={destinationId}
            onReviewSubmit={handleReviewSubmit}
            onCancel={() => setShowReviewForm(false)}
          />
        </div>
      )}

      {/* Reviews Overview */}
      {stats && stats.total_reviews > 0 ? (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 mb-8">
          {/* Rating Summary */}
          <div className="lg:col-span-2">
            <div className="bg-white rounded-lg border border-gray-200 p-6">
              <RatingSummary
                rating={stats.average_rating}
                totalReviews={stats.total_reviews}
                size="lg"
                className="mb-6"
              />
              
              <div className="text-sm text-gray-600">
                <p className="mb-2">
                  <span className="font-medium">{Math.round((stats.rating_distribution[5] + stats.rating_distribution[4]) / stats.total_reviews * 100)}%</span> of guests recommend this destination
                </p>
                <p>
                  Most recent reviews mention: excellent service, beautiful location, great value
                </p>
              </div>
            </div>
          </div>

          {/* Rating Distribution */}
          <div className="bg-white rounded-lg border border-gray-200 p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Rating Breakdown</h3>
            <RatingDistribution
              distribution={stats.rating_distribution}
              totalReviews={stats.total_reviews}
            />
          </div>
        </div>
      ) : (
        <div className="bg-gray-50 rounded-lg p-8 text-center mb-8">
          <MessageSquare className="h-16 w-16 text-gray-300 mx-auto mb-4" />
          <h3 className="text-xl font-semibold text-gray-900 mb-2">No reviews yet</h3>
          <p className="text-gray-600 mb-4">
            Be the first to share your experience with this destination!
          </p>
          {user && (
            <button
              onClick={() => setShowReviewForm(true)}
              className="bg-primary text-primary-foreground px-6 py-3 rounded-lg hover:bg-primary/90 transition-colors"
            >
              Write the first review
            </button>
          )}
        </div>
      )}

      {/* Reviews List */}
      {reviews.length > 0 && (
        <div>
          {/* Sort and Filter */}
          <div className="flex items-center justify-between mb-6">
            <h3 className="text-lg font-semibold text-gray-900">
              {totalReviews} Review{totalReviews !== 1 ? 's' : ''}
            </h3>
            
            <div className="flex items-center space-x-4">
              <div className="relative">
                <select
                  value={sortBy}
                  onChange={(e) => {
                    setSortBy(e.target.value)
                    setCurrentPage(1)
                  }}
                  className="appearance-none bg-white border border-gray-300 rounded-lg px-4 py-2 pr-8 focus:ring-2 focus:ring-primary focus:border-transparent"
                >
                  <option value="newest">Newest First</option>
                  <option value="oldest">Oldest First</option>
                  <option value="highest">Highest Rated</option>
                  <option value="lowest">Lowest Rated</option>
                  <option value="helpful">Most Helpful</option>
                </select>
                <ChevronDown className="absolute right-2 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400 pointer-events-none" />
              </div>
            </div>
          </div>

          {/* Reviews */}
          <div className="space-y-6">
            {reviews.map((review) => (
              <ReviewCard
                key={review.id}
                review={review}
                currentUserId={user?.id}
                onReviewDelete={handleReviewDelete}
              />
            ))}
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-center space-x-2 mt-8">
              <button
                onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                disabled={currentPage === 1}
                className="px-3 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                Previous
              </button>
              
              {Array.from({ length: totalPages }, (_, i) => i + 1).map((page) => (
                <button
                  key={page}
                  onClick={() => setCurrentPage(page)}
                  className={`px-3 py-2 rounded-lg transition-colors ${
                    currentPage === page
                      ? 'bg-primary text-primary-foreground'
                      : 'border border-gray-300 text-gray-700 hover:bg-gray-50'
                  }`}
                >
                  {page}
                </button>
              ))}
              
              <button
                onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
                disabled={currentPage === totalPages}
                className="px-3 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                Next
              </button>
            </div>
          )}
        </div>
      )}

      {/* Login Prompt for Non-authenticated Users */}
      {!user && reviews.length > 0 && (
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-6 mt-8">
          <div className="text-center">
            <h3 className="text-lg font-semibold text-blue-900 mb-2">
              Share Your Experience
            </h3>
            <p className="text-blue-700 mb-4">
              Have you visited this destination? Sign in to write a review and help other travelers.
            </p>
            <div className="space-x-3">
              <a
                href="/auth/login"
                className="inline-flex items-center px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors"
              >
                Sign In to Review
              </a>
              <a
                href="/auth/register"
                className="inline-flex items-center px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors"
              >
                Create Account
              </a>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
