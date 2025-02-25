"use client";

import Bridge from "../form";
import { useNativeBridgeNavigationStore } from "@/stores/nativeBridgeNavigationStore";
import TransactionHistory from "../transaction-history";

export default function BridgeLayout() {
  const isTransactionHistoryOpen = useNativeBridgeNavigationStore.useIsTransactionHistoryOpen();

  if (isTransactionHistoryOpen) {
    return <TransactionHistory />;
  }

  return <Bridge />;
}
