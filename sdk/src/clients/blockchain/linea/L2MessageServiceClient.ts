import {
  Overrides,
  ContractTransactionResponse,
  JsonRpcProvider,
  TransactionReceipt,
  TransactionRequest,
  TransactionResponse,
  Signer,
  Block,
} from "ethers";
import { L2MessageService, L2MessageService__factory } from "../typechain";
import { GasEstimationError } from "../../../core/errors/GasFeeErrors";
import { MessageProps } from "../../../core/entities/Message";
import { OnChainMessageStatus } from "../../../core/enums/MessageEnums";
import { IL2MessageServiceClient } from "../../../core/clients/blockchain/linea/IL2MessageServiceClient";
import { ZERO_ADDRESS } from "../../../core/constants";
import { BaseError } from "../../../core/errors/Base";
import { SDKMode } from "../../../sdk/config";
import { formatMessageStatus } from "../../../core/utils/message";
import { IGasProvider, LineaGasFees } from "../../../core/clients/blockchain/IGasProvider";
import { IMessageRetriever } from "../../../core/clients/blockchain/IMessageRetriever";
import { MessageSent } from "../../../core/types/Events";
import { IL2ChainQuerier } from "../../../core/clients/blockchain/linea/IL2ChainQuerier";

export class L2MessageServiceClient
  implements
    IL2MessageServiceClient<Overrides, TransactionReceipt, TransactionResponse, ContractTransactionResponse, Signer>
{
  private readonly contract: L2MessageService;

  /**
   * Initializes a new instance of the `L2MessageServiceClient`.
   *
   * @param {IL2ChainQuerier} chainQuerier - The chain querier for interacting with the blockchain.
   * @param {string} contractAddress - The address of the L2 message service contract.
   * @param {IMessageRetriever<TransactionReceipt>} messageRetriever - An instance of a class implementing the `IMessageRetriever` interface for retrieving messages.
   * @param {IGasProvider<TransactionRequest>} gasFeeProvider - An instance of a class implementing the `IGasProvider` interface for providing gas fee estimates.
   * @param {SDKMode} mode - The mode in which the SDK is operating, e.g., `read-only` or `read-write`.
   * @param {Signer} [signer] - An optional Ethers.js signer object for signing transactions.
   */
  constructor(
    private readonly chainQuerier: IL2ChainQuerier<
      TransactionReceipt,
      Block,
      TransactionRequest,
      TransactionResponse,
      JsonRpcProvider
    >,
    private readonly contractAddress: string,
    private readonly messageRetriever: IMessageRetriever<TransactionReceipt>,
    private readonly gasFeeProvider: IGasProvider<TransactionRequest>,
    private readonly mode: SDKMode,
    private readonly signer?: Signer,
  ) {
    this.contract = this.getContract(this.contractAddress, this.signer);
  }

  public getSigner(): Signer | undefined {
    return this.signer;
  }

  public getContractAddress(): string {
    return this.contractAddress;
  }

  /**
   * Retrieves message information by message hash.
   *
   * @param {string} messageHash - The hash of the message sent on L2.
   * @returns {Promise<MessageSent | null>} The message information or null if not found.
   */
  public async getMessageByMessageHash(messageHash: string): Promise<MessageSent | null> {
    return this.messageRetriever.getMessageByMessageHash(messageHash);
  }

  /**
   * Retrieves messages information by the transaction hash.
   *
   * @param {string} transactionHash - The hash of the `sendMessage` transaction on L2.
   * @returns {Promise<MessageSent[] | null>} An array of message information or null if not found.
   */
  public async getMessagesByTransactionHash(transactionHash: string): Promise<MessageSent[] | null> {
    return this.messageRetriever.getMessagesByTransactionHash(transactionHash);
  }

  /**
   * Retrieves the transaction receipt by message hash.
   *
   * @param {string} messageHash - The hash of the message sent on L2.
   * @returns {Promise<TransactionReceipt | null>} The `sendMessage` transaction receipt or null if not found.
   */
  public async getTransactionReceiptByMessageHash(messageHash: string): Promise<TransactionReceipt | null> {
    return this.messageRetriever.getTransactionReceiptByMessageHash(messageHash);
  }

  /**
   * Retrieves the `L2MessageService` contract instance.
   *
   * @param {string} contractAddress - Address of the L2 contract.
   * @param {Signer} [signer] - L2 ethers signer instance.
   * @returns {L2MessageService} The `L2MessageService` contract instance.
   * @private
   */
  private getContract(contractAddress: string, signer?: Signer): L2MessageService {
    if (this.mode === "read-only") {
      return L2MessageService__factory.connect(contractAddress, this.chainQuerier.getProvider());
    }

    if (!signer) {
      throw new BaseError("Please provide a signer.");
    }

    return L2MessageService__factory.connect(contractAddress, signer);
  }

  /**
   * Retrieves the L1 message status on L2.
   *
   * @param {string} messageHash - The hash of the message sent on L1.
   * @param {Overrides} [overrides={}] - Ethers call overrides. Defaults to `{}` if not specified.
   * @returns {Promise<OnChainMessageStatus>} Message status (CLAIMED, CLAIMABLE, UNKNOWN).
   */
  public async getMessageStatus(messageHash: string, overrides: Overrides = {}): Promise<OnChainMessageStatus> {
    const status = await this.contract.inboxL1L2MessageStatus(messageHash, overrides);
    return formatMessageStatus(status);
  }

  /**
   * Estimates the gas required for the `claimMessage` transaction.
   *
   * @param {MessageProps & { feeRecipient?: string }} message - The message information object.
   * @param {Overrides} [overrides={}] - Ethers payable overrides. Defaults to `{}` if not specified.
   * @returns {Promise<bigint>} The `claimMessage` transaction gas estimation.
   */
  public async estimateClaimGasFees(
    message: MessageProps & { feeRecipient?: string },
    overrides: Overrides = {},
  ): Promise<LineaGasFees> {
    if (this.mode === "read-only") {
      throw new BaseError("'EstimateClaimGasFees' function not callable using readOnly mode.");
    }

    try {
      const transactionData = this.encodeClaimMessageTransactionData(message);

      return (await this.gasFeeProvider.getGasFees({
        // eslint-disable-next-line @typescript-eslint/no-non-null-asserted-optional-chain
        from: await this.signer?.getAddress()!,
        to: await this.contract.getAddress(),
        value: 0n,
        data: transactionData,
        ...overrides,
      })) as LineaGasFees;
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (e: any) {
      throw new GasEstimationError(e, message);
    }
  }

  /**
   * Claims the message on L2.
   *
   * @param {MessageProps & { feeRecipient?: string }} message - The message information object.
   * @param {Overrides} [overrides] - Ethers payable overrides. Defaults to `{}` if not specified.
   * @returns {Promise<ContractTransactionResponse>} The claimMessage transaction info.
   */
  public async claim(
    message: MessageProps & { feeRecipient?: string },
    overrides: Overrides = {},
  ): Promise<ContractTransactionResponse> {
    if (this.mode === "read-only") {
      throw new BaseError("'claim' function not callable using readOnly mode.");
    }

    const { messageSender, destination, fee, value, calldata, messageNonce, feeRecipient } = message;
    const l2FeeRecipient = feeRecipient ?? ZERO_ADDRESS;

    return await this.contract.claimMessage(
      messageSender,
      destination,
      fee,
      value,
      l2FeeRecipient,
      calldata,
      messageNonce,
      {
        ...overrides,
      },
    );
  }

  /**
   * Retries a specific transaction with a higher fee.
   *
   * @param {string} transactionHash - The hash of the transaction.
   * @param {number} [priceBumpPercent=10] - The percentage of price increase to retry the transaction. Defaults to `10` if not specified.
   * @returns {Promise<TransactionResponse>} The transaction information.
   */
  public async retryTransactionWithHigherFee(
    transactionHash: string,
    priceBumpPercent: number = 10,
  ): Promise<TransactionResponse> {
    if (!Number.isInteger(priceBumpPercent)) {
      throw new Error("'priceBumpPercent' must be an integer");
    }

    if (this.mode === "read-only") {
      throw new BaseError("'retryTransactionWithHigherFee' function not callable using readOnly mode.");
    }

    const transaction = await this.chainQuerier.getTransaction(transactionHash);

    if (!transaction) {
      throw new BaseError(`Transaction with hash ${transactionHash} not found.`);
    }

    let maxPriorityFeePerGas;
    let maxFeePerGas;

    if (!transaction.maxPriorityFeePerGas || !transaction.maxFeePerGas) {
      const txFees = await this.chainQuerier.getFees();
      maxPriorityFeePerGas = txFees.maxPriorityFeePerGas;
      maxFeePerGas = txFees.maxFeePerGas;
    } else {
      maxPriorityFeePerGas = (transaction.maxPriorityFeePerGas * (BigInt(priceBumpPercent) + 100n)) / 100n;
      maxFeePerGas = (transaction.maxFeePerGas * (BigInt(priceBumpPercent) + 100n)) / 100n;
      const maxFeePerGasFromConfig = this.gasFeeProvider.getMaxFeePerGas();
      if (maxPriorityFeePerGas > maxFeePerGasFromConfig) {
        maxPriorityFeePerGas = maxFeePerGasFromConfig;
      }
      if (maxFeePerGas > maxFeePerGasFromConfig) {
        maxFeePerGas = maxFeePerGasFromConfig;
      }
    }

    const updatedTransaction: TransactionRequest = {
      to: transaction.to,
      value: transaction.value,
      data: transaction.data,
      nonce: transaction.nonce,
      gasLimit: transaction.gasLimit,
      chainId: transaction.chainId,
      type: 2,
      maxPriorityFeePerGas,
      maxFeePerGas,
    };

    const signedTransaction = await this.signer!.signTransaction(updatedTransaction);
    return await this.chainQuerier.broadcastTransaction(signedTransaction);
  }

  /**
   * Checks if the rate limit for sending a message has been exceeded based on the provided message fee and value.
   *
   * @param {bigint} _messageFee - The fee associated with the message.
   * @param {bigint} _messageValue - The value being sent with the message.
   * @returns {Promise<boolean>} A promise that resolves to `true` if the rate limit has been exceeded, otherwise `false`.
   */
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  public async isRateLimitExceeded(_messageFee: bigint, _messageValue: bigint): Promise<boolean> {
    return false;
  }

  /**
   * Determines if a transaction failed due to exceeding the rate limit.
   *
   * @param {string} transactionHash - The hash of the transaction to check.
   * @returns {Promise<boolean>} A promise that resolves to `true` if the transaction failed due to a rate limit exceedance, otherwise `false`.
   */
  public async isRateLimitExceededError(transactionHash: string): Promise<boolean> {
    try {
      const tx = await this.chainQuerier.getTransaction(transactionHash);
      const errorEncodedData = await this.chainQuerier.ethCall({
        to: tx?.to,
        from: tx?.from,
        nonce: tx?.nonce,
        gasLimit: tx?.gasLimit,
        data: tx?.data,
        value: tx?.value,
        chainId: tx?.chainId,
        accessList: tx?.accessList,
        maxPriorityFeePerGas: tx?.maxPriorityFeePerGas,
        maxFeePerGas: tx?.maxFeePerGas,
      });
      const error = this.contract.interface.parseError(errorEncodedData);
      return error?.name === "RateLimitExceeded";
    } catch (e) {
      return false;
    }
  }

  /**
   * Encodes the transaction data for claiming a message.
   *
   * @param {MessageProps & { feeRecipient?: string }} message - The message properties including an optional fee recipient.
   * @param {string} message.messageSender - The address of the message sender.
   * @param {string} message.destination - The destination address of the message.
   * @param {bigint} message.fee - The fee associated with the message.
   * @param {bigint} message.value - The value associated with the message.
   * @param {string} message.calldata - The calldata associated with the message.
   * @param {bigint} message.messageNonce - The nonce of the message.
   * @param {string} [message.feeRecipient] - The optional address of the fee recipient. Defaults to ZERO_ADDRESS if not provided.
   * @returns {string} The encoded transaction data for claiming the message.
   */
  public encodeClaimMessageTransactionData(message: MessageProps & { feeRecipient?: string }): string {
    const { messageSender, destination, fee, value, calldata, messageNonce, feeRecipient } = message;
    const l2FeeRecipient = feeRecipient ?? ZERO_ADDRESS;

    return this.contract.interface.encodeFunctionData("claimMessage", [
      messageSender,
      destination,
      fee,
      value,
      l2FeeRecipient,
      calldata,
      messageNonce,
    ]);
  }
}
