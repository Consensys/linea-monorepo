import { describe, it } from "@jest/globals";
import { DataSource } from "typeorm";
import { SnakeNamingStrategy } from "typeorm-naming-strategies";
import { mockClear } from "jest-mock-extended";
import { PostmanServiceClient } from "../PostmanServiceClient";
import {
  TEST_CONTRACT_ADDRESS_1,
  TEST_CONTRACT_ADDRESS_2,
  TEST_L1_SIGNER_PRIVATE_KEY,
  TEST_L2_SIGNER_PRIVATE_KEY,
  TEST_RPC_URL,
} from "../../utils/testHelpers/constants";
import { L1SentEventListener, L2AnchoredEventListener } from "../listeners";
import { L2ClaimStatusWatcher, L2ClaimTxSender } from "../transactions";
import { LineaLogger } from "../../logger";
import { PostmanConfig } from "../utils/types";
import { MessageEntity } from "../entity/Message.entity";
import { InitialDatabaseSetup1685985945638 } from "../migrations/1685985945638-InitialDatabaseSetup";
import { AddNewColumns1687890326970 } from "../migrations/1687890326970-AddNewColumns";
import { UpdateStatusColumn1687890694496 } from "../migrations/1687890694496-UpdateStatusColumn";
import { RemoveUniqueConstraint1689084924789 } from "../migrations/1689084924789-RemoveUniqueConstraint";

jest.mock("ethers", () => {
  const allAutoMocked = jest.createMockFromModule("ethers");
  const actual = jest.requireActual("ethers");
  return {
    __esModules: true,
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    ...allAutoMocked,
    Wallet: actual.Wallet,
  };
});

const postmanServiceClientConfig: PostmanConfig = {
  l1Config: {
    rpcUrl: TEST_RPC_URL,
    messageServiceContractAddress: TEST_CONTRACT_ADDRESS_1,
    onlyEOATarget: true,
    listener: {
      pollingInterval: 4000,
      maxFetchMessagesFromDb: 1000,
      maxBlocksToFetchLogs: 1000,
    },
    claiming: {
      signerPrivateKey: TEST_L1_SIGNER_PRIVATE_KEY,
      messageSubmissionTimeout: 300000,
      maxNonceDiff: 10000,
      maxFeePerGas: 100000000000,
      gasEstimationPercentile: 50,
      profitMargin: 1.0,
      maxNumberOfRetries: 100,
      retryDelayInSeconds: 30,
    },
  },
  l2Config: {
    rpcUrl: TEST_RPC_URL,
    messageServiceContractAddress: TEST_CONTRACT_ADDRESS_2,
    onlyEOATarget: true,
    listener: {
      pollingInterval: 4000,
      maxFetchMessagesFromDb: 1000,
      maxBlocksToFetchLogs: 1000,
    },
    claiming: {
      signerPrivateKey: TEST_L2_SIGNER_PRIVATE_KEY,
      messageSubmissionTimeout: 300000,
      maxNonceDiff: 10000,
      maxFeePerGas: 100000000000,
      gasEstimationPercentile: 50,
      profitMargin: 1.0,
      maxNumberOfRetries: 100,
      retryDelayInSeconds: 30,
      maxClaimGasLimit: 100000,
    },
  },
  loggerOptions: {
    silent: true,
  },
  databaseOptions: {
    type: "postgres",
    host: "127.0.0.1",
    port: 5432,
    username: "postgres",
    password: "postgres",
    database: "db_name",
  },
};

describe("PostmanServiceClient", () => {
  let postmanServiceClient: PostmanServiceClient;
  let loggerSpy: unknown;

  beforeEach(() => {
    postmanServiceClient = new PostmanServiceClient(postmanServiceClientConfig);
    loggerSpy = jest.spyOn(LineaLogger.prototype, "info");
  });

  afterEach(() => {
    mockClear(loggerSpy);
  });

  describe("constructor", () => {
    it("should throw an error when at least one private key is invalid", () => {
      const postmanServiceClientConfigWithInvalidPrivateKey: PostmanConfig = {
        ...postmanServiceClientConfig,
        l1Config: {
          ...postmanServiceClientConfig.l1Config,
          claiming: {
            ...postmanServiceClientConfig.l1Config.claiming,
            signerPrivateKey: "",
          },
        },
      };

      expect(() => new PostmanServiceClient(postmanServiceClientConfigWithInvalidPrivateKey)).toThrow(
        new Error(
          "Something went wrong when trying to generate Wallet. Please check your private key and the provider url.",
        ),
      );
    });
  });

  describe("connectDatabase", () => {
    it("should initialize the db", async () => {
      const initializeSpy = jest.spyOn(DataSource.prototype, "initialize").mockResolvedValue(
        new DataSource({
          type: "postgres",
          host: "127.0.0.1",
          port: 5432,
          username: "postgres",
          password: "postgres",
          database: "db_name",
          entities: [MessageEntity],
          namingStrategy: new SnakeNamingStrategy(),
          migrations: [
            InitialDatabaseSetup1685985945638,
            AddNewColumns1687890326970,
            UpdateStatusColumn1687890694496,
            RemoveUniqueConstraint1689084924789,
          ],
          migrationsTableName: "migrations",
          logging: ["error"],
          migrationsRun: true,
        }),
      );
      await postmanServiceClient.connectDatabase();

      expect(initializeSpy).toHaveBeenCalledTimes(1);
    });
  });

  describe("startAllServices", () => {
    it("should start all postman services", async () => {
      jest.spyOn(L1SentEventListener.prototype, "start").mockImplementationOnce(jest.fn());
      jest.spyOn(L2AnchoredEventListener.prototype, "start").mockImplementationOnce(jest.fn());
      jest.spyOn(L2ClaimTxSender.prototype, "start").mockImplementationOnce(jest.fn());
      jest.spyOn(L2ClaimStatusWatcher.prototype, "start").mockImplementationOnce(jest.fn());
      const loggerSpy = jest.spyOn(LineaLogger.prototype, "info");

      postmanServiceClient.startAllServices();

      expect(loggerSpy).toHaveBeenCalledTimes(1);
      expect(loggerSpy).toHaveBeenCalledWith("All listeners and message deliverers have been started.");
    });

    it("should stop all postman services", async () => {
      jest.spyOn(L1SentEventListener.prototype, "stop").mockImplementationOnce(jest.fn());
      jest.spyOn(L2AnchoredEventListener.prototype, "stop").mockImplementationOnce(jest.fn());
      jest.spyOn(L2ClaimTxSender.prototype, "stop").mockImplementationOnce(jest.fn());
      jest.spyOn(L2ClaimStatusWatcher.prototype, "stop").mockImplementationOnce(jest.fn());
      const loggerSpy = jest.spyOn(LineaLogger.prototype, "info");

      postmanServiceClient.stopAllServices();

      expect(loggerSpy).toHaveBeenCalledTimes(1);
      expect(loggerSpy).toHaveBeenCalledWith("All listeners and message deliverers have been stopped.");
    });
  });
});
