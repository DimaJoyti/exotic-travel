import api from './api'
import { BookingsService } from './bookings'
import { NotificationsService } from './notifications'

export interface WebhookEvent {
  id: string
  type: string
  data: {
    object: any
  }
  created: number
  livemode: boolean
  pending_webhooks: number
  request: {
    id: string
    idempotency_key?: string
  }
}

export interface PaymentWebhookData {
  payment_intent_id: string
  amount: number
  currency: string
  status: string
  customer_email?: string
  metadata?: Record<string, string>
}

export class WebhooksService {
  // Process Stripe webhook events
  static async processStripeWebhook(event: WebhookEvent): Promise<void> {
    console.log(`Processing webhook event: ${event.type}`)

    switch (event.type) {
      case 'payment_intent.succeeded':
        await this.handlePaymentSuccess(event.data.object)
        break
      
      case 'payment_intent.payment_failed':
        await this.handlePaymentFailure(event.data.object)
        break
      
      case 'payment_intent.canceled':
        await this.handlePaymentCancellation(event.data.object)
        break
      
      case 'charge.dispute.created':
        await this.handleChargeDispute(event.data.object)
        break
      
      case 'customer.subscription.created':
        await this.handleSubscriptionCreated(event.data.object)
        break
      
      case 'customer.subscription.deleted':
        await this.handleSubscriptionCanceled(event.data.object)
        break
      
      case 'invoice.payment_succeeded':
        await this.handleInvoicePaymentSuccess(event.data.object)
        break
      
      case 'invoice.payment_failed':
        await this.handleInvoicePaymentFailure(event.data.object)
        break
      
      default:
        console.log(`Unhandled webhook event type: ${event.type}`)
    }
  }

  // Handle successful payment
  static async handlePaymentSuccess(paymentIntent: any): Promise<void> {
    const bookingId = paymentIntent.metadata?.booking_id
    const customerEmail = paymentIntent.receipt_email || paymentIntent.metadata?.customer_email

    try {
      // Update booking status to confirmed
      if (bookingId) {
        await BookingsService.updateBooking(Number(bookingId), {
          status: 'confirmed' as any,
          payment_status: 'paid',
          payment_intent_id: paymentIntent.id,
        })

        // Load booking details for notifications
        const booking = await BookingsService.getBooking(Number(bookingId))
        
        // Send confirmation email
        if (customerEmail) {
          await NotificationsService.sendBookingConfirmation(customerEmail, {
            confirmation_number: `ET${bookingId}${Date.now().toString().slice(-4)}`,
            destination_name: paymentIntent.metadata?.destination_name || 'Your Destination',
            check_in_date: booking.check_in_date,
            check_out_date: booking.check_out_date,
            guests: booking.guests,
            total_price: booking.total_price,
          })
        }

        // Send SMS notification if phone number is available
        const customerPhone = paymentIntent.metadata?.customer_phone
        if (customerPhone) {
          await NotificationsService.sendBookingSMS(
            customerPhone,
            `Your booking has been confirmed! Confirmation: ET${bookingId}${Date.now().toString().slice(-4)}`
          )
        }

        // Log successful payment
        console.log(`Payment successful for booking ${bookingId}: ${paymentIntent.id}`)
      }
    } catch (error) {
      console.error('Error handling payment success:', error)
      // In production, you might want to retry or alert administrators
    }
  }

  // Handle failed payment
  static async handlePaymentFailure(paymentIntent: any): Promise<void> {
    const bookingId = paymentIntent.metadata?.booking_id
    const customerEmail = paymentIntent.receipt_email || paymentIntent.metadata?.customer_email

    try {
      // Update booking status to payment failed
      if (bookingId) {
        await BookingsService.updateBooking(Number(bookingId), {
          status: 'payment_failed' as any,
          payment_status: 'failed',
          payment_intent_id: paymentIntent.id,
        })

        // Send payment failure notification
        if (customerEmail) {
          await NotificationsService.sendEmail({
            to: customerEmail,
            template_id: 'payment_failed',
            variables: {
              booking_id: bookingId,
              failure_reason: paymentIntent.last_payment_error?.message || 'Payment was declined',
              retry_link: `${process.env.NEXT_PUBLIC_APP_URL}/booking/retry/${bookingId}`,
            },
          })
        }

        console.log(`Payment failed for booking ${bookingId}: ${paymentIntent.id}`)
      }
    } catch (error) {
      console.error('Error handling payment failure:', error)
    }
  }

  // Handle payment cancellation
  static async handlePaymentCancellation(paymentIntent: any): Promise<void> {
    const bookingId = paymentIntent.metadata?.booking_id

    try {
      if (bookingId) {
        await BookingsService.updateBooking(Number(bookingId), {
          status: 'cancelled' as any,
          payment_status: 'cancelled',
          payment_intent_id: paymentIntent.id,
        })

        console.log(`Payment cancelled for booking ${bookingId}: ${paymentIntent.id}`)
      }
    } catch (error) {
      console.error('Error handling payment cancellation:', error)
    }
  }

  // Handle charge disputes
  static async handleChargeDispute(dispute: any): Promise<void> {
    try {
      // Notify administrators about the dispute
      await NotificationsService.sendEmail({
        to: process.env.ADMIN_EMAIL || 'admin@exotictravel.com',
        template_id: 'charge_dispute',
        variables: {
          dispute_id: dispute.id,
          charge_id: dispute.charge,
          amount: dispute.amount,
          reason: dispute.reason,
          status: dispute.status,
        },
      })

      // Log dispute for investigation
      console.log(`Charge dispute created: ${dispute.id} for charge ${dispute.charge}`)
    } catch (error) {
      console.error('Error handling charge dispute:', error)
    }
  }

  // Handle subscription creation (for future premium features)
  static async handleSubscriptionCreated(subscription: any): Promise<void> {
    try {
      const customerId = subscription.customer
      const customerEmail = subscription.metadata?.customer_email

      // Update user subscription status
      // This would integrate with your user management system

      // Send welcome email for premium subscription
      if (customerEmail) {
        await NotificationsService.sendEmail({
          to: customerEmail,
          template_id: 'subscription_welcome',
          variables: {
            subscription_id: subscription.id,
            plan_name: subscription.items.data[0]?.price?.nickname || 'Premium Plan',
            billing_cycle: subscription.items.data[0]?.price?.recurring?.interval || 'month',
          },
        })
      }

      console.log(`Subscription created: ${subscription.id} for customer ${customerId}`)
    } catch (error) {
      console.error('Error handling subscription creation:', error)
    }
  }

  // Handle subscription cancellation
  static async handleSubscriptionCanceled(subscription: any): Promise<void> {
    try {
      const customerId = subscription.customer
      const customerEmail = subscription.metadata?.customer_email

      // Update user subscription status
      // This would integrate with your user management system

      // Send cancellation confirmation
      if (customerEmail) {
        await NotificationsService.sendEmail({
          to: customerEmail,
          template_id: 'subscription_cancelled',
          variables: {
            subscription_id: subscription.id,
            cancellation_date: new Date(subscription.canceled_at * 1000).toISOString(),
            access_until: new Date(subscription.current_period_end * 1000).toISOString(),
          },
        })
      }

      console.log(`Subscription cancelled: ${subscription.id} for customer ${customerId}`)
    } catch (error) {
      console.error('Error handling subscription cancellation:', error)
    }
  }

  // Handle successful invoice payment
  static async handleInvoicePaymentSuccess(invoice: any): Promise<void> {
    try {
      const customerId = invoice.customer
      const customerEmail = invoice.customer_email

      // Send invoice receipt
      if (customerEmail) {
        await NotificationsService.sendEmail({
          to: customerEmail,
          template_id: 'invoice_receipt',
          variables: {
            invoice_id: invoice.id,
            amount_paid: invoice.amount_paid,
            currency: invoice.currency,
            payment_date: new Date(invoice.status_transitions.paid_at * 1000).toISOString(),
            invoice_url: invoice.hosted_invoice_url,
          },
        })
      }

      console.log(`Invoice payment successful: ${invoice.id} for customer ${customerId}`)
    } catch (error) {
      console.error('Error handling invoice payment success:', error)
    }
  }

  // Handle failed invoice payment
  static async handleInvoicePaymentFailure(invoice: any): Promise<void> {
    try {
      const customerId = invoice.customer
      const customerEmail = invoice.customer_email

      // Send payment failure notification
      if (customerEmail) {
        await NotificationsService.sendEmail({
          to: customerEmail,
          template_id: 'invoice_payment_failed',
          variables: {
            invoice_id: invoice.id,
            amount_due: invoice.amount_due,
            currency: invoice.currency,
            due_date: new Date(invoice.due_date * 1000).toISOString(),
            payment_url: invoice.hosted_invoice_url,
          },
        })
      }

      console.log(`Invoice payment failed: ${invoice.id} for customer ${customerId}`)
    } catch (error) {
      console.error('Error handling invoice payment failure:', error)
    }
  }

  // Verify webhook signature (for production security)
  static verifyWebhookSignature(
    payload: string,
    signature: string,
    secret: string
  ): boolean {
    // In production, implement proper Stripe webhook signature verification
    // This is a simplified version for demonstration
    try {
      // Stripe uses HMAC SHA256 for webhook signatures
      const crypto = require('crypto')
      const expectedSignature = crypto
        .createHmac('sha256', secret)
        .update(payload, 'utf8')
        .digest('hex')
      
      return crypto.timingSafeEqual(
        Buffer.from(signature, 'hex'),
        Buffer.from(expectedSignature, 'hex')
      )
    } catch (error) {
      console.error('Error verifying webhook signature:', error)
      return false
    }
  }

  // Process webhook with error handling and retries
  static async processWebhookSafely(
    event: WebhookEvent,
    maxRetries: number = 3
  ): Promise<boolean> {
    let attempts = 0
    
    while (attempts < maxRetries) {
      try {
        await this.processStripeWebhook(event)
        return true
      } catch (error) {
        attempts++
        console.error(`Webhook processing attempt ${attempts} failed:`, error)
        
        if (attempts >= maxRetries) {
          // Log final failure for manual investigation
          console.error(`Webhook processing failed after ${maxRetries} attempts:`, {
            event_id: event.id,
            event_type: event.type,
            error: error,
          })
          
          // In production, you might want to store failed webhooks for retry
          return false
        }
        
        // Wait before retrying (exponential backoff)
        await new Promise(resolve => setTimeout(resolve, Math.pow(2, attempts) * 1000))
      }
    }
    
    return false
  }
}
