import { useMemo } from "react";

import { encodeFunctionData, zeroAddress } from "viem";
import { useConnection } from "wagmi";

import { LINEA_ROLLUP_YIELD_EXTENSION_ABI } from "@/abis/LineaRollupYieldExtension";
import { MESSAGE_SERVICE_ABI } from "@/abis/MessageService";
import { MESSAGE_TRANSMITTER_V2_ABI } from "@/abis/MessageTransmitterV2";
import {
  BridgeTransactionType,
  CctpV2BridgeMessage,
  Chain,
  ChainLayer,
  NativeBridgeMessage,
  TransactionStatus,
} from "@/types";
import { buildClaimWithProofParams, isCctpV2BridgeMessage, isNativeBridgeMessage } from "@/utils/message";
import { isUndefined, isUndefinedOrEmptyString } from "@/utils/misc";

type UseClaimTxArgsProps = {
  status?: TransactionStatus;
  type?: BridgeTransactionType;
  fromChain?: Chain;
  toChain?: Chain;
  args?: NativeBridgeMessage | CctpV2BridgeMessage;
  lstSimulationPassed?: boolean;
};

const useClaimTxArgs = ({ status, type, fromChain, toChain, args, lstSimulationPassed }: UseClaimTxArgsProps) => {
  const { address } = useConnection();

  const yieldProviderAddress = toChain?.yieldProviderAddress;

  return useMemo(() => {
    if (
      isUndefinedOrEmptyString(address) ||
      isUndefined(status) ||
      isUndefined(type) ||
      isUndefined(fromChain) ||
      isUndefined(toChain) ||
      isUndefined(args)
    )
      return;

    if (status !== TransactionStatus.READY_TO_CLAIM) return;

    let toAddress: `0x${string}`;
    let encodedData: `0x${string}`;

    switch (type) {
      case BridgeTransactionType.ERC20:
      case BridgeTransactionType.ETH: {
        if (
          !isNativeBridgeMessage(args) ||
          isUndefinedOrEmptyString(args.from) ||
          isUndefinedOrEmptyString(args.to) ||
          isUndefined(args.fee) ||
          isUndefined(args.value) ||
          isUndefined(args.nonce) ||
          args.nonce === 0n ||
          isUndefinedOrEmptyString(args.calldata) ||
          isUndefinedOrEmptyString(args.messageHash) ||
          (isUndefined(args.proof) && toChain.layer === ChainLayer.L1)
        ) {
          return;
        }

        toAddress = toChain.messageServiceAddress;
        const claimWithProofParams = buildClaimWithProofParams(args);

        if (toChain.layer === ChainLayer.L1) {
          if (lstSimulationPassed && claimWithProofParams && !!yieldProviderAddress) {
            encodedData = encodeFunctionData({
              abi: LINEA_ROLLUP_YIELD_EXTENSION_ABI,
              functionName: "claimMessageWithProofAndWithdrawLST",
              args: [claimWithProofParams, yieldProviderAddress as `0x${string}`],
            });
          } else if (claimWithProofParams) {
            encodedData = encodeFunctionData({
              abi: MESSAGE_SERVICE_ABI,
              functionName: "claimMessageWithProof",
              args: [claimWithProofParams],
            });
          } else {
            return;
          }
        } else {
          encodedData = encodeFunctionData({
            abi: MESSAGE_SERVICE_ABI,
            functionName: "claimMessage",
            args: [args.from, args.to, args.fee, args.value, zeroAddress, args.calldata as `0x${string}`, args.nonce],
          });
        }

        break;
      }
      case BridgeTransactionType.USDC:
        if (
          !isCctpV2BridgeMessage(args) ||
          isUndefinedOrEmptyString(args.attestation) ||
          isUndefinedOrEmptyString(args.message)
        ) {
          return;
        }
        toAddress = toChain.cctpMessageTransmitterV2Address;
        encodedData = encodeFunctionData({
          abi: MESSAGE_TRANSMITTER_V2_ABI,
          functionName: "receiveMessage",
          args: [args.message as `0x${string}`, args.attestation as `0x${string}`],
        });
        break;
      default:
        return;
    }

    return {
      type: "claim",
      args: {
        to: toAddress,
        data: encodedData,
        value: 0n,
        chainId: toChain.id,
      },
    };
  }, [address, status, type, fromChain, toChain, args, lstSimulationPassed, yieldProviderAddress]);
};

export default useClaimTxArgs;
