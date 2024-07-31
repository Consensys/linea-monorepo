import { Block, JsonRpcProvider, TransactionReceipt, TransactionRequest, TransactionResponse } from "ethers";
import { MessageSent } from "../../../core/types/Events";
import { MESSAGE_SENT_EVENT_SIGNATURE } from "../../../core/constants";
import { isNull } from "../../../core/utils/shared";
import { L2MessageService, L2MessageService__factory } from "../typechain";
import { IMessageRetriever } from "../../../core/clients/blockchain/IMessageRetriever";
import { IL2MessageServiceLogClient } from "../../../core/clients/blockchain/linea/IL2MessageServiceLogClient";
import { IL2ChainQuerier } from "../../../core/clients/blockchain/linea/IL2ChainQuerier";

export class L2MessageServiceMessageRetriever implements IMessageRetriever<TransactionReceipt> {
  private readonly contract: L2MessageService;

  /**
   * Creates an instance of `L2MessageServiceMessageRetriever`.
   *
   * @param {IL2ChainQuerier} chainQuerier - The chain querier for interacting with the blockchain.
   * @param {IL2MessageServiceLogClient} l2MessageServiceLogClient - An instance of a class implementing the `IL2MessageServiceLogClient` interface for fetching events from the blockchain.
   * @param {string} contractAddress - The address of the L2 message service contract.
   */
  constructor(
    private readonly chainQuerier: IL2ChainQuerier<
      TransactionReceipt,
      Block,
      TransactionRequest,
      TransactionResponse,
      JsonRpcProvider
    >,
    private readonly l2MessageServiceLogClient: IL2MessageServiceLogClient,
    private readonly contractAddress: string,
  ) {
    this.contract = L2MessageService__factory.connect(contractAddress, this.chainQuerier.getProvider());
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
    const receipt = await this.chainQuerier.getTransactionReceipt(transactionHash);
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

    const receipt = await this.chainQuerier.getTransactionReceipt(event.transactionHash);

    if (!receipt) {
      return null;
    }

    return receipt;
  }
}
