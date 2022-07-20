const { url } = require('inspector')
const defaultTheme = require('tailwindcss/defaultTheme')

module.exports = {
  content: [
    "./pages/**/*.{js,ts,jsx,tsx}",
    "./components/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      fontFamily: {
        'sans': ['Gilroy-Regular', ...defaultTheme.fontFamily.sans],
        GilroyRegular: ['Gilroy-Regular'],
        GilroyMedium: ['Gilroy-Medium'],
        GilroyLight: ['Gilroy-Light'],
        GilroyHeavy: ['Gilroy-Heavy'],
        GilroyBold: ['Gilroy-Bold'],
      },
      animation: {
        wiggle: 'wiggle 1s ease-in-out infinite',
      },
      keyframes: {
        wiggle: {
          '0%, 100%': { transform: 'rotate(-1deg)' },
          '50%': { transform: 'rotate(1deg)' },
        }
      },
      colors: {
        transparent: "transparent",
        'black': '#1A1A1A',
        'white': '#FFFFFF',
        'primary': '#9A1750',
        'primary-dark': '#5D001E',
        'secondary': '#E3AFBC',
        'secondary-dark': '#EE4C7C'
      },
      fontSize: {
        '4b5': ['2.9rem', '1'],
        'xxl': ['9rem', '1'],
        'xxxl': ['16rem', '1.2'],
      },
      height: {
        '128': '32rem', 
      }
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('tailwind-scrollbar-hide')
  ],
  corePlugins: {
    fontFamily: true,
  },
}