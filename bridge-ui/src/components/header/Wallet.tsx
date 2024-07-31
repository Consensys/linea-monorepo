"use client";

import { useEffect, useRef } from "react";
import Image from "next/image";
import { useAccount, useDisconnect } from "wagmi";
import { MdLogout } from "react-icons/md";
import classNames from "classnames";
import { formatAddress } from "@/utils/format";
import { useWeb3Modal } from "@web3modal/wagmi/react";

type Props = {
  className?: string;
};

export default function Wallet({ className = "" }: Props) {
  const detailsRef = useRef<HTMLDetailsElement>(null);

  const { address, isConnected } = useAccount();
  const { disconnect } = useDisconnect();
  const { open } = useWeb3Modal();

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
      <button
        id="wallet-connect-btn"
        className={classNames(className, {
          "btn btn-primary rounded-full uppercase text-sm md:text-[0.9375rem] font-semibold": !className,
        })}
        onClick={() => open()}
      >
        Connect Wallet
      </button>
    </div>
  );
}
