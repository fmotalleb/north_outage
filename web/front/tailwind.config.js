/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,jsx}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Vazirmatn', 'Inter', 'system-ui', 'sans-serif'],
      },
      colors: {
        ink: {
          950: '#070b14',
          900: '#0c1220',
          800: '#131a2b',
          700: '#1b2438',
          600: '#273149',
          500: '#3a4666',
        },
      },
      keyframes: {
        pulse_soft: {
          '0%, 100%': { opacity: 0.9 },
          '50%': { opacity: 0.55 },
        },
        slide_up: {
          from: { opacity: 0, transform: 'translateY(8px)' },
          to: { opacity: 1, transform: 'translateY(0)' },
        },
      },
      animation: {
        pulse_soft: 'pulse_soft 2.4s ease-in-out infinite',
        slide_up: 'slide_up 0.35s ease-out both',
      },
    },
  },
  plugins: [],
}
