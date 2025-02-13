import { describe, it, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import { TestLogger } from "../../../utils/testing/helpers";
import { Direction, MessageSent } from "@consensys/linea-sdk";
import { MessageStatus } from "../../../core/enums";
import {
  TEST_ADDRESS_1,
  TEST_ADDRESS_2,
  testL1NetworkConfig,
  testMessageSentEvent,
  testMessageSentEventWithCallData,
} from "../../../utils/testing/constants";
import { IProvider } from "../../../core/clients/blockchain/IProvider";
import { MessageSentEventProcessorConfig } from "../../../core/services/processors/IMessageSentEventProcessor";
import { MessageSentEventProcessor } from "../MessageSentEventProcessor";
import { ILineaRollupLogClient } from "../../../core/clients/blockchain/ethereum/ILineaRollupLogClient";
import { MessageFactory } from "../../../core/entities/MessageFactory";
import {
  Block,
  ContractTransactionResponse,
  Interface,
  JsonRpcProvider,
  TransactionReceipt,
  TransactionRequest,
  TransactionResponse,
} from "ethers";
import { EthereumMessageDBService } from "../../persistence/EthereumMessageDBService";
import { IMessageDBService } from "../../../core/persistence/IMessageDBService";
import { ILogger } from "../../../core/utils/logging/ILogger";
import { IL2MessageServiceLogClient } from "../../../core/clients/blockchain/linea/IL2MessageServiceLogClient";

class TestMessageSentEventProcessor extends MessageSentEventProcessor {
  constructor(
    databaseService: IMessageDBService<ContractTransactionResponse>,
    logClient: ILineaRollupLogClient | IL2MessageServiceLogClient,
    provider: IProvider<TransactionReceipt, Block, TransactionRequest, TransactionResponse, JsonRpcProvider>,
    public readonly config: MessageSentEventProcessorConfig,
    logger: ILogger,
  ) {
    super(databaseService, logClient, provider, config, logger);
  }

  public shouldProcessMessage(
    message: MessageSent,
    messageHash: string,
    filters?: { calldataFilter?: string; calldataFunctionInterface?: string },
  ): boolean {
    return super.shouldProcessMessage(message, messageHash, filters);
  }
}

describe("TestMessageSentEventProcessor", () => {
  let messageSentEventProcessor: TestMessageSentEventProcessor;
  const databaseService = mock<EthereumMessageDBService>();
  const l1LogClientMock = mock<ILineaRollupLogClient>();
  const provider =
    mock<IProvider<TransactionReceipt, Block, TransactionRequest, TransactionResponse, JsonRpcProvider>>();
  const logger = new TestLogger(MessageSentEventProcessor.name);

  beforeEach(() => {
    messageSentEventProcessor = new TestMessageSentEventProcessor(
      databaseService,
      l1LogClientMock,
      provider,
      {
        direction: Direction.L1_TO_L2,
        maxBlocksToFetchLogs: testL1NetworkConfig.listener.maxBlocksToFetchLogs,
        blockConfirmation: testL1NetworkConfig.listener.blockConfirmation,
        isEOAEnabled: testL1NetworkConfig.isEOAEnabled,
        isCalldataEnabled: testL1NetworkConfig.isCalldataEnabled,
      },
      logger,
    );
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  describe("process", () => {
    it("Should insert message with status as sent into repository if the message is not excluded", async () => {
      const loggerInfoSpy = jest.spyOn(logger, "info");
      const messageRepositoryInsertSpy = jest.spyOn(databaseService, "insertMessage");
      jest.spyOn(provider, "getBlockNumber").mockResolvedValue(100);
      jest.spyOn(l1LogClientMock, "getMessageSentEvents").mockResolvedValue([testMessageSentEvent]);
      const expectedMessageToInsert = MessageFactory.createMessage({
        ...testMessageSentEvent,
        sentBlockNumber: testMessageSentEvent.blockNumber,
        direction: Direction.L1_TO_L2,
        status: MessageStatus.SENT,
        claimNumberOfRetry: 0,
      });

      await messageSentEventProcessor.process(200, 0);

      expect(loggerInfoSpy).toHaveBeenCalledTimes(3);
      expect(messageRepositoryInsertSpy).toHaveBeenCalledTimes(1);
      expect(messageRepositoryInsertSpy).toHaveBeenCalledWith(expectedMessageToInsert);
    });

    it("Should insert message with status as excluded into repository if the message is excluded", async () => {
      messageSentEventProcessor = new TestMessageSentEventProcessor(
        databaseService,
        l1LogClientMock,
        provider,
        {
          direction: Direction.L1_TO_L2,
          maxBlocksToFetchLogs: testL1NetworkConfig.listener.maxBlocksToFetchLogs,
          blockConfirmation: testL1NetworkConfig.listener.blockConfirmation,
          isEOAEnabled: !testL1NetworkConfig.isEOAEnabled,
          isCalldataEnabled: testL1NetworkConfig.isCalldataEnabled,
        },
        logger,
      );
      const loggerInfoSpy = jest.spyOn(logger, "info");
      const messageRepositoryInsertSpy = jest.spyOn(databaseService, "insertMessage");
      jest.spyOn(provider, "getBlockNumber").mockResolvedValue(100);
      jest.spyOn(l1LogClientMock, "getMessageSentEvents").mockResolvedValue([testMessageSentEvent]);
      const expectedMessageToInsert = MessageFactory.createMessage({
        ...testMessageSentEvent,
        sentBlockNumber: testMessageSentEvent.blockNumber,
        direction: Direction.L1_TO_L2,
        status: MessageStatus.EXCLUDED,
        claimNumberOfRetry: 0,
      });

      await messageSentEventProcessor.process(0, 0);

      expect(loggerInfoSpy).toHaveBeenCalledTimes(3);
      expect(messageRepositoryInsertSpy).toHaveBeenCalledTimes(1);
      expect(messageRepositoryInsertSpy).toHaveBeenCalledWith(expectedMessageToInsert);
    });

    it("Should insert message with status as excluded into repository if the message is excluded becuase of events filters", async () => {
      messageSentEventProcessor = new TestMessageSentEventProcessor(
        databaseService,
        l1LogClientMock,
        provider,
        {
          direction: Direction.L1_TO_L2,
          maxBlocksToFetchLogs: testL1NetworkConfig.listener.maxBlocksToFetchLogs,
          blockConfirmation: testL1NetworkConfig.listener.blockConfirmation,
          isEOAEnabled: testL1NetworkConfig.isEOAEnabled,
          isCalldataEnabled: true,
          eventFilters: {
            fromAddressFilter: TEST_ADDRESS_1,
            toAddressFilter: TEST_ADDRESS_2,
            calldataFilter: `calldata.funcSignature == "0x6463fb2a" and calldata.params.messageNumber == 85805`,
            calldataFunctionInterface:
              "function claimMessageWithProof((bytes32[] proof,uint256 messageNumber,uint32 leafIndex,address from,address to,uint256 fee,uint256 value,address feeRecipient,bytes32 merkleRoot,bytes data) params)",
          },
        },
        logger,
      );
      const loggerInfoSpy = jest.spyOn(logger, "info");
      const messageRepositoryInsertSpy = jest.spyOn(databaseService, "insertMessage");
      jest.spyOn(provider, "getBlockNumber").mockResolvedValue(100);
      jest.spyOn(l1LogClientMock, "getMessageSentEvents").mockResolvedValue([
        testMessageSentEvent,
        {
          ...testMessageSentEvent,
          calldata:
            "0x6463fb2a000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000001400000000000000000000000000000000000000000000000000000000000014f2c0000000000000000000000000000000000000000000000000000000000000008000000000000000000000000c59d8de7f984abc4913f0177bfb7bbdafac41fa6000000000000000000000000c59d8de7f984abc4913f0177bfb7bbdafac41fa6000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000038d7ea4c680000000000000000000000000000000000000000000000000000000000000000000d835920764c09f5b2f8105900efd4bd88344f958eb3425436d27d4689da595e80000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000541e47c68e1d235e121f188e5083acf352df62e8d730c6813910a3e1e51f0d0a3e973e11619685da115b8cb81b850a4278a3efd28870281a3f0932e2c32f98af9b4c11951957c6f8f642c4af61cd6b24640fec6dc7fc607ee8206a99e92410d306189ee58a32992fa49a5f07feccd1895e3b73923f87f5fc4d07961a3b94d0848e58769b32a1beaf1ea27375a44095a0d1fb664ce2dd358e7fcbfb78c26a193440000000000000000000000000000000000000000000000000000000000000000",
        },
      ]);
      const expectedMessage1ToInsert = MessageFactory.createMessage({
        ...testMessageSentEvent,
        sentBlockNumber: testMessageSentEvent.blockNumber,
        direction: Direction.L1_TO_L2,
        status: MessageStatus.SENT,
        claimNumberOfRetry: 0,
      });

      const expectedMessage2ToInsert = MessageFactory.createMessage({
        ...{
          ...testMessageSentEvent,
          calldata:
            "0x6463fb2a000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000001400000000000000000000000000000000000000000000000000000000000014f2c0000000000000000000000000000000000000000000000000000000000000008000000000000000000000000c59d8de7f984abc4913f0177bfb7bbdafac41fa6000000000000000000000000c59d8de7f984abc4913f0177bfb7bbdafac41fa6000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000038d7ea4c680000000000000000000000000000000000000000000000000000000000000000000d835920764c09f5b2f8105900efd4bd88344f958eb3425436d27d4689da595e80000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000541e47c68e1d235e121f188e5083acf352df62e8d730c6813910a3e1e51f0d0a3e973e11619685da115b8cb81b850a4278a3efd28870281a3f0932e2c32f98af9b4c11951957c6f8f642c4af61cd6b24640fec6dc7fc607ee8206a99e92410d306189ee58a32992fa49a5f07feccd1895e3b73923f87f5fc4d07961a3b94d0848e58769b32a1beaf1ea27375a44095a0d1fb664ce2dd358e7fcbfb78c26a193440000000000000000000000000000000000000000000000000000000000000000",
        },
        sentBlockNumber: testMessageSentEvent.blockNumber,
        direction: Direction.L1_TO_L2,
        status: MessageStatus.EXCLUDED,
        claimNumberOfRetry: 0,
      });

      await messageSentEventProcessor.process(0, 0);

      expect(loggerInfoSpy).toHaveBeenCalledTimes(3);
      expect(messageRepositoryInsertSpy).toHaveBeenCalledTimes(2);
      expect(messageRepositoryInsertSpy).toHaveBeenNthCalledWith(1, expectedMessage1ToInsert);
      expect(messageRepositoryInsertSpy).toHaveBeenNthCalledWith(2, expectedMessage2ToInsert);
    });

    it("Should insert message with calldata with status as sent into repository if calldata is enabled", async () => {
      messageSentEventProcessor = new TestMessageSentEventProcessor(
        databaseService,
        l1LogClientMock,
        provider,
        {
          direction: Direction.L1_TO_L2,
          maxBlocksToFetchLogs: testL1NetworkConfig.listener.maxBlocksToFetchLogs,
          blockConfirmation: testL1NetworkConfig.listener.blockConfirmation,
          isEOAEnabled: testL1NetworkConfig.isEOAEnabled,
          isCalldataEnabled: !testL1NetworkConfig.isCalldataEnabled,
        },
        logger,
      );
      const loggerInfoSpy = jest.spyOn(logger, "info");
      const messageRepositoryInsertSpy = jest.spyOn(databaseService, "insertMessage");
      jest.spyOn(provider, "getBlockNumber").mockResolvedValue(100);
      jest.spyOn(l1LogClientMock, "getMessageSentEvents").mockResolvedValue([testMessageSentEventWithCallData]);
      const expectedMessageToInsert = MessageFactory.createMessage({
        ...testMessageSentEventWithCallData,
        sentBlockNumber: testMessageSentEventWithCallData.blockNumber,
        direction: Direction.L1_TO_L2,
        status: MessageStatus.SENT,
        claimNumberOfRetry: 0,
      });

      await messageSentEventProcessor.process(0, 0);

      expect(loggerInfoSpy).toHaveBeenCalledTimes(3);
      expect(messageRepositoryInsertSpy).toHaveBeenCalledTimes(1);
      expect(messageRepositoryInsertSpy).toHaveBeenCalledWith(expectedMessageToInsert);
    });
  });

  describe("shouldProcessMessage", () => {
    const funcFragment =
      "function claimMessageWithProof((bytes32[] proof,uint256 messageNumber,uint32 leafIndex,address from,address to,uint256 fee,uint256 value,address feeRecipient,bytes32 merkleRoot,bytes data) params)";

    const encodedCalldata = new Interface([funcFragment]).encodeFunctionData(funcFragment, [
      [
        [
          "0x41e47c68e1d235e121f188e5083acf352df62e8d730c6813910a3e1e51f0d0a3",
          "0xe973e11619685da115b8cb81b850a4278a3efd28870281a3f0932e2c32f98af9",
          "0xb4c11951957c6f8f642c4af61cd6b24640fec6dc7fc607ee8206a99e92410d30",
          "0x6189ee58a32992fa49a5f07feccd1895e3b73923f87f5fc4d07961a3b94d0848",
          "0xe58769b32a1beaf1ea27375a44095a0d1fb664ce2dd358e7fcbfb78c26a19344",
        ],
        85804n,
        8n,
        "0xc59d8de7f984AbC4913f0177bfb7BBdaFaC41fA6",
        "0xc59d8de7f984AbC4913f0177bfb7BBdaFaC41fA6",
        0n,
        1000000000000000n,
        "0x0000000000000000000000000000000000000000",
        "0xd835920764c09f5b2f8105900efd4bd88344f958eb3425436d27d4689da595e8",
        "0x",
      ],
    ]);

    it("Should return true if calldata is empty and EOA is enabled", () => {
      const result = messageSentEventProcessor.shouldProcessMessage(
        testMessageSentEvent,
        testMessageSentEvent.messageHash,
      );
      expect(result).toBeTruthy();
    });

    it("Should return false if calldata is empty and EOA is disabled", () => {
      messageSentEventProcessor.config.isEOAEnabled = false;
      const result = messageSentEventProcessor.shouldProcessMessage(
        testMessageSentEvent,
        testMessageSentEvent.messageHash,
      );
      expect(result).toBeFalsy();
    });

    it("Should return true if calldata is not empty and calldata option is enabled", () => {
      messageSentEventProcessor.config.isCalldataEnabled = true;
      const result = messageSentEventProcessor.shouldProcessMessage(
        { ...testMessageSentEvent, calldata: "0x1111111111" },
        testMessageSentEvent.messageHash,
      );
      expect(result).toBeTruthy();
    });

    it("Should return false if calldata is not empty and calldata option is disabled", () => {
      const result = messageSentEventProcessor.shouldProcessMessage(
        { ...testMessageSentEvent, calldata: "0x1111111111" },
        testMessageSentEvent.messageHash,
      );
      expect(result).toBeFalsy();
    });

    it("Should return false if EOA and calldata options are disabled", () => {
      messageSentEventProcessor.config.isEOAEnabled = false;
      const result = messageSentEventProcessor.shouldProcessMessage(
        { ...testMessageSentEvent, calldata: "0x1111111111" },
        testMessageSentEvent.messageHash,
      );
      expect(result).toBeFalsy();
    });

    it("Should return false if event filter criteria is not correctly formatted", () => {
      messageSentEventProcessor.config.isCalldataEnabled = true;

      const result = messageSentEventProcessor.shouldProcessMessage(
        {
          ...testMessageSentEvent,
          calldata: encodedCalldata,
        },
        testMessageSentEvent.messageHash,
        {
          calldataFilter: `from == ${TEST_ADDRESS_1} and to == "${TEST_ADDRESS_2}" and calldata.funcSignature == "0x6463fb2a" and calldata.params.messageNumber == 85805`,
          calldataFunctionInterface: funcFragment,
        },
      );
      expect(result).toBeFalsy();
    });

    it("Should return false if event filter criteria is false", () => {
      messageSentEventProcessor.config.isCalldataEnabled = true;

      const result = messageSentEventProcessor.shouldProcessMessage(
        {
          ...testMessageSentEvent,
          calldata: encodedCalldata,
        },
        testMessageSentEvent.messageHash,
        {
          calldataFilter: `from == "${TEST_ADDRESS_1}" and to == "${TEST_ADDRESS_2}" and calldata.funcSignature == "0x6463fb2a" and calldata.params.messageNumber == 85805`,
          calldataFunctionInterface: funcFragment,
        },
      );
      expect(result).toBeFalsy();
    });

    it("Should return true if event filter criteria is true", () => {
      messageSentEventProcessor.config.isCalldataEnabled = true;

      const result = messageSentEventProcessor.shouldProcessMessage(
        {
          ...testMessageSentEvent,
          calldata: encodedCalldata,
        },
        testMessageSentEvent.messageHash,
        {
          calldataFilter: `calldata.funcSignature == "0x6463fb2a" and calldata.params.messageNumber == 85804`,
          calldataFunctionInterface: funcFragment,
        },
      );
      expect(result).toBeTruthy();
    });
  });
});
