import { useMemo } from "react";
import { useAccount } from "wagmi";
import { encodeFunctionData, padHex, zeroHash } from "viem";
import { useFormStore, useChainStore } from "@/stores";
import { isCctp } from "@/utils/tokens";
import { useCctpFee, useCctpDestinationDomain } from "./useCctpUtilHooks";
import { CCTP_MIN_FINALITY_THRESHOLD } from "@/constants";
import { isNull, isUndefined } from "@/utils";

type UseDepositForBurnTxArgs = {
  allowance?: bigint;
};

const useDepositForBurnTxArgs = ({ allowance }: UseDepositForBurnTxArgs) => {
  const { address } = useAccount();
  const { fromChainId, fromChainLayer, cctpTokenMessengerV2Address } = useChainStore((state) => ({
    fromChainId: state.fromChain.id,
    fromChainLayer: state.fromChain.layer,
    cctpTokenMessengerV2Address: state.fromChain.cctpTokenMessengerV2Address,
  }));
  const cctpDestinationDomain = useCctpDestinationDomain();
  const token = useFormStore((state) => state.token);
  const amount = useFormStore((state) => state.amount);
  const recipient = useFormStore((state) => state.recipient);
  const fee = useCctpFee();

  return useMemo(() => {
    if (!address || isNull(amount) || isUndefined(allowance) || allowance < amount || !recipient || !isCctp(token)) {
      return;
    }

    return {
      type: "depositForBurn",
      args: {
        to: cctpTokenMessengerV2Address,
        data: encodeFunctionData({
          abi: [
            {
              type: "function",
              name: "depositForBurn",
              stateMutability: "nonpayable",
              inputs: [
                { name: "amount", type: "uint256" },
                { name: "destinationDomain", type: "uint32" },
                { name: "mintRecipient", type: "bytes32" },
                { name: "burnToken", type: "address" },
                { name: "destinationCaller", type: "bytes32" },
                { name: "maxFee", type: "uint256" },
                { name: "minFinalityThreshold", type: "uint32" },
              ],
              outputs: [],
            },
          ],
          functionName: "depositForBurn",
          args: [
            amount,
            cctpDestinationDomain,
            padHex(recipient),
            token[fromChainLayer],
            zeroHash,
            fee,
            CCTP_MIN_FINALITY_THRESHOLD,
          ],
        }),
        value: 0n,
        chainId: fromChainId,
      },
    };
  }, [
    address,
    amount,
    allowance,
    recipient,
    token,
    cctpTokenMessengerV2Address,
    cctpDestinationDomain,
    fromChainLayer,
    fee,
    fromChainId,
  ]);
};

export default useDepositForBurnTxArgs;
