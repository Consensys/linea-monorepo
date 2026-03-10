import { ILogger } from "@consensys/linea-shared-utils";

import { IDatabaseCleaner } from "../../core/persistence/IDatabaseCleaner";
import { IMessageRepository } from "../../core/persistence/IMessageRepository";

export class DatabaseCleaner implements IDatabaseCleaner {
  constructor(
    private readonly messageRepository: IMessageRepository,
    private readonly logger: ILogger,
  ) {}

  /**
   * Executes the database cleanup routine to delete messages older than a specified duration.
   *
   * @param {number} msBeforeNowToDelete - The duration in milliseconds before the current time. Messages older than this duration will be deleted.
   */
  public async databaseCleanerRoutine(msBeforeNowToDelete: number) {
    try {
      const affected = await this.messageRepository.deleteMessages(msBeforeNowToDelete);
      this.logger.info("Database cleanup result: deleted %s rows", affected);
    } catch (e) {
      this.logger.error(e);
    }
  }
}
