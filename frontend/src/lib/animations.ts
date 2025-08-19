/**
 * Animation System for Exotic Travel Booking Platform
 * 
 * This file provides a comprehensive animation system using Framer Motion
 * with predefined animations, transitions, and utilities for consistent
 * motion design across the application.
 */

import { Variants, Transition } from 'framer-motion'

// Animation Durations (in seconds)
export const durations = {
  instant: 0,
  fast: 0.15,
  normal: 0.3,
  slow: 0.5,
  slower: 0.75,
  slowest: 1.0,
} as const

// Easing Functions
export const easings = {
  linear: [0, 0, 1, 1],
  easeIn: [0.4, 0, 1, 1],
  easeOut: [0, 0, 0.2, 1],
  easeInOut: [0.4, 0, 0.2, 1],
  spring: [0.34, 1.56, 0.64, 1],
  bounce: [0.68, -0.55, 0.265, 1.55],
} as const

// Spring Configurations
export const springs = {
  gentle: {
    type: 'spring' as const,
    stiffness: 120,
    damping: 14,
  },
  wobbly: {
    type: 'spring' as const,
    stiffness: 180,
    damping: 12,
  },
  stiff: {
    type: 'spring' as const,
    stiffness: 210,
    damping: 20,
  },
  slow: {
    type: 'spring' as const,
    stiffness: 280,
    damping: 60,
  },
} as const

// Common Transitions
export const transitions: Record<string, Transition> = {
  fast: { duration: durations.fast, ease: easings.easeOut },
  normal: { duration: durations.normal, ease: easings.easeInOut },
  slow: { duration: durations.slow, ease: easings.easeOut },
  spring: springs.gentle,
  bounce: { ...springs.wobbly, duration: durations.slow },
}

// Fade Animations
export const fadeVariants: Variants = {
  hidden: { 
    opacity: 0,
    transition: transitions.fast,
  },
  visible: { 
    opacity: 1,
    transition: transitions.normal,
  },
  exit: { 
    opacity: 0,
    transition: transitions.fast,
  },
}

export const fadeInUp: Variants = {
  hidden: { 
    opacity: 0, 
    y: 20,
    transition: transitions.fast,
  },
  visible: { 
    opacity: 1, 
    y: 0,
    transition: transitions.normal,
  },
  exit: { 
    opacity: 0, 
    y: -20,
    transition: transitions.fast,
  },
}

export const fadeInDown: Variants = {
  hidden: { 
    opacity: 0, 
    y: -20,
    transition: transitions.fast,
  },
  visible: { 
    opacity: 1, 
    y: 0,
    transition: transitions.normal,
  },
  exit: { 
    opacity: 0, 
    y: 20,
    transition: transitions.fast,
  },
}

export const fadeInLeft: Variants = {
  hidden: { 
    opacity: 0, 
    x: -20,
    transition: transitions.fast,
  },
  visible: { 
    opacity: 1, 
    x: 0,
    transition: transitions.normal,
  },
  exit: { 
    opacity: 0, 
    x: 20,
    transition: transitions.fast,
  },
}

export const fadeInRight: Variants = {
  hidden: { 
    opacity: 0, 
    x: 20,
    transition: transitions.fast,
  },
  visible: { 
    opacity: 1, 
    x: 0,
    transition: transitions.normal,
  },
  exit: { 
    opacity: 0, 
    x: -20,
    transition: transitions.fast,
  },
}

// Scale Animations
export const scaleVariants: Variants = {
  hidden: { 
    scale: 0.8, 
    opacity: 0,
    transition: transitions.fast,
  },
  visible: { 
    scale: 1, 
    opacity: 1,
    transition: transitions.spring,
  },
  exit: { 
    scale: 0.8, 
    opacity: 0,
    transition: transitions.fast,
  },
}

export const scaleInCenter: Variants = {
  hidden: { 
    scale: 0,
    opacity: 0,
    transition: transitions.fast,
  },
  visible: { 
    scale: 1,
    opacity: 1,
    transition: transitions.bounce,
  },
  exit: { 
    scale: 0,
    opacity: 0,
    transition: transitions.fast,
  },
}

// Slide Animations
export const slideInLeft: Variants = {
  hidden: { 
    x: '-100%',
    transition: transitions.fast,
  },
  visible: { 
    x: 0,
    transition: transitions.spring,
  },
  exit: { 
    x: '-100%',
    transition: transitions.normal,
  },
}

export const slideInRight: Variants = {
  hidden: { 
    x: '100%',
    transition: transitions.fast,
  },
  visible: { 
    x: 0,
    transition: transitions.spring,
  },
  exit: { 
    x: '100%',
    transition: transitions.normal,
  },
}

export const slideInUp: Variants = {
  hidden: { 
    y: '100%',
    transition: transitions.fast,
  },
  visible: { 
    y: 0,
    transition: transitions.spring,
  },
  exit: { 
    y: '100%',
    transition: transitions.normal,
  },
}

export const slideInDown: Variants = {
  hidden: { 
    y: '-100%',
    transition: transitions.fast,
  },
  visible: { 
    y: 0,
    transition: transitions.spring,
  },
  exit: { 
    y: '-100%',
    transition: transitions.normal,
  },
}

// Rotation Animations
export const rotateIn: Variants = {
  hidden: { 
    rotate: -180, 
    scale: 0.8, 
    opacity: 0,
    transition: transitions.fast,
  },
  visible: { 
    rotate: 0, 
    scale: 1, 
    opacity: 1,
    transition: transitions.spring,
  },
  exit: { 
    rotate: 180, 
    scale: 0.8, 
    opacity: 0,
    transition: transitions.fast,
  },
}

// Stagger Animations for Lists
export const staggerContainer: Variants = {
  hidden: { opacity: 0 },
  visible: {
    opacity: 1,
    transition: {
      staggerChildren: 0.1,
      delayChildren: 0.1,
    },
  },
  exit: {
    opacity: 0,
    transition: {
      staggerChildren: 0.05,
      staggerDirection: -1,
    },
  },
}

export const staggerItem: Variants = {
  hidden: { 
    opacity: 0, 
    y: 20,
    transition: transitions.fast,
  },
  visible: { 
    opacity: 1, 
    y: 0,
    transition: transitions.normal,
  },
  exit: { 
    opacity: 0, 
    y: -20,
    transition: transitions.fast,
  },
}

// Travel-specific Animations
export const floatAnimation: Variants = {
  animate: {
    y: [-10, 10, -10],
    transition: {
      duration: 3,
      repeat: Infinity,
      ease: 'easeInOut',
    },
  },
}

export const pulseGlow: Variants = {
  animate: {
    boxShadow: [
      '0 0 5px rgb(59 130 246 / 0.5)',
      '0 0 20px rgb(59 130 246 / 0.8)',
      '0 0 5px rgb(59 130 246 / 0.5)',
    ],
    transition: {
      duration: 2,
      repeat: Infinity,
      ease: 'easeInOut',
    },
  },
}

// Page Transition Animations
export const pageTransition: Variants = {
  initial: { 
    opacity: 0, 
    y: 20,
  },
  animate: { 
    opacity: 1, 
    y: 0,
    transition: {
      duration: durations.normal,
      ease: easings.easeOut,
    },
  },
  exit: { 
    opacity: 0, 
    y: -20,
    transition: {
      duration: durations.fast,
      ease: easings.easeIn,
    },
  },
}

// Modal/Dialog Animations
export const modalBackdrop: Variants = {
  hidden: { opacity: 0 },
  visible: { 
    opacity: 1,
    transition: { duration: durations.fast },
  },
  exit: { 
    opacity: 0,
    transition: { duration: durations.fast },
  },
}

export const modalContent: Variants = {
  hidden: { 
    opacity: 0, 
    scale: 0.8, 
    y: 20,
  },
  visible: { 
    opacity: 1, 
    scale: 1, 
    y: 0,
    transition: transitions.spring,
  },
  exit: { 
    opacity: 0, 
    scale: 0.8, 
    y: 20,
    transition: transitions.fast,
  },
}

// Utility Functions
export const createStaggerDelay = (index: number, baseDelay = 0.1) => ({
  delay: index * baseDelay,
})

export const createSpringTransition = (
  stiffness = 120,
  damping = 14,
  mass = 1
): Transition => ({
  type: 'spring',
  stiffness,
  damping,
  mass,
})

export const createBounceTransition = (
  duration = durations.slow,
  bounce = 0.4
): Transition => ({
  type: 'spring',
  duration,
  bounce,
})

// Animation Presets for Common UI Elements
export const buttonHover = {
  scale: 1.02,
  transition: transitions.fast,
}

export const buttonTap = {
  scale: 0.98,
  transition: transitions.fast,
}

export const cardHover = {
  y: -4,
  boxShadow: '0 10px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)',
  transition: transitions.normal,
}

export const imageHover = {
  scale: 1.05,
  transition: transitions.slow,
}
