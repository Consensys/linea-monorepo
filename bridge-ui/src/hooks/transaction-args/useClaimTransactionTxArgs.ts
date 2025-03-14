import { useMemo } from "react";
import { useAccount } from "wagmi";
import { Address, encodeFunctionData, zeroAddress } from "viem";
import { Proof } from "@consensys/linea-sdk/dist/lib/sdk/merkleTree/types";
import MessageService from "@/abis/MessageService.json";
import { Chain, ChainLayer, TransactionStatus } from "@/types";

type UseClaimTxArgsProps = {
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

const useClaimTxArgs = ({ status, fromChain, toChain, args }: UseClaimTxArgsProps) => {
  const { address } = useAccount();

  return useMemo(() => {
    if (
      !address ||
      !status ||
      status !== TransactionStatus.READY_TO_CLAIM ||
      !fromChain ||
      !toChain ||
      !args.from ||
      !args.to ||
      args.fee === undefined ||
      args.value === undefined ||
      !args.nonce ||
      !args.calldata ||
      !args.messageHash ||
      (!args.proof && toChain.layer === ChainLayer.L1)
    ) {
      return;
    }

    const encodedData =
      toChain.layer === ChainLayer.L1
        ? encodeFunctionData({
            abi: MessageService.abi,
            functionName: "claimMessageWithProof",
            args: [
              {
                data: args.calldata,
                fee: args.fee,
                feeRecipient: zeroAddress,
                from: args.from,
                to: args.to,
                leafIndex: args.proof?.leafIndex,
                merkleRoot: args.proof?.root,
                messageNumber: args.nonce,
                proof: args.proof?.proof,
                value: args.value,
              },
            ],
          })
        : encodeFunctionData({
            abi: MessageService.abi,
            functionName: "claimMessage",
            args: [args.from, args.to, args.fee, args.value, zeroAddress, args.calldata, args.nonce],
          });

    return {
      type: "claim",
      args: {
        to: toChain.messageServiceAddress,
        data: encodedData,
        value: 0n,
        chainId: toChain.id,
      },
    };
  }, [
    address,
    args.calldata,
    args.fee,
    args.from,
    args.messageHash,
    args.nonce,
    args.proof,
    args.to,
    args.value,
    fromChain,
    status,
    toChain,
  ]);
};

export default useClaimTxArgs;
