import type { Meta, StoryObj } from '@storybook/react'
import { Mail, Lock, Search, Eye, EyeOff, User, Phone, Calendar, MapPin } from 'lucide-react'
import { Input } from '../src/components/ui/input'
import { useState } from 'react'

const meta: Meta<typeof Input> = {
  title: 'Design System/Input',
  component: Input,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: 'Enhanced Input component with variants, icons, validation states, and accessibility features. Built with class-variance-authority for consistent styling.',
      },
    },
  },
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: { type: 'select' },
      options: ['default', 'error', 'success', 'warning'],
      description: 'Visual variant of the input',
    },
    size: {
      control: { type: 'select' },
      options: ['sm', 'md', 'lg'],
      description: 'Size of the input',
    },
    type: {
      control: { type: 'select' },
      options: ['text', 'email', 'password', 'search', 'tel', 'url', 'number'],
      description: 'HTML input type',
    },
    disabled: {
      control: { type: 'boolean' },
      description: 'Disables the input',
    },
    required: {
      control: { type: 'boolean' },
      description: 'Makes the input required',
    },
  },
}

export default meta
type Story = StoryObj<typeof meta>

// Default story
export const Default: Story = {
  args: {
    placeholder: 'Enter text...',
    variant: 'default',
    size: 'md',
  },
}

// With label and helper text
export const WithLabel: Story = {
  args: {
    label: 'Email Address',
    placeholder: 'Enter your email',
    helperText: 'We\'ll never share your email with anyone else.',
    type: 'email',
  },
}

// All variants
export const Variants: Story = {
  render: () => (
    <div className="space-y-4 w-80">
      <Input
        label="Default Input"
        placeholder="Default variant"
        variant="default"
      />
      <Input
        label="Success Input"
        placeholder="Success variant"
        variant="success"
        helperText="This looks good!"
      />
      <Input
        label="Warning Input"
        placeholder="Warning variant"
        variant="warning"
        helperText="Please double-check this field"
      />
      <Input
        label="Error Input"
        placeholder="Error variant"
        variant="error"
        error="This field is required"
      />
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Different input variants for various states and feedback.',
      },
    },
  },
}

// All sizes
export const Sizes: Story = {
  render: () => (
    <div className="space-y-4 w-80">
      <Input
        label="Small Input"
        placeholder="Small size"
        size="sm"
      />
      <Input
        label="Medium Input"
        placeholder="Medium size (default)"
        size="md"
      />
      <Input
        label="Large Input"
        placeholder="Large size"
        size="lg"
      />
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Different input sizes from small to large.',
      },
    },
  },
}

// With icons
export const WithIcons: Story = {
  render: () => (
    <div className="space-y-4 w-80">
      <Input
        label="Email"
        placeholder="Enter your email"
        type="email"
        leftIcon={<Mail className="h-4 w-4" />}
      />
      <Input
        label="Search"
        placeholder="Search destinations..."
        type="search"
        leftIcon={<Search className="h-4 w-4" />}
      />
      <Input
        label="Phone"
        placeholder="Enter phone number"
        type="tel"
        leftIcon={<Phone className="h-4 w-4" />}
      />
      <Input
        label="Location"
        placeholder="Enter location"
        leftIcon={<MapPin className="h-4 w-4" />}
        rightIcon={<Search className="h-4 w-4" />}
      />
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Inputs with left icons, right icons, or both for better UX.',
      },
    },
  },
}

// Password input with toggle
export const PasswordToggle: Story = {
  render: () => {
    const [showPassword, setShowPassword] = useState(false)
    
    return (
      <div className="w-80">
        <Input
          label="Password"
          placeholder="Enter your password"
          type={showPassword ? 'text' : 'password'}
          leftIcon={<Lock className="h-4 w-4" />}
          rightIcon={showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
          onRightIconClick={() => setShowPassword(!showPassword)}
          helperText="Click the eye icon to toggle visibility"
        />
      </div>
    )
  },
  parameters: {
    docs: {
      description: {
        story: 'Password input with toggle visibility functionality using clickable right icon.',
      },
    },
  },
}

// Form validation states
export const ValidationStates: Story = {
  render: () => (
    <div className="space-y-4 w-80">
      <Input
        label="Required Field"
        placeholder="This field is required"
        required
        error="This field is required"
        leftIcon={<User className="h-4 w-4" />}
      />
      <Input
        label="Valid Email"
        placeholder="user@example.com"
        type="email"
        variant="success"
        helperText="Email format is valid"
        leftIcon={<Mail className="h-4 w-4" />}
        defaultValue="user@example.com"
      />
      <Input
        label="Warning State"
        placeholder="Check this field"
        variant="warning"
        helperText="This field needs attention"
        leftIcon={<Calendar className="h-4 w-4" />}
      />
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Different validation states with appropriate colors and messaging.',
      },
    },
  },
}

// Disabled states
export const DisabledStates: Story = {
  render: () => (
    <div className="space-y-4 w-80">
      <Input
        label="Disabled Input"
        placeholder="This input is disabled"
        disabled
        leftIcon={<User className="h-4 w-4" />}
      />
      <Input
        label="Disabled with Value"
        defaultValue="Cannot edit this value"
        disabled
        leftIcon={<Lock className="h-4 w-4" />}
      />
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Disabled input states with reduced opacity and no interaction.',
      },
    },
  },
}

// Travel booking form example
export const TravelBookingForm: Story = {
  render: () => (
    <div className="space-y-6 w-96">
      <div className="space-y-2">
        <h3 className="text-lg font-semibold">Booking Information</h3>
        <div className="space-y-4">
          <Input
            label="Destination"
            placeholder="Where would you like to go?"
            leftIcon={<Search className="h-4 w-4" />}
            rightIcon={<MapPin className="h-4 w-4" />}
          />
          <div className="grid grid-cols-2 gap-4">
            <Input
              label="Check-in Date"
              type="date"
              leftIcon={<Calendar className="h-4 w-4" />}
            />
            <Input
              label="Check-out Date"
              type="date"
              leftIcon={<Calendar className="h-4 w-4" />}
            />
          </div>
          <Input
            label="Number of Guests"
            type="number"
            placeholder="2"
            min="1"
            max="10"
            leftIcon={<User className="h-4 w-4" />}
          />
        </div>
      </div>
      
      <div className="space-y-2">
        <h3 className="text-lg font-semibold">Contact Information</h3>
        <div className="space-y-4">
          <Input
            label="Email Address"
            type="email"
            placeholder="your@email.com"
            required
            leftIcon={<Mail className="h-4 w-4" />}
          />
          <Input
            label="Phone Number"
            type="tel"
            placeholder="+1 (555) 123-4567"
            leftIcon={<Phone className="h-4 w-4" />}
          />
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Real-world example of inputs in a travel booking form with appropriate types and icons.',
      },
    },
  },
}

// Interactive playground
export const Playground: Story = {
  args: {
    label: 'Playground Input',
    placeholder: 'Type something...',
    variant: 'default',
    size: 'md',
    disabled: false,
    required: false,
  },
  parameters: {
    docs: {
      description: {
        story: 'Interactive playground to test different input configurations.',
      },
    },
  },
}
