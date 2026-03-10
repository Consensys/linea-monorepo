import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";

import { DEFAULT_DB_CLEANING_INTERVAL, DEFAULT_DB_DAYS_BEFORE_NOW_TO_DELETE } from "../../../core/constants";
import { IDatabaseCleaner } from "../../../core/persistence/IDatabaseCleaner";
import { TestLogger } from "../../../utils/testing/helpers";
import { DatabaseCleaningPoller } from "../DatabaseCleaningPoller";

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
      testDatabaseCleaningPoller = new DatabaseCleaningPoller(databaseCleanerMock, logger, {
        enabled: false,
        cleaningInterval: DEFAULT_DB_CLEANING_INTERVAL,
        daysBeforeNowToDelete: DEFAULT_DB_DAYS_BEFORE_NOW_TO_DELETE,
      });
      const loggerWarnSpy = jest.spyOn(logger, "warn");

      await testDatabaseCleaningPoller.start();

      expect(loggerWarnSpy).toHaveBeenCalledTimes(1);
      expect(loggerWarnSpy).toHaveBeenCalledWith("Poller is disabled.", { name: DatabaseCleaningPoller.name });
    });

    it("Should return and log as warning if it has been started", async () => {
      testDatabaseCleaningPoller = new DatabaseCleaningPoller(databaseCleanerMock, logger, {
        enabled: true,
        cleaningInterval: DEFAULT_DB_CLEANING_INTERVAL,
        daysBeforeNowToDelete: DEFAULT_DB_DAYS_BEFORE_NOW_TO_DELETE,
      });
      const loggerWarnSpy = jest.spyOn(logger, "warn");

      testDatabaseCleaningPoller.start();
      await testDatabaseCleaningPoller.start();

      expect(loggerWarnSpy).toHaveBeenCalledTimes(1);
      expect(loggerWarnSpy).toHaveBeenCalledWith("Poller has already started.", { name: DatabaseCleaningPoller.name });
    });

    it("Should call databaseCleanerRoutine and log as info if it started successfully", async () => {
      testDatabaseCleaningPoller = new DatabaseCleaningPoller(databaseCleanerMock, logger, {
        enabled: true,
        cleaningInterval: DEFAULT_DB_CLEANING_INTERVAL,
        daysBeforeNowToDelete: DEFAULT_DB_DAYS_BEFORE_NOW_TO_DELETE,
      });
      const databaseCleanerMockSpy = jest.spyOn(databaseCleanerMock, "databaseCleanerRoutine");
      const loggerInfoSpy = jest.spyOn(logger, "info");

      testDatabaseCleaningPoller.start();

      expect(databaseCleanerMockSpy).toHaveBeenCalled();
      expect(loggerInfoSpy).toHaveBeenCalledTimes(1);
      expect(loggerInfoSpy).toHaveBeenCalledWith("Starting poller.", { name: DatabaseCleaningPoller.name });
    });
  });

  describe("stop", () => {
    it("Should return and log as info if it stopped successfully", async () => {
      testDatabaseCleaningPoller = new DatabaseCleaningPoller(databaseCleanerMock, logger, {
        enabled: true,
        cleaningInterval: DEFAULT_DB_CLEANING_INTERVAL,
        daysBeforeNowToDelete: DEFAULT_DB_DAYS_BEFORE_NOW_TO_DELETE,
      });
      const loggerInfoSpy = jest.spyOn(logger, "info");

      testDatabaseCleaningPoller.stop();

      expect(loggerInfoSpy).toHaveBeenCalledTimes(2);
      expect(loggerInfoSpy).toHaveBeenNthCalledWith(1, "Stopping poller.", { name: DatabaseCleaningPoller.name });
      expect(loggerInfoSpy).toHaveBeenNthCalledWith(2, "Poller stopped.", { name: DatabaseCleaningPoller.name });
    });
  });
});
