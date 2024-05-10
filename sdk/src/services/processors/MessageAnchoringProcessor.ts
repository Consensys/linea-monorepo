import { Overrides, TransactionResponse, ContractTransactionResponse, TransactionReceipt } from "ethers";
import { IMessageAnchoringProcessor } from "../../core/services/processors/IMessageAnchoringProcessor";
import { IMessageRepository } from "../../core/persistence/IMessageRepository";
import { IChainQuerier } from "../../core/clients/blockchain/IChainQuerier";
import { Direction, MessageStatus, OnChainMessageStatus } from "../../core/enums/MessageEnums";
import { ILogger } from "../../core/utils/logging/ILogger";
import { DEFAULT_MAX_FETCH_MESSAGES_FROM_DB } from "../../core/constants";
import { IMessageServiceContract } from "../../core/services/contracts/IMessageServiceContract";
import { L1NetworkConfig, L2NetworkConfig } from "../../application/postman/app/config/config";

export class MessageAnchoringProcessor implements IMessageAnchoringProcessor {
  private readonly maxFetchMessagesFromDb: number;

  /**
   * Constructs a new instance of the `MessageAnchoringProcessor`.
   *
   * @param {IMessageRepository<unknown>} messageRepository - An instance of a class implementing the `IMessageRepository` interface, used for storing and retrieving message data.
   * @param {IMessageServiceContract<Overrides, TransactionReceipt, TransactionResponse, ContractTransactionResponse>} contractClient - An instance of a class implementing the `IMessageServiceContract` interface, used to interact with the blockchain contract.
   * @param {IChainQuerier<unknown>} chainQuerier - An instance of a class implementing the `IChainQuerier` interface, used to query blockchain data.
   * @param {L1NetworkConfig | L2NetworkConfig} config - Configuration settings for the network, including the maximum number of messages to fetch from the database for processing.
   * @param {Direction} direction - The direction of message flow (L1 to L2 or L2 to L1) that this processor is handling.
   * @param {string} originContractAddress - The contract address from which the messages originate.
   * @param {ILogger} logger - An instance of a class implementing the `ILogger` interface, used for logging messages.
   */
  constructor(
    private readonly messageRepository: IMessageRepository<unknown>,
    private readonly contractClient: IMessageServiceContract<
      Overrides,
      TransactionReceipt,
      TransactionResponse,
      ContractTransactionResponse
    >,
    private readonly chainQuerier: IChainQuerier<unknown>,
    config: L1NetworkConfig | L2NetworkConfig,
    private readonly direction: Direction,
    private readonly originContractAddress: string,
    private readonly logger: ILogger,
  ) {
    this.maxFetchMessagesFromDb = Math.max(
      config.listener.maxFetchMessagesFromDb ?? DEFAULT_MAX_FETCH_MESSAGES_FROM_DB,
      0,
    );
  }

  /**
   * Fetches a set number of messages from the database and updates their status based on the latest anchoring information from the blockchain.
   */
  public async getAndUpdateAnchoredMessageStatus() {
    try {
      const messages = await this.messageRepository.getNFirstMessageSent(
        this.direction,
        this.maxFetchMessagesFromDb,
        this.originContractAddress,
      );

      if (messages.length === this.maxFetchMessagesFromDb) {
        this.logger.warn(`Limit of messages sent to listen reached (%s).`, this.maxFetchMessagesFromDb);
      }

      if (messages.length === 0) {
        return;
      }

      const latestBlockNumber = await this.chainQuerier.getCurrentBlockNumber();

      for (const message of messages) {
        const messageStatus = await this.contractClient.getMessageStatus(message.messageHash, {
          blockTag: latestBlockNumber,
        });

        if (messageStatus === OnChainMessageStatus.CLAIMABLE) {
          message.edit({ status: MessageStatus.ANCHORED });
          this.logger.info("Message has been anchored: messageHash=%s", message.messageHash);
        }

        if (messageStatus === OnChainMessageStatus.CLAIMED) {
          message.edit({ status: MessageStatus.CLAIMED_SUCCESS });
          this.logger.info("Message has already been claimed: messageHash=%s", message.messageHash);
        }
      }

      await this.messageRepository.saveMessages(messages);
    } catch (e) {
      this.logger.error(e);
    }
  }
}
