'use client';

import { useRef, useContext, useEffect } from 'react';
import Image from 'next/image';
import { switchNetwork } from '@wagmi/core';
import log from 'loglevel';

import EthereumLogo from 'public/images/logo/ethereum.svg';
import LineaSepoliaLogo from 'public/images/logo/linea-sepolia.svg';
import { ChainContext, NetworkType } from '@/contexts/chain.context';
import ChainLogo from '@/components/widgets/ChainLogo';

export default function Chains() {
  // Context
  const context = useContext(ChainContext);
  const { networkType, activeChain, alternativeChain, resetToken } = context;

  const detailsRef = useRef<HTMLDetailsElement>(null);

  const switchChain = async () => {
    try {
      resetToken();
      alternativeChain &&
        (await switchNetwork({
          chainId: alternativeChain.id,
        }));
    } catch (error) {
      log.error(error);
    }
  };

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (detailsRef.current && !detailsRef.current.contains(e.target as Node)) {
        detailsRef.current.removeAttribute('open');
      }
    };
    document.addEventListener('click', handleClickOutside);
    return () => {
      document.removeEventListener('click', handleClickOutside);
    };
  }, []);

  if (networkType === NetworkType.WRONG_NETWORK) {
    return (
      <details ref={detailsRef}>
        <summary>Wrong Network</summary>
        <ul className="right-0 z-10 p-2 bg-base-100">
          <li>
            <button id="switch-active-chain-btn" onClick={switchChain}>
              <Image src={EthereumLogo} alt="Linea" width={12} height={12} /> {activeChain && activeChain.name}
            </button>
          </li>
          <li>
            <button id="switch-alternative-chain-btn" onClick={switchChain}>
              <Image src={LineaSepoliaLogo} alt="Linea" width={18} height={18} />{' '}
              {alternativeChain && alternativeChain.name}
            </button>
          </li>
        </ul>
      </details>
    );
  }

  return (
    <details ref={detailsRef}>
      <summary className="rounded-full" id="chain-select">
        {activeChain && <ChainLogo chainId={activeChain.id} />}{' '}
        <span className="hidden md:block" id="active-chain-name">
          {activeChain && activeChain.name}
        </span>
      </summary>
      <ul className="right-0 z-10 p-2 bg-base-100">
        <li>
          <button id="switch-alternative-chain-btn" onClick={switchChain} className="rounded-full min-w-max">
            {alternativeChain && <ChainLogo chainId={alternativeChain.id} />}{' '}
            {alternativeChain && alternativeChain.name}
          </button>
        </li>
      </ul>
    </details>
  );
}
