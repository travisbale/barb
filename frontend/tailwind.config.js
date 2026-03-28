/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        bg: 'var(--color-bg)',
        surface: { DEFAULT: 'var(--color-surface)', hover: 'var(--color-surface-hover)' },
        edge: { DEFAULT: 'var(--color-border)', subtle: 'var(--color-border-subtle)' },
        muted: 'var(--color-text-muted)',
        dim: 'var(--color-text-dim)',
        amber: { DEFAULT: 'var(--color-amber)', dim: 'var(--color-amber-dim)', glow: 'var(--color-amber-glow)' },
        danger: { DEFAULT: 'var(--color-red)', dim: 'var(--color-red-dim)' },
        teal: { DEFAULT: 'var(--color-teal)', dim: 'var(--color-teal-dim)' },
      },
      fontFamily: {
        sans: ['DM Sans', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
    },
  },
  plugins: [],
}
