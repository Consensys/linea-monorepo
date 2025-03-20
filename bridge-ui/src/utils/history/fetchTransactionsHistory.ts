import { Address } from "viem";
import { LineaSDK } from "@consensys/linea-sdk";
import { config } from "@/config";
import { BridgeTransaction, Chain, Token } from "@/types";
import { fetchETHBridgeEvents } from "./fetchETHBridgeEvents";
import { fetchERC20BridgeEvents } from "./fetchERC20BridgeEvents";
import { fetchCCTPBridgeEvents } from "./fetchCCTPBridgeEvents";

type TransactionHistoryParams = {
  lineaSDK: LineaSDK;
  fromChain: Chain;
  toChain: Chain;
  address: Address;
  tokens: Token[];
};

export async function fetchTransactionsHistory({
  lineaSDK,
  fromChain,
  toChain,
  address,
  tokens,
}: TransactionHistoryParams): Promise<BridgeTransaction[]> {
  const events = await Promise.all([
    fetchBridgeEvents(lineaSDK, fromChain, toChain, address, tokens),
    fetchBridgeEvents(lineaSDK, toChain, fromChain, address, tokens),
  ]);
  return events.flat().sort((a, b) => Number(b.timestamp.toString()) - Number(a.timestamp.toString()));
}

async function fetchBridgeEvents(
  lineaSDK: LineaSDK,
  fromChain: Chain,
  toChain: Chain,
  address: Address,
  tokens: Token[],
): Promise<BridgeTransaction[]> {
  const [ethEvents, erc20Events, cctpEvents] = await Promise.all([
    fetchETHBridgeEvents(lineaSDK, address, fromChain, toChain, tokens),
    fetchERC20BridgeEvents(lineaSDK, address, fromChain, toChain, tokens),
    // Feature toggle for CCTP, will filter out USDC transactions if isCCTPEnabled == false
    config.isCCTPEnabled ? fetchCCTPBridgeEvents(address, fromChain, toChain, tokens) : [],
  ]);

  return [...ethEvents, ...erc20Events, ...cctpEvents];
}
