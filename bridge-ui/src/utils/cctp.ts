import MessageTransmitterV2 from "@/abis/MessageTransmitterV2.json";
import {
  CCTPMessageReceivedEvent,
  CctpAttestationMessage,
  CctpAttestationMessageStatus,
  TransactionStatus,
} from "@/types";
import { GetPublicClientReturnType } from "@wagmi/core";
import { getAddress } from "viem";
import { eventCCTPMessageReceived } from "./events";
import {
  fetchCctpAttestationByNonce,
  fetchCctpAttestationByTxHash,
  reattestCCTPV2PreFinalityMessage,
} from "@/services/cctp";

// Contract for sending CCTP message
// TODO - Make dynamic for mainnet and testnet chains
export const CCTP_TOKEN_MESSENGER = getAddress("0x8FE6B999Dc680CcFDD5Bf7EB0974218be2542DAA");
// TODO - Find optimal value
export const CCTP_TRANSFER_MAX_FEE = 500n;
// 1000 Fast transfer, 2000 Standard transfer
export const CCTP_MIN_FINALITY_THRESHOLD = 1000;
// Contract for receiving CCTP message
// TODO - Make dynamic for mainnet and testnet chains
export const CCTP_MESSAGE_TRANSMITTER = getAddress("0xE737e5cEBEEBa77EFE34D4aa090756590b1CE275"); // Appears constant for each chain

export const isCCTPNonceUsed = async (client: GetPublicClientReturnType, nonce: string): Promise<boolean> => {
  const resp = await client?.readContract({
    address: CCTP_MESSAGE_TRANSMITTER,
    abi: MessageTransmitterV2.abi,
    functionName: "usedNonces",
    args: [nonce],
  });

  return resp === 1n;
};

export const getCCTPMessageExpiryBlock = (message: string): bigint => {
  // See CCTPV2 message format at https://developers.circle.com/stablecoins/message-format
  const expiryInHex = message.substring(2 + 344 * 2, 2 + 376 * 2);
  // Return bigint because this is also returned by Viem client.getBlockNumber()
  return BigInt(parseInt(expiryInHex, 16));
};

export const getCCTPTransactionStatus = (
  cctpMessageStatus: CctpAttestationMessageStatus,
  isNonceUsed: boolean,
): TransactionStatus => {
  if (cctpMessageStatus === "pending_confirmations") return TransactionStatus.PENDING;
  if (!isNonceUsed) return TransactionStatus.READY_TO_CLAIM;
  return TransactionStatus.COMPLETED;
};

export const getCCTPClaimTx = async (
  client: GetPublicClientReturnType,
  isNonceUsed: boolean,
  nonce: `0x${string}`,
): Promise<string | undefined> => {
  if (!client) return undefined;
  if (!isNonceUsed) return undefined;

  const messageReceivedEvents = <CCTPMessageReceivedEvent[]>await client.getLogs({
    event: eventCCTPMessageReceived,
    // TODO - Find more efficient `fromBlock` param than 'earliest'
    fromBlock: "earliest",
    toBlock: "latest",
    address: CCTP_MESSAGE_TRANSMITTER,
    args: {
      nonce: nonce,
    },
  });

  if (messageReceivedEvents.length === 0) return undefined;
  return messageReceivedEvents[0].transactionHash;
};

export const refreshCCTPMessageIfNeeded = async (
  message: CctpAttestationMessage,
  isNonceUsed: boolean,
  currentToBlock: bigint,
  fromChainCCTPDomain: number,
): Promise<CctpAttestationMessage | undefined> => {
  if (isNonceUsed) return message;
  // Assume that 'pending_confirmations' implies that a message will not be expired
  if (message.status === "pending_confirmations") return message;

  // Check expiry of current message
  const expiryBlock = getCCTPMessageExpiryBlock(message.message);
  if (expiryBlock === 0n) return message;
  if (currentToBlock < expiryBlock) return message;
  console.log("expiryBlock: ", expiryBlock, "currentToBlock", currentToBlock);

  // We have an expired message, reattest
  // TODO - Investigate if this will result in an edge case where a 'READY_TO_CLAIM' tx regresses to a 'PENDING' tx
  await reattestCCTPV2PreFinalityMessage(message.eventNonce);

  const refreshedMessage = await getCCTPMessageByNonce(message.eventNonce, fromChainCCTPDomain);
  return refreshedMessage;
};

export const getCCTPMessageByTxHash = async (
  transactionHash: string,
  fromChainCCTPDomain: number,
): Promise<CctpAttestationMessage | undefined> => {
  const attestationApiResp = await fetchCctpAttestationByTxHash(fromChainCCTPDomain, transactionHash);
  if (!attestationApiResp) return;
  const message = attestationApiResp.messages[0];
  if (!message) return;
  return message;
};

const getCCTPMessageByNonce = async (
  nonce: string,
  fromChainCCTPDomain: number,
): Promise<CctpAttestationMessage | undefined> => {
  const attestationApiResp = await fetchCctpAttestationByNonce(fromChainCCTPDomain, nonce);
  if (!attestationApiResp) return;
  const message = attestationApiResp.messages[0];
  if (!message) return;
  return message;
};
