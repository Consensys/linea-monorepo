"use client";

import { useEffect, useRef, useState } from "react";
import { useForm, FormProvider } from "react-hook-form";
import { useAccount, useWaitForTransactionReceipt } from "wagmi";
import { parseEther } from "viem";
import { toast } from "react-toastify";
import { BridgeForm, Transaction } from "@/models";
import FromChainToChain from "./FromChainToChain";
import Amount from "./Amount";
import Balance from "./Balance";
import { useSwitchNetwork } from "@/hooks";
import Submit from "./Submit";
import Fees from "./Fees";
import TokenModal from "./TokenModal";
import useBridge from "@/hooks/useBridge";
import Recipient from "./Recipient";
import { TokenType } from "@/config/config";
import { useChainStore } from "@/stores/chainStore";
import { useTokenStore } from "@/stores/tokenStore";
import useFetchHistory from "@/hooks/useFetchHistory";

export default function Bridge() {
  const configContextValue = useTokenStore((state) => state.tokensConfig);
  const [waitingTransaction, setWaitingTransaction] = useState<Transaction | undefined>();

  // Dialog Ref
  const tokensModalRef = useRef<HTMLDialogElement>(null);

  // Context
  const { fromChain, token } = useChainStore((state) => ({
    fromChain: state.fromChain,
    token: state.token,
  }));
  const { fetchHistory } = useFetchHistory();

  // Wagmi
  const { address } = useAccount();
  const {
    isLoading: isWaitingLoading,
    isSuccess: isWaitingSuccess,
    isError: isWaitingError,
  } = useWaitForTransactionReceipt({
    hash: waitingTransaction?.txHash,
    chainId: waitingTransaction?.chainId,
  });

  // Form
  const methods = useForm<BridgeForm>({
    defaultValues: {
      token: configContextValue?.UNKNOWN[0],
      claim: token?.type === TokenType.ETH ? "auto" : "manual",
      amount: "",
    },
  });

  // Hooks
  const { switchChain } = useSwitchNetwork(fromChain?.id);
  const { hash, isLoading, bridge, error: bridgeError } = useBridge();

  // Set tx hash to trigger useWaitForTransaction
  useEffect(() => {
    if (hash) {
      setWaitingTransaction({
        txHash: hash,
        chainId: fromChain?.id,
        name: fromChain?.name,
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [hash]);

  // Clear tx waiting when changing account
  useEffect(() => {
    setWaitingTransaction(undefined);
  }, [address]);

  useEffect(() => {
    if (isWaitingSuccess && waitingTransaction) {
      toast.success(`Transaction validated on ${waitingTransaction?.name}.`);
      setWaitingTransaction(undefined);
      fetchHistory();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isWaitingSuccess, waitingTransaction]);

  useEffect(() => {
    if (isWaitingError) {
      toast.error("Token bridging failed.");
      setWaitingTransaction(undefined);
    }
  }, [isWaitingError]);

  useEffect(() => {
    bridgeError &&
      bridgeError.displayInToast &&
      toast.error(
        <>
          {bridgeError.message}
          <a href={bridgeError.link} target="_blank" className="ml-1 underline">
            here
          </a>
        </>,
        {
          autoClose: false,
          draggable: false,
          closeOnClick: false,
          style: { width: 500, marginLeft: -90 },
        },
      );
  }, [bridgeError]);

  // Click on approve
  const onSubmit = async (data: BridgeForm) => {
    if (isLoading || isWaitingLoading) {
      return;
    }
    await switchChain();
    bridge(data.amount, parseEther(data.minFees), data.recipient);
  };

  return (
    <FormProvider {...methods}>
      <form onSubmit={methods.handleSubmit(onSubmit)}>
        <div className="card-body">
          <h2 className="card-title mb-5 font-medium text-white">Bridge</h2>
          <ul className="space-y-6">
            <li>
              <FromChainToChain />
            </li>
            <li>
              <Amount tokensModalRef={tokensModalRef} />
            </li>
            <li className="text-end text-sm">
              <Balance />
            </li>
            <li>
              <Recipient />
            </li>
            <li>
              <div className="divider"></div>
            </li>
            <li className="text-sm">
              <Fees />
            </li>
            <li>
              <Submit isLoading={isLoading ? true : false} isWaitingLoading={isWaitingLoading} />
            </li>
          </ul>
        </div>
      </form>
      <TokenModal tokensModalRef={tokensModalRef} />
    </FormProvider>
  );
}
