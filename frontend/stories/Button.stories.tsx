import type { Meta, StoryObj } from '@storybook/react'
import { Heart, Download, ArrowRight, Mail, Plus, Search } from 'lucide-react'
import { Button } from '../src/components/ui/button'

const meta: Meta<typeof Button> = {
  title: 'Design System/Button',
  component: Button,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: 'Enhanced Button component with multiple variants, sizes, and interactive states. Built with class-variance-authority for consistent styling and Radix UI for accessibility.',
      },
    },
  },
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: { type: 'select' },
      options: ['primary', 'secondary', 'accent', 'outline', 'ghost', 'destructive', 'success', 'warning', 'info', 'gradient', 'glow'],
      description: 'Visual style variant of the button',
    },
    size: {
      control: { type: 'select' },
      options: ['sm', 'md', 'lg', 'xl', 'icon'],
      description: 'Size of the button',
    },
    loading: {
      control: { type: 'boolean' },
      description: 'Shows loading spinner and disables interaction',
    },
    disabled: {
      control: { type: 'boolean' },
      description: 'Disables the button',
    },
    asChild: {
      control: { type: 'boolean' },
      description: 'Render as child element (useful for links)',
    },
  },
}

export default meta
type Story = StoryObj<typeof meta>

// Default story
export const Default: Story = {
  args: {
    children: 'Button',
    variant: 'primary',
    size: 'md',
  },
}

// All variants showcase
export const Variants: Story = {
  render: () => (
    <div className="flex flex-wrap gap-4">
      <Button variant="primary">Primary</Button>
      <Button variant="secondary">Secondary</Button>
      <Button variant="accent">Accent</Button>
      <Button variant="outline">Outline</Button>
      <Button variant="ghost">Ghost</Button>
      <Button variant="destructive">Destructive</Button>
      <Button variant="success">Success</Button>
      <Button variant="warning">Warning</Button>
      <Button variant="info">Info</Button>
      <Button variant="gradient">Gradient</Button>
      <Button variant="glow">Glow Effect</Button>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'All available button variants with their unique styling and hover effects.',
      },
    },
  },
}

// All sizes showcase
export const Sizes: Story = {
  render: () => (
    <div className="flex items-center gap-4">
      <Button size="sm">Small</Button>
      <Button size="md">Medium</Button>
      <Button size="lg">Large</Button>
      <Button size="xl">Extra Large</Button>
      <Button size="icon">
        <Heart className="h-4 w-4" />
      </Button>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Different button sizes from small to extra large, plus icon-only variant.',
      },
    },
  },
}

// With icons
export const WithIcons: Story = {
  render: () => (
    <div className="flex flex-wrap gap-4">
      <Button leftIcon={<Mail className="h-4 w-4" />}>
        Send Email
      </Button>
      <Button rightIcon={<ArrowRight className="h-4 w-4" />}>
        Continue
      </Button>
      <Button 
        leftIcon={<Download className="h-4 w-4" />}
        rightIcon={<ArrowRight className="h-4 w-4" />}
      >
        Download & Continue
      </Button>
      <Button variant="outline" size="icon">
        <Plus className="h-4 w-4" />
      </Button>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Buttons with left icons, right icons, both, or icon-only variants.',
      },
    },
  },
}

// Loading states
export const LoadingStates: Story = {
  render: () => (
    <div className="flex flex-wrap gap-4">
      <Button loading>Loading...</Button>
      <Button variant="secondary" loading>Processing</Button>
      <Button variant="outline" loading>Saving</Button>
      <Button variant="destructive" loading>Deleting</Button>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Loading states with spinner animation and disabled interaction.',
      },
    },
  },
}

// Disabled states
export const DisabledStates: Story = {
  render: () => (
    <div className="flex flex-wrap gap-4">
      <Button disabled>Disabled Primary</Button>
      <Button variant="secondary" disabled>Disabled Secondary</Button>
      <Button variant="outline" disabled>Disabled Outline</Button>
      <Button variant="ghost" disabled>Disabled Ghost</Button>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Disabled button states with reduced opacity and no interaction.',
      },
    },
  },
}

// Travel-themed examples
export const TravelThemed: Story = {
  render: () => (
    <div className="space-y-6">
      <div className="space-y-2">
        <h3 className="text-lg font-semibold">Booking Actions</h3>
        <div className="flex flex-wrap gap-3">
          <Button variant="gradient" size="lg" rightIcon={<ArrowRight className="h-4 w-4" />}>
            Book Now
          </Button>
          <Button variant="outline" leftIcon={<Heart className="h-4 w-4" />}>
            Add to Wishlist
          </Button>
          <Button variant="ghost" leftIcon={<Search className="h-4 w-4" />}>
            View Details
          </Button>
        </div>
      </div>
      
      <div className="space-y-2">
        <h3 className="text-lg font-semibold">Status Actions</h3>
        <div className="flex flex-wrap gap-3">
          <Button variant="success" leftIcon={<Download className="h-4 w-4" />}>
            Download Tickets
          </Button>
          <Button variant="warning">
            Pending Confirmation
          </Button>
          <Button variant="destructive">
            Cancel Booking
          </Button>
        </div>
      </div>
      
      <div className="space-y-2">
        <h3 className="text-lg font-semibold">Special Effects</h3>
        <div className="flex flex-wrap gap-3">
          <Button variant="glow" size="lg">
            Limited Time Offer
          </Button>
          <Button variant="gradient" rightIcon={<ArrowRight className="h-4 w-4" />}>
            Explore Destinations
          </Button>
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Real-world examples of buttons in a travel booking context with appropriate variants and icons.',
      },
    },
  },
}

// Interactive playground
export const Playground: Story = {
  args: {
    children: 'Playground Button',
    variant: 'primary',
    size: 'md',
    loading: false,
    disabled: false,
  },
  parameters: {
    docs: {
      description: {
        story: 'Interactive playground to test different button configurations.',
      },
    },
  },
}
