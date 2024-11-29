import { useAccount } from "wagmi";
import { getPublicClient } from "@wagmi/core";
import { Chain, PublicClient } from "viem";
import { wagmiConfig, ConfigManager, NetworkLayer, NetworkType } from "@/config";
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
import { OnChainMessageStatus } from "@consensys/linea-sdk";
import { getChainNetworkLayer } from "@/utils/chainsUtil";
import { useTokenStore } from "@/stores/tokenStoreProvider";
import useMessageStatus from "./useMessageStatus";
import useTokenFetch from "./useTokenFetch";

const useFetchBridgeTransactions = () => {
  // Wagmi
  const { address } = useAccount();
  const tokensList = useTokenStore((state) => state.tokensList);
  const { getMessageStatuses } = useMessageStatus();
  const { fetchBridgedToken, fillMissingTokenAddress } = useTokenFetch();
  const { updateOrInsertUserTokenList } = useERC20Storage();

  const fetchTransactions = async ({
    networkType,
    l1Chain,
    l2Chain,
    l1FromBlockNumber,
    l2FromBlockNumber,
    transactions,
  }: BlockRange) => {
    if (!l1Chain || !l2Chain) {
      return;
    }

    const [l1TxHistory, l2TxHistory] = await Promise.all([
      fetchBridgeEvents(l1Chain, l2Chain, l1FromBlockNumber, networkType, NetworkLayer.L1, transactions),
      fetchBridgeEvents(l2Chain, l1Chain, l2FromBlockNumber, networkType, NetworkLayer.L2, transactions),
    ]);

    const newTransactions = [...(l1TxHistory ?? []), ...(l2TxHistory ?? [])];

    // Filter out the transactions that already exist
    const uniqueTransactions = newTransactions.filter(
      (newTx) => !transactions.some((existingTx) => existingTx.transactionHash === newTx.transactionHash),
    );

    // Update the messages status for each transactions
    const allTransactions = [...transactions, ...uniqueTransactions];
    allTransactions.sort((a, b) => (b.timestamp < a.timestamp ? -1 : 1));

    await updateMessagesStatus(allTransactions, networkType);
    return allTransactions;
  };

  const updateMessagesStatus = async (transactions: TransactionHistory[], networkType: NetworkType) => {
    const promises = transactions.map(async (transaction, index) => {
      // Only process the transaction that haves messages with unclaimed or unknwon statuses
      const message = transaction.message;
      if (message) {
        const isMessageClaimed = message.status === OnChainMessageStatus.CLAIMED;

        if (isMessageClaimed) {
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
          message: newMessages?.[0],
        };
        transactions[index] = updatedTransaction;
      }
    });

    await Promise.all(promises);
  };

  const fetchBridgeEvents = async (
    fromChain: Chain,
    toChain: Chain,
    fromBlock: bigint,
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

    if (fromBlock < BigInt(0)) {
      return;
    }

    const [ethHistory, erc20History, usdcHistory] = await Promise.all([
      fetchETHTransactions(client, fromChain, toChain, fromBlock, networkType, networkLayer, transactions),
      fetchERC20Transactions(client, fromChain, toChain, fromBlock, networkType, networkLayer, transactions),
      fetchUSDCTransactions(client, fromChain, toChain, fromBlock, networkType, networkLayer, transactions),
    ]);

    return [...ethHistory, ...erc20History, ...usdcHistory];
  };

  const fetchETHTransactions = async (
    client: PublicClient,
    fromChain: Chain,
    toChain: Chain,
    fromBlock: bigint,
    networkType: NetworkType,
    networkLayer: NetworkLayer,
    existingTransactions: TransactionHistory[],
  ) => {
    const messageServiceAddress = await ConfigManager.getMessageServiceAddress(networkType, networkLayer);
    const [ethLogsForSender, ethLogsForRecipient] = await Promise.all([<Promise<ETHEvent[]>>client.getLogs({
        event: eventETH,
        fromBlock,
        toBlock: "latest",
        address: messageServiceAddress,
        args: {
          _from: address,
        },
      }), <Promise<ETHEvent[]>>client.getLogs({
        event: eventETH,
        fromBlock,
        toBlock: "latest",
        address: messageServiceAddress,
        args: {
          _to: address,
        },
      })]);

    const uniqueEthLogs = Array.from(
      new Map(
        [...ethLogsForSender, ...ethLogsForRecipient].map((log) => [
          `${log.args._from}-${log.args._to}-${log.transactionHash.toString()}`,
          log,
        ]),
      ).values(),
    );

    const newEthLogs = uniqueEthLogs.filter(
      (log) => !existingTransactions.some((tx) => tx.transactionHash === log.transactionHash),
    );

    return parseETHEvents(newEthLogs, client, fromChain, toChain, tokensList, networkType);
  };

  const fetchERC20Transactions = async (
    client: PublicClient,
    fromChain: Chain,
    toChain: Chain,
    fromBlock: bigint,
    networkType: NetworkType,
    networkLayer: NetworkLayer,
    existingTransactions: TransactionHistory[],
  ) => {
    const tokenBridgeAddress = await ConfigManager.getTokenBridgeAddress(networkType, networkLayer);
    const [erc20Logs, erc20V2LogsForSender, erc20V2LogsForRecipient] = await Promise.all([
      <Promise<ERC20Event[]>>client.getLogs({
        event: eventERC20,
        fromBlock,
        toBlock: "latest",
        address: tokenBridgeAddress,
        args: {
          sender: address,
        },
      }),
      <Promise<ERC20V2Event[]>>client.getLogs({
        event: eventERC20V2,
        fromBlock,
        toBlock: "latest",
        address: tokenBridgeAddress,
        args: {
          sender: address,
        },
      }),
      <Promise<ERC20V2Event[]>>client.getLogs({
        event: eventERC20V2,
        fromBlock,
        toBlock: "latest",
        address: tokenBridgeAddress,
        args: {
          recipient: address,
        },
      }),
    ]);

    const uniqueERC20V2Logs = Array.from(
      new Map(
        [...erc20V2LogsForSender, ...erc20V2LogsForRecipient].map((log) => [
          `${log.args.sender}-${log.args.recipient}-${log.transactionHash.toString()}`,
          log,
        ]),
      ).values(),
    );

    const filteredERC20Logs = erc20Logs.filter(
      (log) => !existingTransactions.some((tx) => tx.transactionHash === log.transactionHash),
    );

    const filteredERC20V2Logs = uniqueERC20V2Logs.filter(
      (log) => !existingTransactions.some((tx) => tx.transactionHash === log.transactionHash),
    );

    const [erc20History, erc20V2History] = await Promise.all([
      parseERC20Events(filteredERC20Logs, client, fromChain, toChain, tokensList, networkType),
      parseERC20V2Events(filteredERC20V2Logs, client, fromChain, toChain, tokensList, networkType),
    ]);

    const history = [...erc20History, ...erc20V2History];
    await Promise.all(
      history.map(async (transaction) => {
        await fillMissingTokenAddress(transaction.token);
        updateOrInsertUserTokenList(transaction.token, networkType);
      }),
    );

    return history;
  };

  const fetchUSDCTransactions = async (
    client: PublicClient,
    fromChain: Chain,
    toChain: Chain,
    fromBlock: bigint,
    networkType: NetworkType,
    networkLayer: NetworkLayer,
    existingTransactions: TransactionHistory[],
  ) => {
    const usdcBridgeAddress = await ConfigManager.getUSDCBridgeAddress(networkType, networkLayer);
    const [usdcLogsForSender, usdcLogsForRecipient] = await Promise.all([<Promise<USDCEvent[]>>client.getLogs({
        event: eventUSDC,
        fromBlock,
        toBlock: "latest",
        address: usdcBridgeAddress,
        args: {
          depositor: address,
        },
      }), <Promise<USDCEvent[]>>client.getLogs({
        event: eventUSDC,
        fromBlock,
        toBlock: "latest",
        address: usdcBridgeAddress,
        args: {
          to: address,
        },
      })]);

    const uniqueUSDCLogs = Array.from(
      new Map(
        [...usdcLogsForSender, ...usdcLogsForRecipient].map((log) => [
          `${log.args.depositor}-${log.args.to}-${log.transactionHash.toString()}`,
          log,
        ]),
      ).values(),
    );

    const filteredUSDCLogs = uniqueUSDCLogs.filter(
      (log) => !existingTransactions.some((tx) => tx.transactionHash === log.transactionHash),
    );

    return parseUSDCEvents(filteredUSDCLogs, client, fromChain, toChain, tokensList, networkType);
  };

  return { fetchTransactions };
};

export default useFetchBridgeTransactions;
