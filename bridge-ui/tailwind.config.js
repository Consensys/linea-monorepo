/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      backgroundImage: {
        hero: "url('/bridge_bg.png')",
      },
      colors: {
        primary: "#61DFFF",
        card: "#505050",
        cardBg: "#1D1D1D"
      },
      fontFamily: {
        atypText: ["var(--font-atyp-text)"]
      },
    },
  },

  daisyui: {
    themes: [
      {
        dark: {
          ...require('daisyui/src/theming/themes')['[data-theme=dark]'],
          primary: '#61DFFF',
          'primary-content': '#000000',
          info: '#fff',
        },
      },
    ],
  },
  plugins: [require('daisyui'), require('tailwind-scrollbar')],
};
