'use client';

import { Metadata } from 'next';
import Image from 'next/image';
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

import './globals.css';
import atypFont from './font/atyp';
import atypTextFont from './font/atypText';
import Footer from '@/components/footer/Footer';
import Header from '@/components/header/Header';
import { Inter } from 'next/font/google';
import { ChainProvider } from '@/contexts/chain.context';
import SwitchTestnet from '@/components/widgets/SwitchNetwork';
import { config } from '@/config';
import { HistoryProvider } from '@/contexts/history.context';

import OverlayV from 'public/images/picto/overlay-vertical.svg';
import OverlayH from 'public/images/picto/overlay.svg';
import { TokenProvider } from '@/contexts/token.context';
import { UIProvider } from '@/contexts/ui.context';
import { useEffectOnce } from 'usehooks-ts';
import { StorageKeys } from '@/contexts/storage';
import PackageJSON from '@/../package.json';
import { compare } from 'compare-versions';
import { useState } from 'react';
import Script from 'next/script';
import usabillaBeScript from '@/scripts/usabilla';
import { gtmScript, gtmNoScript } from '@/scripts/gtm';
import { Web3ModalContext } from '@/contexts/web3Modal.context';

const inter = Inter({ subsets: ['latin'] });

const metadata: Metadata = {
  title: 'Linea Bridge',
  description: `Linea Bridge is a bridge solution, providing secure and efficient cross-chain transactions between Layer 1 and Linea networks.
  Discover the future of blockchain interaction with Linea Bridge.`,
  icons: {
    icon: './favicon.png',
  },
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  // We have to wait for the storage to clear before starting rendering
  // the child components since they depend on it
  const [storageLoading, setStorageLoading] = useState(true);

  // Reset localStorage if storage version too old
  useEffectOnce(() => {
    const currentVersion = localStorage.getItem(StorageKeys.STORAGE_VERSION);
    const currentVersionNumber = currentVersion || '0.0.0';
    if (compare(currentVersionNumber, config.storage.minVersion, '<')) {
      localStorage.clear();
      localStorage.setItem(StorageKeys.STORAGE_VERSION, PackageJSON.version);
    }
    setStorageLoading(false);
  });

  return (
    <html lang="en" data-theme="dark">
      <title>{metadata.title?.toString()}</title>
      <meta name="description" content={metadata.description?.toString()} key="desc" />

      <body className={inter.className}>
        <noscript dangerouslySetInnerHTML={{ __html: gtmNoScript }} />

        <UIProvider>
          {!storageLoading && (
            <ChainProvider>
              <TokenProvider>
                <Web3ModalContext>
                  <HistoryProvider>
                    <div
                      className={`${atypFont.variable} ${atypTextFont.variable} ${atypFont.className} flex flex-col min-h-screen bg-hero bg-cover bg-no-repeat`}
                    >
                      <ToastContainer
                        position="top-center"
                        autoClose={2000}
                        hideProgressBar={false}
                        pauseOnFocusLoss={false}
                        // icon={false}
                        theme="dark"
                      />
                      <Header />
                      <main className="container flex flex-col items-center justify-center flex-1">
                        <Image
                          src={OverlayV}
                          alt="Linea"
                          width={132}
                          className="absolute hidden pointer-events-none lg:block right-48 -top-8"
                        />
                        <Image
                          src={OverlayH}
                          alt="Linea"
                          width={210}
                          className="absolute hidden pointer-events-none lg:block left-0 bottom-40 4xl:-left-40"
                        />
                        {children}
                      </main>
                      <SwitchTestnet />
                      <Footer />
                    </div>
                  </HistoryProvider>
                </Web3ModalContext>
              </TokenProvider>
            </ChainProvider>
          )}
        </UIProvider>
      </body>

      <Script id="usabilla" dangerouslySetInnerHTML={{ __html: usabillaBeScript }} />
      <Script id="gtm" dangerouslySetInnerHTML={{ __html: gtmScript }} />
    </html>
  );
}
