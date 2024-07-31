import { Metadata } from "next";
import { headers } from "next/headers";
import Script from "next/script";
import { cookieToInitialState } from "wagmi";
import { Inter } from "next/font/google";
import { wagmiConfig } from "@/config";
import usabillaBeScript from "@/scripts/usabilla";
import { gtmScript, gtmNoScript } from "@/scripts/gtm";
import { Providers } from "@/components/layouts/Providers";
import { Layout } from "@/components/layouts/Layout";
import "react-toastify/dist/ReactToastify.css";
import "./globals.css";

const inter = Inter({ subsets: ["latin"] });

const metadata: Metadata = {
  title: "Linea Bridge",
  description: `Linea Bridge is a bridge solution, providing secure and efficient cross-chain transactions between Layer 1 and Linea networks.
  Discover the future of blockchain interaction with Linea Bridge.`,
  icons: {
    icon: "./favicon.png",
  },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  const initialState = cookieToInitialState(wagmiConfig, headers().get("cookie"));

  return (
    <html lang="en" data-theme="dark">
      <title>{metadata.title?.toString()}</title>
      <meta name="description" content={metadata.description?.toString()} key="desc" />

      <body className={inter.className}>
        <noscript dangerouslySetInnerHTML={{ __html: gtmNoScript }} />

        <Providers initialState={initialState}>
          <Layout>{children}</Layout>
        </Providers>
      </body>

      <Script id="usabilla" dangerouslySetInnerHTML={{ __html: usabillaBeScript }} />
      <Script id="gtm" dangerouslySetInnerHTML={{ __html: gtmScript }} />
    </html>
  );
}
