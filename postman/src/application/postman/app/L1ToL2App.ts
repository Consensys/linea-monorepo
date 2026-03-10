import { type PublicClient } from "viem";
import { type LoggerOptions } from "winston";

import { L1NetworkConfig, L2NetworkConfig } from "./config/config";
import { ILineaRollupLogClient } from "../../../core/clients/blockchain/ethereum/ILineaRollupLogClient";
import { IProvider } from "../../../core/clients/blockchain/IProvider";
import { IL2MessageServiceClient } from "../../../core/clients/blockchain/linea/IL2MessageServiceClient";
import { ILineaProvider } from "../../../core/clients/blockchain/linea/ILineaProvider";
import { Direction, MessageStatus } from "../../../core/enums";
import { IErrorParser } from "../../../core/errors/IErrorParser";
import { ISponsorshipMetricsUpdater, ITransactionMetricsUpdater } from "../../../core/metrics";
import { IMessageRepository } from "../../../core/persistence/IMessageRepository";
import { ICalldataDecoder } from "../../../core/services/ICalldataDecoder";
import { ITransactionSigner } from "../../../core/services/ITransactionSigner";
import { IPoller } from "../../../core/services/pollers/IPoller";
import { InlineNonceManager } from "../../../infrastructure/blockchain/viem/signers/InlineNonceManager";
import { L2ClaimTransactionSizeCalculator } from "../../../services/L2ClaimTransactionSizeCalculator";
import { LineaTransactionValidationService } from "../../../services/LineaTransactionValidationService";
import {
  L2ClaimMessageTransactionSizePoller,
  MessageAnchoringPoller,
  MessageClaimingPoller,
  MessagePersistingPoller,
  MessageSentEventPoller,
} from "../../../services/pollers";
import {
  L2ClaimMessageTransactionSizeProcessor,
  MessageAnchoringProcessor,
  MessageClaimingPersister,
  MessageClaimingProcessor,
  MessageSentEventProcessor,
} from "../../../services/processors";
import { PostmanWinstonLogger } from "../../../utils/PostmanWinstonLogger";

export type L1ToL2Deps = {
  l1LogClient: ILineaRollupLogClient;
  l1Provider: IProvider;
  l2MessageServiceClient: IL2MessageServiceClient;
  l2Provider: ILineaProvider;
  l2PublicClient: PublicClient;
  l2SignerAddress: `0x${string}`;
  messageRepository: IMessageRepository;
  calldataDecoder: ICalldataDecoder;
  transactionSigner: ITransactionSigner;
  sponsorshipMetricsUpdater: ISponsorshipMetricsUpdater;
  transactionMetricsUpdater: ITransactionMetricsUpdater;
  errorParser: IErrorParser;
  l1Config: L1NetworkConfig;
  l2Config: L2NetworkConfig;
  loggerOptions?: LoggerOptions;
};

export class L1ToL2App {
  private readonly pollers: IPoller[];

  constructor(deps: L1ToL2Deps) {
    const {
      l1LogClient,
      l1Provider,
      l2MessageServiceClient,
      l2Provider,
      l2PublicClient,
      l2SignerAddress,
      messageRepository,
      calldataDecoder,
      transactionSigner,
      sponsorshipMetricsUpdater,
      transactionMetricsUpdater,
      errorParser,
      l1Config,
      l2Config,
      loggerOptions,
    } = deps;

    const log = (name: string) => new PostmanWinstonLogger(name, loggerOptions);

    const sentEventProcessor = new MessageSentEventProcessor(
      messageRepository,
      l1LogClient,
      l1Provider,
      calldataDecoder,
      {
        direction: Direction.L1_TO_L2,
        maxBlocksToFetchLogs: l1Config.listener.maxBlocksToFetchLogs,
        blockConfirmation: l1Config.listener.blockConfirmation,
        isEOAEnabled: l1Config.isEOAEnabled,
        isCalldataEnabled: l1Config.isCalldataEnabled,
        eventFilters: l1Config.listener.eventFilters,
      },
      log(`L1${MessageSentEventProcessor.name}`),
    );

    const sentEventPoller = new MessageSentEventPoller(
      sentEventProcessor,
      l1Provider,
      messageRepository,
      {
        direction: Direction.L1_TO_L2,
        pollingInterval: l1Config.listener.pollingInterval,
        initialFromBlock: l1Config.listener.initialFromBlock,
        originContractAddress: l1Config.messageServiceContractAddress,
      },
      log(`L1${MessageSentEventPoller.name}`),
    );

    const anchoringProcessor = new MessageAnchoringProcessor(
      l2MessageServiceClient,
      l2Provider,
      messageRepository,
      {
        direction: Direction.L1_TO_L2,
        maxFetchMessagesFromDb: l1Config.listener.maxFetchMessagesFromDb,
        originContractAddress: l1Config.messageServiceContractAddress,
      },
      log(`L2${MessageAnchoringProcessor.name}`),
    );

    const anchoringPoller = new MessageAnchoringPoller(
      anchoringProcessor,
      { direction: Direction.L1_TO_L2, pollingInterval: l2Config.listener.pollingInterval },
      log(`L2${MessageAnchoringPoller.name}`),
    );

    const validationService = new LineaTransactionValidationService(
      {
        profitMargin: l2Config.claiming.profitMargin,
        maxClaimGasLimit: l2Config.claiming.maxClaimGasLimit,
        isPostmanSponsorshipEnabled: l2Config.claiming.isPostmanSponsorshipEnabled,
        maxPostmanSponsorGasLimit: l2Config.claiming.maxPostmanSponsorGasLimit,
      },
      l2Provider,
      l2MessageServiceClient,
      log(`${LineaTransactionValidationService.name}`),
    );

    const nonceManager = new InlineNonceManager(
      messageRepository,
      l2PublicClient,
      l2SignerAddress,
      l2Config.claiming.maxNonceDiff,
      Direction.L1_TO_L2,
      log(`L2${InlineNonceManager.name}`),
    );

    const getNextMessageToClaim = () =>
      messageRepository.getFirstMessageToClaimOnL2(
        Direction.L1_TO_L2,
        l1Config.messageServiceContractAddress,
        [MessageStatus.TRANSACTION_SIZE_COMPUTED, MessageStatus.FEE_UNDERPRICED],
        l2Config.claiming.maxNumberOfRetries,
        l2Config.claiming.retryDelayInSeconds,
      );

    const claimingProcessor = new MessageClaimingProcessor(
      l2MessageServiceClient,
      nonceManager,
      messageRepository,
      getNextMessageToClaim,
      validationService,
      errorParser,
      {
        direction: Direction.L1_TO_L2,
        originContractAddress: l1Config.messageServiceContractAddress,
        feeRecipientAddress: l2Config.claiming.feeRecipientAddress,
        profitMargin: l2Config.claiming.profitMargin,
        maxNumberOfRetries: l2Config.claiming.maxNumberOfRetries,
        retryDelayInSeconds: l2Config.claiming.retryDelayInSeconds,
        maxClaimGasLimit: l2Config.claiming.maxClaimGasLimit,
        claimViaAddress: l2Config.claiming.claimViaAddress,
      },
      log(`L2${MessageClaimingProcessor.name}`),
    );

    const claimingPoller = new MessageClaimingPoller(
      claimingProcessor,
      { direction: Direction.L1_TO_L2, pollingInterval: l2Config.listener.pollingInterval },
      log(`L2${MessageClaimingPoller.name}`),
    );

    const claimingPersister = new MessageClaimingPersister(
      messageRepository,
      l2MessageServiceClient,
      sponsorshipMetricsUpdater,
      transactionMetricsUpdater,
      l2Provider,
      {
        direction: Direction.L1_TO_L2,
        messageSubmissionTimeout: l2Config.claiming.messageSubmissionTimeout,
        maxTxRetries: l2Config.claiming.maxTxRetries,
        receiptPollingTimeout: l2Config.claiming.messageSubmissionTimeout,
        receiptPollingInterval: l2Config.listener.receiptPollingInterval,
      },
      log(`L2${MessageClaimingPersister.name}`),
    );

    const persistingPoller = new MessagePersistingPoller(
      claimingPersister,
      { direction: Direction.L1_TO_L2, pollingInterval: l2Config.listener.receiptPollingInterval },
      log(`L2${MessagePersistingPoller.name}`),
    );

    const sizeCalculator = new L2ClaimTransactionSizeCalculator(l2MessageServiceClient, transactionSigner);
    const sizeProcessor = new L2ClaimMessageTransactionSizeProcessor(
      messageRepository,
      l2MessageServiceClient,
      sizeCalculator,
      { direction: Direction.L1_TO_L2, originContractAddress: l1Config.messageServiceContractAddress },
      log(`${L2ClaimMessageTransactionSizeProcessor.name}`),
    );

    const sizePoller = new L2ClaimMessageTransactionSizePoller(
      sizeProcessor,
      { pollingInterval: l2Config.listener.pollingInterval },
      log(`${L2ClaimMessageTransactionSizePoller.name}`),
    );

    this.pollers = [sentEventPoller, anchoringPoller, claimingPoller, persistingPoller, sizePoller];
  }

  public start(): void {
    this.pollers.forEach((p) => p.start());
  }

  public stop(): void {
    this.pollers.forEach((p) => p.stop());
  }
}
