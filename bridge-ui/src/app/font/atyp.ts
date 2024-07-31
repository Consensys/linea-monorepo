import localFont from "next/font/local";

const atypFont = localFont({
  src: [
    {
      path: "../assets/fonts/AtypDisplay-Light.woff2",
      weight: "300",
      style: "normal",
    },
    {
      path: "../assets/fonts/AtypDisplay-LightItalic.woff2",
      weight: "300",
      style: "italic",
    },
    {
      path: "../assets/fonts/AtypDisplay-Regular.woff2",
      weight: "400",
      style: "normal",
    },
    {
      path: "../assets/fonts/AtypDisplay-Italic.woff2",
      weight: "400",
      style: "italic",
    },
    {
      path: "../assets/fonts/AtypDisplay-Medium.woff2",
      weight: "500",
      style: "normal",
    },
    {
      path: "../assets/fonts/AtypDisplay-MediumItalic.woff2",
      weight: "500",
      style: "italic",
    },
    {
      path: "../assets/fonts/AtypDisplay-Semibold.woff2",
      weight: "600",
      style: "normal",
    },
    {
      path: "../assets/fonts/AtypDisplay-SemiboldItalic.woff2",
      weight: "600",
      style: "italic",
    },
    {
      path: "../assets/fonts/AtypDisplay-Bold.woff2",
      weight: "700",
      style: "normal",
    },
    {
      path: "../assets/fonts/AtypDisplay-BoldItalic.woff2",
      weight: "700",
      style: "italic",
    },
  ],
  variable: "--font-atyp",
});

export default atypFont;
