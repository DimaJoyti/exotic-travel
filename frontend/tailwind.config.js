/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
    './stories/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      // Enhanced color system with design tokens
      colors: {
        // CSS variables for theme switching
        border: 'hsl(var(--border))',
        input: 'hsl(var(--input))',
        ring: 'hsl(var(--ring))',
        background: 'hsl(var(--background))',
        foreground: 'hsl(var(--foreground))',
        primary: {
          DEFAULT: 'hsl(var(--primary))',
          foreground: 'hsl(var(--primary-foreground))',
        },
        secondary: {
          DEFAULT: 'hsl(var(--secondary))',
          foreground: 'hsl(var(--secondary-foreground))',
        },
        destructive: {
          DEFAULT: 'hsl(var(--destructive))',
          foreground: 'hsl(var(--destructive-foreground))',
        },
        muted: {
          DEFAULT: 'hsl(var(--muted))',
          foreground: 'hsl(var(--muted-foreground))',
        },
        accent: {
          DEFAULT: 'hsl(var(--accent))',
          foreground: 'hsl(var(--accent-foreground))',
        },
        popover: {
          DEFAULT: 'hsl(var(--popover))',
          foreground: 'hsl(var(--popover-foreground))',
        },
        card: {
          DEFAULT: 'hsl(var(--card))',
          foreground: 'hsl(var(--card-foreground))',
        },

        // Brand colors from design tokens
        brand: {
          50: 'hsl(210, 100%, 97%)',
          100: 'hsl(210, 100%, 94%)',
          200: 'hsl(210, 100%, 87%)',
          300: 'hsl(210, 100%, 78%)',
          400: 'hsl(210, 100%, 66%)',
          500: 'hsl(210, 100%, 56%)',
          600: 'hsl(210, 100%, 47%)',
          700: 'hsl(210, 100%, 39%)',
          800: 'hsl(210, 100%, 31%)',
          900: 'hsl(210, 100%, 24%)',
          950: 'hsl(210, 100%, 15%)',
        },

        // Semantic colors
        success: {
          50: 'hsl(142, 76%, 96%)',
          100: 'hsl(142, 76%, 91%)',
          200: 'hsl(142, 76%, 81%)',
          300: 'hsl(142, 76%, 69%)',
          400: 'hsl(142, 76%, 55%)',
          500: 'hsl(142, 76%, 45%)',
          600: 'hsl(142, 76%, 36%)',
          700: 'hsl(142, 76%, 29%)',
          800: 'hsl(142, 76%, 24%)',
          900: 'hsl(142, 76%, 20%)',
          950: 'hsl(142, 76%, 12%)',
        },
        warning: {
          50: 'hsl(48, 96%, 96%)',
          100: 'hsl(48, 96%, 89%)',
          200: 'hsl(48, 96%, 76%)',
          300: 'hsl(48, 96%, 61%)',
          400: 'hsl(48, 96%, 53%)',
          500: 'hsl(48, 96%, 47%)',
          600: 'hsl(48, 96%, 37%)',
          700: 'hsl(48, 96%, 27%)',
          800: 'hsl(48, 96%, 23%)',
          900: 'hsl(48, 96%, 20%)',
          950: 'hsl(48, 96%, 11%)',
        },
        error: {
          50: 'hsl(0, 86%, 97%)',
          100: 'hsl(0, 86%, 94%)',
          200: 'hsl(0, 86%, 87%)',
          300: 'hsl(0, 86%, 78%)',
          400: 'hsl(0, 86%, 67%)',
          500: 'hsl(0, 86%, 55%)',
          600: 'hsl(0, 86%, 47%)',
          700: 'hsl(0, 86%, 39%)',
          800: 'hsl(0, 86%, 31%)',
          900: 'hsl(0, 86%, 26%)',
          950: 'hsl(0, 86%, 15%)',
        },
      },

      // Enhanced typography
      fontFamily: {
        display: ['Playfair Display', 'Georgia', 'serif'],
        body: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
        mono: ['JetBrains Mono', 'Menlo', 'Monaco', 'monospace'],
      },

      // Enhanced spacing scale
      spacing: {
        '0.5': '0.125rem',
        '1.5': '0.375rem',
        '2.5': '0.625rem',
        '3.5': '0.875rem',
        '18': '4.5rem',
        '88': '22rem',
        '100': '25rem',
        '112': '28rem',
        '128': '32rem',
      },

      // Enhanced border radius
      borderRadius: {
        lg: 'var(--radius)',
        md: 'calc(var(--radius) - 2px)',
        sm: 'calc(var(--radius) - 4px)',
        '4xl': '2rem',
      },

      // Enhanced shadows
      boxShadow: {
        'glow': '0 0 20px rgb(59 130 246 / 0.5)',
        'warm': '0 8px 32px rgb(251 146 60 / 0.35)',
        'cool': '0 8px 32px rgb(14 165 233 / 0.35)',
      },

      // Enhanced animations
      keyframes: {
        'accordion-down': {
          from: { height: '0' },
          to: { height: 'var(--radix-accordion-content-height)' },
        },
        'accordion-up': {
          from: { height: 'var(--radix-accordion-content-height)' },
          to: { height: '0' },
        },
        'fade-in': {
          '0%': { opacity: '0', transform: 'translateY(10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        'fade-out': {
          '0%': { opacity: '1', transform: 'translateY(0)' },
          '100%': { opacity: '0', transform: 'translateY(-10px)' },
        },
        'slide-in-right': {
          '0%': { transform: 'translateX(100%)' },
          '100%': { transform: 'translateX(0)' },
        },
        'slide-out-right': {
          '0%': { transform: 'translateX(0)' },
          '100%': { transform: 'translateX(100%)' },
        },
        'bounce-in': {
          '0%': { transform: 'scale(0.3)', opacity: '0' },
          '50%': { transform: 'scale(1.05)' },
          '70%': { transform: 'scale(0.9)' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
        'pulse-glow': {
          '0%, 100%': { boxShadow: '0 0 5px rgb(59 130 246 / 0.5)' },
          '50%': { boxShadow: '0 0 20px rgb(59 130 246 / 0.8)' },
        },
      },

      animation: {
        'accordion-down': 'accordion-down 0.2s ease-out',
        'accordion-up': 'accordion-up 0.2s ease-out',
        'fade-in': 'fade-in 0.3s ease-out',
        'fade-out': 'fade-out 0.3s ease-out',
        'slide-in-right': 'slide-in-right 0.3s ease-out',
        'slide-out-right': 'slide-out-right 0.3s ease-out',
        'bounce-in': 'bounce-in 0.6s cubic-bezier(0.68, -0.55, 0.265, 1.55)',
        'pulse-glow': 'pulse-glow 2s ease-in-out infinite',
      },

      // Enhanced transitions
      transitionDuration: {
        '400': '400ms',
        '600': '600ms',
        '800': '800ms',
        '900': '900ms',
      },

      transitionTimingFunction: {
        'spring': 'cubic-bezier(0.34, 1.56, 0.64, 1)',
        'bounce': 'cubic-bezier(0.68, -0.55, 0.265, 1.55)',
      },
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/typography'),
  ],
}
