"use client";

import { useContext } from "react";
import Image from "next/image";
import { useAccount } from "wagmi";
import Wallet from "./Wallet";
import Chains from "./Chains";
import { UIContext } from "@/contexts/ui.context";
import { NetworkType } from "@/config";
import { useChainStore } from "@/stores/chainStore";

export default function Header() {
  // Hooks
  const { isConnected } = useAccount();

  // Context
  const networkType = useChainStore((state) => state.networkType);

  const { toggleShowBridge } = useContext(UIContext);

  return (
    <header className="container navbar py-4">
      <div className="flex-1">
        <button
          className="btn btn-ghost w-32 -space-y-2 text-xl normal-case text-white md:w-52 md:space-y-0"
          onClick={() => toggleShowBridge(false)}
        >
          <Image
            src={"/images/logo/linea.svg"}
            alt="Linea"
            width={0}
            height={0}
            style={{ width: "215px", height: "auto" }}
            priority
          />
        </button>
        {networkType === NetworkType.SEPOLIA && (
          <div className="badge badge-primary badge-outline ml-10 gap-2">TESTNET</div>
        )}
      </div>
      <div className="flex-none">
        <ul className="menu menu-horizontal px-1">
          {isConnected && (
            <li>
              <Chains />
            </li>
          )}
          <li>
            <Wallet />
          </li>
        </ul>
      </div>
    </header>
  );
}
