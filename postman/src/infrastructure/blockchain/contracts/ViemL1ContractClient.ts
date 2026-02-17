import {
  claimOnL1,
  getL2ToL1MessageStatus,
  getMessageProof,
  OnChainMessageStatus as SdkOnChainMessageStatus,
} from "@consensys/linea-sdk-viem";
import { type Address, type Client, type Hex, type PublicClient, type WalletClient, decodeErrorResult } from "viem";

import { type MessageSent, type TransactionResponse, OnChainMessageStatus } from "../../../domain/types";
import { LineaRollupAbi } from "../abis/LineaRollupAbi";

import type { MessageProps } from "../../../domain/message/Message";
import type { IEthereumGasProvider } from "../../../domain/ports/IGasProvider";
import type { IL1ContractClient } from "../../../domain/ports/IL1ContractClient";
import type { ClaimTransactionOverrides } from "../../../domain/ports/IMessageServiceContract";

const ZERO_ADDRESS = "0x0000000000000000000000000000000000000000";
const DEFAULT_RATE_LIMIT_MARGIN = 0.9;

function mapSdkStatus(status: SdkOnChainMessageStatus): OnChainMessageStatus {
  switch (status) {
    case SdkOnChainMessageStatus.CLAIMED:
      return OnChainMessageStatus.CLAIMED;
    case SdkOnChainMessageStatus.CLAIMABLE:
      return OnChainMessageStatus.CLAIMABLE;
    default:
      return OnChainMessageStatus.UNKNOWN;
  }
}

export class ViemL1ContractClient implements IL1ContractClient {
  constructor(
    private readonly l1PublicClient: PublicClient,
    private readonly l1WalletClient: WalletClient,
    private readonly l2PublicClient: PublicClient,
    private readonly contractAddress: Address,
    private readonly l2ContractAddress: Address,
    private readonly gasProvider: IEthereumGasProvider,
  ) {}

  public async getMessageStatus(params: {
    messageHash: Hex;
    messageBlockNumber?: number;
  }): Promise<OnChainMessageStatus> {
    const sdkStatus = await getL2ToL1MessageStatus(this.l1PublicClient, {
      l2Client: this.l2PublicClient,
      messageHash: params.messageHash,
      lineaRollupAddress: this.contractAddress,
      l2MessageServiceAddress: this.l2ContractAddress,
      ...(params.messageBlockNumber
        ? {
            l2LogsBlockRange: {
              fromBlock: BigInt(params.messageBlockNumber),
              toBlock: BigInt(params.messageBlockNumber),
            },
          }
        : {}),
    });

    return mapSdkStatus(sdkStatus);
  }

  public async claim(
    message: (MessageSent | MessageProps) & { feeRecipient?: string; messageBlockNumber?: number },
    opts: { claimViaAddress?: string; overrides?: ClaimTransactionOverrides } = {},
  ): Promise<TransactionResponse> {
    const gasFees = await this.gasProvider.getGasFees();

    const hash = await claimOnL1(this.l1WalletClient, {
      from: message.messageSender as Address,
      to: message.destination as Address,
      fee: message.fee,
      value: message.value,
      messageNonce: message.messageNonce,
      calldata: message.calldata as Hex,
      feeRecipient: (message.feeRecipient ?? ZERO_ADDRESS) as Address,
      l2Client: this.l2PublicClient as Client,
      lineaRollupAddress: (opts.claimViaAddress ?? this.contractAddress) as Address,
      l2MessageServiceAddress: this.l2ContractAddress,
      gas: opts.overrides?.gasLimit,
      nonce: opts.overrides?.nonce,
      maxFeePerGas: opts.overrides?.maxFeePerGas ?? gasFees.maxFeePerGas,
      maxPriorityFeePerGas: opts.overrides?.maxPriorityFeePerGas ?? gasFees.maxPriorityFeePerGas,
    });

    return this.buildTransactionResponse(hash, opts.overrides);
  }

  public async estimateClaimGas(
    message: (MessageSent | MessageProps) & { feeRecipient?: string; messageBlockNumber?: number },
    opts: { claimViaAddress?: string } = {},
  ): Promise<bigint> {
    const { messageSender, destination, fee, value, calldata, messageNonce, feeRecipient } = message;
    const l1FeeRecipient = (feeRecipient ?? ZERO_ADDRESS) as Address;
    const messageHash = (message as MessageSent).messageHash ?? "";

    const proof = await getMessageProof(this.l1PublicClient as Client, {
      l2Client: this.l2PublicClient as Client,
      messageHash: messageHash as Hex,
      lineaRollupAddress: this.contractAddress,
      l2MessageServiceAddress: this.l2ContractAddress,
    });

    const gasFees = await this.gasProvider.getGasFees();
    const targetAddress = (opts.claimViaAddress ?? this.contractAddress) as Address;

    return this.l1PublicClient.estimateContractGas({
      address: targetAddress,
      abi: LineaRollupAbi,
      functionName: "claimMessageWithProof",
      args: [
        {
          from: messageSender as Address,
          to: destination as Address,
          fee: BigInt(fee),
          value: BigInt(value),
          data: calldata as Hex,
          messageNumber: BigInt(messageNonce),
          proof: proof.proof,
          leafIndex: proof.leafIndex,
          merkleRoot: proof.root,
          feeRecipient: l1FeeRecipient,
        },
      ],
      maxFeePerGas: gasFees.maxFeePerGas,
      maxPriorityFeePerGas: gasFees.maxPriorityFeePerGas,
      account: this.l1WalletClient.account!,
    });
  }

  public async isRateLimitExceeded(messageFee: bigint, messageValue: bigint): Promise<boolean> {
    const [limitInWei, currentPeriodAmountInWei] = await Promise.all([
      this.l1PublicClient.readContract({
        address: this.contractAddress,
        abi: LineaRollupAbi,
        functionName: "limitInWei",
      }) as Promise<bigint>,
      this.l1PublicClient.readContract({
        address: this.contractAddress,
        abi: LineaRollupAbi,
        functionName: "currentPeriodAmountInWei",
      }) as Promise<bigint>,
    ]);

    return (
      parseFloat((currentPeriodAmountInWei + messageFee + messageValue).toString()) >
      parseFloat(limitInWei.toString()) * DEFAULT_RATE_LIMIT_MARGIN
    );
  }

  public async isRateLimitExceededError(transactionHash: string): Promise<boolean> {
    const parsedError = await this.parseTransactionError(transactionHash);
    return parsedError === "RateLimitExceeded";
  }

  public async parseTransactionError(transactionHash: string): Promise<string> {
    try {
      const tx = await this.l1PublicClient.getTransaction({ hash: transactionHash as Hex });
      if (!tx) return "0x";

      const errorData = await this.l1PublicClient.call({
        to: tx.to ?? undefined,
        data: tx.input,
        account: tx.from,
      });

      if (!errorData.data || errorData.data === "0x") return "0x";

      try {
        const decoded = decodeErrorResult({ abi: LineaRollupAbi, data: errorData.data });
        return decoded.errorName;
      } catch {
        return errorData.data;
      }
    } catch {
      return "0x";
    }
  }

  public async retryTransactionWithHigherFee(
    transactionHash: string,
    priceBumpPercent = 10,
  ): Promise<TransactionResponse> {
    const tx = await this.l1PublicClient.getTransaction({ hash: transactionHash as Hex });
    if (!tx) throw new Error(`Transaction not found: ${transactionHash}`);

    const bumpMultiplier = BigInt(100 + priceBumpPercent);
    const newMaxFeePerGas = ((tx.maxFeePerGas ?? 0n) * bumpMultiplier) / 100n;
    const newMaxPriorityFeePerGas = ((tx.maxPriorityFeePerGas ?? 0n) * bumpMultiplier) / 100n;

    const hash = await this.l1WalletClient.sendTransaction({
      to: tx.to as Address,
      data: tx.input,
      value: tx.value,
      nonce: tx.nonce,
      gas: tx.gas,
      maxFeePerGas: newMaxFeePerGas,
      maxPriorityFeePerGas: newMaxPriorityFeePerGas,
      chain: this.l1WalletClient.chain ?? null,
      account: this.l1WalletClient.account!,
    });

    return {
      hash,
      gasLimit: tx.gas,
      maxFeePerGas: newMaxFeePerGas,
      maxPriorityFeePerGas: newMaxPriorityFeePerGas,
      nonce: tx.nonce,
    };
  }

  private async buildTransactionResponse(
    hash: Hex,
    overrides?: ClaimTransactionOverrides,
  ): Promise<TransactionResponse> {
    try {
      const tx = await this.l1PublicClient.getTransaction({ hash });
      return {
        hash: tx.hash,
        gasLimit: tx.gas,
        maxFeePerGas: tx.maxFeePerGas,
        maxPriorityFeePerGas: tx.maxPriorityFeePerGas,
        nonce: tx.nonce,
      };
    } catch {
      return {
        hash,
        gasLimit: overrides?.gasLimit ?? 0n,
        maxFeePerGas: overrides?.maxFeePerGas,
        maxPriorityFeePerGas: overrides?.maxPriorityFeePerGas,
        nonce: overrides?.nonce ?? 0,
      };
    }
  }
}
