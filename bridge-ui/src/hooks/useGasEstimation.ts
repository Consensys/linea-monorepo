import { useCallback, useEffect } from "react";
import { getPublicClient } from "@wagmi/core";
import { useQueryClient } from "@tanstack/react-query";
import { useAccount, useBlockNumber, useEstimateFeesPerGas } from "wagmi";
import { Address, parseUnits } from "viem";
import { TokenType, wagmiConfig } from "@/config";
import { useChainStore } from "@/stores/chainStore";
import USDCBridge from "@/abis/USDCBridge.json";
import TokenBridge from "@/abis/TokenBridge.json";
import MessageService from "@/abis/MessageService.json";

const useGasEstimation = () => {
  // Context
  const { token, tokenBridgeAddress, messageServiceAddress, networkLayer, fromChain } = useChainStore((state) => ({
    token: state.token,
    tokenBridgeAddress: state.tokenBridgeAddress,
    messageServiceAddress: state.messageServiceAddress,
    networkLayer: state.networkLayer,
    fromChain: state.fromChain,
  }));

  const { address } = useAccount();

  const queryClient = useQueryClient();
  const { data: blockNumber } = useBlockNumber({ watch: true });
  const { data: feeData, queryKey } = useEstimateFeesPerGas({ chainId: fromChain?.id, type: "legacy" });

  useEffect(() => {
    if (blockNumber && blockNumber % 5n === 0n) {
      queryClient.invalidateQueries({ queryKey });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [blockNumber, queryClient]);

  const estimateGasBridge = useCallback(
    async (_amount: string, _userMinimumFee?: bigint, to?: Address) => {
      if (!address || !tokenBridgeAddress || !token || !messageServiceAddress) {
        return;
      }
      try {
        if (!feeData?.gasPrice) {
          return;
        }

        const publicClient = getPublicClient(wagmiConfig, {
          chainId: fromChain?.id,
        });

        if (!publicClient) {
          return;
        }

        const amountBigInt = parseUnits(_amount, token.decimals);
        const sendTo = to ? to : address;
        let estimatedGasFee;

        // If amount negative, return
        if (amountBigInt <= BigInt(0)) {
          return;
        }

        switch (token.type) {
          case TokenType.USDC:
            estimatedGasFee = await publicClient.estimateContractGas({
              abi: USDCBridge.abi,
              functionName: "depositTo",
              address: tokenBridgeAddress,
              args: [amountBigInt, sendTo],
              value: _userMinimumFee,
              account: address,
            });
            break;
          case TokenType.ERC20:
            estimatedGasFee = await publicClient.estimateContractGas({
              abi: TokenBridge.abi,
              functionName: "bridgeToken",
              address: tokenBridgeAddress,
              args: [token[networkLayer], amountBigInt, sendTo],
              value: _userMinimumFee,
              account: address,
            });
            break;
          case TokenType.ETH:
            estimatedGasFee = await publicClient.estimateContractGas({
              abi: MessageService.abi,
              functionName: "sendMessage",
              address: messageServiceAddress,
              args: [sendTo, _userMinimumFee, "0x"],
              value: _userMinimumFee ? _userMinimumFee + amountBigInt : amountBigInt,
              account: address,
            });
            break;
          default:
            return;
        }

        return estimatedGasFee * feeData.gasPrice;
      } catch (error) {
        // log.error(error);
        return;
      }
    },
    [address, token, tokenBridgeAddress, feeData, messageServiceAddress, networkLayer, fromChain],
  );

  return { estimateGasBridge };
};

export default useGasEstimation;
