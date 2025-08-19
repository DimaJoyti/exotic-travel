import api from './api'
import { EmailTemplates, EmailTemplate as EmailTemplateType } from './email-templates'

export interface EmailNotificationTemplate {
  id: string
  name: string
  subject: string
  html_content: string
  text_content: string
  variables: string[]
}

export interface EmailNotification {
  to: string
  template_id: string
  variables: Record<string, any>
  scheduled_at?: string
}

export interface SMSNotification {
  to: string
  message: string
  scheduled_at?: string
}

export interface PushNotification {
  user_id: number
  title: string
  body: string
  data?: Record<string, any>
  scheduled_at?: string
}

export interface NotificationPreferences {
  email_enabled: boolean
  sms_enabled: boolean
  push_enabled: boolean
  marketing_emails: boolean
  booking_updates: boolean
  travel_reminders: boolean
  promotional_offers: boolean
}

export class NotificationsService {
  // Email notifications with templates
  static async sendTemplatedEmail(
    to: string,
    templateId: string,
    variables: Record<string, any>,
    scheduledAt?: string
  ): Promise<void> {
    const template = EmailTemplates.getTemplateById(templateId)
    if (!template) {
      throw new Error(`Email template '${templateId}' not found`)
    }

    // Validate required variables
    const validation = EmailTemplates.validateVariables(templateId, variables)
    if (!validation.valid) {
      throw new Error(`Missing required variables: ${validation.missing.join(', ')}`)
    }

    // Add default variables
    const allVariables = {
      app_url: process.env.NEXT_PUBLIC_APP_URL || 'http://localhost:3000',
      current_year: new Date().getFullYear(),
      ...variables
    }

    // Replace variables in template
    const subject = EmailTemplates.replaceVariables(template.subject, allVariables)
    const html = EmailTemplates.replaceVariables(template.html, allVariables)
    const text = EmailTemplates.replaceVariables(template.text, allVariables)

    const notification: EmailNotification = {
      to,
      template_id: templateId,
      variables: allVariables,
      scheduled_at: scheduledAt
    }

    try {
      await api.post('/api/notifications/email/templated', {
        ...notification,
        subject,
        html,
        text
      })
    } catch (error) {
      console.warn('API not available, using mock email service:', error)
      await this.sendMockTemplatedEmail(to, templateId, allVariables, subject, html, text)
    }
  }

  // Original email method for backward compatibility
  static async sendEmail(notification: EmailNotification): Promise<void> {
    await api.post('/api/notifications/email', notification)
  }

  static async sendBookingConfirmation(
    email: string,
    bookingData: {
      confirmation_number: string
      destination_name: string
      check_in_date: string
      check_out_date: string
      guests: number
      total_price: number
      booking_id?: number
    }
  ): Promise<void> {
    const variables = {
      ...bookingData,
      check_in_date: new Date(bookingData.check_in_date).toLocaleDateString('en-US', {
        weekday: 'long',
        year: 'numeric',
        month: 'long',
        day: 'numeric'
      }),
      check_out_date: new Date(bookingData.check_out_date).toLocaleDateString('en-US', {
        weekday: 'long',
        year: 'numeric',
        month: 'long',
        day: 'numeric'
      }),
      total_price: bookingData.total_price.toLocaleString()
    }

    await this.sendTemplatedEmail(email, 'booking_confirmation', variables)
  }

  static async sendBookingReminder(
    email: string,
    bookingData: {
      destination_name: string
      check_in_date: string
      confirmation_number: string
      days_until_trip: number
      booking_id?: number
    }
  ): Promise<void> {
    const variables = {
      ...bookingData,
      check_in_date: new Date(bookingData.check_in_date).toLocaleDateString('en-US', {
        weekday: 'long',
        year: 'numeric',
        month: 'long',
        day: 'numeric'
      })
    }

    await this.sendTemplatedEmail(email, 'booking_reminder', variables)
  }

  static async sendWelcomeEmail(
    email: string,
    userData: {
      first_name: string
      last_name: string
    }
  ): Promise<void> {
    await this.sendTemplatedEmail(email, 'welcome', userData)
  }

  static async sendPasswordResetEmail(
    email: string,
    resetData: {
      first_name: string
      reset_link: string
      expires_at: string
    }
  ): Promise<void> {
    const variables = {
      ...resetData,
      expires_at: new Date(resetData.expires_at).toLocaleString()
    }

    await this.sendTemplatedEmail(email, 'password_reset', variables)
  }

  // SMS notifications
  static async sendSMS(notification: SMSNotification): Promise<void> {
    await api.post('/api/notifications/sms', notification)
  }

  static async sendBookingSMS(
    phone: string,
    message: string
  ): Promise<void> {
    const notification: SMSNotification = {
      to: phone,
      message,
    }
    
    await this.sendSMS(notification)
  }

  // Push notifications
  static async sendPushNotification(notification: PushNotification): Promise<void> {
    await api.post('/api/notifications/push', notification)
  }

  // Notification preferences
  static async getNotificationPreferences(userId: number): Promise<NotificationPreferences> {
    const response = await api.get<NotificationPreferences>(`/api/users/${userId}/notification-preferences`)
    return response.data
  }

  static async updateNotificationPreferences(
    userId: number,
    preferences: Partial<NotificationPreferences>
  ): Promise<NotificationPreferences> {
    const response = await api.patch<NotificationPreferences>(
      `/api/users/${userId}/notification-preferences`,
      preferences
    )
    return response.data
  }

  // Email templates
  static async getEmailTemplates(): Promise<EmailNotificationTemplate[]> {
    const response = await api.get<EmailNotificationTemplate[]>('/api/notifications/email/templates')
    return response.data
  }

  static async createEmailTemplate(template: Omit<EmailNotificationTemplate, 'id'>): Promise<EmailNotificationTemplate> {
    const response = await api.post<EmailNotificationTemplate>('/api/notifications/email/templates', template)
    return response.data
  }

  // Mock implementations for development
  static async sendMockTemplatedEmail(
    to: string,
    templateId: string,
    variables: Record<string, any>,
    subject: string,
    html: string,
    text: string
  ): Promise<void> {
    console.log('📧 Mock Templated Email Sent:', {
      to,
      template: templateId,
      subject,
      variables,
      html_preview: html.substring(0, 200) + '...',
      text_preview: text.substring(0, 200) + '...'
    })

    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 500))
  }

  static async sendMockEmail(notification: EmailNotification): Promise<void> {
    console.log('📧 Mock Email Sent:', {
      to: notification.to,
      template: notification.template_id,
      variables: notification.variables,
    })

    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 500))
  }

  static async sendMockBookingConfirmation(
    email: string,
    bookingData: any
  ): Promise<void> {
    console.log('📧 Booking Confirmation Email:', {
      to: email,
      subject: `Booking Confirmed - ${bookingData.destination_name}`,
      content: `Your booking ${bookingData.confirmation_number} has been confirmed!`,
      data: bookingData,
    })
    
    await this.sendMockEmail({
      to: email,
      template_id: 'booking_confirmation',
      variables: bookingData,
    })
  }

  static async sendMockWelcomeEmail(email: string, userData: any): Promise<void> {
    console.log('📧 Welcome Email:', {
      to: email,
      subject: `Welcome to ExoticTravel, ${userData.first_name}!`,
      content: 'Thank you for joining our travel community!',
      data: userData,
    })
    
    await this.sendMockEmail({
      to: email,
      template_id: 'welcome',
      variables: userData,
    })
  }

  static async sendMockSMS(phone: string, message: string): Promise<void> {
    console.log('📱 Mock SMS Sent:', {
      to: phone,
      message,
    })
    
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 300))
  }

  static async sendMockPushNotification(notification: PushNotification): Promise<void> {
    console.log('🔔 Mock Push Notification:', notification)
    
    // Simulate API delay
    await new Promise(resolve => setTimeout(resolve, 200))
  }

  // Utility functions
  static formatPhoneNumber(phone: string): string {
    // Remove all non-digits
    const cleaned = phone.replace(/\D/g, '')
    
    // Add country code if missing
    if (cleaned.length === 10) {
      return `+1${cleaned}`
    }
    
    if (cleaned.length === 11 && cleaned.startsWith('1')) {
      return `+${cleaned}`
    }
    
    return phone
  }

  static validateEmail(email: string): boolean {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    return emailRegex.test(email)
  }

  static validatePhoneNumber(phone: string): boolean {
    const phoneRegex = /^\+?[\d\s\-\(\)]{10,}$/
    return phoneRegex.test(phone)
  }

  // Notification scheduling
  static async scheduleNotification(
    type: 'email' | 'sms' | 'push',
    notification: any,
    scheduledAt: Date
  ): Promise<void> {
    const payload = {
      ...notification,
      scheduled_at: scheduledAt.toISOString(),
    }
    
    await api.post(`/api/notifications/${type}/schedule`, payload)
  }

  static async cancelScheduledNotification(notificationId: string): Promise<void> {
    await api.delete(`/api/notifications/scheduled/${notificationId}`)
  }

  // Bulk notifications
  static async sendBulkEmails(notifications: EmailNotification[]): Promise<void> {
    await api.post('/api/notifications/email/bulk', { notifications })
  }

  static async sendNewsletterToSegment(
    segment: string,
    templateId: string,
    variables: Record<string, any>
  ): Promise<void> {
    await api.post('/api/notifications/newsletter', {
      segment,
      template_id: templateId,
      variables,
    })
  }
}
