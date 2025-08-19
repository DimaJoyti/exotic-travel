'use client'

import React from 'react'
import { motion, HTMLMotionProps } from 'framer-motion'

// Simple fallback animated components
export interface AnimatedDivProps extends HTMLMotionProps<'div'> {
  children: React.ReactNode
  delay?: number
  duration?: number
}

export const AnimatedDiv: React.FC<AnimatedDivProps> = ({
  children,
  className,
  ...props
}) => (
  <motion.div className={className} {...props}>
    {children}
  </motion.div>
)

export interface FadeInProps extends HTMLMotionProps<'div'> {
  direction?: 'up' | 'down' | 'left' | 'right'
  delay?: number
  children: React.ReactNode
}

export const FadeIn: React.FC<FadeInProps> = ({
  direction = 'up',
  delay = 0,
  children,
  className,
  ...props
}) => (
  <motion.div
    initial={{ opacity: 0, y: direction === 'up' ? 20 : direction === 'down' ? -20 : 0, x: direction === 'left' ? -20 : direction === 'right' ? 20 : 0 }}
    whileInView={{ opacity: 1, y: 0, x: 0 }}
    viewport={{ once: true }}
    transition={{ duration: 0.6, delay }}
    className={className}
    {...props}
  >
    {children}
  </motion.div>
)

export interface ScaleInProps extends HTMLMotionProps<'div'> {
  delay?: number
  children: React.ReactNode
}

export const ScaleIn: React.FC<ScaleInProps> = ({
  delay = 0,
  children,
  className,
  ...props
}) => (
  <motion.div
    initial={{ opacity: 0, scale: 0.9 }}
    whileInView={{ opacity: 1, scale: 1 }}
    viewport={{ once: true }}
    transition={{ duration: 0.6, delay }}
    className={className}
    {...props}
  >
    {children}
  </motion.div>
)

export interface SlideInProps extends HTMLMotionProps<'div'> {
  direction: 'left' | 'right' | 'up' | 'down'
  delay?: number
  children: React.ReactNode
}

export const SlideIn: React.FC<SlideInProps> = ({
  direction,
  delay = 0,
  children,
  className,
  ...props
}) => {
  const getInitial = () => {
    switch (direction) {
      case 'left': return { x: -100, opacity: 0 }
      case 'right': return { x: 100, opacity: 0 }
      case 'up': return { y: 100, opacity: 0 }
      case 'down': return { y: -100, opacity: 0 }
    }
  }

  return (
    <motion.div
      initial={getInitial()}
      whileInView={{ x: 0, y: 0, opacity: 1 }}
      viewport={{ once: true }}
      transition={{ duration: 0.6, delay }}
      className={className}
      {...props}
    >
      {children}
    </motion.div>
  )
}

export interface StaggerContainerProps extends HTMLMotionProps<'div'> {
  staggerDelay?: number
  children: React.ReactNode
}

export const StaggerContainer: React.FC<StaggerContainerProps> = ({
  staggerDelay = 0.1,
  children,
  className,
  ...props
}) => (
  <motion.div
    initial={{ opacity: 0 }}
    whileInView={{ opacity: 1 }}
    viewport={{ once: true }}
    transition={{ staggerChildren: staggerDelay, delayChildren: 0.1 }}
    className={className}
    {...props}
  >
    {children}
  </motion.div>
)

export interface StaggerItemProps extends HTMLMotionProps<'div'> {
  delay?: number
  children: React.ReactNode
}

export const StaggerItem: React.FC<StaggerItemProps> = ({
  delay = 0,
  children,
  className,
  ...props
}) => (
  <motion.div
    initial={{ opacity: 0, y: 20 }}
    whileInView={{ opacity: 1, y: 0 }}
    viewport={{ once: true }}
    transition={{ duration: 0.5, delay }}
    className={className}
    {...props}
  >
    {children}
  </motion.div>
)

export const PageTransition: React.FC<HTMLMotionProps<'div'>> = ({
  children,
  className,
  ...props
}) => (
  <motion.div
    initial={{ opacity: 0, y: 20 }}
    animate={{ opacity: 1, y: 0 }}
    exit={{ opacity: 0, y: -20 }}
    transition={{ duration: 0.5 }}
    className={className}
    {...props}
  >
    {children}
  </motion.div>
)

export const ModalBackdrop: React.FC<HTMLMotionProps<'div'>> = ({
  children,
  className,
  ...props
}) => (
  <motion.div
    initial={{ opacity: 0 }}
    animate={{ opacity: 1 }}
    exit={{ opacity: 0 }}
    className={`fixed inset-0 bg-black/50 z-50 ${className}`}
    {...props}
  >
    {children}
  </motion.div>
)

export const ModalContent: React.FC<HTMLMotionProps<'div'>> = ({
  children,
  className,
  ...props
}) => (
  <motion.div
    initial={{ opacity: 0, scale: 0.9, y: 20 }}
    animate={{ opacity: 1, scale: 1, y: 0 }}
    exit={{ opacity: 0, scale: 0.9, y: 20 }}
    className={`fixed left-1/2 top-1/2 z-50 w-full max-w-lg -translate-x-1/2 -translate-y-1/2 bg-white p-6 shadow-lg rounded-lg border ${className}`}
    {...props}
  >
    {children}
  </motion.div>
)

export const FloatingElement: React.FC<HTMLMotionProps<'div'>> = ({
  children,
  className,
  ...props
}) => (
  <motion.div
    animate={{ y: [0, -10, 0] }}
    transition={{ duration: 3, repeat: Infinity, ease: 'easeInOut' }}
    className={className}
    {...props}
  >
    {children}
  </motion.div>
)

export const PulseGlow: React.FC<HTMLMotionProps<'div'>> = ({
  children,
  className,
  ...props
}) => (
  <motion.div
    animate={{ 
      boxShadow: [
        '0 0 5px rgb(59 130 246 / 0.5)',
        '0 0 20px rgb(59 130 246 / 0.8)',
        '0 0 5px rgb(59 130 246 / 0.5)',
      ]
    }}
    transition={{ duration: 2, repeat: Infinity, ease: 'easeInOut' }}
    className={className}
    {...props}
  >
    {children}
  </motion.div>
)

export interface HoverAnimationProps extends HTMLMotionProps<'div'> {
  hoverScale?: number
  hoverY?: number
  tapScale?: number
  children: React.ReactNode
}

export const HoverAnimation: React.FC<HoverAnimationProps> = ({
  hoverScale = 1.02,
  hoverY = -2,
  tapScale = 0.98,
  children,
  className,
  ...props
}) => (
  <motion.div
    whileHover={{ scale: hoverScale, y: hoverY }}
    whileTap={{ scale: tapScale }}
    transition={{ duration: 0.2 }}
    className={className}
    {...props}
  >
    {children}
  </motion.div>
)

export const Parallax: React.FC<{
  offset?: number
  children: React.ReactNode
  className?: string
}> = ({
  offset = 50,
  children,
  className,
}) => (
  <motion.div
    initial={{ y: offset }}
    whileInView={{ y: 0 }}
    viewport={{ once: true, amount: 0.3 }}
    transition={{ duration: 0.6, ease: 'easeOut' }}
    className={className}
  >
    {children}
  </motion.div>
)

export const CountUp: React.FC<{
  from?: number
  to: number
  duration?: number
  className?: string
  suffix?: string
  prefix?: string
}> = ({
  from = 0,
  to,
  duration = 2,
  className,
  suffix = '',
  prefix = '',
}) => {
  const [count, setCount] = React.useState(from)

  React.useEffect(() => {
    let startTime: number
    let animationFrame: number

    const animate = (timestamp: number) => {
      if (!startTime) startTime = timestamp
      const progress = Math.min((timestamp - startTime) / (duration * 1000), 1)

      const currentCount = from + (to - from) * progress
      setCount(Math.round(currentCount))

      if (progress < 1) {
        animationFrame = requestAnimationFrame(animate)
      }
    }

    animationFrame = requestAnimationFrame(animate)
    return () => cancelAnimationFrame(animationFrame)
  }, [from, to, duration])

  return (
    <motion.span
      initial={{ opacity: 0 }}
      whileInView={{ opacity: 1 }}
      viewport={{ once: true }}
      transition={{ duration: 0.5 }}
      className={className}
    >
      {prefix}{count}{suffix}
    </motion.span>
  )
}

export const Typewriter: React.FC<{
  text: string
  delay?: number
  speed?: number
  className?: string
}> = ({
  text,
  delay = 0,
  speed = 50,
  className,
}) => {
  const [displayText, setDisplayText] = React.useState('')
  const [currentIndex, setCurrentIndex] = React.useState(0)

  React.useEffect(() => {
    const timer = setTimeout(() => {
      if (currentIndex < text.length) {
        setDisplayText(prev => prev + text[currentIndex])
        setCurrentIndex(prev => prev + 1)
      }
    }, currentIndex === 0 ? delay : speed)

    return () => clearTimeout(timer)
  }, [currentIndex, text, delay, speed])

  return (
    <motion.span
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ delay: delay / 1000 }}
      className={className}
    >
      {displayText}
      <motion.span
        animate={{ opacity: [1, 0] }}
        transition={{ duration: 0.8, repeat: Infinity, repeatType: 'reverse' }}
      >
        |
      </motion.span>
    </motion.span>
  )
}