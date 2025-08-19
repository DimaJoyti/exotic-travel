'use client'

import { useState } from 'react'
import { Eye, Code, Mail, Download, Send } from 'lucide-react'
import { EmailTemplates, EmailTemplate } from '@/lib/email-templates'

interface EmailPreviewProps {
  templateId: string
  variables?: Record<string, any>
  className?: string
}

export default function EmailPreview({ 
  templateId, 
  variables = {}, 
  className = '' 
}: EmailPreviewProps) {
  const [viewMode, setViewMode] = useState<'preview' | 'html' | 'text'>('preview')
  const [testEmail, setTestEmail] = useState('')
  const [sending, setSending] = useState(false)

  const template = EmailTemplates.getTemplateById(templateId)
  
  if (!template) {
    return (
      <div className={`bg-red-50 border border-red-200 rounded-lg p-4 ${className}`}>
        <p className="text-red-700">Template '{templateId}' not found</p>
      </div>
    )
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

  const handleSendTest = async () => {
    if (!testEmail) {
      alert('Please enter a test email address')
      return
    }

    setSending(true)
    try {
      // In a real implementation, this would send a test email
      console.log('Sending test email to:', testEmail)
      await new Promise(resolve => setTimeout(resolve, 1000)) // Simulate API call
      alert(`Test email sent to ${testEmail}`)
    } catch (error) {
      alert('Failed to send test email')
    } finally {
      setSending(false)
    }
  }

  const handleDownloadHtml = () => {
    const blob = new Blob([html], { type: 'text/html' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${templateId}-email.html`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  return (
    <div className={`bg-white rounded-lg border border-gray-200 ${className}`}>
      {/* Header */}
      <div className="border-b border-gray-200 p-4">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h3 className="text-lg font-semibold text-gray-900">{template.name}</h3>
            <p className="text-sm text-gray-600">Template ID: {template.id}</p>
          </div>
          
          <div className="flex space-x-2">
            <button
              onClick={handleDownloadHtml}
              className="flex items-center px-3 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
            >
              <Download className="h-4 w-4 mr-2" />
              Download HTML
            </button>
          </div>
        </div>

        {/* Subject Line */}
        <div className="mb-4">
          <label className="block text-sm font-medium text-gray-700 mb-1">Subject Line</label>
          <div className="bg-gray-50 border border-gray-200 rounded-lg p-3">
            <p className="text-gray-900 font-medium">{subject}</p>
          </div>
        </div>

        {/* View Mode Tabs */}
        <div className="flex space-x-1">
          <button
            onClick={() => setViewMode('preview')}
            className={`flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-colors ${
              viewMode === 'preview'
                ? 'bg-primary text-primary-foreground'
                : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
            }`}
          >
            <Eye className="h-4 w-4 mr-2" />
            Preview
          </button>
          <button
            onClick={() => setViewMode('html')}
            className={`flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-colors ${
              viewMode === 'html'
                ? 'bg-primary text-primary-foreground'
                : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
            }`}
          >
            <Code className="h-4 w-4 mr-2" />
            HTML
          </button>
          <button
            onClick={() => setViewMode('text')}
            className={`flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-colors ${
              viewMode === 'text'
                ? 'bg-primary text-primary-foreground'
                : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
            }`}
          >
            <Mail className="h-4 w-4 mr-2" />
            Text
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="p-4">
        {viewMode === 'preview' && (
          <div className="border border-gray-200 rounded-lg overflow-hidden">
            <iframe
              srcDoc={html}
              className="w-full h-96"
              title="Email Preview"
              sandbox="allow-same-origin"
            />
          </div>
        )}

        {viewMode === 'html' && (
          <div className="bg-gray-900 text-gray-100 rounded-lg p-4 overflow-auto max-h-96">
            <pre className="text-sm whitespace-pre-wrap">{html}</pre>
          </div>
        )}

        {viewMode === 'text' && (
          <div className="bg-gray-50 border border-gray-200 rounded-lg p-4 max-h-96 overflow-auto">
            <pre className="text-sm whitespace-pre-wrap text-gray-900">{text}</pre>
          </div>
        )}
      </div>

      {/* Test Email Section */}
      <div className="border-t border-gray-200 p-4">
        <h4 className="text-sm font-medium text-gray-900 mb-3">Send Test Email</h4>
        <div className="flex space-x-3">
          <input
            type="email"
            value={testEmail}
            onChange={(e) => setTestEmail(e.target.value)}
            placeholder="Enter test email address"
            className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
          />
          <button
            onClick={handleSendTest}
            disabled={sending || !testEmail}
            className="flex items-center px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
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

      {/* Variables Info */}
      {template.variables.length > 0 && (
        <div className="border-t border-gray-200 p-4">
          <h4 className="text-sm font-medium text-gray-900 mb-3">Template Variables</h4>
          <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
            {template.variables.map((variable) => (
              <div key={variable} className="bg-gray-50 rounded-lg p-2">
                <code className="text-sm text-gray-700">{`{{${variable}}}`}</code>
                <div className="text-xs text-gray-500 mt-1">
                  {(allVariables as any)[variable] || 'Not provided'}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

// Component for listing all available templates
export function EmailTemplatesList({ className = '' }: { className?: string }) {
  const templates = EmailTemplates.getAllTemplates()

  return (
    <div className={`space-y-4 ${className}`}>
      <h2 className="text-xl font-semibold text-gray-900">Email Templates</h2>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {templates.map((template) => (
          <div key={template.id} className="bg-white border border-gray-200 rounded-lg p-4">
            <div className="flex items-center justify-between mb-2">
              <h3 className="font-semibold text-gray-900">{template.name}</h3>
              <span className="text-xs bg-gray-100 text-gray-600 px-2 py-1 rounded">
                {template.id}
              </span>
            </div>
            
            <p className="text-sm text-gray-600 mb-3">{template.subject}</p>
            
            <div className="flex items-center justify-between">
              <div className="text-xs text-gray-500">
                {template.variables.length} variable{template.variables.length !== 1 ? 's' : ''}
              </div>
              
              <button className="text-sm text-primary hover:text-primary/80 font-medium">
                Preview →
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
