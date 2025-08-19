'use client'

import React, { useState, useRef, useEffect } from 'react'
import { Mic, MicOff, Volume2, VolumeX, Zap, Brain, Sparkles, Loader2 } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { Button } from '@/components/ui/button'
import { FadeIn, ScaleIn } from '@/components/ui/animated'

interface VoiceSearchProps {
  onSearch: (query: string) => void
  onVoiceCommand: (command: VoiceCommand) => void
  className?: string
  language?: string
}

interface VoiceCommand {
  type: 'search' | 'filter' | 'navigate' | 'action'
  command: string
  parameters: Record<string, any>
  confidence: number
}

interface VoiceSession {
  isListening: boolean
  isProcessing: boolean
  isSupported: boolean
  hasPermission: boolean
  error: string | null
  transcript: string
  interimTranscript: string
}

interface VoiceResponse {
  text: string
  action?: () => void
}

export default function VoiceSearch({ 
  onSearch, 
  onVoiceCommand, 
  className = '', 
  language = 'en-US' 
}: VoiceSearchProps) {
  const recognitionRef = useRef<any>(null)
  const synthRef = useRef<SpeechSynthesis | null>(null)
  const timeoutRef = useRef<NodeJS.Timeout>()
  
  const [voiceSession, setVoiceSession] = useState<VoiceSession>({
    isListening: false,
    isProcessing: false,
    isSupported: false,
    hasPermission: false,
    error: null,
    transcript: '',
    interimTranscript: ''
  })
  
  const [voiceEnabled, setVoiceEnabled] = useState(true)
  const [lastCommand, setLastCommand] = useState<VoiceCommand | null>(null)
  const [conversationHistory, setConversationHistory] = useState<string[]>([])

  // Initialize speech recognition
  useEffect(() => {
    initializeSpeechRecognition()
    initializeSpeechSynthesis()
    
    return () => {
      if (recognitionRef.current) {
        recognitionRef.current.stop()
      }
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
  }, [])

  const initializeSpeechRecognition = () => {
    if ('webkitSpeechRecognition' in window || 'SpeechRecognition' in window) {
      const SpeechRecognition = (window as any).SpeechRecognition || (window as any).webkitSpeechRecognition
      recognitionRef.current = new SpeechRecognition()
      
      const recognition = recognitionRef.current
      recognition.continuous = true
      recognition.interimResults = true
      recognition.lang = language
      recognition.maxAlternatives = 3

      recognition.onstart = () => {
        setVoiceSession(prev => ({ 
          ...prev, 
          isListening: true, 
          error: null,
          transcript: '',
          interimTranscript: ''
        }))
      }

      recognition.onresult = (event: any) => {
        let interimTranscript = ''
        let finalTranscript = ''

        for (let i = event.resultIndex; i < event.results.length; i++) {
          const transcript = event.results[i][0].transcript
          if (event.results[i].isFinal) {
            finalTranscript += transcript
          } else {
            interimTranscript += transcript
          }
        }

        setVoiceSession(prev => ({
          ...prev,
          transcript: finalTranscript,
          interimTranscript: interimTranscript
        }))

        if (finalTranscript) {
          processVoiceInput(finalTranscript)
        }
      }

      recognition.onerror = (event: any) => {
        setVoiceSession(prev => ({
          ...prev,
          isListening: false,
          error: `Speech recognition error: ${event.error}`
        }))
      }

      recognition.onend = () => {
        setVoiceSession(prev => ({ ...prev, isListening: false }))
      }

      setVoiceSession(prev => ({ ...prev, isSupported: true }))
    } else {
      setVoiceSession(prev => ({ 
        ...prev, 
        isSupported: false,
        error: 'Speech recognition not supported in this browser'
      }))
    }
  }

  const initializeSpeechSynthesis = () => {
    if ('speechSynthesis' in window) {
      synthRef.current = window.speechSynthesis
    }
  }

  const startListening = async () => {
    if (!recognitionRef.current) return

    try {
      // Request microphone permission
      await navigator.mediaDevices.getUserMedia({ audio: true })
      
      setVoiceSession(prev => ({ ...prev, hasPermission: true }))
      recognitionRef.current.start()
      
      // Auto-stop after 10 seconds of silence
      timeoutRef.current = setTimeout(() => {
        stopListening()
      }, 10000)
      
    } catch (error) {
      setVoiceSession(prev => ({
        ...prev,
        error: 'Microphone permission denied'
      }))
    }
  }

  const stopListening = () => {
    if (recognitionRef.current) {
      recognitionRef.current.stop()
    }
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
    }
  }

  const processVoiceInput = async (transcript: string) => {
    setVoiceSession(prev => ({ ...prev, isProcessing: true }))
    
    try {
      const command = await parseVoiceCommand(transcript)
      setLastCommand(command)
      
      // Add to conversation history
      setConversationHistory(prev => [...prev.slice(-4), transcript])
      
      // Execute command
      switch (command.type) {
        case 'search':
          onSearch(command.parameters.query)
          speakResponse(`Searching for ${command.parameters.query}`)
          break
        case 'filter':
          onVoiceCommand(command)
          speakResponse(`Applied ${command.parameters.filterType} filter`)
          break
        case 'navigate':
          onVoiceCommand(command)
          speakResponse(`Navigating to ${command.parameters.destination}`)
          break
        case 'action':
          onVoiceCommand(command)
          speakResponse(`Executing ${command.command}`)
          break
        default:
          speakResponse("I didn't understand that command. Try saying 'search for beaches' or 'show me luxury destinations'")
      }
      
    } catch (error) {
      console.error('Voice processing error:', error)
      speakResponse("Sorry, I couldn't process that command. Please try again.")
    } finally {
      setVoiceSession(prev => ({ ...prev, isProcessing: false }))
    }
  }

  const parseVoiceCommand = async (transcript: string): Promise<VoiceCommand> => {
    const lowerTranscript = transcript.toLowerCase()
    
    // Search commands
    if (lowerTranscript.includes('search for') || lowerTranscript.includes('find') || lowerTranscript.includes('show me')) {
      const query = extractSearchQuery(lowerTranscript)
      return {
        type: 'search',
        command: transcript,
        parameters: { query },
        confidence: 0.9
      }
    }
    
    // Filter commands
    if (lowerTranscript.includes('filter by') || lowerTranscript.includes('show only')) {
      const filterType = extractFilterType(lowerTranscript)
      const filterValue = extractFilterValue(lowerTranscript)
      return {
        type: 'filter',
        command: transcript,
        parameters: { filterType, filterValue },
        confidence: 0.8
      }
    }
    
    // Navigation commands
    if (lowerTranscript.includes('go to') || lowerTranscript.includes('navigate to') || lowerTranscript.includes('open')) {
      const destination = extractNavigationDestination(lowerTranscript)
      return {
        type: 'navigate',
        command: transcript,
        parameters: { destination },
        confidence: 0.85
      }
    }
    
    // Action commands
    if (lowerTranscript.includes('book') || lowerTranscript.includes('add to wishlist') || lowerTranscript.includes('compare')) {
      const action = extractAction(lowerTranscript)
      return {
        type: 'action',
        command: transcript,
        parameters: { action },
        confidence: 0.75
      }
    }
    
    // Default to search if no specific command detected
    return {
      type: 'search',
      command: transcript,
      parameters: { query: transcript },
      confidence: 0.5
    }
  }

  const extractSearchQuery = (transcript: string): string => {
    const patterns = [
      /search for (.+)/,
      /find (.+)/,
      /show me (.+)/,
      /looking for (.+)/
    ]
    
    for (const pattern of patterns) {
      const match = transcript.match(pattern)
      if (match) return match[1]
    }
    
    return transcript
  }

  const extractFilterType = (transcript: string): string => {
    if (transcript.includes('price')) return 'price'
    if (transcript.includes('duration')) return 'duration'
    if (transcript.includes('country')) return 'country'
    if (transcript.includes('rating')) return 'rating'
    return 'general'
  }

  const extractFilterValue = (transcript: string): string => {
    const match = transcript.match(/\$?(\d+(?:,\d{3})*(?:\.\d{2})?)/);
    if (match) return match[1]
    
    const words = transcript.split(' ')
    return words[words.length - 1]
  }

  const extractNavigationDestination = (transcript: string): string => {
    const patterns = [
      /go to (.+)/,
      /navigate to (.+)/,
      /open (.+)/
    ]
    
    for (const pattern of patterns) {
      const match = transcript.match(pattern)
      if (match) return match[1]
    }
    
    return 'home'
  }

  const extractAction = (transcript: string): string => {
    if (transcript.includes('book')) return 'book'
    if (transcript.includes('wishlist')) return 'wishlist'
    if (transcript.includes('compare')) return 'compare'
    if (transcript.includes('share')) return 'share'
    return 'unknown'
  }

  const speakResponse = (text: string) => {
    if (!voiceEnabled || !synthRef.current) return
    
    // Cancel any ongoing speech
    synthRef.current.cancel()
    
    const utterance = new SpeechSynthesisUtterance(text)
    utterance.rate = 0.9
    utterance.pitch = 1
    utterance.volume = 0.8
    utterance.lang = language
    
    synthRef.current.speak(utterance)
  }

  const toggleVoice = () => {
    setVoiceEnabled(!voiceEnabled)
    if (voiceEnabled) {
      synthRef.current?.cancel()
    }
  }

  return (
    <div className={`space-y-4 ${className}`}>
      {/* Voice Control Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-2">
          <Mic className="h-5 w-5 text-brand-500" />
          <h3 className="font-semibold text-gray-900">Voice Search</h3>
          {voiceSession.isSupported && (
            <span className="text-xs bg-green-100 text-green-700 px-2 py-1 rounded-full">
              Supported
            </span>
          )}
        </div>
        
        <Button
          variant="ghost"
          size="sm"
          onClick={toggleVoice}
          className="text-gray-600"
        >
          {voiceEnabled ? <Volume2 className="h-4 w-4" /> : <VolumeX className="h-4 w-4" />}
        </Button>
      </div>

      {/* Voice Interface */}
      <div className="relative">
        {!voiceSession.isSupported ? (
          <div className="text-center p-6 bg-gray-50 rounded-lg">
            <MicOff className="h-12 w-12 text-gray-400 mx-auto mb-2" />
            <p className="text-gray-600">Voice search not supported in this browser</p>
          </div>
        ) : (
          <div className="text-center p-6 bg-gradient-to-br from-brand-50 to-accent-50 rounded-lg">
            <AnimatePresence mode="wait">
              {voiceSession.isListening ? (
                <motion.div
                  key="listening"
                  initial={{ scale: 0.8, opacity: 0 }}
                  animate={{ scale: 1, opacity: 1 }}
                  exit={{ scale: 0.8, opacity: 0 }}
                >
                  <div className="relative">
                    <div className="w-16 h-16 bg-red-500 rounded-full flex items-center justify-center mx-auto mb-4 animate-pulse">
                      <Mic className="h-8 w-8 text-white" />
                    </div>
                    <div className="absolute inset-0 w-16 h-16 bg-red-500 rounded-full mx-auto animate-ping opacity-30" />
                  </div>
                  <p className="text-gray-900 font-medium mb-2">Listening...</p>
                  {voiceSession.interimTranscript && (
                    <p className="text-gray-600 text-sm italic">"{voiceSession.interimTranscript}"</p>
                  )}
                </motion.div>
              ) : voiceSession.isProcessing ? (
                <motion.div
                  key="processing"
                  initial={{ scale: 0.8, opacity: 0 }}
                  animate={{ scale: 1, opacity: 1 }}
                  exit={{ scale: 0.8, opacity: 0 }}
                >
                  <div className="w-16 h-16 bg-brand-500 rounded-full flex items-center justify-center mx-auto mb-4">
                    <Brain className="h-8 w-8 text-white animate-pulse" />
                  </div>
                  <p className="text-gray-900 font-medium">Processing...</p>
                </motion.div>
              ) : (
                <motion.div
                  key="ready"
                  initial={{ scale: 0.8, opacity: 0 }}
                  animate={{ scale: 1, opacity: 1 }}
                  exit={{ scale: 0.8, opacity: 0 }}
                >
                  <Button
                    onClick={startListening}
                    className="w-16 h-16 rounded-full bg-brand-500 hover:bg-brand-600 mb-4"
                    disabled={!voiceSession.hasPermission && !voiceSession.isSupported}
                  >
                    <Mic className="h-8 w-8" />
                  </Button>
                  <p className="text-gray-900 font-medium mb-1">Tap to speak</p>
                  <p className="text-gray-600 text-sm">
                    Try: "Search for beaches", "Show me luxury hotels", "Filter by price under $2000"
                  </p>
                </motion.div>
              )}
            </AnimatePresence>

            {voiceSession.error && (
              <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-lg">
                <p className="text-red-700 text-sm">{voiceSession.error}</p>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Last Command Display */}
      {lastCommand && (
        <FadeIn>
          <div className="bg-white border border-gray-200 rounded-lg p-4">
            <div className="flex items-center space-x-2 mb-2">
              <Sparkles className="h-4 w-4 text-brand-500" />
              <span className="text-sm font-medium text-gray-900">Last Command</span>
              <span className="text-xs bg-gray-100 text-gray-600 px-2 py-1 rounded-full">
                {Math.round(lastCommand.confidence * 100)}% confidence
              </span>
            </div>
            <p className="text-gray-700 text-sm">"{lastCommand.command}"</p>
            <p className="text-gray-500 text-xs mt-1">
              Interpreted as: {lastCommand.type} → {JSON.stringify(lastCommand.parameters)}
            </p>
          </div>
        </FadeIn>
      )}

      {/* Voice Commands Help */}
      <details className="bg-gray-50 rounded-lg">
        <summary className="p-3 cursor-pointer text-sm font-medium text-gray-700">
          Voice Commands Help
        </summary>
        <div className="px-3 pb-3 space-y-2 text-xs text-gray-600">
          <div><strong>Search:</strong> "Search for beaches", "Find luxury hotels", "Show me cultural destinations"</div>
          <div><strong>Filter:</strong> "Filter by price under $2000", "Show only 5-star hotels", "Filter by duration 7 days"</div>
          <div><strong>Navigate:</strong> "Go to wishlist", "Open booking page", "Navigate to profile"</div>
          <div><strong>Actions:</strong> "Add to wishlist", "Book this destination", "Compare destinations"</div>
        </div>
      </details>
    </div>
  )
}
