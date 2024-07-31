import localFont from "next/font/local";

const atypTextFont = localFont({
  src: [
    {
      path: "../assets/fonts/AtypText-Light.woff2",
      weight: "300",
      style: "normal",
    },
    {
      path: "../assets/fonts/AtypText-LightItalic.woff2",
      weight: "300",
      style: "italic",
    },
    {
      path: "../assets/fonts/AtypText-Regular.woff2",
      weight: "400",
      style: "normal",
    },
    {
      path: "../assets/fonts/AtypText-Italic.woff2",
      weight: "400",
      style: "italic",
    },
    {
      path: "../assets/fonts/AtypText-Medium.woff2",
      weight: "500",
      style: "normal",
    },
    {
      path: "../assets/fonts/AtypText-MediumItalic.woff2",
      weight: "500",
      style: "italic",
    },
    {
      path: "../assets/fonts/AtypText-Semibold.woff2",
      weight: "600",
      style: "normal",
    },
    {
      path: "../assets/fonts/AtypText-SemiboldItalic.woff2",
      weight: "600",
      style: "italic",
    },
    {
      path: "../assets/fonts/AtypText-Bold.woff2",
      weight: "700",
      style: "normal",
    },
    {
      path: "../assets/fonts/AtypText-BoldItalic.woff2",
      weight: "700",
      style: "italic",
    },
  ],
  variable: "--font-atyp-text",
});

export default atypTextFont;
