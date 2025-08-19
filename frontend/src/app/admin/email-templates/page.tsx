'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { Mail, Eye, Send, Settings, RefreshCw } from 'lucide-react'
import { useAuth } from '@/contexts/auth-context'
import { EmailTemplates } from '@/lib/email-templates'
import { NotificationsService } from '@/lib/notifications'
import AdminLayout from '@/components/admin/admin-layout'
import EmailPreview from '@/components/email/email-preview'

export default function EmailTemplatesPage() {
  const { user } = useAuth()
  const router = useRouter()
  const [selectedTemplate, setSelectedTemplate] = useState('welcome')
  const [testVariables, setTestVariables] = useState<Record<string, string>>({})
  const [testEmail, setTestEmail] = useState('')
  const [sending, setSending] = useState(false)

  useEffect(() => {
    if (!user) {
      router.push('/auth/login')
      return
    }

    if (user.role !== 'admin') {
      router.push('/dashboard')
      return
    }

    // Initialize test variables for the selected template
    initializeTestVariables()
  }, [user, router, selectedTemplate])

  const initializeTestVariables = () => {
    const template = EmailTemplates.getTemplateById(selectedTemplate)
    if (!template) return

    const defaultVariables: Record<string, string> = {
      first_name: 'John',
      last_name: 'Doe',
      destination_name: 'Maldives Paradise Resort',
      confirmation_number: 'ET123456789',
      check_in_date: '2024-06-15',
      check_out_date: '2024-06-22',
      guests: '2',
      total_price: '5000',
      booking_id: '123',
      days_until_trip: '7',
      reset_link: 'https://exotictravel.com/reset-password?token=abc123',
      expires_at: new Date(Date.now() + 3600000).toISOString(), // 1 hour from now
    }

    const variables: Record<string, string> = {}
    template.variables.forEach(variable => {
      variables[variable] = defaultVariables[variable] || `[${variable}]`
    })

    setTestVariables(variables)
  }

  const handleSendTestEmail = async () => {
    if (!testEmail) {
      alert('Please enter a test email address')
      return
    }

    setSending(true)
    try {
      // Convert string values to appropriate types
      const processedVariables: Record<string, any> = {}
      Object.keys(testVariables).forEach(key => {
        const value = testVariables[key]
        // Try to convert numbers
        if (!isNaN(Number(value)) && value !== '') {
          processedVariables[key] = Number(value)
        } else {
          processedVariables[key] = value
        }
      })

      await NotificationsService.sendTemplatedEmail(
        testEmail,
        selectedTemplate,
        processedVariables
      )

      alert(`Test email sent successfully to ${testEmail}`)
    } catch (error) {
      console.error('Error sending test email:', error)
      alert('Failed to send test email. Check console for details.')
    } finally {
      setSending(false)
    }
  }

  const templates = EmailTemplates.getAllTemplates()

  return (
    <AdminLayout>
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Email Templates</h1>
            <p className="text-gray-600 mt-1">Preview and test email templates</p>
          </div>
          <div className="flex space-x-3">
            <button
              onClick={initializeTestVariables}
              className="flex items-center px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
            >
              <RefreshCw className="h-4 w-4 mr-2" />
              Reset Variables
            </button>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
        {/* Template Selection */}
        <div className="lg:col-span-1">
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Templates</h3>
            
            <div className="space-y-2">
              {templates.map((template) => (
                <button
                  key={template.id}
                  onClick={() => setSelectedTemplate(template.id)}
                  className={`w-full text-left p-3 rounded-lg transition-colors ${
                    selectedTemplate === template.id
                      ? 'bg-primary text-primary-foreground'
                      : 'hover:bg-gray-50 text-gray-700'
                  }`}
                >
                  <div className="font-medium">{template.name}</div>
                  <div className="text-sm opacity-75">{template.id}</div>
                </button>
              ))}
            </div>
          </div>

          {/* Test Email Form */}
          <div className="bg-white rounded-lg border border-gray-200 p-4 mt-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Send Test Email</h3>
            
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Test Email Address
                </label>
                <input
                  type="email"
                  value={testEmail}
                  onChange={(e) => setTestEmail(e.target.value)}
                  placeholder="test@example.com"
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
                />
              </div>
              
              <button
                onClick={handleSendTestEmail}
                disabled={sending || !testEmail}
                className="w-full flex items-center justify-center px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {sending ? (
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                ) : (
                  <Send className="h-4 w-4 mr-2" />
                )}
                {sending ? 'Sending...' : 'Send Test'}
              </button>
            </div>
          </div>
        </div>

        {/* Template Variables */}
        <div className="lg:col-span-1">
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Template Variables</h3>
            
            <div className="space-y-3">
              {Object.keys(testVariables).map((variable) => (
                <div key={variable}>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    {variable}
                  </label>
                  <input
                    type="text"
                    value={testVariables[variable]}
                    onChange={(e) => setTestVariables(prev => ({
                      ...prev,
                      [variable]: e.target.value
                    }))}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent text-sm"
                  />
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Email Preview */}
        <div className="lg:col-span-2">
          <EmailPreview
            templateId={selectedTemplate}
            variables={testVariables}
          />
        </div>
      </div>

      {/* Email Statistics */}
      <div className="mt-12">
        <h2 className="text-2xl font-bold text-gray-900 mb-6">Email Statistics</h2>
        
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-2 bg-blue-100 rounded-lg">
                <Mail className="h-6 w-6 text-blue-600" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">Emails Sent Today</p>
                <p className="text-2xl font-bold text-gray-900">247</p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-2 bg-green-100 rounded-lg">
                <Eye className="h-6 w-6 text-green-600" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">Open Rate</p>
                <p className="text-2xl font-bold text-gray-900">68.5%</p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-2 bg-purple-100 rounded-lg">
                <Settings className="h-6 w-6 text-purple-600" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">Click Rate</p>
                <p className="text-2xl font-bold text-gray-900">12.3%</p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-2 bg-red-100 rounded-lg">
                <RefreshCw className="h-6 w-6 text-red-600" />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">Bounce Rate</p>
                <p className="text-2xl font-bold text-gray-900">2.1%</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </AdminLayout>
  )
}
