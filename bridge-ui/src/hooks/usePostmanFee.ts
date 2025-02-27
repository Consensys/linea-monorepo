import { useCallback, useEffect } from "react";
import { useAccount, useBlockNumber, useEstimateFeesPerGas, useReadContract } from "wagmi";
import { Address, encodeAbiParameters, encodeFunctionData, keccak256, parseUnits, zeroAddress } from "viem";
import { TokenType, wagmiConfig } from "@/config";
import { useChainStore } from "@/stores/chainStore";
import { getPublicClient } from "@wagmi/core";
import { useQueryClient } from "@tanstack/react-query";
import TokenBridge from "@/abis/TokenBridge.json";
import MessageService from "@/abis/MessageService.json";
import { ChainLayer } from "@/types";
import { useSelectedToken } from "./useSelectedToken";

function computeMessageHash(
  from: Address,
  to: Address,
  fee: bigint,
  value: bigint,
  nonce: bigint,
  calldata: `0x${string}` = "0x",
) {
  return keccak256(
    encodeAbiParameters(
      [
        { name: "from", type: "address" },
        { name: "to", type: "address" },
        { name: "fee", type: "uint256" },
        { name: "value", type: "uint256" },
        { name: "nonce", type: "uint256" },
        { name: "calldata", type: "bytes" },
      ],
      [from, to, fee, value, nonce, calldata],
    ),
  );
}

function computeMessageStorageSlot(messageHash: `0x${string}`) {
  const mappingSlot = 176n;

  return keccak256(
    encodeAbiParameters(
      [
        { name: "messageHash", type: "bytes32" },
        { name: "mappingSlot", type: "uint256" },
      ],
      [messageHash, mappingSlot],
    ),
  );
}

type UsePostmanFeeProps = {
  claimingType?: string;
};

const usePostmanFee = ({ claimingType }: UsePostmanFeeProps) => {
  const { address } = useAccount();

  const token = useSelectedToken();
  const toChain = useChainStore.useToChain();
  const fromChain = useChainStore.useFromChain();

  const queryClient = useQueryClient();
  const { data: blockNumber } = useBlockNumber({ watch: true });
  const { data: feeData, queryKey } = useEstimateFeesPerGas({ chainId: toChain?.id, type: "legacy" });

  useEffect(() => {
    if (blockNumber && blockNumber % 5n === 0n) {
      queryClient.invalidateQueries({ queryKey });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [blockNumber, queryClient]);

  const { data: nextMessageNumber } = useReadContract({
    address: fromChain.messageServiceAddress,
    abi: MessageService.abi,
    functionName: "nextMessageNumber",
    chainId: fromChain?.id,
    query: {
      enabled:
        !!fromChain?.id &&
        !!fromChain.messageServiceAddress &&
        fromChain.layer === ChainLayer.L1 &&
        claimingType === "auto",
    },
  });

  const calculatePostmanFee = useCallback(
    async (amount: string, recipient?: Address): Promise<bigint> => {
      if (!address || !fromChain.tokenBridgeAddress || !token || !nextMessageNumber) {
        return 0n;
      }

      if (fromChain.layer !== ChainLayer.L1) {
        return 0n;
      }

      try {
        if (!feeData?.gasPrice) {
          return 0n;
        }

        const destinationChainPublicClient = getPublicClient(wagmiConfig, {
          chainId: toChain?.id,
        });

        if (!destinationChainPublicClient) {
          return 0n;
        }

        const toMessageServiceAddress = toChain.messageServiceAddress;
        const toTokenBridgeAddress = toChain.tokenBridgeAddress;

        const amountBigInt = parseUnits(amount, token.decimals);
        const toAddress = recipient || address;
        let estimatedGasFee;

        // If amount negative, return
        if (amountBigInt <= BigInt(0)) {
          return 0n;
        }

        if (token.type === TokenType.ERC20) {
          const originChainPublicClient = getPublicClient(wagmiConfig, {
            chainId: fromChain?.id,
          });

          if (!originChainPublicClient) {
            return 0n;
          }

          const nativeToken = (await originChainPublicClient.readContract({
            account: address,
            address: fromChain.tokenBridgeAddress,
            abi: TokenBridge.abi,
            functionName: "bridgedToNativeToken",
            args: [token[fromChain.layer]],
          })) as Address;

          let tokenAddress = token[fromChain.layer];
          let chainId = fromChain?.id;
          let tokenMetadata = encodeAbiParameters(
            [
              { name: "tokenName", type: "string" },
              { name: "tokenSymbol", type: "string" },
              { name: "tokenDecimals", type: "uint8" },
            ],
            [token.name, token.symbol, token.decimals],
          );

          if (nativeToken !== zeroAddress) {
            tokenAddress = nativeToken;
            chainId = toChain?.id;
            tokenMetadata = "0x";
          }

          const encodedData = encodeFunctionData({
            abi: TokenBridge.abi,
            functionName: "completeBridging",
            args: [tokenAddress, amountBigInt, toAddress, chainId, tokenMetadata],
          });

          const messageHash = computeMessageHash(
            fromChain.tokenBridgeAddress,
            toTokenBridgeAddress,
            0n,
            0n,
            nextMessageNumber as bigint,
            encodedData,
          );

          const storageSlot = computeMessageStorageSlot(messageHash);

          estimatedGasFee = await destinationChainPublicClient.estimateContractGas({
            abi: MessageService.abi,
            functionName: "claimMessage",
            address: toMessageServiceAddress,
            args: [
              fromChain.tokenBridgeAddress,
              toTokenBridgeAddress,
              0n,
              0n,
              zeroAddress,
              encodedData,
              nextMessageNumber as bigint,
            ],
            value: 0n,
            account: address,
            stateOverride: [
              {
                address: toMessageServiceAddress,
                stateDiff: [
                  {
                    slot: storageSlot,
                    value: "0x0000000000000000000000000000000000000000000000000000000000000001",
                  },
                ],
              },
            ],
          });
        } else if (token.type === TokenType.ETH) {
          const messageHash = computeMessageHash(
            address,
            toAddress,
            0n,
            amountBigInt,
            nextMessageNumber as bigint,
            "0x",
          );

          const storageSlot = computeMessageStorageSlot(messageHash);

          estimatedGasFee = await destinationChainPublicClient.estimateContractGas({
            abi: MessageService.abi,
            functionName: "claimMessage",
            address: toMessageServiceAddress,
            args: [address, toAddress, 0n, amountBigInt, zeroAddress, "0x", nextMessageNumber as bigint],
            value: 0n,
            account: address,
            stateOverride: [
              {
                address: toMessageServiceAddress,
                stateDiff: [
                  {
                    slot: storageSlot,
                    value: "0x0000000000000000000000000000000000000000000000000000000000000001",
                  },
                ],
              },
            ],
          });
        } else {
          return 0n;
        }

        return feeData.gasPrice * (estimatedGasFee + fromChain.gasLimitSurplus) * fromChain.profitMargin;
      } catch (error) {
        console.error(error);
        return 0n;
      }
    },
    [address, fromChain.layer, nextMessageNumber, feeData?.gasPrice, toChain?.id, token, fromChain.tokenBridgeAddress],
  );

  return { calculatePostmanFee };
};

export default usePostmanFee;
