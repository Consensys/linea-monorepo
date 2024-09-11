"use client";

import { useEffect, useRef } from "react";
import Image from "next/image";
import { switchChain } from "@wagmi/core";
import log from "loglevel";
import { config, NetworkLayer, NetworkType, wagmiConfig } from "@/config";
import { useChainStore } from "@/stores/chainStore";
import DropdownItem from "@/components/DropdownItem";

const networks = config.networks;

export function Chains() {
  const { networkType, networkLayer, resetToken } = useChainStore((state) => ({
    networkType: state.networkType,
    networkLayer: state.networkLayer,
    resetToken: state.resetToken,
  }));

  const detailsRef = useRef<HTMLDetailsElement>(null);

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

  const switchNetworkHandler = async (chainId: number) => {
    if (networkLayer === NetworkLayer.UNKNOWN) {
      return;
    }

    try {
      resetToken();
      await switchChain(wagmiConfig, {
        chainId,
      });
    } catch (error) {
      log.error(error);
    }
  };

  if (networkType == NetworkType.SEPOLIA || networkType == NetworkType.MAINNET) {
    return (
      <details className="dropdown relative" ref={detailsRef}>
        <summary className="flex cursor-pointer items-center gap-2 rounded-full border-2 border-card p-2 px-3">
          <Image
            src={
              networkType === NetworkType.SEPOLIA ? "/images/logo/linea-sepolia.svg" : "/images/logo/linea-mainnet.svg"
            }
            alt="MetaMask"
            width={0}
            height={0}
            style={{ width: "18px", height: "auto" }}
          />

          <span className="hidden md:block">{networkType === NetworkType.SEPOLIA ? "Linea Sepolia" : "Linea"}</span>
        </summary>
        <ul className="menu dropdown-content absolute right-0 z-10 mt-2 min-w-max border-2 border-card bg-cardBg p-0 shadow">
          <DropdownItem
            title={config.networks.MAINNET.L2.name}
            iconPath="/images/logo/linea-mainnet.svg"
            onClick={() => switchNetworkHandler(networks.MAINNET.L1.chainId)}
          />
          <DropdownItem
            title={config.networks.SEPOLIA.L2.name}
            iconPath="/images/logo/linea-sepolia.svg"
            onClick={() => switchNetworkHandler(networks.SEPOLIA.L1.chainId)}
          />
        </ul>
      </details>
    );
  }
}
