"use client";

import { useRef, useEffect } from "react";
import Image from "next/image";
import { switchChain } from "@wagmi/core";
import log from "loglevel";
import ChainLogo from "@/components/widgets/ChainLogo";
import { NetworkType, wagmiConfig } from "@/config";
import { useChainStore } from "@/stores/chainStore";

export default function Chains() {
  // Context
  const { networkType, fromChain, toChain, resetToken } = useChainStore((state) => ({
    networkType: state.networkType,
    fromChain: state.fromChain,
    toChain: state.toChain,
    resetToken: state.resetToken,
  }));

  const detailsRef = useRef<HTMLDetailsElement>(null);

  const switchNetwork = async () => {
    try {
      resetToken();
      toChain &&
        (await switchChain(wagmiConfig, {
          chainId: toChain.id,
        }));
    } catch (error) {
      log.error(error);
    }
  };

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (detailsRef.current && !detailsRef.current.contains(e.target as Node)) {
        detailsRef.current.removeAttribute("open");
      }
    };
    document.addEventListener("click", handleClickOutside);
    return () => {
      document.removeEventListener("click", handleClickOutside);
    };
  }, []);

  if (networkType === NetworkType.WRONG_NETWORK) {
    return (
      <details ref={detailsRef}>
        <summary>Wrong Network</summary>
        <ul className="right-0 z-10 bg-base-100 p-2">
          <li>
            <button id="switch-active-chain-btn" onClick={switchNetwork}>
              <Image
                src={"/images/logo/ethereum.svg"}
                alt="Linea"
                width={0}
                height={0}
                style={{ width: "12px", height: "auto" }}
              />{" "}
              {fromChain && fromChain.name}
            </button>
          </li>
          <li>
            <button id="switch-alternative-chain-btn" onClick={switchNetwork}>
              <Image
                src={"/images/logo/linea-sepolia.svg"}
                alt="Linea"
                width={0}
                height={0}
                style={{ width: "18px", height: "auto" }}
              />{" "}
              {toChain && toChain.name}
            </button>
          </li>
        </ul>
      </details>
    );
  }

  return (
    <details ref={detailsRef}>
      <summary className="rounded-full" id="chain-select">
        {fromChain && <ChainLogo chainId={fromChain.id} />}{" "}
        <span className="hidden md:block" id="active-chain-name">
          {fromChain && fromChain.name}
        </span>
      </summary>
      <ul className="right-0 z-10 bg-base-100 p-2">
        <li>
          <button id="switch-alternative-chain-btn" onClick={switchNetwork} className="min-w-max rounded-full">
            {toChain && <ChainLogo chainId={toChain.id} />} {toChain && toChain.name}
          </button>
        </li>
      </ul>
    </details>
  );
}
