import {
  BigNumber,
  BytesLike,
  CallOverrides,
  ContractTransaction,
  Event,
  EventFilter,
  PayableOverrides,
  Signer,
  ethers,
} from "ethers";
import {
  BlockTag,
  JsonRpcProvider,
  TransactionReceipt,
  TransactionRequest,
  TransactionResponse,
} from "@ethersproject/providers";
import { Message, SDKMode, ParsedEvent } from "../utils/types";
import { EventParser } from "./EventParser";
import { L2MessageService, L2MessageService__factory } from "../../typechain";
import { EIP1559GasProvider } from "./EIP1559GasProvider";
import { GasEstimationError } from "../utils/errors";
import { formatMessageStatus } from "../utils/helpers";
import { OnChainMessageStatus } from "../utils/enum";
import { mapMessageSentEventOrLogToMessage } from "../utils/mappers";
import { MESSAGE_SENT_EVENT_SIGNATURE } from "../utils/constants";

export class L2MessageServiceContract extends EIP1559GasProvider {
  public readonly contract: L2MessageService;

  constructor(
    public readonly provider: JsonRpcProvider,
    public readonly contractAddress: string,
    private readonly mode: SDKMode,
    public readonly signer?: Signer,
    maxFeePerGas?: number,
    gasEstimationPercentile?: number,
  ) {
    super(provider, maxFeePerGas, gasEstimationPercentile);
    this.contract = this.getContract(this.contractAddress, this.signer);
  }

  public getContractAbi() {
    return L2MessageService__factory.abi;
  }

  public async getMessageByMessageHash(messageHash: string): Promise<Message | null> {
    const messageSentFilter = this.contract.filters.MessageSent(null, null, null, null, null, null, messageHash);
    const [event] = await this.contract.queryFilter(messageSentFilter, 0, "latest");

    if (!event) {
      return null;
    }

    return mapMessageSentEventOrLogToMessage(event);
  }

  public async getMessagesByTransactionHash(transactionHash: string): Promise<Message[] | null> {
    const receipt = await this.provider.getTransactionReceipt(transactionHash);

    if (!receipt) {
      return null;
    }

    return receipt.logs
      .filter((log) => log.address === this.contractAddress && log.topics[0] === MESSAGE_SENT_EVENT_SIGNATURE)
      .map((log) => this.contract.interface.parseLog(log))
      .map(mapMessageSentEventOrLogToMessage);
  }

  public async getTransactionReceiptByMessageHash(messageHash: string): Promise<TransactionReceipt | null> {
    const messageSentFilter = this.contract.filters.MessageSent(null, null, null, null, null, null, messageHash);
    const [event] = await this.contract.queryFilter(messageSentFilter, 0, "latest");

    if (!event) {
      return null;
    }

    const receipt = await this.provider.getTransactionReceipt(event.transactionHash);

    if (!receipt) {
      return null;
    }

    return receipt;
  }

  public getContract(contractAddress: string, signer?: Signer): L2MessageService {
    if (this.mode === "read-only") {
      return L2MessageService__factory.connect(contractAddress, this.provider);
    }

    if (!signer) {
      throw new Error("Please provide a signer.");
    }

    return L2MessageService__factory.connect(contractAddress, signer);
  }

  public async getCurrentNonce(accountAddress?: string): Promise<number> {
    if (this.mode === "read-only") {
      if (!accountAddress) {
        throw new Error("Please provider an account address.");
      }
      return await this.provider.getTransactionCount(accountAddress);
    }

    if (!this.signer) {
      throw new Error("Please provide a signer.");
    }

    return await this.provider.getTransactionCount(await this.signer.getAddress());
  }

  public async getCurrentBlockNumber(): Promise<number> {
    return await this.provider.getBlockNumber();
  }

  public async getEvents<TEventFilter extends EventFilter>(
    eventFilter: TEventFilter,
    fromBlock?: number,
    toBlock?: BlockTag,
    fromBlockLogIndex?: number,
  ): Promise<ParsedEvent<Event>[]> {
    let events = await this.contract.queryFilter(eventFilter, fromBlock, toBlock);
    events = events.filter((event) => {
      if (typeof fromBlockLogIndex === "undefined" || typeof fromBlock === "undefined") {
        return true;
      }
      if (event.blockNumber === fromBlock && event.logIndex < fromBlockLogIndex) {
        return false;
      }
      return true;
    });
    if (events.length > 0) {
      return EventParser.filterAndParseEvents(events);
    }

    return [];
  }

  public async getMessageStatus(messageHash: BytesLike, overrides: CallOverrides = {}): Promise<OnChainMessageStatus> {
    const status = await this.contract.inboxL1L2MessageStatus(messageHash, overrides);
    return formatMessageStatus(status);
  }

  public async estimateClaimGas(
    message: Message & { feeRecipient?: string },
    overrides: PayableOverrides = {},
  ): Promise<BigNumber> {
    if (this.mode === "read-only") {
      throw new Error("'EstimateClaimGas' function not callable using readOnly mode.");
    }

    const { messageSender, destination, fee, value, calldata, messageNonce, feeRecipient } = message;
    const l2FeeRecipient = feeRecipient || ethers.constants.AddressZero;
    try {
      return await this.contract.estimateGas.claimMessage(
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
    } catch (e) {
      throw new GasEstimationError(e, message);
    }
  }

  public async claim(
    message: Message & { feeRecipient?: string },
    overrides: PayableOverrides = {},
  ): Promise<ContractTransaction> {
    if (this.mode === "read-only") {
      throw new Error("'claim' function not callable using readOnly mode.");
    }

    const { messageSender, destination, fee, value, calldata, messageNonce, feeRecipient } = message;
    const l2FeeRecipient = feeRecipient || ethers.constants.AddressZero;

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

  public async retryTransactionWithHigherFee(
    transactionHash?: string,
    overrides: { maxPriorityFeePerGas?: BigNumber; maxFeePerGas?: BigNumber } = {},
  ): Promise<TransactionResponse | null> {
    if (this.mode === "read-only") {
      throw new Error("'retryTransactionWithHigherFee' function not callable using readOnly mode.");
    }

    if (!this.signer) {
      throw new Error("Please provide a signer.");
    }

    if (!transactionHash) return null;
    const transaction = await this.provider.getTransaction(transactionHash);
    const updatedTransaction: TransactionRequest = {
      to: transaction.to,
      value: transaction.value,
      data: transaction.data,
      nonce: transaction.nonce,
      gasLimit: transaction.gasLimit,
      chainId: transaction.chainId,
      type: 2,
      ...overrides,
    };
    const signedTransaction = await this.signer.signTransaction(updatedTransaction);
    return await this.provider.sendTransaction(signedTransaction);
  }
}
