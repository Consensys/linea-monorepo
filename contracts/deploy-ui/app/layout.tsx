import type { Metadata } from "next";
import { ReactNode } from "react";

export const metadata: Metadata = {
  title: "Contracts Deploy UI",
  description: "Local wallet signing UI for contracts deployments",
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body style={{ margin: 0, backgroundColor: "#0b1020", color: "#e5e7eb", fontFamily: "Arial, sans-serif" }}>
        {children}
      </body>
    </html>
  );
}
