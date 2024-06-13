import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import { DatabaseCleaningPoller } from "../DatabaseCleaningPoller";
import { IDatabaseCleaner } from "../../../core/persistence/IDatabaseCleaner";
import { TestLogger } from "../../../utils/testing/helpers";

describe("TestDatabaseCleaningPoller", () => {
  let testDatabaseCleaningPoller: DatabaseCleaningPoller;
  const databaseCleanerMock = mock<IDatabaseCleaner>();
  const logger = new TestLogger(DatabaseCleaningPoller.name);

  beforeEach(() => {});

  afterEach(() => {
    jest.resetAllMocks();
  });

  describe("start", () => {
    it("Should return log as warning if not enabled", async () => {
      testDatabaseCleaningPoller = new DatabaseCleaningPoller(databaseCleanerMock, logger);
      const loggerWarnSpy = jest.spyOn(logger, "warn");

      await testDatabaseCleaningPoller.start();

      expect(loggerWarnSpy).toHaveBeenCalledTimes(1);
      expect(loggerWarnSpy).toHaveBeenCalledWith("%s is disabled", DatabaseCleaningPoller.name);
    });

    it("Should return and log as warning if it has been started", async () => {
      testDatabaseCleaningPoller = new DatabaseCleaningPoller(databaseCleanerMock, logger, { enabled: true });
      const loggerWarnSpy = jest.spyOn(logger, "warn");

      testDatabaseCleaningPoller.start();
      await testDatabaseCleaningPoller.start();

      expect(loggerWarnSpy).toHaveBeenCalledTimes(1);
      expect(loggerWarnSpy).toHaveBeenCalledWith("%s has already started.", DatabaseCleaningPoller.name);
    });

    it("Should call databaseCleanerRoutine and log as info if it started successfully", async () => {
      testDatabaseCleaningPoller = new DatabaseCleaningPoller(databaseCleanerMock, logger, { enabled: true });
      const databaseCleanerMockSpy = jest.spyOn(databaseCleanerMock, "databaseCleanerRoutine");
      const loggerInfoSpy = jest.spyOn(logger, "info");

      testDatabaseCleaningPoller.start();

      expect(databaseCleanerMockSpy).toHaveBeenCalled();
      expect(loggerInfoSpy).toHaveBeenCalledTimes(1);
      expect(loggerInfoSpy).toHaveBeenCalledWith("Starting %s...", DatabaseCleaningPoller.name);
    });
  });

  describe("stop", () => {
    it("Should return and log as info if it stopped successfully", async () => {
      testDatabaseCleaningPoller = new DatabaseCleaningPoller(databaseCleanerMock, logger, { enabled: true });
      const loggerInfoSpy = jest.spyOn(logger, "info");

      testDatabaseCleaningPoller.stop();

      expect(loggerInfoSpy).toHaveBeenCalledTimes(2);
      expect(loggerInfoSpy).toHaveBeenNthCalledWith(1, "Stopping %s...", DatabaseCleaningPoller.name);
      expect(loggerInfoSpy).toHaveBeenNthCalledWith(2, "%s stopped.", DatabaseCleaningPoller.name);
    });
  });
});
