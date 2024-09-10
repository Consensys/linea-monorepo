import type { Config } from "tailwindcss";
import daisyui from "daisyui";
import tailwindScrollbar from "tailwind-scrollbar";
import daisyuiThemes from "daisyui/src/theming/themes";

const config: Config = {
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: "#61DFFF",
        secondary: "#FF62E6",
        card: "#505050",
        cardBg: "#1D1D1D",
        success: "#C1FF14",
      },
      fontFamily: {
        atypText: ["var(--font-atyp-text)"],
      },
    },
  },

  daisyui: {
    themes: [
      {
        dark: {
          ...daisyuiThemes.dark,
          primary: "#61DFFF",
          secondary: "#FF62E6",
          "primary-content": "#000000",
          info: "#61DFFF",
          success: "#C1FF14",
        },
      },
    ],
  },
  plugins: [daisyui, tailwindScrollbar],
};

export default config;
