"use client";

import { useIsLoggedIn } from "@dynamic-labs/sdk-react-core";
import { useAccount } from "wagmi";
import { FormProvider, useForm } from "react-hook-form";
import Bridge from "../form";
import { useNativeBridgeNavigationStore } from "@/stores/nativeBridgeNavigationStore";
import TransactionHistory from "../transaction-history";
import { supportedChainIds } from "@/lib/wagmi";
import { useTokens } from "@/hooks/useTokens";
import { useChainStore } from "@/stores/chainStore";
import { BridgeType } from "@/config/config";
import { BridgeForm } from "@/models";
import { ChainLayer } from "@/types";

export default function BridgeLayout() {
  const isTransactionHistoryOpen = useNativeBridgeNavigationStore.useIsTransactionHistoryOpen();
  const { chain, address } = useAccount();
  const isLoggedIn = useIsLoggedIn();
  const tokens = useTokens();
  const fromChain = useChainStore.useFromChain();

  const methods = useForm<BridgeForm>({
    disabled: !isLoggedIn || isTransactionHistoryOpen,
    defaultValues: {
      token: tokens[0],
      claim: fromChain?.layer === ChainLayer.L1 ? "auto" : "manual",
      amount: "",
      minFees: 0n,
      gasFees: 0n,
      bridgingAllowed: false,
      balance: "0",
      mode: BridgeType.NATIVE,
      destinationAddress: address,
    },
  });

  if (isLoggedIn && (!chain?.id || !supportedChainIds.includes(chain.id))) {
    return null;
  }

  if (isTransactionHistoryOpen) {
    return <TransactionHistory />;
  }

  return (
    <FormProvider {...methods}>
      <Bridge />
    </FormProvider>
  );
}
