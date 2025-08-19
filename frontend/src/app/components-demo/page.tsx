'use client'

import { useState } from 'react'
import { Mail, Lock, Search, Heart, Star, User } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card'

export default function ComponentsDemo() {
  const [loading, setLoading] = useState(false)

  const handleLoadingDemo = () => {
    setLoading(true)
    setTimeout(() => setLoading(false), 2000)
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 py-12">
      <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold text-gray-900 mb-4">
            UI Components Showcase
          </h1>
          <p className="text-xl text-gray-600 max-w-3xl mx-auto">
            Beautifully designed components with improved TailwindCSS styling
          </p>
        </div>

        <div className="space-y-12">
          {/* Buttons Section */}
          <Card>
            <CardHeader>
              <CardTitle>Button Components</CardTitle>
              <CardDescription>
                Various button styles and states with improved hover effects
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                <div className="space-y-3">
                  <h4 className="font-semibold text-gray-900">Primary Buttons</h4>
                  <div className="space-y-2">
                    <Button size="sm">Small Button</Button>
                    <Button size="md">Medium Button</Button>
                    <Button size="lg">Large Button</Button>
                    <Button size="xl">Extra Large</Button>
                  </div>
                </div>
                
                <div className="space-y-3">
                  <h4 className="font-semibold text-gray-900">Button Variants</h4>
                  <div className="space-y-2">
                    <Button variant="primary">Primary</Button>
                    <Button variant="secondary">Secondary</Button>
                    <Button variant="outline">Outline</Button>
                    <Button variant="ghost">Ghost</Button>
                    <Button variant="destructive">Destructive</Button>
                  </div>
                </div>
                
                <div className="space-y-3">
                  <h4 className="font-semibold text-gray-900">Button States</h4>
                  <div className="space-y-2">
                    <Button loading={loading} onClick={handleLoadingDemo}>
                      {loading ? 'Loading...' : 'Click for Loading'}
                    </Button>
                    <Button disabled>Disabled Button</Button>
                    <Button>
                      <Heart className="h-4 w-4 mr-2" />
                      With Icon
                    </Button>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Input Components */}
          <Card>
            <CardHeader>
              <CardTitle>Input Components</CardTitle>
              <CardDescription>
                Form inputs with better styling and validation states
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="space-y-4">
                  <Input
                    id="demo-email"
                    label="Email Address"
                    type="email"
                    placeholder="Enter your email"
                    leftIcon={<Mail />}
                    helperText="We'll never share your email"
                  />

                  <Input
                    id="demo-password"
                    label="Password"
                    type="password"
                    placeholder="Enter your password"
                    leftIcon={<Lock />}
                    error="Password must be at least 8 characters"
                  />

                  <Input
                    id="demo-search"
                    label="Search"
                    type="text"
                    placeholder="Search destinations..."
                    leftIcon={<Search />}
                    rightIcon={<Star />}
                  />
                </div>
                
                <div className="space-y-4">
                  <Input
                    id="demo-fullname"
                    label="Full Name"
                    type="text"
                    placeholder="John Doe"
                    leftIcon={<User />}
                  />

                  <Input
                    id="demo-disabled"
                    label="Disabled Input"
                    type="text"
                    placeholder="This is disabled"
                    disabled
                  />

                  <Input
                    id="demo-no-label"
                    type="text"
                    placeholder="Input without label"
                    helperText="This input has no label"
                  />
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Card Components */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            <Card className="hover:shadow-lg transition-shadow duration-200">
              <CardHeader>
                <CardTitle>Maldives Paradise</CardTitle>
                <CardDescription>
                  Luxury overwater bungalows in crystal clear waters
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="aspect-video bg-gradient-to-br from-blue-400 to-teal-500 rounded-lg mb-4"></div>
                <p className="text-sm text-gray-600">
                  Experience the ultimate tropical getaway with pristine beaches and world-class amenities.
                </p>
              </CardContent>
              <CardFooter>
                <Button className="w-full">Book Now</Button>
              </CardFooter>
            </Card>

            <Card className="hover:shadow-lg transition-shadow duration-200">
              <CardHeader>
                <CardTitle>Swiss Alps Adventure</CardTitle>
                <CardDescription>
                  Mountain peaks and alpine lakes await
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="aspect-video bg-gradient-to-br from-green-400 to-blue-500 rounded-lg mb-4"></div>
                <p className="text-sm text-gray-600">
                  Discover breathtaking mountain vistas and charming alpine villages.
                </p>
              </CardContent>
              <CardFooter>
                <Button variant="outline" className="w-full">Learn More</Button>
              </CardFooter>
            </Card>

            <Card className="hover:shadow-lg transition-shadow duration-200">
              <CardHeader>
                <CardTitle>Tokyo City Lights</CardTitle>
                <CardDescription>
                  Modern metropolis meets ancient culture
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="aspect-video bg-gradient-to-br from-purple-400 to-pink-500 rounded-lg mb-4"></div>
                <p className="text-sm text-gray-600">
                  Immerse yourself in the vibrant energy of Japan's capital city.
                </p>
              </CardContent>
              <CardFooter>
                <Button variant="secondary" className="w-full">Explore</Button>
              </CardFooter>
            </Card>
          </div>

          {/* Color Palette */}
          <Card>
            <CardHeader>
              <CardTitle>Color System</CardTitle>
              <CardDescription>
                Updated color palette with better contrast and accessibility
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div className="space-y-2">
                  <div className="h-16 bg-primary rounded-lg"></div>
                  <p className="text-sm font-medium">Primary</p>
                </div>
                <div className="space-y-2">
                  <div className="h-16 bg-secondary rounded-lg"></div>
                  <p className="text-sm font-medium">Secondary</p>
                </div>
                <div className="space-y-2">
                  <div className="h-16 bg-accent rounded-lg"></div>
                  <p className="text-sm font-medium">Accent</p>
                </div>
                <div className="space-y-2">
                  <div className="h-16 bg-muted rounded-lg"></div>
                  <p className="text-sm font-medium">Muted</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
