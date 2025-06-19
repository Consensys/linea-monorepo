import {
  ContractTransactionResponse,
  ErrorDescription,
  Overrides,
  Signer,
  TransactionReceipt,
  TransactionResponse,
} from "ethers";
import { OnChainMessageStatus } from "@consensys/linea-sdk";
import { IL1ToL2MessageStatusService } from "../core/services/IL1ToL2MessageStatusService";
import { IL2MessageServiceClient } from "../core/clients/blockchain/linea/IL2MessageServiceClient";

export class L1ToL2MessageStatusService implements IL1ToL2MessageStatusService<Overrides> {
  /**
   * Constructs a new instance of the `L1ToL2MessageStatusService`.
   *
   * @param {IL2MessageServiceClient} l2MessageServiceClient - An instance of a class implementing the `IL2MessageServiceClient` interface, used to interact with the L2MessageService.
   */
  constructor(
    private readonly l2MessageServiceClient: IL2MessageServiceClient<
      Overrides,
      TransactionReceipt,
      TransactionResponse,
      ContractTransactionResponse,
      Signer,
      ErrorDescription
    >,
  ) {}

  /**
   * Retrieves the status of a message on the L2MessageService by its hash.
   *
   * @param {string} messageHash - The hash of the message to check.
   * @param {Overrides} [overrides={}] - Optional transaction overrides.
   * @returns {Promise<OnChainMessageStatus>} A promise that resolves to the status of the message.
   */
  public async getMessageStatus(messageHash: string, overrides: Overrides = {}): Promise<OnChainMessageStatus> {
    const messageStatus = await this.l2MessageServiceClient.getMessageStatus(messageHash, overrides);
    return messageStatus;
  }
}
