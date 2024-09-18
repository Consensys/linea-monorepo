import localFont from "next/font/local";

const atypTextFont = localFont({
  display: "swap",
  src: [
    {
      path: "../../../public/fonts/AtypText-Light-subset.woff2",
      weight: "300",
      style: "normal",
    },
    {
      path: "../../../public/fonts/AtypText-LightItalic.woff2",
      weight: "300",
      style: "italic",
    },
    {
      path: "../../../public/fonts/AtypText-Light.woff2",
      weight: "300",
      style: "normal",
    },
    {
      path: "../../../public/fonts/AtypText-Regular-subset.woff2",
      weight: "400",
      style: "normal",
    },
    {
      path: "../../../public/fonts/AtypText-Regular.woff2",
      weight: "400",
      style: "normal",
    },
    {
      path: "../../../public/fonts/AtypText-Italic-subset.woff2",
      weight: "400",
      style: "italic",
    },
    {
      path: "../../../public/fonts/AtypText-Italic.woff2",
      weight: "400",
      style: "italic",
    },
    {
      path: "../../../public/fonts/AtypText-Medium-subset.woff2",
      weight: "500",
      style: "normal",
    },
    {
      path: "../../../public/fonts/AtypText-Medium.woff2",
      weight: "500",
      style: "normal",
    },
    {
      path: "../../../public/fonts/AtypText-MediumItalic.woff2",
      weight: "500",
      style: "italic",
    },
    {
      path: "../../../public/fonts/AtypText-Semibold-subset.woff2",
      weight: "600",
      style: "normal",
    },
    {
      path: "../../../public/fonts/AtypText-Semibold.woff2",
      weight: "600",
      style: "normal",
    },
    {
      path: "../../../public/fonts/AtypText-SemiboldItalic.woff2",
      weight: "600",
      style: "italic",
    },
    {
      path: "../../../public/fonts/AtypText-Bold-subset.woff2",
      weight: "700",
      style: "normal",
    },
    {
      path: "../../../public/fonts/AtypText-Bold.woff2",
      weight: "700",
      style: "normal",
    },
    {
      path: "../../../public/fonts/AtypText-BoldItalic.woff2",
      weight: "700",
      style: "italic",
    },
  ],
  variable: "--font-atyp-text",
});

export default atypTextFont;
