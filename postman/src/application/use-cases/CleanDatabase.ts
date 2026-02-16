import type { ILogger } from "../../domain/ports/ILogger";
import type { IMessageDBService } from "../../domain/ports/IMessageDBService";

export class CleanDatabase {
  constructor(
    private readonly databaseService: IMessageDBService,
    private readonly logger: ILogger,
  ) {}

  public async databaseCleanerRoutine(msBeforeNowToDelete: number): Promise<void> {
    try {
      const affected = await this.databaseService.deleteMessages(msBeforeNowToDelete);
      this.logger.info("Database cleanup result: deleted %s rows", affected);
    } catch (e) {
      this.logger.error(e);
    }
  }
}
