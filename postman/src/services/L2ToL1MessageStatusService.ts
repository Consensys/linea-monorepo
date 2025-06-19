import {
  ContractTransactionResponse,
  ErrorDescription,
  Overrides,
  TransactionReceipt,
  TransactionResponse,
} from "ethers";
import { OnChainMessageStatus } from "@consensys/linea-sdk";
import { ILineaRollupClient } from "../core/clients/blockchain/ethereum/ILineaRollupClient";
import {
  IL2ToL1MessageStatusService,
  L2ToL1MessageStatusServiceConfig,
} from "../core/services/IL2ToL1MessageStatusService";
import { IMessageDBService } from "../core/persistence/IMessageDBService";
import { getStartingBlocksForLogsFetching } from "../utils/logs";

export class L2ToL1MessageStatusService implements IL2ToL1MessageStatusService<Overrides> {
  /**
   * Constructs a new instance of the `L2ToL1MessageStatusService`.
   *
   * @param {ILineaRollupClient} lineaRollupClient - An instance of a class implementing the `ILineaRollupClient` interface, used to interact with the LineaRollup.
   */
  constructor(
    private readonly lineaRollupClient: ILineaRollupClient<
      Overrides,
      TransactionReceipt,
      TransactionResponse,
      ContractTransactionResponse,
      ErrorDescription
    >,
    private readonly databaseService: IMessageDBService<ContractTransactionResponse>,
    private readonly config: L2ToL1MessageStatusServiceConfig,
  ) {}

  /**
   * Retrieves the status of a message on the Linea rollup by its hash.
   *
   * @param {string} messageHash - The hash of the message to check.
   * @param {Overrides} [overrides={}] - Optional transaction overrides.
   * @returns {Promise<OnChainMessageStatus>} A promise that resolves to the status of the message.
   */
  public async getMessageStatus(messageHash: string, overrides: Overrides = {}): Promise<OnChainMessageStatus> {
    const { l1LogsFromBlock, l2LogsFromBlock } = await getStartingBlocksForLogsFetching(
      this.config,
      this.databaseService,
    );

    const messageStatus = await this.lineaRollupClient.getMessageStatus(messageHash, {
      overrides,
      l1LogsFromBlock,
      l2LogsFromBlock,
    });

    return messageStatus;
  }
}
