import type { Metadata } from "next";
import { ReactNode } from "react";

import "./globals.css";

export const metadata: Metadata = {
  title: "Linea · Contracts deploy UI",
  description: "Local wallet signing UI for Linea contracts deployments (Hardhat DEPLOY_WITH_UI).",
  icons: {
    icon: [{ url: "https://linea.build/favicon-32x32.png", sizes: "32x32", type: "image/png" }],
  },
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body className="deploy-ui-body">{children}</body>
    </html>
  );
}
