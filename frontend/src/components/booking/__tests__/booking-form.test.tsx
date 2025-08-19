import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import BookingForm from '../booking-form'
import { AuthProvider } from '@/contexts/auth-context'

// Mock the booking service
jest.mock('@/lib/bookings', () => ({
  BookingsService: {
    createBooking: jest.fn(),
  },
}))

// Mock the destinations service
jest.mock('@/lib/destinations', () => ({
  DestinationsService: {
    getDestination: jest.fn(),
  },
}))

const mockDestination = {
  id: 1,
  name: 'Maldives Paradise Resort',
  country: 'Maldives',
  description: 'Beautiful tropical paradise',
  price_per_night: 500,
  rating: 4.8,
  images: ['image1.jpg', 'image2.jpg'],
  amenities: ['WiFi', 'Pool', 'Spa'],
  location: {
    latitude: 3.2028,
    longitude: 73.2207,
  },
}

const MockAuthProvider = ({ children }: { children: React.ReactNode }) => {
  const mockUser = {
    id: 1,
    email: 'test@example.com',
    firstName: 'John',
    lastName: 'Doe',
    role: 'user' as const,
  }

  return (
    <AuthProvider value={{ user: mockUser, login: jest.fn(), logout: jest.fn(), loading: false }}>
      {children}
    </AuthProvider>
  )
}

describe('BookingForm', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    const { DestinationsService } = require('@/lib/destinations')
    DestinationsService.getDestination.mockResolvedValue(mockDestination)
  })

  it('renders booking form correctly', async () => {
    render(
      <MockAuthProvider>
        <BookingForm destinationId={1} />
      </MockAuthProvider>
    )

    await waitFor(() => {
      expect(screen.getByText(/book your stay/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/check-in date/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/check-out date/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/guests/i)).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /book now/i })).toBeInTheDocument()
    })
  })

  it('shows destination information', async () => {
    render(
      <MockAuthProvider>
        <BookingForm destinationId={1} />
      </MockAuthProvider>
    )

    await waitFor(() => {
      expect(screen.getByText(mockDestination.name)).toBeInTheDocument()
      expect(screen.getByText(`$${mockDestination.price_per_night}`)).toBeInTheDocument()
    })
  })

  it('validates required fields', async () => {
    const user = userEvent.setup()
    
    render(
      <MockAuthProvider>
        <BookingForm destinationId={1} />
      </MockAuthProvider>
    )

    const bookButton = await screen.findByRole('button', { name: /book now/i })
    await user.click(bookButton)

    await waitFor(() => {
      expect(screen.getByText(/check-in date is required/i)).toBeInTheDocument()
      expect(screen.getByText(/check-out date is required/i)).toBeInTheDocument()
    })
  })

  it('validates date range', async () => {
    const user = userEvent.setup()
    
    render(
      <MockAuthProvider>
        <BookingForm destinationId={1} />
      </MockAuthProvider>
    )

    const checkInInput = await screen.findByLabelText(/check-in date/i)
    const checkOutInput = screen.getByLabelText(/check-out date/i)

    // Set check-out date before check-in date
    const tomorrow = new Date()
    tomorrow.setDate(tomorrow.getDate() + 1)
    const today = new Date()

    await user.type(checkInInput, tomorrow.toISOString().split('T')[0])
    await user.type(checkOutInput, today.toISOString().split('T')[0])

    const bookButton = screen.getByRole('button', { name: /book now/i })
    await user.click(bookButton)

    await waitFor(() => {
      expect(screen.getByText(/check-out date must be after check-in date/i)).toBeInTheDocument()
    })
  })

  it('validates minimum stay duration', async () => {
    const user = userEvent.setup()
    
    render(
      <MockAuthProvider>
        <BookingForm destinationId={1} />
      </MockAuthProvider>
    )

    const checkInInput = await screen.findByLabelText(/check-in date/i)
    const checkOutInput = screen.getByLabelText(/check-out date/i)

    const today = new Date()
    const tomorrow = new Date()
    tomorrow.setDate(today.getDate() + 1)

    await user.type(checkInInput, today.toISOString().split('T')[0])
    await user.type(checkOutInput, tomorrow.toISOString().split('T')[0])

    const bookButton = screen.getByRole('button', { name: /book now/i })
    await user.click(bookButton)

    await waitFor(() => {
      expect(screen.getByText(/minimum stay is 1 night/i)).toBeInTheDocument()
    })
  })

  it('calculates total price correctly', async () => {
    const user = userEvent.setup()
    
    render(
      <MockAuthProvider>
        <BookingForm destinationId={1} />
      </MockAuthProvider>
    )

    const checkInInput = await screen.findByLabelText(/check-in date/i)
    const checkOutInput = screen.getByLabelText(/check-out date/i)
    const guestsInput = screen.getByLabelText(/guests/i)

    const today = new Date()
    const threeDaysLater = new Date()
    threeDaysLater.setDate(today.getDate() + 3)

    await user.type(checkInInput, today.toISOString().split('T')[0])
    await user.type(checkOutInput, threeDaysLater.toISOString().split('T')[0])
    await user.clear(guestsInput)
    await user.type(guestsInput, '2')

    // 3 nights * $500 = $1500
    await waitFor(() => {
      expect(screen.getByText(/\$1,500/)).toBeInTheDocument()
    })
  })

  it('handles guest count changes', async () => {
    const user = userEvent.setup()
    
    render(
      <MockAuthProvider>
        <BookingForm destinationId={1} />
      </MockAuthProvider>
    )

    const guestsInput = await screen.findByLabelText(/guests/i)
    const incrementButton = screen.getByRole('button', { name: /increase guests/i })
    const decrementButton = screen.getByRole('button', { name: /decrease guests/i })

    expect(guestsInput).toHaveValue(1)

    await user.click(incrementButton)
    expect(guestsInput).toHaveValue(2)

    await user.click(incrementButton)
    expect(guestsInput).toHaveValue(3)

    await user.click(decrementButton)
    expect(guestsInput).toHaveValue(2)

    await user.click(decrementButton)
    expect(guestsInput).toHaveValue(1)

    // Should not go below 1
    await user.click(decrementButton)
    expect(guestsInput).toHaveValue(1)
  })

  it('submits booking successfully', async () => {
    const user = userEvent.setup()
    const { BookingsService } = require('@/lib/bookings')
    
    BookingsService.createBooking.mockResolvedValue({
      id: 1,
      destination_id: 1,
      user_id: 1,
      check_in_date: '2024-01-01',
      check_out_date: '2024-01-04',
      guests: 2,
      total_price: 1500,
      status: 'confirmed',
    })

    render(
      <MockAuthProvider>
        <BookingForm destinationId={1} />
      </MockAuthProvider>
    )

    const checkInInput = await screen.findByLabelText(/check-in date/i)
    const checkOutInput = screen.getByLabelText(/check-out date/i)
    const guestsInput = screen.getByLabelText(/guests/i)

    await user.type(checkInInput, '2024-01-01')
    await user.type(checkOutInput, '2024-01-04')
    await user.clear(guestsInput)
    await user.type(guestsInput, '2')

    const bookButton = screen.getByRole('button', { name: /book now/i })
    await user.click(bookButton)

    await waitFor(() => {
      expect(BookingsService.createBooking).toHaveBeenCalledWith({
        destination_id: 1,
        check_in_date: '2024-01-01',
        check_out_date: '2024-01-04',
        guests: 2,
        total_price: 1500,
      })
    })
  })

  it('shows loading state during booking submission', async () => {
    const user = userEvent.setup()
    const { BookingsService } = require('@/lib/bookings')
    
    BookingsService.createBooking.mockImplementation(() => 
      new Promise(resolve => setTimeout(resolve, 1000))
    )

    render(
      <MockAuthProvider>
        <BookingForm destinationId={1} />
      </MockAuthProvider>
    )

    const checkInInput = await screen.findByLabelText(/check-in date/i)
    const checkOutInput = screen.getByLabelText(/check-out date/i)

    await user.type(checkInInput, '2024-01-01')
    await user.type(checkOutInput, '2024-01-04')

    const bookButton = screen.getByRole('button', { name: /book now/i })
    await user.click(bookButton)

    expect(screen.getByText(/processing booking/i)).toBeInTheDocument()
    expect(bookButton).toBeDisabled()
  })

  it('shows error message on booking failure', async () => {
    const user = userEvent.setup()
    const { BookingsService } = require('@/lib/bookings')
    
    BookingsService.createBooking.mockRejectedValue(new Error('Booking failed'))

    render(
      <MockAuthProvider>
        <BookingForm destinationId={1} />
      </MockAuthProvider>
    )

    const checkInInput = await screen.findByLabelText(/check-in date/i)
    const checkOutInput = screen.getByLabelText(/check-out date/i)

    await user.type(checkInInput, '2024-01-01')
    await user.type(checkOutInput, '2024-01-04')

    const bookButton = screen.getByRole('button', { name: /book now/i })
    await user.click(bookButton)

    await waitFor(() => {
      expect(screen.getByText(/booking failed/i)).toBeInTheDocument()
    })
  })

  it('shows special requests textarea', async () => {
    render(
      <MockAuthProvider>
        <BookingForm destinationId={1} />
      </MockAuthProvider>
    )

    await waitFor(() => {
      expect(screen.getByLabelText(/special requests/i)).toBeInTheDocument()
    })
  })

  it('handles special requests input', async () => {
    const user = userEvent.setup()
    
    render(
      <MockAuthProvider>
        <BookingForm destinationId={1} />
      </MockAuthProvider>
    )

    const specialRequestsInput = await screen.findByLabelText(/special requests/i)
    
    await user.type(specialRequestsInput, 'Please arrange airport pickup')
    
    expect(specialRequestsInput).toHaveValue('Please arrange airport pickup')
  })
})
