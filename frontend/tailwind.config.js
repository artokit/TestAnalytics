/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        base: '#12141c',
        surface: '#1a1d29',
        surface2: '#23273a',
        line: '#2a2e3f',
        muted: '#8b90a7',
        faint: '#5b5f77',
        accent: '#6c7bff',
        ok: '#4fd48e',
        warn: '#f0b43c',
        danger: '#eb5a5a',
        ink: '#e7e9f2',
      },
    },
  },
  plugins: [],
}
