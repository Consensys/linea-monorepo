import MessageTransmitterV2 from "@/abis/MessageTransmitterV2.json";
import { CctpAttestationMessage, CctpAttestationMessageStatus, TransactionStatus } from "@/types";
import { GetPublicClientReturnType } from "@wagmi/core";
import {
  fetchCctpAttestationByNonce,
  fetchCctpAttestationByTxHash,
  reattestCCTPV2PreFinalityMessage,
} from "@/services/cctp";
import { keccak256 } from "viem";

// TODO - Find optimal value
export const CCTP_TRANSFER_MAX_FEE_FALLBACK = 100n;
// Value we add to 'minimumFee' returned by CCTP API to nudge for timely attestation
export const CCTP_TRANSFER_FEE_BUFFER = 50n;
// 1000 Fast transfer, 2000 Standard transfer
export const CCTP_MIN_FINALITY_THRESHOLD = 1000;

// keccak256("MessageSent(bytes)")
const MessageSentTopic0 = "0x8c5261668696ce22758910d05bab8f186d6eb247ceac2af2e82c7dc17669b036";

// Deterministic nonce for CCTPV2 - https://developers.circle.com/stablecoins/message-format
// txHash, messageIndex, messageHash
// TODO - Get further clarification from Circle on encoding scheme
// We can then streamline CCTP API calls like so: compute Nonce -> check if nonce used
//   -> (NonceUsed) Assign Complete Status
//   -> (NonceNotUsed) Status is either PENDING or READY_TO_CLAIM, consult CCTP API
export const getCCTPNonce = async (
  client: GetPublicClientReturnType,
  depositTxHash: `0x${string}`,
  nonce?: string,
): Promise<string | undefined> => {
  // Get txReceipt
  const txReceipt = await client?.getTransactionReceipt({ hash: depositTxHash });
  if (!txReceipt) return;
  const messageSentEventLog = txReceipt.logs.find((log) => log.topics[0] === MessageSentTopic0);
  if (!messageSentEventLog) return;
  const messageIndex = messageSentEventLog.logIndex;
  const message = messageSentEventLog.data;
  const messageHash = keccak256(message);
  console.log(
    `actualNonce: ${nonce}, depositTxHash: ${depositTxHash}, messageIndex: ${messageIndex}, messageHash: ${messageHash}`,
  );
  console.log("0xdeadbeaf");
  return "";
};

export const isCCTPNonceUsed = async (
  client: GetPublicClientReturnType,
  nonce: string,
  cctpMessageTransmitterV2Address: `0x${string}`,
): Promise<boolean> => {
  const resp = await client?.readContract({
    address: cctpMessageTransmitterV2Address,
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

export const refreshCCTPMessageIfNeeded = async (
  message: CctpAttestationMessage,
  status: TransactionStatus,
  currentToBlock: bigint,
  fromChainCCTPDomain: number,
): Promise<CctpAttestationMessage | undefined> => {
  if (status === TransactionStatus.COMPLETED) return message;
  // Assume that 'pending_confirmations' implies that a message will not be expired
  if (message.status === "pending_confirmations") return message;

  // Check expiry of current message
  const expiryBlock = getCCTPMessageExpiryBlock(message.message);
  if (expiryBlock === 0n) return message;
  if (currentToBlock < expiryBlock) return message;

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

export const getCCTPMessageByNonce = async (
  nonce: string,
  fromChainCCTPDomain: number,
): Promise<CctpAttestationMessage | undefined> => {
  const attestationApiResp = await fetchCctpAttestationByNonce(fromChainCCTPDomain, nonce);
  if (!attestationApiResp) return;
  const message = attestationApiResp.messages[0];
  if (!message) return;
  return message;
};
