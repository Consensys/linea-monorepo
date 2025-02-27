"use client";

import { useIsLoggedIn } from "@dynamic-labs/sdk-react-core";
import { useAccount } from "wagmi";
import Bridge from "../form";
import { useNativeBridgeNavigationStore } from "@/stores/nativeBridgeNavigationStore";
import TransactionHistory from "../transaction-history";
import { supportedChainIds } from "@/lib/wagmi";

export default function BridgeLayout() {
  const isTransactionHistoryOpen = useNativeBridgeNavigationStore.useIsTransactionHistoryOpen();
  const { chain } = useAccount();
  const isLoggedIn = useIsLoggedIn();

  if (isLoggedIn && (!chain?.id || !supportedChainIds.includes(chain.id))) {
    return null;
  }

  if (isTransactionHistoryOpen) {
    return <TransactionHistory />;
  }

  return <Bridge />;
}
