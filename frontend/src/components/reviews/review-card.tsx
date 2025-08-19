'use client'

import { useState } from 'react'
import { 
  ThumbsUp, 
  Flag, 
  Calendar, 
  CheckCircle, 
  MoreHorizontal,
  Edit,
  Trash2
} from 'lucide-react'
import { Review } from '@/types'
import { ReviewsService } from '@/lib/reviews'
import StarRating from './star-rating'

interface ReviewCardProps {
  review: Review
  currentUserId?: number
  onReviewUpdate?: (review: Review) => void
  onReviewDelete?: (reviewId: number) => void
  className?: string
}

export default function ReviewCard({
  review,
  currentUserId,
  onReviewUpdate,
  onReviewDelete,
  className = ''
}: ReviewCardProps) {
  const [isHelpful, setIsHelpful] = useState(false)
  const [helpfulCount, setHelpfulCount] = useState(review.helpful_count)
  const [showActions, setShowActions] = useState(false)
  const [loading, setLoading] = useState(false)

  const isOwnReview = currentUserId === review.user_id

  const handleMarkHelpful = async () => {
    if (loading || isOwnReview) return

    setLoading(true)
    try {
      await ReviewsService.markMockReviewHelpful(review.id)
      setIsHelpful(!isHelpful)
      setHelpfulCount(prev => isHelpful ? prev - 1 : prev + 1)
    } catch (error) {
      console.error('Error marking review as helpful:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleReport = async () => {
    if (loading || isOwnReview) return

    const reason = prompt('Please provide a reason for reporting this review:')
    if (!reason) return

    setLoading(true)
    try {
      await ReviewsService.reportReview(review.id, reason)
      alert('Review reported successfully. Thank you for your feedback.')
    } catch (error) {
      console.error('Error reporting review:', error)
      alert('Failed to report review. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  const handleEdit = () => {
    // This would open an edit modal or navigate to edit page
    console.log('Edit review:', review.id)
  }

  const handleDelete = async () => {
    if (!confirm('Are you sure you want to delete this review?')) return

    setLoading(true)
    try {
      await ReviewsService.deleteReview(review.id)
      onReviewDelete?.(review.id)
    } catch (error) {
      console.error('Error deleting review:', error)
      alert('Failed to delete review. Please try again.')
    } finally {
      setLoading(false)
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    })
  }

  const formatTravelDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short'
    })
  }

  return (
    <div className={`bg-white rounded-lg border border-gray-200 p-6 ${className}`}>
      {/* Header */}
      <div className="flex items-start justify-between mb-4">
        <div className="flex items-start space-x-4">
          {/* User Avatar */}
          <div className="h-10 w-10 rounded-full bg-primary flex items-center justify-center flex-shrink-0">
            <span className="text-sm font-medium text-primary-foreground">
              {review.user?.first_name?.charAt(0)}{review.user?.last_name?.charAt(0)}
            </span>
          </div>
          
          {/* User Info and Rating */}
          <div className="flex-1">
            <div className="flex items-center space-x-2 mb-1">
              <h4 className="font-semibold text-gray-900">
                {review.user?.first_name} {review.user?.last_name}
              </h4>
              {review.verified_booking && (
                <div className="flex items-center text-green-600">
                  <CheckCircle className="h-4 w-4 mr-1" />
                  <span className="text-xs font-medium">Verified Stay</span>
                </div>
              )}
            </div>
            
            <div className="flex items-center space-x-3 text-sm text-gray-600">
              <StarRating rating={review.rating} size="sm" />
              <span>•</span>
              <div className="flex items-center">
                <Calendar className="h-4 w-4 mr-1" />
                <span>Traveled {formatTravelDate(review.travel_date)}</span>
              </div>
            </div>
          </div>
        </div>

        {/* Actions Menu */}
        <div className="relative">
          <button
            onClick={() => setShowActions(!showActions)}
            className="p-1 text-gray-400 hover:text-gray-600 transition-colors"
          >
            <MoreHorizontal className="h-5 w-5" />
          </button>
          
          {showActions && (
            <div className="absolute right-0 top-8 bg-white border border-gray-200 rounded-lg shadow-lg py-1 z-10 min-w-[120px]">
              {isOwnReview ? (
                <>
                  <button
                    onClick={handleEdit}
                    className="flex items-center w-full px-3 py-2 text-sm text-gray-700 hover:bg-gray-50"
                  >
                    <Edit className="h-4 w-4 mr-2" />
                    Edit
                  </button>
                  <button
                    onClick={handleDelete}
                    className="flex items-center w-full px-3 py-2 text-sm text-red-600 hover:bg-gray-50"
                  >
                    <Trash2 className="h-4 w-4 mr-2" />
                    Delete
                  </button>
                </>
              ) : (
                <button
                  onClick={handleReport}
                  className="flex items-center w-full px-3 py-2 text-sm text-gray-700 hover:bg-gray-50"
                >
                  <Flag className="h-4 w-4 mr-2" />
                  Report
                </button>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Review Title */}
      {review.title && (
        <h3 className="text-lg font-semibold text-gray-900 mb-3">
          {review.title}
        </h3>
      )}

      {/* Review Content */}
      <div className="space-y-4 mb-4">
        {/* Comment */}
        <p className="text-gray-700 leading-relaxed">
          {review.comment}
        </p>

        {/* Pros and Cons */}
        {(review.pros?.length || review.cons?.length) && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {review.pros && review.pros.length > 0 && (
              <div>
                <h5 className="text-sm font-semibold text-green-700 mb-2">👍 Pros</h5>
                <ul className="space-y-1">
                  {review.pros.map((pro, index) => (
                    <li key={index} className="text-sm text-gray-600 flex items-start">
                      <span className="text-green-500 mr-2">•</span>
                      {pro}
                    </li>
                  ))}
                </ul>
              </div>
            )}

            {review.cons && review.cons.length > 0 && (
              <div>
                <h5 className="text-sm font-semibold text-red-700 mb-2">👎 Cons</h5>
                <ul className="space-y-1">
                  {review.cons.map((con, index) => (
                    <li key={index} className="text-sm text-gray-600 flex items-start">
                      <span className="text-red-500 mr-2">•</span>
                      {con}
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Footer */}
      <div className="flex items-center justify-between pt-4 border-t border-gray-100">
        <div className="flex items-center space-x-4">
          {/* Helpful Button */}
          {!isOwnReview && (
            <button
              onClick={handleMarkHelpful}
              disabled={loading}
              className={`flex items-center space-x-1 text-sm transition-colors ${
                isHelpful 
                  ? 'text-primary' 
                  : 'text-gray-600 hover:text-primary'
              } disabled:opacity-50`}
            >
              <ThumbsUp className={`h-4 w-4 ${isHelpful ? 'fill-current' : ''}`} />
              <span>Helpful ({helpfulCount})</span>
            </button>
          )}
        </div>

        {/* Review Date */}
        <span className="text-sm text-gray-500">
          {formatDate(review.created_at)}
        </span>
      </div>
    </div>
  )
}

// Component for displaying review summary
interface ReviewSummaryProps {
  reviews: Review[]
  className?: string
}

export function ReviewSummary({ reviews, className = '' }: ReviewSummaryProps) {
  if (reviews.length === 0) {
    return (
      <div className={`text-center py-8 ${className}`}>
        <p className="text-gray-600">No reviews yet. Be the first to share your experience!</p>
      </div>
    )
  }

  const averageRating = ReviewsService.calculateAverageRating(reviews)
  const distribution = ReviewsService.getRatingDistribution(reviews)

  return (
    <div className={`bg-gray-50 rounded-lg p-6 ${className}`}>
      <div className="text-center mb-6">
        <div className="text-4xl font-bold text-gray-900 mb-2">
          {averageRating.toFixed(1)}
        </div>
        <StarRating rating={averageRating} size="lg" className="justify-center mb-2" />
        <p className="text-gray-600">
          Based on {reviews.length} review{reviews.length !== 1 ? 's' : ''}
        </p>
      </div>

      <div className="space-y-2">
        {[5, 4, 3, 2, 1].map((rating) => {
          const count = distribution[rating as keyof typeof distribution]
          const percentage = reviews.length > 0 ? (count / reviews.length) * 100 : 0
          
          return (
            <div key={rating} className="flex items-center space-x-3">
              <span className="text-sm text-gray-600 w-8">{rating} ⭐</span>
              <div className="flex-1 bg-gray-200 rounded-full h-2">
                <div
                  className="bg-yellow-400 h-2 rounded-full transition-all duration-300"
                  style={{ width: `${percentage}%` }}
                />
              </div>
              <span className="text-sm text-gray-600 w-8 text-right">{count}</span>
            </div>
          )
        })}
      </div>
    </div>
  )
}
