import { describe, it, expect, beforeEach, afterEach } from "@jest/globals";
import { MockProxy } from "jest-mock-extended";
import { DataSource, EntityManager } from "typeorm";

import { buildTestPostmanOptions } from "../../../../utils/testing/fixtures";
import {
  mockApplication,
  mockDataSource,
  mockLogger,
  mockMessageMetricsUpdater,
  mockMetricsService,
  mockPoller,
  mockSponsorshipMetricsUpdater,
  mockTransactionMetricsUpdater,
} from "../../../../utils/testing/mocks";
import { PostmanApp } from "../PostmanApp";
import { PostmanServices } from "../PostmanContainer";

const mockBuildPostmanServices = jest.fn();
const mockGetConfig = jest.fn();
const mockDBCreate = jest.fn();
const mockCreatePostmanApi = jest.fn();

jest.mock("@consensys/linea-shared-utils", () => ({
  WinstonLogger: jest.fn().mockImplementation(() => ({
    info: jest.fn(),
    warn: jest.fn(),
    error: jest.fn(),
    debug: jest.fn(),
    name: "test",
  })),
}));

jest.mock("../config/utils", () => ({
  getConfig: (...args: unknown[]) => mockGetConfig(...args),
}));

jest.mock("../PostmanContainer", () => ({
  buildPostmanServices: (...args: unknown[]) => mockBuildPostmanServices(...args),
}));

jest.mock("../../../../infrastructure/persistence/dataSource", () => ({
  DB: { create: (...args: unknown[]) => mockDBCreate(...args) },
}));

jest.mock("../../../../infrastructure/api/PostmanApi", () => ({
  createPostmanApi: (...args: unknown[]) => mockCreatePostmanApi(...args),
}));

jest.mock("../../../../infrastructure/metrics/PostmanMetricsService", () => ({
  PostmanMetricsService: jest.fn().mockImplementation(() => mockMetricsService()),
}));

jest.mock("../../../../infrastructure/metrics/MessageMetricsUpdater", () => ({
  MessageMetricsUpdater: jest.fn().mockImplementation(() => mockMessageMetricsUpdater()),
}));

jest.mock("../../../../infrastructure/metrics/SponsorshipMetricsUpdater", () => ({
  SponsorshipMetricsUpdater: jest.fn().mockImplementation(() => mockSponsorshipMetricsUpdater()),
}));

jest.mock("../../../../infrastructure/metrics/TransactionMetricsUpdater", () => ({
  TransactionMetricsUpdater: jest.fn().mockImplementation(() => mockTransactionMetricsUpdater()),
}));

describe("PostmanApp", () => {
  let mockDs: MockProxy<DataSource>;
  let services: PostmanServices;

  beforeEach(() => {
    mockDs = mockDataSource();
    mockDs.manager = {} as EntityManager;
    mockDs.initialize = jest.fn().mockResolvedValue(undefined);
    mockDs.destroy = jest.fn().mockResolvedValue(undefined);
    mockDs.subscribers = [];

    services = {
      l1ToL2App: { start: jest.fn(), stop: jest.fn() },
      l2ToL1App: { start: jest.fn(), stop: jest.fn() },
      databaseCleaningPoller: mockPoller(),
    };

    mockGetConfig.mockReturnValue({
      l1Config: {},
      l2Config: {},
      l1L2AutoClaimEnabled: true,
      l2L1AutoClaimEnabled: true,
      databaseOptions: { type: "postgres" },
      databaseCleanerConfig: { enabled: false },
      apiConfig: { port: 3000 },
    });
    mockDBCreate.mockReturnValue(mockDs);
    mockBuildPostmanServices.mockResolvedValue(services);
    mockCreatePostmanApi.mockReturnValue(mockApplication());
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  it("should construct without errors", () => {
    const options = buildTestPostmanOptions();
    expect(() => new PostmanApp(options)).not.toThrow();
  });

  it("should call getConfig during construction", () => {
    const options = buildTestPostmanOptions();
    new PostmanApp(options);
    expect(mockGetConfig).toHaveBeenCalledWith(options);
  });

  describe("start", () => {
    it("should initialize the database", async () => {
      const app = new PostmanApp(buildTestPostmanOptions());
      await app.start();
      expect(mockDs.initialize).toHaveBeenCalled();
    });

    it("should build postman services", async () => {
      const app = new PostmanApp(buildTestPostmanOptions());
      await app.start();
      expect(mockBuildPostmanServices).toHaveBeenCalled();
    });

    it("should create and start the API", async () => {
      const mockApi = mockApplication();
      mockCreatePostmanApi.mockReturnValue(mockApi);

      const app = new PostmanApp(buildTestPostmanOptions());
      await app.start();

      expect(mockCreatePostmanApi).toHaveBeenCalled();
      expect(mockApi.start).toHaveBeenCalled();
    });

    it("should start l1ToL2App when present", async () => {
      const app = new PostmanApp(buildTestPostmanOptions());
      await app.start();
      expect(services.l1ToL2App!.start).toHaveBeenCalled();
    });

    it("should start l2ToL1App when present", async () => {
      const app = new PostmanApp(buildTestPostmanOptions());
      await app.start();
      expect(services.l2ToL1App!.start).toHaveBeenCalled();
    });

    it("should start database cleaning poller when present", async () => {
      const app = new PostmanApp(buildTestPostmanOptions());
      await app.start();
      expect(services.databaseCleaningPoller!.start).toHaveBeenCalled();
    });

    it("should not throw when optional services are undefined", async () => {
      mockBuildPostmanServices.mockResolvedValue({});

      const app = new PostmanApp(buildTestPostmanOptions());
      await expect(app.start()).resolves.not.toThrow();
    });
  });

  describe("stop", () => {
    it("should stop all services and destroy the database", async () => {
      const mockApi = mockApplication();
      mockCreatePostmanApi.mockReturnValue(mockApi);

      const app = new PostmanApp(buildTestPostmanOptions());
      await app.start();
      await app.stop();

      expect(services.l1ToL2App!.stop).toHaveBeenCalled();
      expect(services.l2ToL1App!.stop).toHaveBeenCalled();
      expect(services.databaseCleaningPoller!.stop).toHaveBeenCalled();
      expect(mockApi.stop).toHaveBeenCalled();
      expect(mockDs.destroy).toHaveBeenCalled();
    });

    it("should not throw when stop is called before start", async () => {
      const app = new PostmanApp(buildTestPostmanOptions());
      await expect(app.stop()).resolves.not.toThrow();
    });
  });
});
