"use client";

import Bridge from "../bridge/Bridge";
import { useTokenStore } from "@/stores/tokenStoreProvider";
import { FormProvider, useForm } from "react-hook-form";
import { BridgeForm } from "@/models";
import { useChainStore } from "@/stores/chainStore";
import { TokenType } from "@/config";

export default function BridgeLayout() {
  const configContextValue = useTokenStore((state) => state.tokensList);
  const token = useChainStore((state) => state.token);

  const methods = useForm<BridgeForm>({
    defaultValues: {
      token: configContextValue?.UNKNOWN[0],
      claim: token?.type === TokenType.ETH ? "auto" : "manual",
      amount: "",
      minFees: 0n,
      gasFees: 0n,
      bridgingAllowed: false,
      balance: "0",
    },
  });

  return (
    <FormProvider {...methods}>
      <Bridge />
    </FormProvider>
  );
}
