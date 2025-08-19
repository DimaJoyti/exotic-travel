import api from './api'
import { Review } from '@/types'

export interface ReviewStats {
  average_rating: number
  total_reviews: number
  rating_distribution: {
    5: number
    4: number
    3: number
    2: number
    1: number
  }
}

export interface CreateReviewData {
  destination_id: number
  booking_id?: number
  rating: number
  title: string
  comment: string
  pros?: string[]
  cons?: string[]
  travel_date: string
}

export interface UpdateReviewData {
  rating?: number
  title?: string
  comment?: string
  pros?: string[]
  cons?: string[]
}

export class ReviewsService {
  // Get reviews for a destination
  static async getDestinationReviews(
    destinationId: number,
    page: number = 1,
    limit: number = 10,
    sortBy: string = 'newest'
  ): Promise<{ reviews: Review[]; total: number; stats: ReviewStats }> {
    const response = await api.get<{ reviews: Review[]; total: number; stats: ReviewStats }>(
      `/api/destinations/${destinationId}/reviews`,
      {
        params: { page, limit, sort: sortBy }
      }
    )
    return response.data
  }

  // Get user's reviews
  static async getUserReviews(userId: number): Promise<Review[]> {
    const response = await api.get<Review[]>(`/api/users/${userId}/reviews`)
    return response.data
  }

  // Get single review
  static async getReview(reviewId: number): Promise<Review> {
    const response = await api.get<Review>(`/api/reviews/${reviewId}`)
    return response.data
  }

  // Create a new review
  static async createReview(reviewData: CreateReviewData): Promise<Review> {
    const response = await api.post<Review>('/api/reviews', reviewData)
    return response.data
  }

  // Update a review
  static async updateReview(reviewId: number, reviewData: UpdateReviewData): Promise<Review> {
    const response = await api.patch<Review>(`/api/reviews/${reviewId}`, reviewData)
    return response.data
  }

  // Delete a review
  static async deleteReview(reviewId: number): Promise<void> {
    await api.delete(`/api/reviews/${reviewId}`)
  }

  // Mark review as helpful
  static async markReviewHelpful(reviewId: number): Promise<void> {
    await api.post(`/api/reviews/${reviewId}/helpful`)
  }

  // Report a review
  static async reportReview(reviewId: number, reason: string): Promise<void> {
    await api.post(`/api/reviews/${reviewId}/report`, { reason })
  }

  // Get review statistics for a destination
  static async getDestinationReviewStats(destinationId: number): Promise<ReviewStats> {
    const response = await api.get<ReviewStats>(`/api/destinations/${destinationId}/reviews/stats`)
    return response.data
  }

  // Mock implementations for development
  static async getMockDestinationReviews(
    destinationId: number,
    page: number = 1,
    limit: number = 10,
    sortBy: string = 'newest'
  ): Promise<{ reviews: Review[]; total: number; stats: ReviewStats }> {
    // Mock reviews data
    const mockReviews: Review[] = [
      {
        id: 1,
        user_id: 1,
        destination_id: destinationId,
        booking_id: 1,
        rating: 5,
        title: "Absolutely Amazing Experience!",
        comment: "This was the trip of a lifetime! The destination exceeded all my expectations. The accommodations were luxurious, the staff was incredibly friendly, and the activities were perfectly organized. I would definitely recommend this to anyone looking for an unforgettable adventure.",
        pros: ["Stunning scenery", "Excellent service", "Great value for money", "Well organized"],
        cons: ["Weather could have been better"],
        travel_date: "2024-02-15",
        verified_booking: true,
        helpful_count: 12,
        created_at: "2024-02-20T10:00:00Z",
        updated_at: "2024-02-20T10:00:00Z",
        user: {
          id: 1,
          email: "john.doe@email.com",
          first_name: "John",
          last_name: "D.",
          role: "user",
          created_at: "2024-01-15T10:00:00Z",
          updated_at: "2024-01-15T10:00:00Z",
        }
      },
      {
        id: 2,
        user_id: 2,
        destination_id: destinationId,
        booking_id: 2,
        rating: 4,
        title: "Great destination with minor issues",
        comment: "Overall, this was a fantastic trip. The location is breathtaking and the activities were well-planned. However, there were some minor issues with the accommodation that prevented it from being perfect. The food was excellent and the guides were very knowledgeable.",
        pros: ["Beautiful location", "Knowledgeable guides", "Excellent food"],
        cons: ["Room maintenance issues", "Limited WiFi"],
        travel_date: "2024-01-20",
        verified_booking: true,
        helpful_count: 8,
        created_at: "2024-01-25T10:00:00Z",
        updated_at: "2024-01-25T10:00:00Z",
        user: {
          id: 2,
          email: "jane.smith@email.com",
          first_name: "Jane",
          last_name: "S.",
          role: "user",
          created_at: "2024-01-01T10:00:00Z",
          updated_at: "2024-01-01T10:00:00Z",
        }
      },
      {
        id: 3,
        user_id: 3,
        destination_id: destinationId,
        booking_id: 3,
        rating: 5,
        title: "Perfect for families!",
        comment: "We traveled with our two kids and this destination was perfect for families. There were plenty of activities for children, the staff was very accommodating, and the safety measures were excellent. Our kids are already asking when we can go back!",
        pros: ["Family-friendly", "Safe environment", "Kids activities", "Accommodating staff"],
        cons: [],
        travel_date: "2024-03-01",
        verified_booking: true,
        helpful_count: 15,
        created_at: "2024-03-05T10:00:00Z",
        updated_at: "2024-03-05T10:00:00Z",
        user: {
          id: 3,
          email: "mike.johnson@email.com",
          first_name: "Mike",
          last_name: "J.",
          role: "user",
          created_at: "2024-02-01T10:00:00Z",
          updated_at: "2024-02-01T10:00:00Z",
        }
      },
      {
        id: 4,
        user_id: 4,
        destination_id: destinationId,
        rating: 3,
        title: "Good but not great",
        comment: "The destination has potential but there are areas for improvement. The location is nice and the activities are decent, but the service could be better and the value for money is questionable. It's an okay experience but I expected more for the price.",
        pros: ["Nice location", "Decent activities"],
        cons: ["Overpriced", "Service could be better", "Limited dining options"],
        travel_date: "2024-01-10",
        verified_booking: false,
        helpful_count: 3,
        created_at: "2024-01-15T10:00:00Z",
        updated_at: "2024-01-15T10:00:00Z",
        user: {
          id: 4,
          email: "sarah.wilson@email.com",
          first_name: "Sarah",
          last_name: "W.",
          role: "user",
          created_at: "2024-01-05T10:00:00Z",
          updated_at: "2024-01-05T10:00:00Z",
        }
      }
    ]

    const mockStats: ReviewStats = {
      average_rating: 4.25,
      total_reviews: 4,
      rating_distribution: {
        5: 2,
        4: 1,
        3: 1,
        2: 0,
        1: 0
      }
    }

    // Apply sorting
    const sortedReviews = [...mockReviews]
    switch (sortBy) {
      case 'newest':
        sortedReviews.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
        break
      case 'oldest':
        sortedReviews.sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime())
        break
      case 'highest':
        sortedReviews.sort((a, b) => b.rating - a.rating)
        break
      case 'lowest':
        sortedReviews.sort((a, b) => a.rating - b.rating)
        break
      case 'helpful':
        sortedReviews.sort((a, b) => b.helpful_count - a.helpful_count)
        break
    }

    // Apply pagination
    const startIndex = (page - 1) * limit
    const paginatedReviews = sortedReviews.slice(startIndex, startIndex + limit)

    return {
      reviews: paginatedReviews,
      total: mockReviews.length,
      stats: mockStats
    }
  }

  static async createMockReview(reviewData: CreateReviewData): Promise<Review> {
    console.log('📝 Mock Review Created:', reviewData)
    
    const mockReview: Review = {
      id: Date.now(),
      user_id: 1, // Current user
      destination_id: reviewData.destination_id,
      booking_id: reviewData.booking_id,
      rating: reviewData.rating,
      title: reviewData.title,
      comment: reviewData.comment,
      pros: reviewData.pros,
      cons: reviewData.cons,
      travel_date: reviewData.travel_date,
      verified_booking: !!reviewData.booking_id,
      helpful_count: 0,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    }

    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 500))
    
    return mockReview
  }

  static async markMockReviewHelpful(reviewId: number): Promise<void> {
    console.log('👍 Mock Review Marked Helpful:', reviewId)
    
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 200))
  }

  // Utility functions
  static calculateAverageRating(reviews: Review[]): number {
    if (reviews.length === 0) return 0
    const sum = reviews.reduce((acc, review) => acc + review.rating, 0)
    return Math.round((sum / reviews.length) * 10) / 10
  }

  static getRatingDistribution(reviews: Review[]): ReviewStats['rating_distribution'] {
    const distribution = { 5: 0, 4: 0, 3: 0, 2: 0, 1: 0 }
    reviews.forEach(review => {
      distribution[review.rating as keyof typeof distribution]++
    })
    return distribution
  }

  static formatReviewDate(dateString: string): string {
    const date = new Date(dateString)
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    })
  }

  static getReviewSummary(rating: number): string {
    if (rating >= 4.5) return 'Excellent'
    if (rating >= 4.0) return 'Very Good'
    if (rating >= 3.5) return 'Good'
    if (rating >= 3.0) return 'Average'
    if (rating >= 2.0) return 'Poor'
    return 'Terrible'
  }
}
