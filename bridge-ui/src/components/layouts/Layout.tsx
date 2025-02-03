"use client";

import Header from "../v2/header";
import { useInitialiseChain } from "@/hooks";
import { useAccount } from "wagmi";
import { linea, lineaSepolia, mainnet, sepolia } from "viem/chains";
import WrongNetwork from "./WrongNetwork";
import { Theme } from "@/types";
import Image from "next/image";
import { usePathname } from "next/navigation";
import clsx from "clsx";

export function Layout({ children }: { children: React.ReactNode }) {
  useInitialiseChain();

  const { chainId } = useAccount();
  const pathname = usePathname();

  return chainId &&
    ![mainnet.id, sepolia.id, linea.id, lineaSepolia.id].includes(chainId as 1 | 11155111 | 59144 | 59141) ? (
    <WrongNetwork />
  ) : (
    <div className="layout">
      <div className="container-v2">
        <Header theme={Theme.navy} />
        <main>{children}</main>
      </div>

      <div>
        <Image
          className="left-illustration"
          src={"/images/illustration/illustration-left.svg"}
          role="presentation"
          alt="illustration left"
          width={300}
          height={445}
        />
        <Image
          className="right-illustration"
          src={"/images/illustration/illustration-right.svg"}
          role="presentation"
          alt="illustration right"
          width={610}
          height={842}
        />
        <Image
          className={clsx("mobile-illustration", { hidden: pathname === "/faq" })}
          src={"/images/illustration/illustration-mobile.svg"}
          role="presentation"
          alt="illustration mobile"
          width={428}
          height={359}
        />
      </div>
    </div>
  );
}
