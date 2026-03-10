import { type PublicClient } from "viem";
import { type LoggerOptions } from "winston";

import { L1NetworkConfig, L2NetworkConfig } from "./config/config";
import { ILineaRollupClient } from "../../../core/clients/blockchain/ethereum/ILineaRollupClient";
import { IEthereumGasProvider } from "../../../core/clients/blockchain/IGasProvider";
import { IProvider } from "../../../core/clients/blockchain/IProvider";
import { IL2MessageServiceLogClient } from "../../../core/clients/blockchain/linea/IL2MessageServiceLogClient";
import { Direction } from "../../../core/enums";
import { IErrorParser } from "../../../core/errors/IErrorParser";
import { ISponsorshipMetricsUpdater, ITransactionMetricsUpdater } from "../../../core/metrics";
import { IMessageRepository } from "../../../core/persistence/IMessageRepository";
import { ICalldataDecoder } from "../../../core/services/ICalldataDecoder";
import { IPoller } from "../../../core/services/pollers/IPoller";
import { InlineNonceManager } from "../../../infrastructure/blockchain/viem/signers/InlineNonceManager";
import { EthereumTransactionValidationService } from "../../../services/EthereumTransactionValidationService";
import {
  MessageAnchoringPoller,
  MessageClaimingPoller,
  MessagePersistingPoller,
  MessageSentEventPoller,
} from "../../../services/pollers";
import {
  MessageAnchoringProcessor,
  MessageClaimingPersister,
  MessageClaimingProcessor,
  MessageSentEventProcessor,
} from "../../../services/processors";
import { PostmanWinstonLogger } from "../../../utils/PostmanWinstonLogger";

export type L2ToL1Deps = {
  l2LogClient: IL2MessageServiceLogClient;
  l2Provider: IProvider;
  lineaRollupClient: ILineaRollupClient;
  l1Provider: IProvider;
  l1PublicClient: PublicClient;
  l1SignerAddress: `0x${string}`;
  messageRepository: IMessageRepository;
  l1GasProvider: IEthereumGasProvider;
  calldataDecoder: ICalldataDecoder;
  sponsorshipMetricsUpdater: ISponsorshipMetricsUpdater;
  transactionMetricsUpdater: ITransactionMetricsUpdater;
  errorParser: IErrorParser;
  l1Config: L1NetworkConfig;
  l2Config: L2NetworkConfig;
  loggerOptions?: LoggerOptions;
};

export class L2ToL1App {
  private readonly pollers: IPoller[];

  constructor(deps: L2ToL1Deps) {
    const {
      l2LogClient,
      l2Provider,
      lineaRollupClient,
      l1Provider,
      l1PublicClient,
      l1SignerAddress,
      messageRepository,
      l1GasProvider,
      calldataDecoder,
      sponsorshipMetricsUpdater,
      transactionMetricsUpdater,
      errorParser,
      l1Config,
      l2Config,
      loggerOptions,
    } = deps;

    const log = (name: string) => new PostmanWinstonLogger(name, loggerOptions);
    const logWithErrorParser = (name: string) => new PostmanWinstonLogger(name, loggerOptions, errorParser);

    const sentEventProcessor = new MessageSentEventProcessor(
      messageRepository,
      l2LogClient,
      l2Provider,
      calldataDecoder,
      {
        direction: Direction.L2_TO_L1,
        maxBlocksToFetchLogs: l2Config.listener.maxBlocksToFetchLogs,
        blockConfirmation: l2Config.listener.blockConfirmation,
        isEOAEnabled: l2Config.isEOAEnabled,
        isCalldataEnabled: l2Config.isCalldataEnabled,
        eventFilters: l2Config.listener.eventFilters,
      },
      log(`L2${MessageSentEventProcessor.name}`),
    );

    const sentEventPoller = new MessageSentEventPoller(
      sentEventProcessor,
      l2Provider,
      messageRepository,
      {
        direction: Direction.L2_TO_L1,
        pollingInterval: l2Config.listener.pollingInterval,
        initialFromBlock: l2Config.listener.initialFromBlock,
        originContractAddress: l2Config.messageServiceContractAddress,
      },
      logWithErrorParser(`L2${MessageSentEventPoller.name}`),
    );

    const anchoringProcessor = new MessageAnchoringProcessor(
      lineaRollupClient,
      l1Provider,
      messageRepository,
      {
        direction: Direction.L2_TO_L1,
        maxFetchMessagesFromDb: l1Config.listener.maxFetchMessagesFromDb,
        originContractAddress: l2Config.messageServiceContractAddress,
      },
      logWithErrorParser(`L1${MessageAnchoringProcessor.name}`),
    );

    const anchoringPoller = new MessageAnchoringPoller(
      anchoringProcessor,
      { direction: Direction.L2_TO_L1, pollingInterval: l1Config.listener.pollingInterval },
      log(`L1${MessageAnchoringPoller.name}`),
    );

    const validationService = new EthereumTransactionValidationService(
      lineaRollupClient,
      l1GasProvider,
      {
        profitMargin: l1Config.claiming.profitMargin,
        maxClaimGasLimit: l1Config.claiming.maxClaimGasLimit,
        isPostmanSponsorshipEnabled: l1Config.claiming.isPostmanSponsorshipEnabled,
        maxPostmanSponsorGasLimit: l1Config.claiming.maxPostmanSponsorGasLimit,
      },
      log(`${EthereumTransactionValidationService.name}`),
    );

    const nonceManager = new InlineNonceManager(
      messageRepository,
      l1PublicClient,
      l1SignerAddress,
      l1Config.claiming.maxNonceDiff,
      Direction.L2_TO_L1,
      log(`L1${InlineNonceManager.name}`),
    );

    const getNextMessageToClaim = async () => {
      const { maxFeePerGas } = await l1GasProvider.getGasFees();
      return messageRepository.getFirstMessageToClaimOnL1(
        Direction.L2_TO_L1,
        l2Config.messageServiceContractAddress,
        maxFeePerGas,
        l1Config.claiming.profitMargin,
        l1Config.claiming.maxNumberOfRetries,
        l1Config.claiming.retryDelayInSeconds,
      );
    };

    const claimingProcessor = new MessageClaimingProcessor(
      lineaRollupClient,
      nonceManager,
      messageRepository,
      getNextMessageToClaim,
      validationService,
      errorParser,
      {
        direction: Direction.L2_TO_L1,
        originContractAddress: l2Config.messageServiceContractAddress,
        feeRecipientAddress: l1Config.claiming.feeRecipientAddress,
        profitMargin: l1Config.claiming.profitMargin,
        maxNumberOfRetries: l1Config.claiming.maxNumberOfRetries,
        retryDelayInSeconds: l1Config.claiming.retryDelayInSeconds,
        maxClaimGasLimit: l1Config.claiming.maxClaimGasLimit,
        claimViaAddress: l1Config.claiming.claimViaAddress,
      },
      logWithErrorParser(`L1${MessageClaimingProcessor.name}`),
    );

    const claimingPoller = new MessageClaimingPoller(
      claimingProcessor,
      { direction: Direction.L2_TO_L1, pollingInterval: l1Config.listener.pollingInterval },
      log(`L1${MessageClaimingPoller.name}`),
    );

    const claimingPersister = new MessageClaimingPersister(
      messageRepository,
      lineaRollupClient,
      sponsorshipMetricsUpdater,
      transactionMetricsUpdater,
      l1Provider,
      {
        direction: Direction.L2_TO_L1,
        messageSubmissionTimeout: l1Config.claiming.messageSubmissionTimeout,
        maxTxRetries: l1Config.claiming.maxTxRetries,
        receiptPollingTimeout: l1Config.claiming.messageSubmissionTimeout,
        receiptPollingInterval: l1Config.listener.receiptPollingInterval,
      },
      logWithErrorParser(`L1${MessageClaimingPersister.name}`),
    );

    const persistingPoller = new MessagePersistingPoller(
      claimingPersister,
      { direction: Direction.L2_TO_L1, pollingInterval: l1Config.listener.receiptPollingInterval },
      log(`L1${MessagePersistingPoller.name}`),
    );

    this.pollers = [sentEventPoller, anchoringPoller, claimingPoller, persistingPoller];
  }

  public start(): void {
    this.pollers.forEach((p) => p.start());
  }

  public stop(): void {
    this.pollers.forEach((p) => p.stop());
  }
}
