import { useAccount } from "wagmi";
import { getPublicClient } from "@wagmi/core";
import { Chain } from "viem";
import { config, wagmiConfig, ConfigManager, NetworkLayer, NetworkType } from "@/config";
import {
  eventERC20,
  eventERC20V2,
  eventETH,
  eventUSDC,
  parseERC20Events,
  parseERC20V2Events,
  parseETHEvents,
  parseUSDCEvents,
} from "@/utils/transactionParsers";
import { ERC20Event, ERC20V2Event, ETHEvent, USDCEvent } from "@/models";
import useERC20Storage from "./useERC20Storage";
import { BlockRange, TransactionHistory } from "@/models/history";
import useFetchAnchoringEvents from "./useFetchAnchoringEvents";
import { OnChainMessageStatus } from "@consensys/linea-sdk";
import useBridge from "./useBridge";
import { getChainNetworkLayer } from "@/utils/chainsUtil";
import { useTokenStore } from "@/stores/tokenStore";
import useMessageStatus from "./useMessageStatus";

const useFetchBridgeTransactions = () => {
  // Wagmi
  const { address } = useAccount();
  const tokensConfig = useTokenStore((state) => state.tokensConfig);
  const { fetchAnchoringMessageHashes } = useFetchAnchoringEvents();
  const { getMessageStatuses } = useMessageStatus();
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

        const newMessages = await getMessageStatuses(txHash, fromLayer);

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
    const client = getPublicClient(wagmiConfig, {
      chainId: fromChain.id,
    });

    if (!client) {
      return;
    }

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

      const [ethLogsForSender, ethLogsForRecipient] = await Promise.all([<Promise<ETHEvent[]>>client.getLogs({
          event: eventETH,
          fromBlock,
          toBlock,
          address: messageServiceAddress,
          args: {
            _from: address,
          },
        }), <Promise<ETHEvent[]>>client.getLogs({
          event: eventETH,
          fromBlock,
          toBlock,
          address: messageServiceAddress,
          args: {
            _to: address,
          },
        })]);

      // Remove duplicates
      const uniqueMap = new Map<string, ETHEvent>();

      [...ethLogsForSender, ...ethLogsForRecipient].forEach((obj) => {
        const uniqueKey = `${obj.args._from}-${obj.args._to}-${obj.transactionHash.toString()}`;
        if (!uniqueMap.has(uniqueKey)) {
          uniqueMap.set(uniqueKey, obj);
        }
      });

      const ethLogs = Array.from(uniqueMap.values());

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

      const [erc20Logs, erc20V2LogsForSender, erc20V2LogsForRecipient] = await Promise.all([
        <Promise<ERC20Event[]>>client.getLogs({
          event: eventERC20,
          fromBlock,
          toBlock,
          address: tokenBridgeAddress,
          args: {
            sender: address,
          },
        }),
        <Promise<ERC20V2Event[]>>client.getLogs({
          event: eventERC20V2,
          fromBlock,
          toBlock,
          address: tokenBridgeAddress,
          args: {
            sender: address,
          },
        }),
        <Promise<ERC20V2Event[]>>client.getLogs({
          event: eventERC20V2,
          fromBlock,
          toBlock,
          address: tokenBridgeAddress,
          args: {
            recipient: address,
          },
        }),
      ]);

      // Remove duplicates
      const uniqueERC20V2EventsMap = new Map<string, ERC20V2Event>();

      [...erc20V2LogsForSender, ...erc20V2LogsForRecipient].forEach((obj) => {
        const uniqueKey = `${obj.args.sender}-${obj.args.recipient}-${obj.transactionHash.toString()}`;
        if (!uniqueERC20V2EventsMap.has(uniqueKey)) {
          uniqueERC20V2EventsMap.set(uniqueKey, obj);
        }
      });

      // Filter out logs that already exist in transactions
      const filteredERC20Logs = erc20Logs.filter(
        (log) => !existingTransactions.some((tx) => tx.transactionHash === log.transactionHash),
      );

      const filteredERC20V2Logs = Array.from(uniqueERC20V2EventsMap.values()).filter(
        (log) => !existingTransactions.some((tx) => tx.transactionHash === log.transactionHash),
      );

      const er20History = await parseERC20Events(
        filteredERC20Logs,
        client,
        fromChain,
        toChain,
        tokensConfig,
        networkType,
      );
      const erc20V2History = await parseERC20V2Events(
        filteredERC20V2Logs,
        client,
        fromChain,
        toChain,
        tokensConfig,
        networkType,
      );

      const history = [...er20History, ...erc20V2History];

      // Add the missing tokens to the user's token list
      history.map(async (transaction) => {
        await fillMissingTokenAddress(transaction.token);
        updateOrInsertUserTokenList(transaction.token, networkType);
      });

      return history;
    };

    /**
     * Fetch USDC history
     * @returns
     */
    const fetchUSDCTransactions = async (existingTransactions: TransactionHistory[]) => {
      const usdcBridgeAddress = await ConfigManager.getUSDCBridgeAddress(networkType, networkLayer);

      const [usdcLogsForSender, usdcLogsForRecipient] = await Promise.all([<Promise<USDCEvent[]>>client.getLogs({
          event: eventUSDC,
          fromBlock,
          toBlock,
          address: usdcBridgeAddress,
          args: {
            depositor: address,
          },
        }), <Promise<USDCEvent[]>>client.getLogs({
          event: eventUSDC,
          fromBlock,
          toBlock,
          address: usdcBridgeAddress,
          args: {
            to: address,
          },
        })]);

      // Remove duplicates
      const uniqueMap = new Map<string, USDCEvent>();

      [...usdcLogsForSender, ...usdcLogsForRecipient].forEach((obj) => {
        const uniqueKey = `${obj.args.depositor}-${obj.args.to}-${obj.transactionHash.toString()}`;
        if (!uniqueMap.has(uniqueKey)) {
          uniqueMap.set(uniqueKey, obj);
        }
      });

      const usdcLogs = Array.from(uniqueMap.values());

      // Filter out logs that already exist in transactions
      const filteredUSDCLogs = usdcLogs.filter(
        (log) => !existingTransactions.some((tx) => tx.transactionHash === log.transactionHash),
      );

      const usdcHistory = await parseUSDCEvents(
        filteredUSDCLogs,
        client,
        fromChain,
        toChain,
        tokensConfig,
        networkType,
      );

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
