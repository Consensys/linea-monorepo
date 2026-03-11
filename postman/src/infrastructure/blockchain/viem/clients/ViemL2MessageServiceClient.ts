import {
  claimOnL2,
  getL1ToL2MessageStatus,
  getTransactionReceiptByMessageHash as sdkGetTransactionReceiptByMessageHash,
} from "@consensys/linea-sdk-viem";
import { type PublicClient, type WalletClient, decodeErrorResult, encodeFunctionData } from "viem";
import { getContractEvents } from "viem/actions";

import { ILineaGasProvider, LineaGasFees } from "../../../../core/clients/blockchain/IGasProvider";
import { IL2MessageServiceClient } from "../../../../core/clients/blockchain/linea/IL2MessageServiceClient";
import { ZERO_ADDRESS } from "../../../../core/constants/blockchain";
import { MessageProps } from "../../../../core/entities/Message";
import { OnChainMessageStatus } from "../../../../core/enums";
import {
  Address,
  Hash,
  Hex,
  ErrorDescription,
  MessageSent,
  Overrides,
  TransactionReceipt,
  TransactionSubmission,
} from "../../../../core/types";
import { L2MessageServiceAbi } from "../../abis/L2MessageServiceAbi";
import { mapViemReceiptToCoreReceipt } from "../mappers";

export class ViemL2MessageServiceClient implements IL2MessageServiceClient {
  constructor(
    private readonly publicClient: PublicClient,
    private readonly walletClient: WalletClient,
    private readonly contractAddress: Address,
    private readonly gasProvider: ILineaGasProvider,
    private readonly signerAddress: Address,
  ) {}

  public getContractAddress(): Address {
    return this.contractAddress;
  }

  public async getMessageStatus(params: {
    messageHash: Hash;
    messageBlockNumber?: number;
  }): Promise<OnChainMessageStatus> {
    return getL1ToL2MessageStatus(this.publicClient, {
      messageHash: params.messageHash,
      l2MessageServiceAddress: this.contractAddress,
    });
  }

  public encodeClaimMessageTransactionData(message: MessageProps & { feeRecipient?: Address }): Hex {
    const feeRecipient = message.feeRecipient ?? ZERO_ADDRESS;
    return encodeFunctionData({
      abi: L2MessageServiceAbi,
      functionName: "claimMessage",
      args: [
        message.messageSender,
        message.destination,
        message.fee,
        message.value,
        feeRecipient,
        message.calldata,
        message.messageNonce,
      ],
    });
  }

  public async estimateClaimGasFees(
    message: (MessageSent | MessageProps) & { feeRecipient?: Address },
    opts: { claimViaAddress?: Address; overrides?: Overrides } = {},
  ): Promise<LineaGasFees> {
    const transactionData = this.encodeClaimMessageTransactionDataFromMessage(message);
    const contractAddress = opts.claimViaAddress ?? this.contractAddress;

    return this.gasProvider.getGasFees({
      from: this.signerAddress,
      to: contractAddress,
      value: 0n,
      data: transactionData,
    });
  }

  public async claim(
    message: (MessageSent | MessageProps) & { feeRecipient?: Address; messageBlockNumber?: number },
    opts: { claimViaAddress?: Address; overrides?: Overrides } = {},
  ): Promise<TransactionSubmission> {
    const from = message.messageSender;
    const to = message.destination;
    const calldata = message.calldata;
    const nonce = message.messageNonce;
    const feeRecipient = message.feeRecipient ?? ZERO_ADDRESS;
    const contractAddress = opts.claimViaAddress ?? this.contractAddress;

    const txHash = await claimOnL2(this.walletClient, {
      from,
      to,
      fee: message.fee,
      value: message.value,
      messageNonce: nonce,
      calldata,
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

  public async getMessageByMessageHash(messageHash: Hash): Promise<MessageSent | null> {
    const events = await getContractEvents(this.publicClient, {
      address: this.contractAddress,
      abi: L2MessageServiceAbi,
      eventName: "MessageSent",
      args: { _messageHash: messageHash },
      fromBlock: "earliest",
      toBlock: "latest",
    });
    if (!events.length) return null;
    const e = events[0];
    const args = e.args as {
      _messageHash: Hash;
      _from: Address;
      _to: Address;
      _fee: bigint;
      _value: bigint;
      _nonce: bigint;
      _calldata: Hex;
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

  public async getTransactionReceiptByMessageHash(messageHash: Hash): Promise<TransactionReceipt | null> {
    try {
      const receipt = await sdkGetTransactionReceiptByMessageHash(this.publicClient, {
        messageHash,
        messageServiceAddress: this.contractAddress,
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

  public async isRateLimitExceededError(transactionHash: Hash): Promise<boolean> {
    const parsedError = await this.parseTransactionError(transactionHash);
    if (typeof parsedError === "string") return false;
    return parsedError.name === "RateLimitExceeded";
  }

  public async parseTransactionError(transactionHash: Hash): Promise<ErrorDescription | string> {
    let errorEncodedData: Hash = "0x";
    try {
      const tx = await this.publicClient.getTransaction({ hash: transactionHash });
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
    message: (MessageSent | MessageProps) & { feeRecipient?: Address },
  ): Hex {
    const from = message.messageSender;
    const to = message.destination;
    const calldata = message.calldata;
    const nonce = message.messageNonce;
    const feeRecipient = message.feeRecipient ?? ZERO_ADDRESS;

    return encodeFunctionData({
      abi: L2MessageServiceAbi,
      functionName: "claimMessage",
      args: [from, to, message.fee, message.value, feeRecipient, calldata, nonce],
    });
  }
}
