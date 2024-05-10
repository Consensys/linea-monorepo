import {
  Overrides,
  ContractTransactionResponse,
  JsonRpcProvider,
  TransactionReceipt,
  TransactionRequest,
  TransactionResponse,
  Signer,
} from "ethers";
import { L2MessageService, L2MessageService__factory } from "../typechain";
import { EIP1559GasProvider } from "../EIP1559GasProvider";
import { GasEstimationError } from "../../../core/errors/GasFeeErrors";
import { MessageProps } from "../../../core/entities/Message";
import { OnChainMessageStatus } from "../../../core/enums/MessageEnums";
import { IL2MessageServiceClient } from "../../../core/clients/blockchain/linea/IL2MessageServiceClient";
import { MESSAGE_SENT_EVENT_SIGNATURE, DEFAULT_RATE_LIMIT_MARGIN, ZERO_ADDRESS } from "../../../core/constants";
import { IL2MessageServiceLogClient } from "../../../core/clients/blockchain/linea/IL2MessageServiceLogClient";
import { MessageSent } from "../../../core/types/Events";
import { BaseError } from "../../../core/errors/Base";
import { isNull } from "../../../core/utils/shared";
import { SDKMode } from "../../../sdk/config";
import { formatMessageStatus } from "../../../core/utils/message";

export class L2MessageServiceClient extends EIP1559GasProvider implements IL2MessageServiceClient {
  private readonly contract: L2MessageService;

  /**
   * Initializes a new instance of the `L2MessageServiceClient`.
   *
   * @param {JsonRpcProvider} provider - The JSON RPC provider for interacting with the Ethereum network.
   * @param {string} contractAddress - The address of the L2 message service contract.
   * @param {IL2MessageServiceLogClient} l2MessageServiceLogClient - An instance of a class implementing the `IL2MessageServiceLogClient` interface for fetching events from the blockchain.
   * @param {SDKMode} mode - The mode in which the SDK is operating, e.g., `read-only` or `read-write`.
   * @param {Signer} [signer] - An optional Ethers.js signer object for signing transactions.
   * @param {bigint} [maxFeePerGas] - The maximum fee per gas that the user is willing to pay.
   * @param {number} [gasEstimationPercentile] - The percentile to sample from each block's effective priority fees.
   * @param {boolean} [enforceMaxGasFee] - Whether to enforce the maximum gas fee.
   */
  constructor(
    provider: JsonRpcProvider,
    private readonly contractAddress: string,
    private readonly l2MessageServiceLogClient: IL2MessageServiceLogClient,
    private readonly mode: SDKMode,
    private readonly signer?: Signer,
    maxFeePerGas?: bigint,
    gasEstimationPercentile?: number,
    enforceMaxGasFee?: boolean,
  ) {
    super(provider, maxFeePerGas, gasEstimationPercentile, enforceMaxGasFee);
    this.contract = this.getContract(this.contractAddress, this.signer);
  }

  /**
   * Retrieves message information by message hash.
   *
   * @param {string} messageHash - The hash of the message sent on L2.
   * @returns {Promise<MessageSent | null>} The message information or null if not found.
   */
  public async getMessageByMessageHash(messageHash: string): Promise<MessageSent | null> {
    const [event] = await this.l2MessageServiceLogClient.getMessageSentEvents({
      filters: { messageHash },
      fromBlock: 0,
      toBlock: "latest",
    });
    return event ?? null;
  }

  /**
   * Retrieves messages information by the transaction hash.
   *
   * @param {string} transactionHash - The hash of the `sendMessage` transaction on L2.
   * @returns {Promise<MessageSent[] | null>} An array of message information or null if not found.
   */
  public async getMessagesByTransactionHash(transactionHash: string): Promise<MessageSent[] | null> {
    const receipt = await this.provider.getTransactionReceipt(transactionHash);
    if (!receipt) {
      return null;
    }

    const messageSentEvents = await Promise.all(
      receipt.logs
        .filter((log) => log.address === this.contractAddress && log.topics[0] === MESSAGE_SENT_EVENT_SIGNATURE)
        .map((log) => this.contract.interface.parseLog(log))
        .filter((log) => !isNull(log))
        .map((log) => this.getMessageByMessageHash(log!.args._messageHash)),
    );
    return messageSentEvents.filter((log) => !isNull(log)) as MessageSent[];
  }

  /**
   * Retrieves the transaction receipt by message hash.
   *
   * @param {string} messageHash - The hash of the message sent on L2.
   * @returns {Promise<TransactionReceipt | null>} The `sendMessage` transaction receipt or null if not found.
   */
  public async getTransactionReceiptByMessageHash(messageHash: string): Promise<TransactionReceipt | null> {
    const [event] = await this.l2MessageServiceLogClient.getMessageSentEvents({
      filters: { messageHash },
      fromBlock: 0,
      toBlock: "latest",
    });

    if (!event) {
      return null;
    }

    const receipt = await this.provider.getTransactionReceipt(event.transactionHash);

    if (!receipt) {
      return null;
    }

    return receipt;
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
      return L2MessageService__factory.connect(contractAddress, this.provider);
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
  public async estimateClaimGas(
    message: MessageProps & { feeRecipient?: string },
    overrides: Overrides = {},
  ): Promise<bigint> {
    if (this.mode === "read-only") {
      throw new BaseError("'EstimateClaimGas' function not callable using readOnly mode.");
    }

    const { messageSender, destination, fee, value, calldata, messageNonce, feeRecipient } = message;
    const l2FeeRecipient = feeRecipient ?? ZERO_ADDRESS;
    try {
      return await this.contract.claimMessage.estimateGas(
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

    const transaction = await this.provider.getTransaction(transactionHash);

    if (!transaction) {
      throw new BaseError(`Transaction with hash ${transactionHash} not found.`);
    }

    let maxPriorityFeePerGas;
    let maxFeePerGas;

    if (!transaction.maxPriorityFeePerGas || !transaction.maxFeePerGas) {
      const txFees = await this.get1559Fees();
      maxPriorityFeePerGas = txFees.maxPriorityFeePerGas;
      maxFeePerGas = txFees.maxFeePerGas;
    } else {
      maxPriorityFeePerGas = (transaction.maxPriorityFeePerGas * (BigInt(priceBumpPercent) + 100n)) / 100n;
      maxFeePerGas = (transaction.maxFeePerGas * (BigInt(priceBumpPercent) + 100n)) / 100n;
      if (maxPriorityFeePerGas > this.maxFeePerGas) {
        maxPriorityFeePerGas = this.maxFeePerGas;
      }
      if (maxFeePerGas > this.maxFeePerGas) {
        maxFeePerGas = this.maxFeePerGas;
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
    return await this.provider.broadcastTransaction(signedTransaction);
  }

  /**
   * Checks if the rate limit for sending a message has been exceeded based on the provided message fee and value.
   *
   * @param {bigint} messageFee - The fee associated with the message.
   * @param {bigint} messageValue - The value being sent with the message.
   * @returns {Promise<boolean>} A promise that resolves to `true` if the rate limit has been exceeded, otherwise `false`.
   */
  public async isRateLimitExceeded(messageFee: bigint, messageValue: bigint): Promise<boolean> {
    const rateLimitInWei = await this.contract.limitInWei();
    const currentPeriodAmountInWei = await this.contract.currentPeriodAmountInWei();

    return (
      parseFloat((currentPeriodAmountInWei + BigInt(messageFee) + BigInt(messageValue)).toString()) >
      parseFloat(rateLimitInWei.toString()) * DEFAULT_RATE_LIMIT_MARGIN
    );
  }

  /**
   * Determines if a transaction failed due to exceeding the rate limit.
   *
   * @param {string} transactionHash - The hash of the transaction to check.
   * @returns {Promise<boolean>} A promise that resolves to `true` if the transaction failed due to a rate limit exceedance, otherwise `false`.
   */
  public async isRateLimitExceededError(transactionHash: string): Promise<boolean> {
    try {
      const tx = await this.provider.getTransaction(transactionHash);
      const errorEncodedData = await this.provider.call({
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
}
