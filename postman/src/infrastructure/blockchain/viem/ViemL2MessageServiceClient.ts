import {
  claimOnL2,
  getL1ToL2MessageStatus,
  getTransactionReceiptByMessageHash as sdkGetTransactionReceiptByMessageHash,
} from "@consensys/linea-sdk-viem";
import { type Hex, type PublicClient, type WalletClient, decodeErrorResult, encodeFunctionData } from "viem";
import { getContractEvents } from "viem/actions";

import { mapViemReceiptToCoreReceipt } from "./mappers";
import { ILineaGasProvider, LineaGasFees } from "../../../core/clients/blockchain/IGasProvider";
import { IL2MessageServiceClient } from "../../../core/clients/blockchain/linea/IL2MessageServiceClient";
import { ZERO_ADDRESS } from "../../../core/constants/blockchain";
import { MessageProps } from "../../../core/entities/Message";
import { OnChainMessageStatus } from "../../../core/enums";
import {
  ErrorDescription,
  MessageSent,
  Overrides,
  TransactionReceipt,
  TransactionSubmission,
} from "../../../core/types";
import { L2MessageServiceAbi } from "../abis/L2MessageServiceAbi";

export class ViemL2MessageServiceClient implements IL2MessageServiceClient {
  constructor(
    private readonly publicClient: PublicClient,
    private readonly walletClient: WalletClient,
    private readonly contractAddress: string,
    private readonly gasProvider: ILineaGasProvider,
    private readonly signerAddress: string,
  ) {}

  public getContractAddress(): string {
    return this.contractAddress;
  }

  public async getMessageStatus(params: {
    messageHash: string;
    messageBlockNumber?: number;
    overrides?: Overrides;
  }): Promise<OnChainMessageStatus> {
    return getL1ToL2MessageStatus(this.publicClient, {
      messageHash: params.messageHash as Hex,
      l2MessageServiceAddress: this.contractAddress as Hex,
    });
  }

  public encodeClaimMessageTransactionData(message: MessageProps & { feeRecipient?: string }): string {
    const feeRecipient = message.feeRecipient ?? ZERO_ADDRESS;
    return encodeFunctionData({
      abi: L2MessageServiceAbi,
      functionName: "claimMessage",
      args: [
        message.messageSender as Hex,
        message.destination as Hex,
        message.fee,
        message.value,
        feeRecipient as Hex,
        message.calldata as Hex,
        message.messageNonce,
      ],
    });
  }

  public async estimateClaimGasFees(
    message: (MessageSent | MessageProps) & { feeRecipient?: string },
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    _opts: { claimViaAddress?: string; overrides?: Overrides } = {},
  ): Promise<LineaGasFees> {
    const transactionData = this.encodeClaimMessageTransactionDataFromMessage(message);

    return this.gasProvider.getGasFees({
      from: this.signerAddress,
      to: this.contractAddress,
      value: 0n,
      data: transactionData,
    });
  }

  public async claim(
    message: (MessageSent | MessageProps) & { feeRecipient?: string; messageBlockNumber?: number },
    opts: { claimViaAddress?: string; overrides?: Overrides } = {},
  ): Promise<TransactionSubmission> {
    const from = message.messageSender;
    const to = message.destination;
    const calldata = message.calldata;
    const nonce = message.messageNonce;
    const feeRecipient = (message.feeRecipient ?? ZERO_ADDRESS) as Hex;
    const contractAddress = (opts.claimViaAddress ?? this.contractAddress) as Hex;

    const txHash = await claimOnL2(this.walletClient, {
      from: from as Hex,
      to: to as Hex,
      fee: message.fee,
      value: message.value,
      messageNonce: nonce,
      calldata: calldata as Hex,
      feeRecipient,
      l2MessageServiceAddress: contractAddress,
      nonce: opts.overrides?.nonce,
      gas: opts.overrides?.gasLimit,
      maxFeePerGas: opts.overrides?.maxFeePerGas,
      maxPriorityFeePerGas: opts.overrides?.maxPriorityFeePerGas,
    });

    return {
      hash: txHash,
      nonce: opts.overrides?.nonce ?? 0,
      gasLimit: opts.overrides?.gasLimit ?? 0n,
      maxFeePerGas: opts.overrides?.maxFeePerGas,
      maxPriorityFeePerGas: opts.overrides?.maxPriorityFeePerGas,
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
      const fees = await this.publicClient.estimateFeesPerGas();
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

  public async getMessageByMessageHash(messageHash: string): Promise<MessageSent | null> {
    const events = await getContractEvents(this.publicClient, {
      address: this.contractAddress as Hex,
      abi: L2MessageServiceAbi,
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

  // L2 has no rate limit
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  public async isRateLimitExceeded(_messageFee: bigint, _messageValue: bigint): Promise<boolean> {
    return false;
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
        const errData = (callError as { data?: `0x${string}` }).data;
        if (errData) errorEncodedData = errData;
      }

      if (errorEncodedData === "0x") return errorEncodedData;

      const decoded = decodeErrorResult({
        abi: L2MessageServiceAbi,
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

  private encodeClaimMessageTransactionDataFromMessage(
    message: (MessageSent | MessageProps) & { feeRecipient?: string },
  ): string {
    const from = message.messageSender;
    const to = message.destination;
    const calldata = message.calldata;
    const nonce = message.messageNonce;
    const feeRecipient = message.feeRecipient ?? ZERO_ADDRESS;

    return encodeFunctionData({
      abi: L2MessageServiceAbi,
      functionName: "claimMessage",
      args: [from as Hex, to as Hex, message.fee, message.value, feeRecipient as Hex, calldata as Hex, nonce],
    });
  }
}
