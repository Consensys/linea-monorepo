import { useCallback, useEffect } from "react";
import { useAccount, useBlockNumber, useEstimateFeesPerGas, useReadContract } from "wagmi";
import { Address, concat, encodeFunctionData, keccak256, parseUnits, toHex, zeroAddress } from "viem";
import { config, NetworkLayer, TokenType, wagmiConfig } from "@/config";
import { useChainStore } from "@/stores/chainStore";
import { getPublicClient } from "@wagmi/core";
import { useQueryClient } from "@tanstack/react-query";
import USDCBridge from "@/abis/USDCBridge.json";
import TokenBridge from "@/abis/TokenBridge.json";
import MessageService from "@/abis/MessageService.json";

type UsePostmanFeeProps = {
  currentLayer: NetworkLayer;
  claimingType?: string;
};

const usePostmanFee = ({ currentLayer, claimingType }: UsePostmanFeeProps) => {
  const { address } = useAccount();

  const token = useChainStore((state) => state.token);
  const toChain = useChainStore((state) => state.toChain);
  const fromChain = useChainStore((state) => state.fromChain);
  const fromMessageServiceAddress = useChainStore((state) => state.messageServiceAddress);
  const tokenBridgeAddress = useChainStore((state) => state.tokenBridgeAddress);
  const networkType = useChainStore((state) => state.networkType);

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
    address: fromMessageServiceAddress ?? "0x",
    abi: MessageService.abi,
    functionName: "nextMessageNumber",
    chainId: fromChain?.id,
    query: {
      enabled:
        !!fromChain?.id && !!fromMessageServiceAddress && currentLayer === NetworkLayer.L1 && claimingType === "auto",
    },
  });

  const calculatePostmanFee = useCallback(
    async (amount: string, recipient?: Address): Promise<bigint> => {
      if (!address || !tokenBridgeAddress || !token || !nextMessageNumber) {
        return 0n;
      }

      if (currentLayer !== NetworkLayer.L1) {
        return 0n;
      }

      try {
        if (!feeData?.gasPrice) {
          return 0n;
        }

        const publicClient = getPublicClient(wagmiConfig, {
          chainId: toChain?.id,
        });

        if (!publicClient) {
          return 0n;
        }

        const toMessageServiceAddress = config.networks[networkType][NetworkLayer.L2].messageServiceAddress;
        const toUSDCBridgeAddress = config.networks[networkType][NetworkLayer.L2].usdcBridgeAddress;
        const toTokenBridgeAddress = config.networks[networkType][NetworkLayer.L2].tokenBridgeAddress;

        const amountBigInt = parseUnits(amount, token.decimals);
        const toAddress = recipient || address;
        let estimatedGasFee;

        // If amount negative, return
        if (amountBigInt <= BigInt(0)) {
          return 0n;
        }

        if (token.type === TokenType.USDC) {
          const encodedData = encodeFunctionData({
            abi: USDCBridge.abi,
            functionName: "receiveFromOtherLayer",
            args: [toAddress, amountBigInt],
          });

          const storageSlot = keccak256(
            concat([
              address,
              toUSDCBridgeAddress,
              toHex(0n),
              toHex(0n),
              toHex(nextMessageNumber as bigint),
              encodedData,
            ]),
          );

          estimatedGasFee = await publicClient.estimateContractGas({
            abi: MessageService.abi,
            functionName: "claimMessage",
            address: toMessageServiceAddress,
            args: [address, toUSDCBridgeAddress, 0n, 0n, zeroAddress, encodedData, nextMessageNumber as bigint],
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
        } else if (token.type === TokenType.ERC20) {
          // TODO: fix encoded data
          const encodedData = encodeFunctionData({
            abi: TokenBridge.abi,
            functionName: "completeBridging",
            args: [toAddress, amountBigInt],
          });

          const storageSlot = keccak256(
            concat([
              address,
              toTokenBridgeAddress,
              toHex(0n),
              toHex(0n),
              toHex(nextMessageNumber as bigint),
              encodedData,
            ]),
          );

          estimatedGasFee = await publicClient.estimateContractGas({
            abi: MessageService.abi,
            functionName: "claimMessage",
            address: toMessageServiceAddress,
            args: [address, toUSDCBridgeAddress, 0n, 0n, zeroAddress, encodedData, nextMessageNumber as bigint],
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
          const storageSlot = keccak256(
            concat([address, toAddress, toHex(0n), toHex(amountBigInt), toHex(nextMessageNumber as bigint), "0x"]),
          );

          estimatedGasFee = await publicClient.estimateContractGas({
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

        return (
          feeData.gasPrice *
          (estimatedGasFee + config.networks[networkType].gasLimitSurplus) *
          config.networks[networkType].profitMargin
        );
      } catch (error) {
        console.error(error);
        return 0n;
      }
    },
    [address, currentLayer, nextMessageNumber, feeData?.gasPrice, networkType, toChain?.id, token, tokenBridgeAddress],
  );

  return { calculatePostmanFee };
};

export default usePostmanFee;
