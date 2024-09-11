"use client";

import { ToastContainer } from "react-toastify";
import atypTextFont from "@/app/font/atypText";
import atypFont from "@/app/font/atyp";
import { Header } from "./header";
import { useInitialiseChain, useInitialiseToken } from "@/hooks";
import Sidebar from "./Sidebar";

export function Layout({ children }: { children: React.ReactNode }) {
  useInitialiseChain();
  useInitialiseToken();

  return (
    <div
      className={`${atypFont.variable} ${atypTextFont.variable} ${atypFont.className} flex min-h-screen flex-col bg-cover bg-no-repeat`}
    >
      <ToastContainer
        position="top-center"
        autoClose={2000}
        hideProgressBar={false}
        pauseOnFocusLoss={false}
        theme="dark"
      />
      <Sidebar />
      <div className="md:ml-64">
        <Header />
      </div>
      <main className="m-0 flex-1 p-3 md:ml-64 md:p-10">{children}</main>
    </div>
  );
}
