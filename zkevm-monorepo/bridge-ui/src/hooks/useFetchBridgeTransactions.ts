import { useAccount } from 'wagmi';
import { getPublicClient } from '@wagmi/core';
import { Chain } from 'viem';

import { config } from '@/config';
import { NetworkLayer, NetworkType } from '@/contexts/chain.context';
import { ConfigManager } from '@/config/config';
import {
  eventERC20,
  eventETH,
  eventUSDC,
  parseERC20Events,
  parseETHEvents,
  parseUSDCEvents,
} from '@/utils/transactionParsers';
import { ERC20Event, ETHEvent, USDCEvent } from '@/models';
import useERC20Storage from './useERC20Storage';
import { BlockRange, TransactionHistory } from '@/models/history';
import useFetchAnchoringEvents from './useFetchAnchoringEvents';
import { OnChainMessageStatus } from '@consensys/linea-sdk';
import useMessageService from './useMessageService';
import useBridge from './useBridge';
import { getChainNetworkLayer } from '@/utils/chainsUtil';
import { useConfigContext } from '@/contexts/token.context';

const useFetchBridgeTransactions = () => {
  // Wagmi
  const { address } = useAccount();

  const { tokensConfig } = useConfigContext();

  const { fetchAnchoringMessageHashes } = useFetchAnchoringEvents();
  const { getMessagesStatusesByTransactionHash } = useMessageService();
  const { fetchBridgedToken, fillMissingTokenAddress } = useBridge();
  const { updateOrInsertUserTokenList } = useERC20Storage();

  const fetchTransactions = async ({
    networkType,
    l1Chain,
    l2Chain,
    l1FromBlockNumber,
    l1ToBlockNumber,
    l2FromBlockNumber,
    l2ToBlockNumber,
    transactions,
  }: BlockRange) => {
    if (!l1Chain || !l2Chain) {
      return;
    }

    const l1TxHistory = await fetchBridgeEvents(
      l1Chain,
      l2Chain,
      l1FromBlockNumber,
      l1ToBlockNumber,
      networkType,
      NetworkLayer.L1,
      transactions,
    );

    const l2TxHistory = await fetchBridgeEvents(
      l2Chain,
      l1Chain,
      l2FromBlockNumber,
      l2ToBlockNumber,
      networkType,
      NetworkLayer.L2,
      transactions,
    );

    const newTransactions = [...(l1TxHistory ?? []), ...(l2TxHistory ?? [])];

    // Filter out the transactions that already exist
    const uniqueTransactions = newTransactions.filter(
      (newTx) => !transactions.some((existingTx) => existingTx.transactionHash === newTx.transactionHash),
    );

    // Get currently anchoring messages
    const messageHashes = await fetchAnchoringMessageHashes(
      l2Chain,
      config.networks[networkType].L2.messageServiceAddress,
    );

    // Update the messages status for each transactions
    const allTransactions = [...transactions, ...uniqueTransactions];
    allTransactions.sort((a, b) => (b.timestamp < a.timestamp ? -1 : 1));

    await updateMessagesStatus(allTransactions, messageHashes, networkType);
    return allTransactions;
  };

  const updateMessagesStatus = async (
    transactions: TransactionHistory[],
    messageHashes: string[],
    networkType: NetworkType,
  ) => {
    const promises = transactions.map(async (transaction, index) => {
      // Only process the transaction that haves messages with unclaimed or unknwon statuses
      const messages = transaction.messages;
      if (messages && messages.length > 0) {
        const hasUnClaimedMessage = messages.filter((message) => message.status !== OnChainMessageStatus.CLAIMED);

        if (hasUnClaimedMessage.length === 0) {
          // We skip this one since all the messages have been claimed
          return;
        }
      }

      const txHash = transaction.transactionHash;
      const fromLayer = getChainNetworkLayer(transaction.fromChain);
      const toLayer = getChainNetworkLayer(transaction.toChain);
      if (fromLayer && toLayer) {
        // Update message status and the token address on the destination for ERC20s
        const fromLayerToken = transaction.token[fromLayer];
        let toLayerToken = transaction.token[toLayer];
        if (!toLayerToken && fromLayerToken) {
          toLayerToken = (await fetchBridgedToken(transaction.fromChain, transaction.toChain, fromLayerToken)) || null;
          // Update or add the token in the user's token list
          transaction.token[toLayer] = toLayerToken;
          updateOrInsertUserTokenList(transaction.token, networkType);
        }

        const newMessages = await getMessagesStatusesByTransactionHash(txHash, fromLayer);
        const updatedTransaction = {
          ...transaction,
          token: {
            ...transaction.token,
            [toLayer]: toLayerToken,
          },
          messages: newMessages,
        };
        transactions[index] = updatedTransaction;
      }

      // Check pending anchoring to remove claim button
      for (const transaction of transactions) {
        if (!transaction.messages?.length) {
          continue;
        }
        for (const message of transaction.messages) {
          if (messageHashes.find((messageHash) => messageHash === message.messageHash)) {
            // Set to UNKNOWN if recent anchoring found, should directly mutate _transactions
            message.status = OnChainMessageStatus.UNKNOWN;
          }
        }
      }
    });

    await Promise.all(promises);
  };

  const fetchBridgeEvents = async (
    fromChain: Chain,
    toChain: Chain,
    fromBlock: bigint,
    toBlock: bigint,
    networkType: NetworkType,
    networkLayer: NetworkLayer,
    transactions: TransactionHistory[],
  ) => {
    const client = getPublicClient({
      chainId: fromChain.id,
    });

    if (networkLayer !== NetworkLayer.L1 && networkLayer !== NetworkLayer.L2) {
      return;
    }

    if (fromBlock < BigInt(0) || toBlock < BigInt(0) || fromBlock > toBlock) {
      return;
    }

    let history: TransactionHistory[] = [];

    /**
     * Fetch ETH history
     * @returns
     */
    const fetchETHTransactions = async (existingTransactions: TransactionHistory[]) => {
      const messageServiceAddress = await ConfigManager.getMessageServiceAddress(networkType, networkLayer);

      const ethLogs = <ETHEvent[]>await client.getLogs({
        event: eventETH,
        fromBlock,
        toBlock,
        address: messageServiceAddress,
        args: {
          _from: address,
        },
      });

      // Filter out logs that already exist in transactions
      const newEthLogs = ethLogs.filter(
        (log) => !existingTransactions.some((tx) => tx.transactionHash === log.transactionHash),
      );

      const ethHistory = await parseETHEvents(newEthLogs, client, fromChain, toChain, tokensConfig, networkType);

      return ethHistory;
    };

    /**
     * Fetch ERC20 history
     * @returns
     */
    const fetchERC20Transactions = async (existingTransactions: TransactionHistory[]) => {
      const tokenBridgeAddress = await ConfigManager.getTokenBridgeAddress(networkType, networkLayer);

      const erc20Logs = <ERC20Event[]>await client.getLogs({
        event: eventERC20,
        fromBlock,
        toBlock,
        address: tokenBridgeAddress,
        args: {
          sender: address,
        },
      });

      // Filter out logs that already exist in transactions
      const newErc20Logs = erc20Logs.filter(
        (log) => !existingTransactions.some((tx) => tx.transactionHash === log.transactionHash),
      );

      const er20History = await parseERC20Events(newErc20Logs, client, fromChain, toChain, tokensConfig, networkType);

      // Add the missing tokens to the user's token list
      er20History.map(async (transaction) => {
        await fillMissingTokenAddress(transaction.token);
        updateOrInsertUserTokenList(transaction.token, networkType);
      });

      return er20History;
    };

    /**
     * Fetch USDC history
     * @returns
     */
    const fetchUSDCTransactions = async (existingTransactions: TransactionHistory[]) => {
      const usdcBridgeAddress = await ConfigManager.getUSDCBridgeAddress(networkType, networkLayer);

      const usdcLogs = <USDCEvent[]>await client.getLogs({
        event: eventUSDC,
        fromBlock,
        toBlock,
        address: usdcBridgeAddress,
        args: {
          depositor: address,
        },
      });

      // Filter out logs that already exist in transactions
      const newUsdcLogs = usdcLogs.filter(
        (log) => !existingTransactions.some((tx) => tx.transactionHash === log.transactionHash),
      );

      const usdcHistory = await parseUSDCEvents(newUsdcLogs, client, fromChain, toChain, tokensConfig, networkType);

      return usdcHistory;
    };

    const ethHistory = await fetchETHTransactions(transactions);
    const erc20History = await fetchERC20Transactions(transactions);
    const usdcHistory = await fetchUSDCTransactions(transactions);

    history = [...ethHistory, ...erc20History, ...usdcHistory];

    return history;
  };

  return { fetchTransactions };
};

export default useFetchBridgeTransactions;
