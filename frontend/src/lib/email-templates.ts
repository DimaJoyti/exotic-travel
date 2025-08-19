export interface EmailTemplate {
  id: string
  name: string
  subject: string
  html: string
  text: string
  variables: string[]
}

export class EmailTemplates {
  // Welcome email template
  static getWelcomeTemplate(): EmailTemplate {
    return {
      id: 'welcome',
      name: 'Welcome Email',
      subject: 'Welcome to ExoticTravel, {{first_name}}!',
      html: `
        <!DOCTYPE html>
        <html>
        <head>
          <meta charset="utf-8">
          <meta name="viewport" content="width=device-width, initial-scale=1.0">
          <title>Welcome to ExoticTravel</title>
          <style>
            body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; }
            .container { max-width: 600px; margin: 0 auto; padding: 20px; }
            .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
            .content { background: white; padding: 30px; border: 1px solid #e1e5e9; }
            .footer { background: #f8f9fa; padding: 20px; text-align: center; border-radius: 0 0 8px 8px; }
            .btn { display: inline-block; background: #667eea; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
            .feature { margin: 20px 0; padding: 15px; background: #f8f9fa; border-radius: 6px; }
          </style>
        </head>
        <body>
          <div class="container">
            <div class="header">
              <h1>🌍 Welcome to ExoticTravel!</h1>
              <p>Your gateway to extraordinary adventures</p>
            </div>
            
            <div class="content">
              <h2>Hello {{first_name}}!</h2>
              
              <p>Welcome to ExoticTravel! We're thrilled to have you join our community of adventurous travelers who seek unique and unforgettable experiences around the world.</p>
              
              <div class="feature">
                <h3>🏝️ Discover Amazing Destinations</h3>
                <p>Explore our curated collection of exotic destinations, from tropical paradises to cultural wonders.</p>
              </div>
              
              <div class="feature">
                <h3>✈️ Easy Booking Process</h3>
                <p>Book your dream vacation in just a few clicks with our streamlined booking system.</p>
              </div>
              
              <div class="feature">
                <h3>🎯 Personalized Recommendations</h3>
                <p>Get destination recommendations tailored to your preferences and travel style.</p>
              </div>
              
              <a href="{{app_url}}/destinations" class="btn">Start Exploring</a>
              
              <p>If you have any questions, our support team is here to help. Just reply to this email or visit our help center.</p>
              
              <p>Happy travels!<br>The ExoticTravel Team</p>
            </div>
            
            <div class="footer">
              <p>&copy; 2024 ExoticTravel. All rights reserved.</p>
              <p><a href="{{app_url}}/unsubscribe">Unsubscribe</a> | <a href="{{app_url}}/contact">Contact Us</a></p>
            </div>
          </div>
        </body>
        </html>
      `,
      text: `
        Welcome to ExoticTravel, {{first_name}}!
        
        We're thrilled to have you join our community of adventurous travelers who seek unique and unforgettable experiences around the world.
        
        What you can do with ExoticTravel:
        
        🏝️ Discover Amazing Destinations
        Explore our curated collection of exotic destinations, from tropical paradises to cultural wonders.
        
        ✈️ Easy Booking Process
        Book your dream vacation in just a few clicks with our streamlined booking system.
        
        🎯 Personalized Recommendations
        Get destination recommendations tailored to your preferences and travel style.
        
        Start exploring: {{app_url}}/destinations
        
        If you have any questions, our support team is here to help. Just reply to this email or visit our help center.
        
        Happy travels!
        The ExoticTravel Team
        
        © 2024 ExoticTravel. All rights reserved.
        Unsubscribe: {{app_url}}/unsubscribe
      `,
      variables: ['first_name', 'last_name', 'app_url']
    }
  }

  // Booking confirmation template
  static getBookingConfirmationTemplate(): EmailTemplate {
    return {
      id: 'booking_confirmation',
      name: 'Booking Confirmation',
      subject: 'Booking Confirmed: {{destination_name}} - {{confirmation_number}}',
      html: `
        <!DOCTYPE html>
        <html>
        <head>
          <meta charset="utf-8">
          <meta name="viewport" content="width=device-width, initial-scale=1.0">
          <title>Booking Confirmation</title>
          <style>
            body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; }
            .container { max-width: 600px; margin: 0 auto; padding: 20px; }
            .header { background: linear-gradient(135deg, #10b981 0%, #059669 100%); color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
            .content { background: white; padding: 30px; border: 1px solid #e1e5e9; }
            .footer { background: #f8f9fa; padding: 20px; text-align: center; border-radius: 0 0 8px 8px; }
            .booking-details { background: #f0fdf4; border: 1px solid #bbf7d0; padding: 20px; border-radius: 6px; margin: 20px 0; }
            .detail-row { display: flex; justify-content: space-between; margin: 10px 0; padding: 8px 0; border-bottom: 1px solid #e5e7eb; }
            .detail-label { font-weight: bold; color: #374151; }
            .detail-value { color: #1f2937; }
            .total { background: #667eea; color: white; padding: 15px; border-radius: 6px; text-align: center; font-size: 18px; font-weight: bold; }
            .btn { display: inline-block; background: #10b981; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
          </style>
        </head>
        <body>
          <div class="container">
            <div class="header">
              <h1>✅ Booking Confirmed!</h1>
              <p>Your adventure awaits</p>
            </div>
            
            <div class="content">
              <h2>Great news! Your booking is confirmed.</h2>
              
              <p>Thank you for choosing ExoticTravel. We're excited to help you create unforgettable memories on your upcoming trip.</p>
              
              <div class="booking-details">
                <h3>📋 Booking Details</h3>
                
                <div class="detail-row">
                  <span class="detail-label">Confirmation Number:</span>
                  <span class="detail-value">{{confirmation_number}}</span>
                </div>
                
                <div class="detail-row">
                  <span class="detail-label">Destination:</span>
                  <span class="detail-value">{{destination_name}}</span>
                </div>
                
                <div class="detail-row">
                  <span class="detail-label">Check-in Date:</span>
                  <span class="detail-value">{{check_in_date}}</span>
                </div>
                
                <div class="detail-row">
                  <span class="detail-label">Check-out Date:</span>
                  <span class="detail-value">{{check_out_date}}</span>
                </div>
                
                <div class="detail-row">
                  <span class="detail-label">Number of Guests:</span>
                  <span class="detail-value">{{guests}}</span>
                </div>
                
                <div class="total">
                  Total Amount: $\${total_price}
                </div>
              </div>
              
              <h3>📱 What's Next?</h3>
              <ul>
                <li>You'll receive detailed travel information 7 days before departure</li>
                <li>Check your email for any updates or changes</li>
                <li>Contact us if you have any questions or special requests</li>
                <li>Don't forget to check passport and visa requirements</li>
              </ul>
              
              <a href="{{app_url}}/bookings/{{booking_id}}" class="btn">View Booking Details</a>
              
              <p>We can't wait to help you create amazing memories!</p>
              
              <p>Safe travels,<br>The ExoticTravel Team</p>
            </div>
            
            <div class="footer">
              <p>&copy; 2024 ExoticTravel. All rights reserved.</p>
              <p><a href="{{app_url}}/contact">Contact Support</a> | <a href="{{app_url}}/bookings">My Bookings</a></p>
            </div>
          </div>
        </body>
        </html>
      `,
      text: `
        Booking Confirmed!
        
        Great news! Your booking is confirmed.
        
        Thank you for choosing ExoticTravel. We're excited to help you create unforgettable memories on your upcoming trip.
        
        BOOKING DETAILS:
        Confirmation Number: {{confirmation_number}}
        Destination: {{destination_name}}
        Check-in Date: {{check_in_date}}
        Check-out Date: {{check_out_date}}
        Number of Guests: {{guests}}
        Total Amount: $\${total_price}
        
        What's Next?
        - You'll receive detailed travel information 7 days before departure
        - Check your email for any updates or changes
        - Contact us if you have any questions or special requests
        - Don't forget to check passport and visa requirements
        
        View your booking: {{app_url}}/bookings/{{booking_id}}
        
        We can't wait to help you create amazing memories!
        
        Safe travels,
        The ExoticTravel Team
        
        © 2024 ExoticTravel. All rights reserved.
        Contact Support: {{app_url}}/contact
      `,
      variables: ['confirmation_number', 'destination_name', 'check_in_date', 'check_out_date', 'guests', 'total_price', 'booking_id', 'app_url']
    }
  }

  // Booking reminder template
  static getBookingReminderTemplate(): EmailTemplate {
    return {
      id: 'booking_reminder',
      name: 'Booking Reminder',
      subject: 'Your trip to {{destination_name}} is coming up! - {{confirmation_number}}',
      html: `
        <!DOCTYPE html>
        <html>
        <head>
          <meta charset="utf-8">
          <meta name="viewport" content="width=device-width, initial-scale=1.0">
          <title>Trip Reminder</title>
          <style>
            body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; }
            .container { max-width: 600px; margin: 0 auto; padding: 20px; }
            .header { background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%); color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
            .content { background: white; padding: 30px; border: 1px solid #e1e5e9; }
            .footer { background: #f8f9fa; padding: 20px; text-align: center; border-radius: 0 0 8px 8px; }
            .countdown { background: #fef3c7; border: 1px solid #fbbf24; padding: 20px; border-radius: 6px; text-align: center; margin: 20px 0; }
            .checklist { background: #f0f9ff; border: 1px solid #7dd3fc; padding: 20px; border-radius: 6px; margin: 20px 0; }
            .btn { display: inline-block; background: #f59e0b; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
          </style>
        </head>
        <body>
          <div class="container">
            <div class="header">
              <h1>🎒 Trip Reminder</h1>
              <p>Your adventure is almost here!</p>
            </div>
            
            <div class="content">
              <h2>Get ready for {{destination_name}}!</h2>
              
              <div class="countdown">
                <h3>⏰ Only {{days_until_trip}} days to go!</h3>
                <p>Your departure date: {{check_in_date}}</p>
              </div>
              
              <p>We're excited that your trip is approaching! Here's everything you need to know to prepare for your amazing adventure.</p>
              
              <div class="checklist">
                <h3>📝 Pre-Travel Checklist</h3>
                <ul>
                  <li>✈️ Check passport expiration (should be valid for 6+ months)</li>
                  <li>🛂 Verify visa requirements for your destination</li>
                  <li>💉 Review recommended vaccinations</li>
                  <li>🧳 Start packing (check weather forecast)</li>
                  <li>💳 Notify your bank of travel plans</li>
                  <li>📱 Download offline maps and translation apps</li>
                  <li>🏥 Consider travel insurance</li>
                  <li>📋 Make copies of important documents</li>
                </ul>
              </div>
              
              <h3>📞 Need Help?</h3>
              <p>Our travel experts are here to assist you with any questions or concerns. Don't hesitate to reach out!</p>
              
              <a href="{{app_url}}/bookings/{{booking_id}}" class="btn">View Trip Details</a>
              
              <p>We can't wait for you to experience this incredible destination!</p>
              
              <p>Safe travels,<br>The ExoticTravel Team</p>
            </div>
            
            <div class="footer">
              <p>&copy; 2024 ExoticTravel. All rights reserved.</p>
              <p><a href="{{app_url}}/contact">Contact Support</a> | <a href="{{app_url}}/bookings">My Bookings</a></p>
            </div>
          </div>
        </body>
        </html>
      `,
      text: `
        Trip Reminder - Your adventure is almost here!
        
        Get ready for {{destination_name}}!
        
        Only {{days_until_trip}} days to go!
        Your departure date: {{check_in_date}}
        
        We're excited that your trip is approaching! Here's everything you need to know to prepare for your amazing adventure.
        
        PRE-TRAVEL CHECKLIST:
        ✈️ Check passport expiration (should be valid for 6+ months)
        🛂 Verify visa requirements for your destination
        💉 Review recommended vaccinations
        🧳 Start packing (check weather forecast)
        💳 Notify your bank of travel plans
        📱 Download offline maps and translation apps
        🏥 Consider travel insurance
        📋 Make copies of important documents
        
        Need Help?
        Our travel experts are here to assist you with any questions or concerns. Don't hesitate to reach out!
        
        View trip details: {{app_url}}/bookings/{{booking_id}}
        
        We can't wait for you to experience this incredible destination!
        
        Safe travels,
        The ExoticTravel Team
        
        © 2024 ExoticTravel. All rights reserved.
        Contact Support: {{app_url}}/contact
      `,
      variables: ['destination_name', 'days_until_trip', 'check_in_date', 'confirmation_number', 'booking_id', 'app_url']
    }
  }

  // Password reset template
  static getPasswordResetTemplate(): EmailTemplate {
    return {
      id: 'password_reset',
      name: 'Password Reset',
      subject: 'Reset your ExoticTravel password',
      html: `
        <!DOCTYPE html>
        <html>
        <head>
          <meta charset="utf-8">
          <meta name="viewport" content="width=device-width, initial-scale=1.0">
          <title>Password Reset</title>
          <style>
            body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; }
            .container { max-width: 600px; margin: 0 auto; padding: 20px; }
            .header { background: linear-gradient(135deg, #dc2626 0%, #b91c1c 100%); color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
            .content { background: white; padding: 30px; border: 1px solid #e1e5e9; }
            .footer { background: #f8f9fa; padding: 20px; text-align: center; border-radius: 0 0 8px 8px; }
            .btn { display: inline-block; background: #dc2626; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; margin: 20px 0; }
            .warning { background: #fef2f2; border: 1px solid #fecaca; padding: 15px; border-radius: 6px; margin: 20px 0; }
          </style>
        </head>
        <body>
          <div class="container">
            <div class="header">
              <h1>🔐 Password Reset</h1>
              <p>Secure your account</p>
            </div>
            
            <div class="content">
              <h2>Hello {{first_name}},</h2>
              
              <p>We received a request to reset the password for your ExoticTravel account. If you made this request, click the button below to reset your password.</p>
              
              <a href="{{reset_link}}" class="btn">Reset Password</a>
              
              <div class="warning">
                <h3>⚠️ Important Security Information</h3>
                <ul>
                  <li>This link will expire in 1 hour for security reasons</li>
                  <li>If you didn't request this reset, please ignore this email</li>
                  <li>Never share this link with anyone</li>
                  <li>ExoticTravel will never ask for your password via email</li>
                </ul>
              </div>
              
              <p>If the button doesn't work, copy and paste this link into your browser:</p>
              <p style="word-break: break-all; color: #667eea;">{{reset_link}}</p>
              
              <p>If you didn't request a password reset, you can safely ignore this email. Your password will remain unchanged.</p>
              
              <p>For security questions, contact our support team.</p>
              
              <p>Best regards,<br>The ExoticTravel Security Team</p>
            </div>
            
            <div class="footer">
              <p>&copy; 2024 ExoticTravel. All rights reserved.</p>
              <p><a href="{{app_url}}/contact">Contact Support</a> | <a href="{{app_url}}/security">Security Center</a></p>
            </div>
          </div>
        </body>
        </html>
      `,
      text: `
        Password Reset - ExoticTravel
        
        Hello {{first_name}},
        
        We received a request to reset the password for your ExoticTravel account. If you made this request, use the link below to reset your password.
        
        Reset your password: {{reset_link}}
        
        IMPORTANT SECURITY INFORMATION:
        - This link will expire in 1 hour for security reasons
        - If you didn't request this reset, please ignore this email
        - Never share this link with anyone
        - ExoticTravel will never ask for your password via email
        
        If you didn't request a password reset, you can safely ignore this email. Your password will remain unchanged.
        
        For security questions, contact our support team.
        
        Best regards,
        The ExoticTravel Security Team
        
        © 2024 ExoticTravel. All rights reserved.
        Contact Support: {{app_url}}/contact
      `,
      variables: ['first_name', 'reset_link', 'expires_at', 'app_url']
    }
  }

  // Get all templates
  static getAllTemplates(): EmailTemplate[] {
    return [
      this.getWelcomeTemplate(),
      this.getBookingConfirmationTemplate(),
      this.getBookingReminderTemplate(),
      this.getPasswordResetTemplate(),
    ]
  }

  // Get template by ID
  static getTemplateById(id: string): EmailTemplate | null {
    const templates = this.getAllTemplates()
    return templates.find(template => template.id === id) || null
  }

  // Replace variables in template
  static replaceVariables(template: string, variables: Record<string, any>): string {
    let result = template
    
    Object.keys(variables).forEach(key => {
      const regex = new RegExp(`{{${key}}}`, 'g')
      result = result.replace(regex, variables[key] || '')
    })
    
    return result
  }

  // Validate template variables
  static validateVariables(templateId: string, variables: Record<string, any>): { valid: boolean; missing: string[] } {
    const template = this.getTemplateById(templateId)
    if (!template) {
      return { valid: false, missing: ['Template not found'] }
    }
    
    const missing = template.variables.filter(variable => 
      variables[variable] === undefined || variables[variable] === null
    )
    
    return {
      valid: missing.length === 0,
      missing
    }
  }
}
