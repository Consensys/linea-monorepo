'use client'
import { WalletHooksProvider, useWallet } from "@layerswap/widget";
import useCustomEVM from "./useCustomeEvm";
import { use } from "react";


export default function ({ children }: { children: JSX.Element | JSX.Element[] }) {
    const customEvm = useCustomEVM()
    return <WalletHooksProvider overides={{ evm: customEvm }}>
        {children}
    </WalletHooksProvider>
}