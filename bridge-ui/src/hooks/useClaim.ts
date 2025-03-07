import { useSendTransaction, useWaitForTransactionReceipt } from "wagmi";
import { Address } from "viem";
import { Proof } from "@consensys/linea-sdk/dist/lib/sdk/merkleTree/types";
import useClaimTxArgs from "./transaction-args/useClaimTransactionTxArgs";
import { Chain } from "@/types";
import { TransactionStatus } from "@/types/transaction";

type UseClaimProps = {
  status?: TransactionStatus;
  fromChain?: Chain;
  toChain?: Chain;
  args: {
    from?: Address;
    to?: Address;
    fee?: bigint;
    value?: bigint;
    nonce?: bigint;
    calldata?: string;
    messageHash?: string;
    proof?: Proof;
  };
};

const useClaim = (props: UseClaimProps) => {
  const transactionArgs = useClaimTxArgs(props);

  const { data: hash, sendTransaction, isPending, error, isSuccess } = useSendTransaction();

  const { isLoading: isConfirming, isSuccess: isConfirmed } = useWaitForTransactionReceipt({
    hash,
  });

  return {
    transactionType: "claim",
    claim: transactionArgs ? () => sendTransaction(transactionArgs.args) : undefined,
    isPending: isPending,
    isConfirming,
    isConfirmed,
    error,
    isSuccess,
  };
};

export default useClaim;
