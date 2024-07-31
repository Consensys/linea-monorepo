"use client";

import Image from "next/image";
import { ToastContainer } from "react-toastify";
import atypTextFont from "@/app/font/atypText";
import atypFont from "@/app/font/atyp";
import Header from "../header/Header";
import SwitchNetwork from "../widgets/SwitchNetwork";
import Footer from "../footer/Footer";
import useInitialiseChain from "@/hooks/useInitialiseChain";
import useInitialiseToken from "@/hooks/useInitialiseToken";

export function Layout({ children }: { children: React.ReactNode }) {
  useInitialiseChain();
  useInitialiseToken();

  return (
    <div
      className={`${atypFont.variable} ${atypTextFont.variable} ${atypFont.className} flex min-h-screen flex-col bg-hero bg-cover bg-no-repeat`}
    >
      <ToastContainer
        position="top-center"
        autoClose={2000}
        hideProgressBar={false}
        pauseOnFocusLoss={false}
        theme="dark"
      />
      <Header />
      <main className="container flex flex-1 flex-col items-center justify-center">
        <Image
          src={"/images/picto/overlay-vertical.svg"}
          alt="Linea"
          height={0}
          width={0}
          style={{ width: "132px", height: "auto" }}
          className="pointer-events-none absolute -top-8 right-48 hidden lg:block"
        />
        <Image
          src={"/images/picto/overlay.svg"}
          alt="Linea"
          height={0}
          width={0}
          style={{ width: "210px", height: "auto" }}
          className="4xl:-left-40 pointer-events-none absolute bottom-40 left-0 hidden lg:block"
        />
        {children}
      </main>
      <SwitchNetwork />
      <Footer />
    </div>
  );
}
