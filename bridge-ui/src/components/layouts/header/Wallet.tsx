"use client";

import { useEffect, useRef } from "react";
import Image from "next/image";
import { useAccount, useDisconnect } from "wagmi";
import { MdLogout } from "react-icons/md";
import { formatAddress } from "@/utils/format";
import ConnectButton from "@/components/ConnectButton";

export default function Wallet() {
  const detailsRef = useRef<HTMLDetailsElement>(null);

  const { address, isConnected } = useAccount();
  const { disconnect } = useDisconnect();

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

  if (isConnected) {
    return (
      <details ref={detailsRef}>
        <summary className="rounded-full">
          <Image
            src={"/images/logo/metamask.svg"}
            alt="MetaMask"
            width={0}
            height={0}
            style={{ width: "18px", height: "auto" }}
          />
          <span className="hidden md:block">{formatAddress(address)}</span>
        </summary>
        <ul className="right-0 z-10 bg-base-100 p-2">
          <li>
            <button id="wallet-disconnect-btn" onClick={() => disconnect()} className="rounded-full">
              <MdLogout className="text-xl" />
              Disconnect
            </button>
          </li>
        </ul>
      </details>
      //
    );
  }

  return (
    <div className="p-0">
      <ConnectButton />
    </div>
  );
}
