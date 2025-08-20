'use client'

import React, { useState } from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { 
  BarChart3, 
  Target, 
  PenTool, 
  Palette, 
  Users, 
  Settings, 
  Menu, 
  X,
  Brain,
  Zap,
  TrendingUp,
  Image,
  Mail,
  Share2
} from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

interface MarketingLayoutProps {
  children: React.ReactNode
}

interface NavigationItem {
  name: string
  href: string
  icon: React.ReactNode
  badge?: string
  description: string
}

export default function MarketingLayout({ children }: MarketingLayoutProps) {
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const pathname = usePathname()

  const navigation: NavigationItem[] = [
    {
      name: 'Dashboard',
      href: '/marketing/dashboard',
      icon: <BarChart3 className="h-5 w-5" />,
      description: 'Overview and key metrics'
    },
    {
      name: 'Campaigns',
      href: '/marketing/campaigns',
      icon: <Target className="h-5 w-5" />,
      badge: '8',
      description: 'Manage marketing campaigns'
    },
    {
      name: 'Content Studio',
      href: '/marketing/content',
      icon: <PenTool className="h-5 w-5" />,
      badge: 'AI',
      description: 'AI-powered content generation'
    },
    {
      name: 'Visual Assets',
      href: '/marketing/assets',
      icon: <Image className="h-5 w-5" />,
      description: 'AI-generated visual content'
    },
    {
      name: 'Analytics',
      href: '/marketing/analytics',
      icon: <TrendingUp className="h-5 w-5" />,
      description: 'Performance insights'
    },
    {
      name: 'Audience',
      href: '/marketing/audience',
      icon: <Users className="h-5 w-5" />,
      description: 'Audience segments and insights'
    },
    {
      name: 'Brand Manager',
      href: '/marketing/brand',
      icon: <Palette className="h-5 w-5" />,
      description: 'Brand identity and guidelines'
    },
    {
      name: 'Integrations',
      href: '/marketing/integrations',
      icon: <Share2 className="h-5 w-5" />,
      description: 'Platform connections'
    },
    {
      name: 'Settings',
      href: '/marketing/settings',
      icon: <Settings className="h-5 w-5" />,
      description: 'Marketing preferences'
    }
  ]

  const isActive = (href: string) => {
    if (href === '/marketing/dashboard') {
      return pathname === '/marketing' || pathname === '/marketing/dashboard'
    }
    return pathname.startsWith(href)
  }

  const getBadgeColor = (badge: string) => {
    switch (badge) {
      case 'AI':
        return 'bg-gradient-to-r from-purple-500 to-pink-500 text-white'
      default:
        return 'bg-blue-500 text-white'
    }
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-blue-50">
      {/* Mobile sidebar backdrop */}
      <AnimatePresence>
        {sidebarOpen && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-40 bg-black bg-opacity-50 lg:hidden"
            onClick={() => setSidebarOpen(false)}
          />
        )}
      </AnimatePresence>

      {/* Sidebar */}
      <AnimatePresence>
        <motion.div
          initial={{ x: -300 }}
          animate={{ x: sidebarOpen ? 0 : -300 }}
          transition={{ type: "spring", stiffness: 300, damping: 30 }}
          className={`fixed inset-y-0 left-0 z-50 w-80 bg-white/95 backdrop-blur-xl border-r border-gray-200 shadow-xl lg:translate-x-0 lg:static lg:inset-0 ${
            sidebarOpen ? 'translate-x-0' : '-translate-x-full'
          } lg:translate-x-0 transition-transform duration-300 ease-in-out`}
        >
          <div className="flex h-full flex-col">
            {/* Header */}
            <div className="flex items-center justify-between p-6 border-b border-gray-200">
              <div className="flex items-center">
                <div className="flex items-center justify-center w-10 h-10 bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg mr-3">
                  <Brain className="h-6 w-6 text-white" />
                </div>
                <div>
                  <h1 className="text-lg font-bold text-gray-900">Marketing AI</h1>
                  <p className="text-sm text-gray-600">Powered by AI</p>
                </div>
              </div>
              <Button
                variant="ghost"
                size="sm"
                className="lg:hidden"
                onClick={() => setSidebarOpen(false)}
              >
                <X className="h-5 w-5" />
              </Button>
            </div>

            {/* Navigation */}
            <nav className="flex-1 px-4 py-6 space-y-2 overflow-y-auto">
              {navigation.map((item) => (
                <Link key={item.name} href={item.href}>
                  <motion.div
                    whileHover={{ scale: 1.02 }}
                    whileTap={{ scale: 0.98 }}
                    className={`group flex items-center px-4 py-3 text-sm font-medium rounded-xl transition-all duration-200 ${
                      isActive(item.href)
                        ? 'bg-gradient-to-r from-blue-600 to-purple-600 text-white shadow-lg'
                        : 'text-gray-700 hover:bg-gray-100 hover:text-gray-900'
                    }`}
                  >
                    <div className={`mr-3 ${isActive(item.href) ? 'text-white' : 'text-gray-400 group-hover:text-gray-600'}`}>
                      {item.icon}
                    </div>
                    <div className="flex-1">
                      <div className="flex items-center justify-between">
                        <span>{item.name}</span>
                        {item.badge && (
                          <Badge 
                            className={`ml-2 text-xs ${
                              isActive(item.href) 
                                ? 'bg-white/20 text-white' 
                                : getBadgeColor(item.badge)
                            }`}
                          >
                            {item.badge}
                          </Badge>
                        )}
                      </div>
                      <p className={`text-xs mt-1 ${
                        isActive(item.href) ? 'text-white/80' : 'text-gray-500'
                      }`}>
                        {item.description}
                      </p>
                    </div>
                  </motion.div>
                </Link>
              ))}
            </nav>

            {/* Footer */}
            <div className="p-4 border-t border-gray-200">
              <div className="bg-gradient-to-r from-blue-50 to-purple-50 rounded-lg p-4">
                <div className="flex items-center mb-2">
                  <Zap className="h-5 w-5 text-blue-600 mr-2" />
                  <span className="text-sm font-medium text-gray-900">AI Credits</span>
                </div>
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-xs text-gray-600">2,847 / 5,000 used</p>
                    <div className="w-24 h-2 bg-gray-200 rounded-full mt-1">
                      <div className="w-14 h-2 bg-gradient-to-r from-blue-500 to-purple-500 rounded-full"></div>
                    </div>
                  </div>
                  <Button size="sm" variant="outline" className="text-xs">
                    Upgrade
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </motion.div>
      </AnimatePresence>

      {/* Main content */}
      <div className="lg:pl-80">
        {/* Mobile header */}
        <div className="sticky top-0 z-30 flex h-16 items-center gap-x-4 border-b border-gray-200 bg-white/95 backdrop-blur-xl px-4 shadow-sm lg:hidden">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => setSidebarOpen(true)}
          >
            <Menu className="h-5 w-5" />
          </Button>
          <div className="flex items-center">
            <div className="flex items-center justify-center w-8 h-8 bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg mr-2">
              <Brain className="h-4 w-4 text-white" />
            </div>
            <span className="text-lg font-bold text-gray-900">Marketing AI</span>
          </div>
        </div>

        {/* Page content */}
        <main className="min-h-screen">
          {children}
        </main>
      </div>
    </div>
  )
}
