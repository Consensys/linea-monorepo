import {
  claimOnL1,
  getL2ToL1MessageStatus,
  getMessageProof as sdkGetMessageProof,
  getTransactionReceiptByMessageHash as sdkGetTransactionReceiptByMessageHash,
} from "@consensys/linea-sdk-viem";
import { type Hex, type PublicClient, type WalletClient, decodeErrorResult } from "viem";
import { estimateContractGas, getContractEvents, readContract } from "viem/actions";

import { mapViemReceiptToCoreReceipt } from "./mappers";
import { ILineaRollupClient } from "../../../core/clients/blockchain/ethereum/ILineaRollupClient";
import { Proof } from "../../../core/clients/blockchain/ethereum/IMerkleTreeService";
import { IEthereumGasProvider } from "../../../core/clients/blockchain/IGasProvider";
import { ZERO_ADDRESS } from "../../../core/constants/blockchain";
import { DEFAULT_RATE_LIMIT_MARGIN } from "../../../core/constants/common";
import { MessageProps } from "../../../core/entities/Message";
import { OnChainMessageStatus } from "../../../core/enums";
import {
  ErrorDescription,
  MessageSent,
  Overrides,
  TransactionReceipt,
  TransactionSubmission,
} from "../../../core/types";
import { LineaRollupAbi } from "../abis/LineaRollupAbi";

export class ViemLineaRollupClient implements ILineaRollupClient {
  constructor(
    private readonly publicClient: PublicClient,
    private readonly walletClient: WalletClient,
    private readonly contractAddress: string,
    private readonly l2PublicClient: PublicClient,
    private readonly l2ContractAddress: string,
    private readonly gasProvider: IEthereumGasProvider,
  ) {}

  public async getMessageStatus(params: {
    messageHash: string;
    messageBlockNumber?: number;
    overrides?: Overrides;
  }): Promise<OnChainMessageStatus> {
    return getL2ToL1MessageStatus(this.publicClient, {
      l2Client: this.l2PublicClient,
      messageHash: params.messageHash as Hex,
      lineaRollupAddress: this.contractAddress as Hex,
      l2MessageServiceAddress: this.l2ContractAddress as Hex,
      l2LogsBlockRange: params.messageBlockNumber
        ? { fromBlock: BigInt(params.messageBlockNumber), toBlock: BigInt(params.messageBlockNumber) }
        : undefined,
    });
  }

  public async getMessageProof(messageHash: string, messageBlockNumber?: number): Promise<Proof> {
    return sdkGetMessageProof(this.publicClient, {
      l2Client: this.l2PublicClient,
      messageHash: messageHash as Hex,
      lineaRollupAddress: this.contractAddress as Hex,
      l2MessageServiceAddress: this.l2ContractAddress as Hex,
      l2LogsBlockRange: messageBlockNumber
        ? { fromBlock: BigInt(messageBlockNumber), toBlock: BigInt(messageBlockNumber) }
        : undefined,
    });
  }

  public async estimateClaimGas(
    message: (MessageSent | MessageProps) & { feeRecipient?: string; messageBlockNumber?: number },
    opts: { claimViaAddress?: string; overrides?: Overrides } = {},
  ): Promise<bigint> {
    const { proof, leafIndex, root } = await this.getMessageProof(
      message.messageHash,
      "messageBlockNumber" in message ? message.messageBlockNumber : undefined,
    );

    const from = message.messageSender;
    const to = message.destination;
    const calldata = message.calldata;
    const nonce = message.messageNonce;
    const feeRecipient = message.feeRecipient ?? ZERO_ADDRESS;

    const gasFees = await this.gasProvider.getGasFees();
    const contractAddress = (opts.claimViaAddress ?? this.contractAddress) as Hex;
    const [account] = await this.walletClient.getAddresses();

    return estimateContractGas(this.publicClient, {
      address: contractAddress,
      abi: LineaRollupAbi,
      functionName: "claimMessageWithProof",
      args: [
        {
          proof: proof as `0x${string}`[],
          messageNumber: nonce,
          leafIndex,
          from: from as Hex,
          to: to as Hex,
          fee: message.fee,
          value: message.value,
          feeRecipient: feeRecipient as Hex,
          merkleRoot: root as Hex,
          data: calldata as Hex,
        },
      ],
      account,
      maxFeePerGas: opts.overrides?.maxFeePerGas ?? gasFees.maxFeePerGas,
      maxPriorityFeePerGas: opts.overrides?.maxPriorityFeePerGas ?? gasFees.maxPriorityFeePerGas,
    });
  }

  public async claim(
    message: (MessageSent | MessageProps) & { feeRecipient?: string; messageBlockNumber?: number },
    opts: { claimViaAddress?: string; overrides?: Overrides } = {},
  ): Promise<TransactionSubmission> {
    const { proof, leafIndex, root } = await this.getMessageProof(
      message.messageHash,
      "messageBlockNumber" in message ? message.messageBlockNumber : undefined,
    );

    const from = message.messageSender;
    const to = message.destination;
    const calldata = message.calldata;
    const nonce = message.messageNonce;
    const feeRecipient = (message.feeRecipient ?? ZERO_ADDRESS) as Hex;

    const gasFees = await this.gasProvider.getGasFees();
    const contractAddress = (opts.claimViaAddress ?? this.contractAddress) as Hex;

    const txHash = await claimOnL1(this.walletClient, {
      from: from as Hex,
      to: to as Hex,
      fee: message.fee,
      value: message.value,
      messageNonce: nonce,
      calldata: calldata as Hex,
      feeRecipient,
      messageProof: { proof: proof as `0x${string}`[], root: root as `0x${string}`, leafIndex },
      lineaRollupAddress: contractAddress,
      nonce: opts.overrides?.nonce,
      gas: opts.overrides?.gasLimit,
      maxFeePerGas: opts.overrides?.maxFeePerGas ?? gasFees.maxFeePerGas,
      maxPriorityFeePerGas: opts.overrides?.maxPriorityFeePerGas ?? gasFees.maxPriorityFeePerGas,
    });

    return {
      hash: txHash,
      nonce: opts.overrides?.nonce ?? 0,
      gasLimit: opts.overrides?.gasLimit ?? 0n,
      maxFeePerGas: opts.overrides?.maxFeePerGas ?? gasFees.maxFeePerGas,
      maxPriorityFeePerGas: opts.overrides?.maxPriorityFeePerGas ?? gasFees.maxPriorityFeePerGas,
    };
  }

  public async retryTransactionWithHigherFee(
    transactionHash: string,
    priceBumpPercent: number = 10,
  ): Promise<TransactionSubmission> {
    if (!Number.isInteger(priceBumpPercent)) {
      throw new Error("'priceBumpPercent' must be an integer");
    }

    const transaction = await this.publicClient.getTransaction({ hash: transactionHash as Hex });

    if (!transaction) {
      throw new Error(`Transaction with hash ${transactionHash} not found.`);
    }

    let maxFeePerGas: bigint;
    let maxPriorityFeePerGas: bigint;

    if (!transaction.maxFeePerGas || !transaction.maxPriorityFeePerGas) {
      const fees = await this.gasProvider.getGasFees();
      maxFeePerGas = fees.maxFeePerGas;
      maxPriorityFeePerGas = fees.maxPriorityFeePerGas;
    } else {
      const bump = BigInt(priceBumpPercent) + 100n;
      maxFeePerGas = (transaction.maxFeePerGas * bump) / 100n;
      maxPriorityFeePerGas = (transaction.maxPriorityFeePerGas * bump) / 100n;
      const cap = this.gasProvider.getMaxFeePerGas();
      if (maxFeePerGas > cap) maxFeePerGas = cap;
      if (maxPriorityFeePerGas > cap) maxPriorityFeePerGas = cap;
    }

    const txHash = await this.walletClient.sendTransaction({
      account: transaction.from,
      to: transaction.to ?? undefined,
      value: transaction.value,
      data: transaction.input,
      nonce: transaction.nonce,
      gas: transaction.gas,
      maxFeePerGas,
      maxPriorityFeePerGas,
      chain: null,
    });

    return {
      hash: txHash,
      nonce: transaction.nonce,
      gasLimit: transaction.gas,
      maxFeePerGas,
      maxPriorityFeePerGas,
    };
  }

  public async isRateLimitExceeded(messageFee: bigint, messageValue: bigint): Promise<boolean> {
    const [limitInWei, currentPeriodAmountInWei] = await Promise.all([
      readContract(this.publicClient, {
        address: this.contractAddress as Hex,
        abi: LineaRollupAbi,
        functionName: "limitInWei",
      }),
      readContract(this.publicClient, {
        address: this.contractAddress as Hex,
        abi: LineaRollupAbi,
        functionName: "currentPeriodAmountInWei",
      }),
    ]);

    return (
      parseFloat((currentPeriodAmountInWei + messageFee + messageValue).toString()) >
      parseFloat(limitInWei.toString()) * DEFAULT_RATE_LIMIT_MARGIN
    );
  }

  public async isRateLimitExceededError(transactionHash: string): Promise<boolean> {
    const parsedError = await this.parseTransactionError(transactionHash);
    if (typeof parsedError === "string") return false;
    return parsedError.name === "RateLimitExceeded";
  }

  public async parseTransactionError(transactionHash: string): Promise<ErrorDescription | string> {
    let errorEncodedData: `0x${string}` = "0x";
    try {
      const tx = await this.publicClient.getTransaction({ hash: transactionHash as Hex });
      if (!tx) return errorEncodedData;

      try {
        const { data } = await this.publicClient.call({
          to: tx.to ?? undefined,
          account: tx.from,
          nonce: tx.nonce,
          gas: tx.gas,
          data: tx.input,
          value: tx.value,
          maxFeePerGas: tx.maxFeePerGas ?? undefined,
          maxPriorityFeePerGas: tx.maxPriorityFeePerGas ?? undefined,
        });
        errorEncodedData = data ?? "0x";
      } catch (callError) {
        // Extract revert data from the call error if possible
        const errData = (callError as { data?: `0x${string}` }).data;
        if (errData) errorEncodedData = errData;
      }

      if (errorEncodedData === "0x") return errorEncodedData;

      const decoded = decodeErrorResult({
        abi: LineaRollupAbi,
        data: errorEncodedData,
      });

      return {
        name: decoded.errorName,
        args: decoded.args ? [...decoded.args] : [],
      };
    } catch {
      return errorEncodedData;
    }
  }

  public async getMessageStatusUsingMerkleTree(params: {
    messageHash: string;
    messageBlockNumber?: number;
    overrides?: Overrides;
  }): Promise<OnChainMessageStatus> {
    return this.getMessageStatus(params);
  }

  public async getMessageStatusUsingMessageHash(
    messageHash: string,
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    _opts: Record<string, unknown>,
  ): Promise<OnChainMessageStatus> {
    const statusValue = await readContract(this.publicClient, {
      address: this.contractAddress as Hex,
      abi: LineaRollupAbi,
      functionName: "inboxL2L1MessageStatus",
      args: [messageHash as Hex],
    });

    if (statusValue === 2n) return OnChainMessageStatus.CLAIMED;
    if (statusValue === 1n) return OnChainMessageStatus.CLAIMABLE;

    const claimedEvents = await getContractEvents(this.publicClient, {
      address: this.contractAddress as Hex,
      abi: LineaRollupAbi,
      eventName: "MessageClaimed",
      args: { _messageHash: messageHash as Hex },
    });

    if (claimedEvents.length > 0) return OnChainMessageStatus.CLAIMED;
    return OnChainMessageStatus.UNKNOWN;
  }

  public getMessageSiblings(messageHash: string, messages: string[], treeDepth: number): string[] {
    const batchSize = 2 ** treeDepth;
    if (!messages.includes(messageHash)) {
      throw new Error("Message hash not found in messages");
    }
    const result = [messageHash];
    for (const h of messages) {
      if (h !== messageHash && result.length < batchSize) {
        result.push(h);
      }
    }
    const zeroHash = "0x" + "00".repeat(32);
    while (result.length < batchSize) {
      result.push(zeroHash);
    }
    return result;
  }

  public async getFinalizationMessagingInfo(transactionHash: string): Promise<{
    treeDepth: number;
    l2MerkleRoots: string[];
    l2MessagingBlocksRange: { startingBlock: number; endBlock: number };
  }> {
    const L2_MERKLE_ROOT_ADDED_SIG = "0x300e6f978eee6a4b0bba78dd8400dc64fd5652dbfc868a2258e16d0977be222b";
    const L2_MESSAGING_BLOCK_ANCHORED_SIG = "0x3c116827db9db3a30c1a25db8b0ee4bab9d2b223560209cfd839601b621c726d";

    const receipt = await this.publicClient.getTransactionReceipt({ hash: transactionHash as Hex });
    if (!receipt || receipt.logs.length === 0) {
      throw new Error("Transaction does not exist or no logs found");
    }

    const merkleRootLogs = receipt.logs.filter((log) => log.topics[0] === L2_MERKLE_ROOT_ADDED_SIG);
    if (merkleRootLogs.length === 0) {
      throw new Error("No L2MerkleRootAdded events found");
    }

    const l2MerkleRoots = merkleRootLogs.map((log) => log.topics[1] as string);
    const treeDepth = Number(BigInt(merkleRootLogs[0].topics[2] as string));

    const blockAnchoredLogs = receipt.logs.filter((log) => log.topics[0] === L2_MESSAGING_BLOCK_ANCHORED_SIG);
    const blocks = blockAnchoredLogs.map((log) => Number(BigInt(log.topics[1] as string)));
    const startingBlock = Math.min(...blocks);
    const endBlock = Math.max(...blocks);

    return { treeDepth, l2MerkleRoots, l2MessagingBlocksRange: { startingBlock, endBlock } };
  }

  public async getMessageByMessageHash(messageHash: string): Promise<MessageSent | null> {
    const events = await getContractEvents(this.publicClient, {
      address: this.contractAddress as Hex,
      abi: LineaRollupAbi,
      eventName: "MessageSent",
      args: { _messageHash: messageHash as Hex },
      fromBlock: "earliest",
      toBlock: "latest",
    });
    if (!events.length) return null;
    const e = events[0];
    const args = e.args as {
      _messageHash: string;
      _from: string;
      _to: string;
      _fee: bigint;
      _value: bigint;
      _nonce: bigint;
      _calldata: string;
    };
    return {
      messageHash: args._messageHash,
      messageSender: args._from,
      destination: args._to,
      fee: args._fee,
      value: args._value,
      messageNonce: args._nonce,
      calldata: args._calldata,
      contractAddress: this.contractAddress,
      blockNumber: Number(e.blockNumber),
      transactionHash: e.transactionHash!,
      logIndex: e.logIndex!,
    };
  }

  public async getTransactionReceiptByMessageHash(messageHash: string): Promise<TransactionReceipt | null> {
    try {
      const receipt = await sdkGetTransactionReceiptByMessageHash(this.publicClient, {
        messageHash: messageHash as Hex,
        messageServiceAddress: this.contractAddress as Hex,
      });
      return mapViemReceiptToCoreReceipt(receipt);
    } catch {
      return null;
    }
  }
}
