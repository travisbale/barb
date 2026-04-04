/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        bg: 'rgb(var(--color-bg) / <alpha-value>)',
        surface: {
          DEFAULT: 'rgb(var(--color-surface) / <alpha-value>)',
          hover: 'rgb(var(--color-surface-hover) / <alpha-value>)',
        },
        edge: {
          DEFAULT: 'rgb(var(--color-border) / <alpha-value>)',
          subtle: 'rgb(var(--color-border-subtle) / <alpha-value>)',
        },
        primary: 'rgb(var(--color-text) / <alpha-value>)',
        muted: 'rgb(var(--color-text-muted) / <alpha-value>)',
        dim: 'rgb(var(--color-text-dim) / <alpha-value>)',
        amber: {
          DEFAULT: 'rgb(var(--color-amber) / <alpha-value>)',
          dim: 'rgb(var(--color-amber-dim) / <alpha-value>)',
          glow: 'rgb(var(--color-amber-glow) / 0.08)',
        },
        danger: {
          DEFAULT: 'rgb(var(--color-red) / <alpha-value>)',
          dim: 'rgb(var(--color-red-dim) / 0.06)',
        },
        teal: {
          DEFAULT: 'rgb(var(--color-teal) / <alpha-value>)',
          dim: 'rgb(var(--color-teal-dim) / 0.06)',
        },
        green: {
          DEFAULT: 'rgb(var(--color-green) / <alpha-value>)',
          dim: 'rgb(var(--color-green-dim) / 0.06)',
        },
      },
      fontFamily: {
        sans: ['DM Sans', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
    },
  },
  plugins: [],
}
