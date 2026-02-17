import {
  claimOnL2,
  getL1ToL2MessageStatus,
  OnChainMessageStatus as SdkOnChainMessageStatus,
} from "@consensys/linea-sdk-viem";
import {
  type Address,
  type Hex,
  type PublicClient,
  type WalletClient,
  decodeErrorResult,
  encodeFunctionData,
} from "viem";

import {
  type LineaGasFees,
  type MessageSent,
  type TransactionResponse,
  OnChainMessageStatus,
} from "../../../domain/types";
import { L2MessageServiceAbi } from "../abis/L2MessageServiceAbi";

import type { MessageProps } from "../../../domain/message/Message";
import type { ILineaGasProvider } from "../../../domain/ports/IGasProvider";
import type { IL2ContractClient } from "../../../domain/ports/IL2ContractClient";
import type { ClaimTransactionOverrides } from "../../../domain/ports/IMessageServiceContract";

const ZERO_ADDRESS = "0x0000000000000000000000000000000000000000";

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

export class ViemL2ContractClient implements IL2ContractClient {
  constructor(
    private readonly publicClient: PublicClient,
    private readonly walletClient: WalletClient,
    private readonly contractAddress: Address,
    private readonly gasProvider: ILineaGasProvider,
  ) {}

  public async getMessageStatus(params: {
    messageHash: Hex;
    messageBlockNumber?: number;
  }): Promise<OnChainMessageStatus> {
    const sdkStatus = await getL1ToL2MessageStatus(this.publicClient, {
      messageHash: params.messageHash,
      l2MessageServiceAddress: this.contractAddress,
    });

    return mapSdkStatus(sdkStatus);
  }

  public async claim(
    message: (MessageSent | MessageProps) & { feeRecipient?: string; messageBlockNumber?: number },
    opts: { claimViaAddress?: string; overrides?: ClaimTransactionOverrides } = {},
  ): Promise<TransactionResponse> {
    const hash = await claimOnL2(this.walletClient, {
      from: message.messageSender as Address,
      to: message.destination as Address,
      fee: BigInt(message.fee),
      value: BigInt(message.value),
      messageNonce: BigInt(message.messageNonce),
      calldata: message.calldata as Hex,
      feeRecipient: (message.feeRecipient ?? ZERO_ADDRESS) as Address,
      l2MessageServiceAddress: (opts.claimViaAddress ?? this.contractAddress) as Address,
      gas: opts.overrides?.gasLimit,
      nonce: opts.overrides?.nonce,
      maxFeePerGas: opts.overrides?.maxFeePerGas,
      maxPriorityFeePerGas: opts.overrides?.maxPriorityFeePerGas,
    });

    return this.buildTransactionResponse(hash, opts.overrides);
  }

  public encodeClaimMessageTransactionData(message: MessageProps & { feeRecipient?: string }): string {
    const { messageSender, destination, fee, value, calldata, messageNonce, feeRecipient } = message;
    const l2FeeRecipient = (feeRecipient ?? ZERO_ADDRESS) as Address;

    return encodeFunctionData({
      abi: L2MessageServiceAbi,
      functionName: "claimMessage",
      args: [
        messageSender as Address,
        destination as Address,
        BigInt(fee),
        BigInt(value),
        l2FeeRecipient,
        calldata as Hex,
        BigInt(messageNonce),
      ],
    });
  }

  public async estimateClaimGasFees(
    message: (MessageSent | MessageProps) & { feeRecipient?: string },
  ): Promise<LineaGasFees> {
    const transactionData = this.encodeClaimMessageTransactionData(message as MessageProps & { feeRecipient?: string });
    return this.gasProvider.getGasFees(transactionData);
  }

  public getSigner(): string | undefined {
    return this.walletClient.account?.address;
  }

  public getContractAddress(): string {
    return this.contractAddress;
  }

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  public async isRateLimitExceeded(_messageFee: bigint, _messageValue: bigint): Promise<boolean> {
    return false;
  }

  public async isRateLimitExceededError(transactionHash: string): Promise<boolean> {
    const parsedError = await this.parseTransactionError(transactionHash);
    return parsedError === "RateLimitExceeded";
  }

  public async parseTransactionError(transactionHash: string): Promise<string> {
    try {
      const tx = await this.publicClient.getTransaction({ hash: transactionHash as Hex });
      if (!tx) return "0x";

      const errorData = await this.publicClient.call({
        to: tx.to ?? undefined,
        data: tx.input,
        account: tx.from,
      });

      if (!errorData.data || errorData.data === "0x") return "0x";

      try {
        const decoded = decodeErrorResult({ abi: L2MessageServiceAbi, data: errorData.data });
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
    const tx = await this.publicClient.getTransaction({ hash: transactionHash as Hex });
    if (!tx) throw new Error(`Transaction not found: ${transactionHash}`);

    const bumpMultiplier = BigInt(100 + priceBumpPercent);
    const newMaxFeePerGas = ((tx.maxFeePerGas ?? 0n) * bumpMultiplier) / 100n;
    const newMaxPriorityFeePerGas = ((tx.maxPriorityFeePerGas ?? 0n) * bumpMultiplier) / 100n;

    const hash = await this.walletClient.sendTransaction({
      to: tx.to as Address,
      data: tx.input,
      value: tx.value,
      nonce: tx.nonce,
      gas: tx.gas,
      maxFeePerGas: newMaxFeePerGas,
      maxPriorityFeePerGas: newMaxPriorityFeePerGas,
      chain: this.walletClient.chain ?? null,
      account: this.walletClient.account!,
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
      const tx = await this.publicClient.getTransaction({ hash });
      return {
        hash: tx.hash,
        gasLimit: tx.gas,
        maxFeePerGas: tx.maxFeePerGas ?? undefined,
        maxPriorityFeePerGas: tx.maxPriorityFeePerGas ?? undefined,
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
