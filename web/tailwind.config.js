/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx,svelte}",
  ],
  theme: {
    extend: {
      colors: {
        monokai: {
          pink: '#FF6188',
          orange: '#FC9867',
          cyan: '#78DCE8',
          green: '#A9DC76',
          yellow: '#FFD866',
          purple: '#AB9DF2',
          bg: '#000000',
          'bg-light': '#1a1a1a',
          text: '#FCFCFA',
          'text-muted': '#939293',
          'text-dim': '#5c5c5c',
          border: '#2d2d2d',
        },
        light: {
          pink: '#D91B5B',
          orange: '#D97B3C',
          cyan: '#0E7490',
          green: '#16803C',
          yellow: '#A16207',
          purple: '#7C3AED',
          bg: '#FAFAFA',
          'bg-light': '#F4F4F5',
          text: '#18181B',
          'text-muted': '#71717A',
          'text-dim': '#A1A1AA',
          border: '#E4E4E7',
        },
      },
    },
  },
  plugins: [],
  darkMode: 'class',
}
