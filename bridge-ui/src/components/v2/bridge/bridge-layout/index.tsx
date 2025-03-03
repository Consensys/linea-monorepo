"use client";

import { useIsLoggedIn } from "@dynamic-labs/sdk-react-core";
import { useAccount } from "wagmi";
import Bridge from "../form";
import { useNativeBridgeNavigationStore } from "@/stores/nativeBridgeNavigationStore";
import TransactionHistory from "../transaction-history";
import { supportedChainIds } from "@/lib/wagmi";
import { useTokens } from "@/hooks/useTokens";
import { useChainStore } from "@/stores/chainStore";
import { BridgeType } from "@/config/config";
import { ChainLayer } from "@/types";
import { FormStoreProvider } from "@/stores/formStoreProvider";
import { FormState } from "@/stores/formStore";

export default function BridgeLayout() {
  const isTransactionHistoryOpen = useNativeBridgeNavigationStore.useIsTransactionHistoryOpen();
  const { chain, address } = useAccount();
  const isLoggedIn = useIsLoggedIn();
  const tokens = useTokens();
  const fromChain = useChainStore.useFromChain();

  if (isLoggedIn && (!chain?.id || !supportedChainIds.includes(chain.id))) {
    return null;
  }

  if (isTransactionHistoryOpen) {
    return <TransactionHistory />;
  }

  const initialFormState: FormState = {
    token: tokens[0],
    claim: fromChain?.layer === ChainLayer.L1 ? "auto" : "manual",
    amount: null,
    minimumFees: 0n,
    gasFees: 0n,
    bridgingFees: 0n,
    balance: 0n,
    mode: BridgeType.NATIVE,
    recipient: address || "0x",
  };

  return (
    <FormStoreProvider initialState={initialFormState}>
      <Bridge />
    </FormStoreProvider>
  );
}
