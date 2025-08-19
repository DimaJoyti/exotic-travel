'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Plus, X, Calendar } from 'lucide-react'
import { ReviewsService, CreateReviewData } from '@/lib/reviews'
import { Review } from '@/types'
import StarRating from './star-rating'

const reviewSchema = z.object({
  rating: z.number().min(1, 'Please select a rating').max(5),
  title: z.string().min(1, 'Title is required').max(100, 'Title must be less than 100 characters'),
  comment: z.string().min(10, 'Comment must be at least 10 characters').max(1000, 'Comment must be less than 1000 characters'),
  travel_date: z.string().min(1, 'Travel date is required'),
  pros: z.array(z.string()).optional(),
  cons: z.array(z.string()).optional(),
})

type ReviewFormData = z.infer<typeof reviewSchema>

interface ReviewFormProps {
  destinationId: number
  bookingId?: number
  onReviewSubmit: (review: Review) => void
  onCancel?: () => void
  className?: string
}

export default function ReviewForm({
  destinationId,
  bookingId,
  onReviewSubmit,
  onCancel,
  className = ''
}: ReviewFormProps) {
  const [submitting, setSubmitting] = useState(false)
  const [pros, setPros] = useState<string[]>([])
  const [cons, setCons] = useState<string[]>([])
  const [newPro, setNewPro] = useState('')
  const [newCon, setNewCon] = useState('')

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
  } = useForm<ReviewFormData>({
    resolver: zodResolver(reviewSchema),
    defaultValues: {
      rating: 0,
      title: '',
      comment: '',
      travel_date: '',
      pros: [],
      cons: [],
    },
  })

  const rating = watch('rating')
  const comment = watch('comment')

  const onSubmit = async (data: ReviewFormData) => {
    setSubmitting(true)

    try {
      const reviewData: CreateReviewData = {
        destination_id: destinationId,
        booking_id: bookingId,
        rating: data.rating,
        title: data.title,
        comment: data.comment,
        travel_date: data.travel_date,
        pros: pros.length > 0 ? pros : undefined,
        cons: cons.length > 0 ? cons : undefined,
      }

      const newReview = await ReviewsService.createMockReview(reviewData)
      onReviewSubmit(newReview)
    } catch (error) {
      console.error('Error submitting review:', error)
      alert('Failed to submit review. Please try again.')
    } finally {
      setSubmitting(false)
    }
  }

  const handleRatingChange = (newRating: number) => {
    setValue('rating', newRating)
  }

  const addPro = () => {
    if (newPro.trim() && !pros.includes(newPro.trim())) {
      const updatedPros = [...pros, newPro.trim()]
      setPros(updatedPros)
      setValue('pros', updatedPros)
      setNewPro('')
    }
  }

  const removePro = (index: number) => {
    const updatedPros = pros.filter((_, i) => i !== index)
    setPros(updatedPros)
    setValue('pros', updatedPros)
  }

  const addCon = () => {
    if (newCon.trim() && !cons.includes(newCon.trim())) {
      const updatedCons = [...cons, newCon.trim()]
      setCons(updatedCons)
      setValue('cons', updatedCons)
      setNewCon('')
    }
  }

  const removeCon = (index: number) => {
    const updatedCons = cons.filter((_, i) => i !== index)
    setCons(updatedCons)
    setValue('cons', updatedCons)
  }

  const getRatingText = (rating: number): string => {
    switch (rating) {
      case 1: return 'Terrible'
      case 2: return 'Poor'
      case 3: return 'Average'
      case 4: return 'Very Good'
      case 5: return 'Excellent'
      default: return 'Select a rating'
    }
  }

  return (
    <div className={`bg-white rounded-lg border border-gray-200 p-6 ${className}`}>
      <h3 className="text-xl font-semibold text-gray-900 mb-6">Write a Review</h3>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        {/* Rating */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-3">
            Overall Rating *
          </label>
          <div className="flex items-center space-x-4">
            <StarRating
              rating={rating}
              interactive
              onRatingChange={handleRatingChange}
              size="lg"
            />
            <span className="text-lg font-medium text-gray-700">
              {getRatingText(rating)}
            </span>
          </div>
          {errors.rating && (
            <p className="mt-1 text-sm text-red-600">{errors.rating.message}</p>
          )}
        </div>

        {/* Travel Date */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            When did you travel? *
          </label>
          <div className="relative">
            <Calendar className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
            <input
              type="date"
              {...register('travel_date')}
              max={new Date().toISOString().split('T')[0]}
              className="w-full pl-10 pr-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
            />
          </div>
          {errors.travel_date && (
            <p className="mt-1 text-sm text-red-600">{errors.travel_date.message}</p>
          )}
        </div>

        {/* Title */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Review Title *
          </label>
          <input
            type="text"
            {...register('title')}
            placeholder="Summarize your experience in a few words"
            className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          />
          {errors.title && (
            <p className="mt-1 text-sm text-red-600">{errors.title.message}</p>
          )}
        </div>

        {/* Comment */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Your Review *
          </label>
          <textarea
            {...register('comment')}
            rows={5}
            placeholder="Share your experience with other travelers. What did you love? What could be improved?"
            className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent resize-none"
          />
          <div className="flex justify-between mt-1">
            {errors.comment ? (
              <p className="text-sm text-red-600">{errors.comment.message}</p>
            ) : (
              <p className="text-sm text-gray-500">Minimum 10 characters</p>
            )}
            <p className="text-sm text-gray-500">{comment?.length || 0}/1000</p>
          </div>
        </div>

        {/* Pros and Cons */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Pros */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-3">
              👍 What did you like?
            </label>
            <div className="space-y-2">
              {pros.map((pro, index) => (
                <div key={index} className="flex items-center justify-between bg-green-50 border border-green-200 rounded-lg px-3 py-2">
                  <span className="text-sm text-green-800">{pro}</span>
                  <button
                    type="button"
                    onClick={() => removePro(index)}
                    className="text-green-600 hover:text-green-800"
                  >
                    <X className="h-4 w-4" />
                  </button>
                </div>
              ))}
              <div className="flex space-x-2">
                <input
                  type="text"
                  value={newPro}
                  onChange={(e) => setNewPro(e.target.value)}
                  onKeyPress={(e) => e.key === 'Enter' && (e.preventDefault(), addPro())}
                  placeholder="Add a positive point"
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent text-sm"
                />
                <button
                  type="button"
                  onClick={addPro}
                  className="px-3 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
                >
                  <Plus className="h-4 w-4" />
                </button>
              </div>
            </div>
          </div>

          {/* Cons */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-3">
              👎 What could be improved?
            </label>
            <div className="space-y-2">
              {cons.map((con, index) => (
                <div key={index} className="flex items-center justify-between bg-red-50 border border-red-200 rounded-lg px-3 py-2">
                  <span className="text-sm text-red-800">{con}</span>
                  <button
                    type="button"
                    onClick={() => removeCon(index)}
                    className="text-red-600 hover:text-red-800"
                  >
                    <X className="h-4 w-4" />
                  </button>
                </div>
              ))}
              <div className="flex space-x-2">
                <input
                  type="text"
                  value={newCon}
                  onChange={(e) => setNewCon(e.target.value)}
                  onKeyPress={(e) => e.key === 'Enter' && (e.preventDefault(), addCon())}
                  placeholder="Add an area for improvement"
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent text-sm"
                />
                <button
                  type="button"
                  onClick={addCon}
                  className="px-3 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
                >
                  <Plus className="h-4 w-4" />
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Submit Buttons */}
        <div className="flex space-x-4 pt-4">
          <button
            type="submit"
            disabled={submitting}
            className="flex-1 bg-primary text-primary-foreground py-3 px-6 rounded-lg font-medium hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {submitting ? (
              <div className="flex items-center justify-center">
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                Submitting...
              </div>
            ) : (
              'Submit Review'
            )}
          </button>
          
          {onCancel && (
            <button
              type="button"
              onClick={onCancel}
              className="px-6 py-3 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
            >
              Cancel
            </button>
          )}
        </div>
      </form>
    </div>
  )
}
