import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { AuthProvider } from '@/contexts/auth-context'
import LoginForm from '../login-form'

// Mock the auth service
jest.mock('@/lib/auth', () => ({
  AuthService: {
    login: jest.fn(),
  },
}))

// Mock the router
const mockPush = jest.fn()
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockPush,
  }),
}))

const MockAuthProvider = ({ children }: { children: React.ReactNode }) => {
  return (
    <AuthProvider>
      {children}
    </AuthProvider>
  )
}

describe('LoginForm', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders login form correctly', () => {
    render(
      <MockAuthProvider>
        <LoginForm />
      </MockAuthProvider>
    )

    expect(screen.getByLabelText(/email/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
    expect(screen.getByText(/don't have an account/i)).toBeInTheDocument()
  })

  it('shows validation errors for empty fields', async () => {
    const user = userEvent.setup()
    
    render(
      <MockAuthProvider>
        <LoginForm />
      </MockAuthProvider>
    )

    const submitButton = screen.getByRole('button', { name: /sign in/i })
    await user.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText(/email is required/i)).toBeInTheDocument()
      expect(screen.getByText(/password is required/i)).toBeInTheDocument()
    })
  })

  it('shows validation error for invalid email format', async () => {
    const user = userEvent.setup()
    
    render(
      <MockAuthProvider>
        <LoginForm />
      </MockAuthProvider>
    )

    const emailInput = screen.getByLabelText(/email/i)
    const submitButton = screen.getByRole('button', { name: /sign in/i })

    await user.type(emailInput, 'invalid-email')
    await user.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText(/please enter a valid email/i)).toBeInTheDocument()
    })
  })

  it('submits form with valid data', async () => {
    const user = userEvent.setup()
    const { AuthService } = require('@/lib/auth')
    
    AuthService.login.mockResolvedValue({
      user: {
        id: 1,
        email: 'test@example.com',
        firstName: 'John',
        lastName: 'Doe',
        role: 'user',
      },
      accessToken: 'mock-token',
      refreshToken: 'mock-refresh-token',
    })

    render(
      <MockAuthProvider>
        <LoginForm />
      </MockAuthProvider>
    )

    const emailInput = screen.getByLabelText(/email/i)
    const passwordInput = screen.getByLabelText(/password/i)
    const submitButton = screen.getByRole('button', { name: /sign in/i })

    await user.type(emailInput, 'test@example.com')
    await user.type(passwordInput, 'password123')
    await user.click(submitButton)

    await waitFor(() => {
      expect(AuthService.login).toHaveBeenCalledWith({
        email: 'test@example.com',
        password: 'password123',
      })
    })
  })

  it('shows error message on login failure', async () => {
    const user = userEvent.setup()
    const { AuthService } = require('@/lib/auth')
    
    AuthService.login.mockRejectedValue(new Error('Invalid credentials'))

    render(
      <MockAuthProvider>
        <LoginForm />
      </MockAuthProvider>
    )

    const emailInput = screen.getByLabelText(/email/i)
    const passwordInput = screen.getByLabelText(/password/i)
    const submitButton = screen.getByRole('button', { name: /sign in/i })

    await user.type(emailInput, 'test@example.com')
    await user.type(passwordInput, 'wrongpassword')
    await user.click(submitButton)

    await waitFor(() => {
      expect(screen.getByText(/invalid credentials/i)).toBeInTheDocument()
    })
  })

  it('shows loading state during submission', async () => {
    const user = userEvent.setup()
    const { AuthService } = require('@/lib/auth')
    
    // Mock a delayed response
    AuthService.login.mockImplementation(() => 
      new Promise(resolve => setTimeout(resolve, 1000))
    )

    render(
      <MockAuthProvider>
        <LoginForm />
      </MockAuthProvider>
    )

    const emailInput = screen.getByLabelText(/email/i)
    const passwordInput = screen.getByLabelText(/password/i)
    const submitButton = screen.getByRole('button', { name: /sign in/i })

    await user.type(emailInput, 'test@example.com')
    await user.type(passwordInput, 'password123')
    await user.click(submitButton)

    expect(screen.getByText(/signing in/i)).toBeInTheDocument()
    expect(submitButton).toBeDisabled()
  })

  it('toggles password visibility', async () => {
    const user = userEvent.setup()
    
    render(
      <MockAuthProvider>
        <LoginForm />
      </MockAuthProvider>
    )

    const passwordInput = screen.getByLabelText(/password/i)
    const toggleButton = screen.getByRole('button', { name: /toggle password visibility/i })

    expect(passwordInput).toHaveAttribute('type', 'password')

    await user.click(toggleButton)
    expect(passwordInput).toHaveAttribute('type', 'text')

    await user.click(toggleButton)
    expect(passwordInput).toHaveAttribute('type', 'password')
  })

  it('navigates to register page when clicking sign up link', async () => {
    const user = userEvent.setup()
    
    render(
      <MockAuthProvider>
        <LoginForm />
      </MockAuthProvider>
    )

    const signUpLink = screen.getByText(/sign up/i)
    await user.click(signUpLink)

    expect(mockPush).toHaveBeenCalledWith('/auth/register')
  })

  it('handles remember me checkbox', async () => {
    const user = userEvent.setup()
    
    render(
      <MockAuthProvider>
        <LoginForm />
      </MockAuthProvider>
    )

    const rememberMeCheckbox = screen.getByLabelText(/remember me/i)
    
    expect(rememberMeCheckbox).not.toBeChecked()
    
    await user.click(rememberMeCheckbox)
    expect(rememberMeCheckbox).toBeChecked()
    
    await user.click(rememberMeCheckbox)
    expect(rememberMeCheckbox).not.toBeChecked()
  })

  it('shows forgot password link', () => {
    render(
      <MockAuthProvider>
        <LoginForm />
      </MockAuthProvider>
    )

    const forgotPasswordLink = screen.getByText(/forgot your password/i)
    expect(forgotPasswordLink).toBeInTheDocument()
    expect(forgotPasswordLink).toHaveAttribute('href', '/auth/forgot-password')
  })

  it('disables form during submission', async () => {
    const user = userEvent.setup()
    const { AuthService } = require('@/lib/auth')
    
    // Mock a delayed response
    AuthService.login.mockImplementation(() => 
      new Promise(resolve => setTimeout(resolve, 1000))
    )

    render(
      <MockAuthProvider>
        <LoginForm />
      </MockAuthProvider>
    )

    const emailInput = screen.getByLabelText(/email/i)
    const passwordInput = screen.getByLabelText(/password/i)
    const submitButton = screen.getByRole('button', { name: /sign in/i })

    await user.type(emailInput, 'test@example.com')
    await user.type(passwordInput, 'password123')
    await user.click(submitButton)

    expect(emailInput).toBeDisabled()
    expect(passwordInput).toBeDisabled()
    expect(submitButton).toBeDisabled()
  })
})
