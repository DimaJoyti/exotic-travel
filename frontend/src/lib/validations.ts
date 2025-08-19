import { z } from 'zod'

export const loginSchema = z.object({
  email: z
    .string()
    .min(1, 'Email is required')
    .email('Please enter a valid email address'),
  password: z
    .string()
    .min(1, 'Password is required')
    .min(8, 'Password must be at least 8 characters'),
})

export const registerSchema = z.object({
  first_name: z
    .string()
    .min(1, 'First name is required')
    .min(2, 'First name must be at least 2 characters')
    .max(50, 'First name must be less than 50 characters'),
  last_name: z
    .string()
    .min(1, 'Last name is required')
    .min(2, 'Last name must be at least 2 characters')
    .max(50, 'Last name must be less than 50 characters'),
  email: z
    .string()
    .min(1, 'Email is required')
    .email('Please enter a valid email address'),
  password: z
    .string()
    .min(1, 'Password is required')
    .min(8, 'Password must be at least 8 characters')
    .regex(
      /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/,
      'Password must contain at least one uppercase letter, one lowercase letter, and one number'
    ),
  confirmPassword: z
    .string()
    .min(1, 'Please confirm your password'),
}).refine((data) => data.password === data.confirmPassword, {
  message: 'Passwords do not match',
  path: ['confirmPassword'],
})

export const forgotPasswordSchema = z.object({
  email: z
    .string()
    .min(1, 'Email is required')
    .email('Please enter a valid email address'),
})

export const resetPasswordSchema = z.object({
  password: z
    .string()
    .min(1, 'Password is required')
    .min(8, 'Password must be at least 8 characters')
    .regex(
      /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/,
      'Password must contain at least one uppercase letter, one lowercase letter, and one number'
    ),
  confirmPassword: z
    .string()
    .min(1, 'Please confirm your password'),
}).refine((data) => data.password === data.confirmPassword, {
  message: 'Passwords do not match',
  path: ['confirmPassword'],
})

export const guestDetailSchema = z.object({
  first_name: z
    .string()
    .min(1, 'First name is required')
    .min(2, 'First name must be at least 2 characters'),
  last_name: z
    .string()
    .min(1, 'Last name is required')
    .min(2, 'Last name must be at least 2 characters'),
  email: z
    .string()
    .min(1, 'Email is required')
    .email('Please enter a valid email address'),
  phone: z
    .string()
    .optional(),
  date_of_birth: z
    .string()
    .optional(),
  passport_number: z
    .string()
    .optional(),
  dietary_requirements: z
    .string()
    .optional(),
})

export const bookingSchema = z.object({
  destination_id: z.number().min(1, 'Destination is required'),
  check_in_date: z
    .string()
    .min(1, 'Check-in date is required')
    .refine((date) => {
      const checkIn = new Date(date)
      const today = new Date()
      today.setHours(0, 0, 0, 0)
      return checkIn >= today
    }, 'Check-in date must be today or in the future'),
  check_out_date: z
    .string()
    .min(1, 'Check-out date is required'),
  guests: z
    .number()
    .min(1, 'At least 1 guest is required')
    .max(20, 'Maximum 20 guests allowed'),
  guest_details: z
    .array(guestDetailSchema)
    .min(1, 'At least one guest detail is required'),
  special_requests: z
    .string()
    .optional(),
}).refine((data) => {
  const checkIn = new Date(data.check_in_date)
  const checkOut = new Date(data.check_out_date)
  return checkOut > checkIn
}, {
  message: 'Check-out date must be after check-in date',
  path: ['check_out_date'],
}).refine((data) => {
  return data.guest_details.length === data.guests
}, {
  message: 'Number of guest details must match number of guests',
  path: ['guest_details'],
})

export type LoginFormData = z.infer<typeof loginSchema>
export type RegisterFormData = z.infer<typeof registerSchema>
export type ForgotPasswordFormData = z.infer<typeof forgotPasswordSchema>
export type ResetPasswordFormData = z.infer<typeof resetPasswordSchema>
export type GuestDetailFormData = z.infer<typeof guestDetailSchema>
export type BookingFormData = z.infer<typeof bookingSchema>
