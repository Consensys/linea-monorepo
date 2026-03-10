import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";

import { Direction } from "../../../core/enums";
import { IPoller } from "../../../core/services/pollers/IPoller";
import { testL2NetworkConfig } from "../../../utils/testing/constants";
import { TestLogger } from "../../../utils/testing/helpers";
import { IntervalPoller, IProcessable } from "../IntervalPoller";

describe("IntervalPoller", () => {
  let poller: IPoller;
  const processorMock = mock<IProcessable>();
  const loggerName = "TestIntervalPoller";
  const logger = new TestLogger(loggerName);

  beforeEach(() => {
    poller = new IntervalPoller(
      processorMock,
      {
        direction: Direction.L1_TO_L2,
        pollingInterval: testL2NetworkConfig.listener.pollingInterval,
      },
      logger,
    );
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  describe("start", () => {
    it("Should return and log as warning if it has been started", async () => {
      const loggerWarnSpy = jest.spyOn(logger, "warn");

      poller.start();
      await poller.start();

      expect(loggerWarnSpy).toHaveBeenCalledTimes(1);
      expect(loggerWarnSpy).toHaveBeenCalledWith("Poller has already started.", { name: loggerName });
    });

    it("Should call process and log as info if it started successfully", async () => {
      const processorSpy = jest.spyOn(processorMock, "process");
      const loggerInfoSpy = jest.spyOn(logger, "info");

      poller.start();

      expect(processorSpy).toHaveBeenCalled();
      expect(loggerInfoSpy).toHaveBeenCalledTimes(1);
      expect(loggerInfoSpy).toHaveBeenCalledWith("Starting poller.", {
        direction: Direction.L1_TO_L2,
        name: loggerName,
      });
    });

    it("Should omit direction from logs when not configured", async () => {
      const pollerNoDirection = new IntervalPoller(
        processorMock,
        { pollingInterval: testL2NetworkConfig.listener.pollingInterval },
        logger,
      );
      const loggerInfoSpy = jest.spyOn(logger, "info");

      pollerNoDirection.start();

      expect(loggerInfoSpy).toHaveBeenCalledWith("Starting poller.", {
        name: loggerName,
      });
    });
  });

  describe("stop", () => {
    it("Should return and log as info if it stopped successfully", async () => {
      const loggerInfoSpy = jest.spyOn(logger, "info");

      poller.stop();

      expect(loggerInfoSpy).toHaveBeenCalledTimes(2);
      expect(loggerInfoSpy).toHaveBeenNthCalledWith(1, "Stopping poller.", {
        direction: Direction.L1_TO_L2,
        name: loggerName,
      });
      expect(loggerInfoSpy).toHaveBeenNthCalledWith(2, "Poller stopped.", {
        direction: Direction.L1_TO_L2,
        name: loggerName,
      });
    });
  });
});
