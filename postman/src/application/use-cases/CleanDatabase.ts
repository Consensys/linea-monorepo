import type { ILogger } from "../../domain/ports/ILogger";
import type { IMessageRepository } from "../../domain/ports/IMessageRepository";

export class CleanDatabase {
  constructor(
    private readonly repository: IMessageRepository,
    private readonly logger: ILogger,
  ) {}

  public async databaseCleanerRoutine(msBeforeNowToDelete: number): Promise<void> {
    try {
      const affected = await this.repository.deleteMessages(msBeforeNowToDelete);
      this.logger.info("Database cleanup result: deleted %s rows", affected);
    } catch (e) {
      this.logger.error(e);
    }
  }
}
