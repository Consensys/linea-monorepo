import { useState, useCallback, useContext } from 'react';
import { prepareWriteContract, writeContract, getPublicClient, readContract } from '@wagmi/core';
import { useAccount, useFeeData, useWaitForTransaction } from 'wagmi';
import { Address, Chain, parseUnits, zeroAddress } from 'viem';
import log from 'loglevel';

import USDCBridge from '@/abis/USDCBridge.json';
import TokenBridge from '@/abis/TokenBridge.json';
import MessageService from '@/abis/MessageService.json';
import { ChainContext } from '@/contexts/chain.context';
import useMessageService from './useMessageService';
import { TokenInfo, TokenType, config } from '@/config/config';
import { BridgeError, BridgeErrors, Transaction } from '@/models';
import { getChainNetworkLayer } from '@/utils/chainsUtil';
import { FieldErrors, FieldValues } from 'react-hook-form';

type UseBridge = {
  hash: Address | undefined;
  isLoading: boolean;
  bridge: (amount: string, userMinimumFee?: bigint, to?: Address) => Promise<void>;
  estimateGasBridge: (amount: string, userMinimumFee?: bigint, to?: Address) => Promise<bigint | undefined>;
  isError: boolean;
  error: BridgeError | null;
  bridgeEnabled: (amount: string, allowance: bigint, errors: FieldErrors<FieldValues>) => boolean;
  fetchBridgedToken: (fromChain: Chain, toChain: Chain, nativeToken: Address) => Promise<Address | undefined>;
  fetchNativeToken: (chain: Chain, bridgedToken: Address) => Promise<Address | undefined>;
  fillMissingTokenAddress: (token: TokenInfo) => Promise<void>;
};

const useBridge = (): UseBridge => {
  const [transaction, setTransaction] = useState<Transaction | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<BridgeError | null>(null);

  // Context
  const context = useContext(ChainContext);
  const {
    token,
    tokenBridgeAddress,
    messageServiceAddress,
    networkLayer,
    networkType,
    fromChain,
    activeChain,
    alternativeChain,
  } = context;

  const { minimumFee } = useMessageService();
  const { address, isConnected } = useAccount();
  const { data: feeData } = useFeeData({
    chainId: fromChain?.id,
    watch: true,
  });
  const { isLoading: isTxLoading } = useWaitForTransaction({
    hash: transaction?.txHash,
    chainId: transaction?.chainId,
    onSuccess() {
      setTransaction(null);
    },
    onError() {
      setTransaction(null);
    },
  });

  const handleError = (_error: Error) => {
    const cause = (_error.cause as any).data?.abiItem?.name;
    switch (cause) {
      case BridgeErrors.ReservedToken:
        setError({
          name: cause,
          message:
            'The token you have selected is not currently available. Please select another token. For more details see',
          link: 'https://docs.linea.build/use-mainnet/bridges-of-linea',
          displayInToast: true,
        });
        break;
      case BridgeErrors.RateLimitExceeded:
        setError({
          name: cause,
          message:
            "The message service has reached it's daily withdrawal limit. Please try again in 24 hours. For more details about this security measure see",
          link: 'https://docs.linea.build/use-mainnet/bridges-of-linea',
          displayInToast: true,
        });
        break;
      default:
        setError({
          name: cause,
          message: '',
          link: '',
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
            return await prepareWriteContract({
              address: tokenBridgeAddress,
              abi: USDCBridge.abi,
              functionName: 'depositTo',
              args: [amountBigInt, _address],
              value: _minimumFee,
            });
          case TokenType.ERC20:
            return await prepareWriteContract({
              address: tokenBridgeAddress,
              abi: TokenBridge.abi,
              functionName: 'bridgeToken',
              args: [token[networkLayer], amountBigInt, _address],
              value: _minimumFee,
            });
          case TokenType.ETH:
            const valueToSend = _minimumFee + amountBigInt;
            return await prepareWriteContract({
              address: messageServiceAddress,
              abi: MessageService.abi,
              functionName: 'sendMessage',
              args: [_address, _minimumFee, '0x'],
              value: valueToSend,
            });
          default:
            throw new Error('TokenType not found');
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
          const { hash } = await writeContract(config);
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

  const estimateGasBridge = useCallback(
    async (_amount: string, _userMinimumFee?: bigint, to?: Address) => {
      if (!address || !tokenBridgeAddress || !token || !messageServiceAddress) {
        return;
      }
      try {
        if (!feeData?.gasPrice) {
          return;
        }

        const publicClient = getPublicClient({
          chainId: fromChain?.id,
        });

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
              functionName: 'depositTo',
              address: tokenBridgeAddress,
              args: [amountBigInt, sendTo],
              value: _userMinimumFee,
              account: address,
            });
            break;
          case TokenType.ERC20:
            estimatedGasFee = await publicClient.estimateContractGas({
              abi: TokenBridge.abi,
              functionName: 'bridgeToken',
              address: tokenBridgeAddress,
              args: [token[networkLayer], amountBigInt, sendTo],
              value: _userMinimumFee,
              account: address,
            });
            break;
          case TokenType.ETH:
            const valueToSend = _userMinimumFee ? _userMinimumFee + amountBigInt : amountBigInt;
            estimatedGasFee = await publicClient.estimateContractGas({
              abi: MessageService.abi,
              functionName: 'sendMessage',
              address: messageServiceAddress,
              args: [sendTo, _userMinimumFee, '0x'],
              value: valueToSend,
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

  const fetchBridgedToken = useCallback(
    async (fromChain: Chain, toChain: Chain, nativeToken: Address) => {
      const fromLayer = getChainNetworkLayer(fromChain);
      const toLayer = getChainNetworkLayer(toChain);
      if (!toLayer || !fromLayer) {
        return;
      }

      const _tokenBridgeAddress = config.networks[networkType][toLayer].tokenBridgeAddress;

      if (!_tokenBridgeAddress) {
        return;
      }

      try {
        const bridgedToken = (await readContract({
          address: _tokenBridgeAddress,
          abi: TokenBridge.abi,
          functionName: 'nativeToBridgedToken',
          args: [fromChain.id, nativeToken],
          chainId: toChain.id,
        })) as Address;

        return bridgedToken;
      } catch (error) {
        log.warn('Error fetching bridged token address');
      }
    },
    [networkType],
  );

  const fetchNativeToken = useCallback(
    async (chain: Chain, bridgedToken: Address) => {
      const layer = getChainNetworkLayer(chain);
      if (!layer) {
        return;
      }

      const _tokenBridgeAddress = config.networks[networkType][layer].tokenBridgeAddress;

      if (!_tokenBridgeAddress) {
        return;
      }

      try {
        const nativeToken = (await readContract({
          address: _tokenBridgeAddress,
          abi: TokenBridge.abi,
          functionName: 'bridgedToNativeToken',
          args: [bridgedToken],
          chainId: chain.id,
        })) as Address;

        return nativeToken;
      } catch (error) {
        log.warn('Error fetching native token address');
      }
    },
    [networkType],
  );

  const fillMissingTokenAddress = useCallback(
    async (token: TokenInfo) => {
      if (!activeChain || !alternativeChain) {
        return;
      }

      // Since we don't if a token is native or bridged for a chain we try all the combinations
      // possible to find its counterpart on the other chain
      if (!token.L1 && token.L2) {
        token.L1 = (await fetchNativeToken(activeChain, token.L2)) || null;
        if (!token.L1 || token.L1 !== zeroAddress) return;

        token.L1 = (await fetchNativeToken(alternativeChain, token.L2)) || null;
        if (!token.L1 || token.L1 !== zeroAddress) return;

        token.L1 = (await fetchBridgedToken(activeChain, alternativeChain, token.L2)) || null;
        if (!token.L1 || token.L1 !== zeroAddress) return;

        token.L1 = (await fetchBridgedToken(alternativeChain, activeChain, token.L2)) || null;
      } else if (token.L1) {
        token.L2 = (await fetchNativeToken(activeChain, token.L1)) || null;
        if (!token.L2 || token.L2 !== zeroAddress) return;

        token.L2 = (await fetchNativeToken(alternativeChain, token.L1)) || null;
        if (!token.L2 || token.L2 !== zeroAddress) return;

        token.L2 = (await fetchBridgedToken(activeChain, alternativeChain, token.L1)) || null;
        if (!token.L2 || token.L2 !== zeroAddress) return;

        token.L2 = (await fetchBridgedToken(alternativeChain, activeChain, token.L1)) || null;
      }

      if (token.L1 === zeroAddress) token.L1 = null;
      if (token.L2 === zeroAddress) token.L2 = null;
    },
    [activeChain, alternativeChain, fetchBridgedToken, fetchNativeToken],
  );

  const isEmptyObject = (obj: object): boolean => {
    return Object.keys(obj).length === 0 && obj.constructor === Object;
  };

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
    estimateGasBridge,
    bridgeEnabled,
    fetchBridgedToken,
    fetchNativeToken,
    fillMissingTokenAddress,
  };
};

export default useBridge;
