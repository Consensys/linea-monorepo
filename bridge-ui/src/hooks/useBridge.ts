import { useState, useCallback, useEffect } from "react";
import { writeContract, simulateContract } from "@wagmi/core";
import { useAccount, useWaitForTransactionReceipt } from "wagmi";
import { Address, parseUnits } from "viem";
import log from "loglevel";
import USDCBridge from "@/abis/USDCBridge.json";
import TokenBridge from "@/abis/TokenBridge.json";
import MessageService from "@/abis/MessageService.json";
import { TokenType } from "@/config/config";
import { BridgeError, BridgeErrors, Transaction } from "@/models";
import { FieldErrors, FieldValues } from "react-hook-form";
import { wagmiConfig } from "@/config";
import { useChainStore } from "@/stores/chainStore";
import useMinimumFee from "./useMinimumFee";
import { isEmptyObject } from "@/utils/utils";

type UseBridge = {
  hash: Address | undefined;
  isLoading: boolean;
  bridge: (amount: string, userMinimumFee?: bigint, to?: Address) => Promise<void>;
  isError: boolean;
  error: BridgeError | null;
  bridgeEnabled: (amount: string, allowance: bigint, errors: FieldErrors<FieldValues>) => boolean;
};

const useBridge = (): UseBridge => {
  const [transaction, setTransaction] = useState<Transaction | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<BridgeError | null>(null);

  // Context
  const { token, tokenBridgeAddress, messageServiceAddress, networkLayer, fromChain } = useChainStore((state) => ({
    token: state.token,
    tokenBridgeAddress: state.tokenBridgeAddress,
    messageServiceAddress: state.messageServiceAddress,
    networkLayer: state.networkLayer,
    fromChain: state.fromChain,
  }));

  const { minimumFee } = useMinimumFee();
  const { address, isConnected } = useAccount();

  const {
    isLoading: isTxLoading,
    isSuccess: isTxSuccess,
    isError: isTxError,
  } = useWaitForTransactionReceipt({
    hash: transaction?.txHash,
    chainId: transaction?.chainId,
  });

  useEffect(() => {
    if (isTxSuccess || isTxError) {
      setTransaction(null);
    }
  }, [isTxSuccess, isTxError]);

  const handleError = (_error: Error) => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const cause = (_error.cause as any).data?.abiItem?.name;
    switch (cause) {
      case BridgeErrors.ReservedToken:
        setError({
          name: cause,
          message:
            "The token you have selected is not currently available. Please select another token. For more details see",
          link: "https://docs.linea.build/use-mainnet/bridges-of-linea",
          displayInToast: true,
        });
        break;
      case BridgeErrors.RateLimitExceeded:
        setError({
          name: cause,
          message:
            "The message service has reached it's daily withdrawal limit. Please try again in 24 hours. For more details about this security measure see",
          link: "https://docs.linea.build/use-mainnet/bridges-of-linea",
          displayInToast: true,
        });
        break;
      default:
        setError({
          name: cause,
          message: "",
          link: "",
          displayInToast: false,
        });
        break;
    }
  };

  const getWriteConfig = useCallback(
    async (_amount: string, _address: Address, _minimumFee: bigint) => {
      if (tokenBridgeAddress && messageServiceAddress && token) {
        const amountBigInt = parseUnits(_amount, token.decimals);

        switch (token.type) {
          case TokenType.USDC:
            return await simulateContract(wagmiConfig, {
              address: tokenBridgeAddress,
              abi: USDCBridge.abi,
              functionName: "depositTo",
              args: [amountBigInt, _address],
              value: _minimumFee,
            });
          case TokenType.ERC20:
            return await simulateContract(wagmiConfig, {
              address: tokenBridgeAddress,
              abi: TokenBridge.abi,
              functionName: "bridgeToken",
              args: [token[networkLayer], amountBigInt, _address],
              value: _minimumFee,
            });
          case TokenType.ETH:
            return await simulateContract(wagmiConfig, {
              address: messageServiceAddress,
              abi: MessageService.abi,
              functionName: "sendMessage",
              args: [_address, _minimumFee, "0x"],
              value: _minimumFee + amountBigInt,
            });
          default:
            throw new Error("TokenType not found");
        }
      }

      return;
    },
    [messageServiceAddress, tokenBridgeAddress, token, networkLayer],
  );

  const bridge = useCallback(
    async (amount: string, userMinimumFee?: bigint, to?: Address) => {
      setError(null);
      setIsLoading(true);

      if (!amount || !tokenBridgeAddress || !token) {
        setIsLoading(false);
        return;
      }

      const sendTo = to ? to : address;
      if (!sendTo) {
        setIsLoading(false);
        return;
      }

      try {
        const config = await getWriteConfig(amount, sendTo, userMinimumFee ?? minimumFee);
        if (config) {
          const hash = await writeContract(wagmiConfig, config.request);
          setTransaction({
            txHash: hash,
            chainId: fromChain?.id,
            name: fromChain?.name,
          });
        }
      } catch (error) {
        log.error(error);
        handleError(error as Error);
      }

      setIsLoading(false);
    },
    [minimumFee, token, tokenBridgeAddress, address, getWriteConfig, fromChain],
  );

  const bridgeEnabled = useCallback(
    (amount: string, allowance: bigint, errors: FieldErrors<FieldValues>) => {
      if (!token || !amount || isLoading || isTxLoading || !isConnected) {
        return false;
      }

      // Check form errors
      if (!isEmptyObject(errors)) {
        return false;
      }

      // Check allowance
      const amountBigInt = parseUnits(amount, token.decimals);
      if (token.type !== TokenType.ETH && (!allowance || amountBigInt > allowance)) {
        return false;
      }

      return true;
    },
    [isConnected, isLoading, isTxLoading, token],
  );

  return {
    hash: transaction?.txHash,
    isLoading,
    bridge,
    isError: error !== null,
    error,
    bridgeEnabled,
  };
};

export default useBridge;
