import { DataSource } from "typeorm";
import { BigNumber } from "ethers";
import { MessageRepository } from "../repositories/MessageRepository";
import { L1NetworkConfig, MessageInDb, L2NetworkConfig } from "../utils/types";
import { OnChainMessageStatus } from "../../utils/enum";
import { LineaLogger } from "../../logger";
import { Direction, MessageStatus } from "../utils/enums";
import { L1MessageServiceContract, L2MessageServiceContract } from "../../contracts";
import { MessageEntity } from "../entity/Message.entity";
import { ErrorParser, ParsableError } from "../../errorHandlers";
import {
  DEFAULT_LISTENER_INTERVAL,
  DEFAULT_MAX_CLAIM_GAS_LIMIT,
  DEFAULT_MAX_FEE_PER_GAS,
  DEFAULT_MAX_NONCE_DIFF,
  DEFAULT_MAX_NUMBER_OF_RETRIES,
  DEFAULT_PROFIT_MARGIN,
  DEFAULT_RETRY_DELAY_IN_SECONDS,
  PROFIT_MARGIN_MULTIPLIER,
} from "../../utils/constants";
import { wait } from "../utils/helpers";

export abstract class ClaimTxSender<
  TMessageServiceContract extends L1MessageServiceContract | L2MessageServiceContract,
> {
  protected logger: LineaLogger;
  protected maxNonceDiff: number;
  protected shouldStopListening: boolean;
  protected messageRepository: MessageRepository;
  protected originContractAddress: string;
  protected pollingInterval: number;
  protected feeRecipient?: string;
  protected profitMargin: number;
  protected maxRetry: number;
  protected retryDelayInSeconds: number;
  protected maxFeePerGasFromConfig: BigNumber;
  protected maxClaimGasLimit: BigNumber;

  constructor(
    private dataSource: DataSource,
    private readonly messageServiceContract: TMessageServiceContract,
    config: L1NetworkConfig | L2NetworkConfig,
    protected readonly direction: Direction,
  ) {
    this.maxNonceDiff = Math.max(config.claiming.maxNonceDiff ?? DEFAULT_MAX_NONCE_DIFF, 0);
    this.shouldStopListening = false;
    this.messageRepository = new MessageRepository(this.dataSource);
    this.feeRecipient = config.claiming.feeRecipientAddress;
    this.pollingInterval = config.listener.pollingInterval ?? DEFAULT_LISTENER_INTERVAL;
    this.profitMargin = config.claiming.profitMargin ?? DEFAULT_PROFIT_MARGIN;
    this.maxRetry = config.claiming.maxNumberOfRetries ?? DEFAULT_MAX_NUMBER_OF_RETRIES;
    this.retryDelayInSeconds = config.claiming.retryDelayInSeconds ?? DEFAULT_RETRY_DELAY_IN_SECONDS;
    this.maxFeePerGasFromConfig = BigNumber.from(config.claiming.maxFeePerGas ?? DEFAULT_MAX_FEE_PER_GAS);
    this.maxClaimGasLimit = BigNumber.from(config.claiming.maxClaimGasLimit ?? DEFAULT_MAX_CLAIM_GAS_LIMIT);
  }

  public async start() {
    while (!this.shouldStopListening) {
      await this.listenForReadyToBeClaimedMessages(this.pollingInterval);
    }
  }

  public stop() {
    this.shouldStopListening = true;
  }

  protected async listenForReadyToBeClaimedMessages(interval: number) {
    let nextMessageToClaim: MessageInDb | null = null;

    try {
      const nonce = await this.getNonce();

      if (!nonce && nonce !== 0) {
        this.logger.error(`Nonce returned from getNonce is an invalid value (e.g. null or undefined)`);
        return;
      }

      const { maxFeePerGas } = await this.messageServiceContract.get1559Fees();
      nextMessageToClaim = await this.messageRepository.getFirstMessageToClaim(
        this.direction,
        this.originContractAddress,
        // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
        maxFeePerGas!,
        this.profitMargin,
        this.maxRetry,
        this.retryDelayInSeconds,
      );

      if (!nextMessageToClaim) {
        await wait(this.pollingInterval);
        return;
      }

      if (BigNumber.from(nextMessageToClaim.fee).eq(0) && this.profitMargin !== 0) {
        this.logger.warn(`Zero fee found in this message: ${JSON.stringify(nextMessageToClaim)}`);
        await this.messageRepository.updateMessage(nextMessageToClaim.messageHash, nextMessageToClaim.direction, {
          status: MessageStatus.ZERO_FEE,
        });
        return;
      }

      const messageStatus = await this.messageServiceContract.getMessageStatus(nextMessageToClaim.messageHash);

      if (messageStatus === OnChainMessageStatus.CLAIMED) {
        this.logger.info(`Found already claimed message: ${JSON.stringify(nextMessageToClaim)}`);
        await this.messageRepository.updateMessage(nextMessageToClaim.messageHash, nextMessageToClaim.direction, {
          status: MessageStatus.CLAIMED_SUCCESS,
        });
        return;
      }

      const { estimatedGasLimit, threshold } = await this.calculateGasEstimationAndThresHold(nextMessageToClaim);
      const txGasLimit = this.getGasLimit(estimatedGasLimit);

      if (!txGasLimit) {
        await this.messageRepository.updateMessage(nextMessageToClaim.messageHash, nextMessageToClaim.direction, {
          status: MessageStatus.NON_EXECUTABLE,
        });
        this.logger.warn(
          `Estimated gas limit (${estimatedGasLimit}) is higher than the max allowed gas limit (${this.maxClaimGasLimit.toNumber()}) for this message: ${JSON.stringify(
            nextMessageToClaim,
          )}`,
        );
        return;
      }

      await this.messageRepository.updateMessage(nextMessageToClaim.messageHash, nextMessageToClaim.direction, {
        claimGasEstimationThreshold: threshold,
      });

      const isTxUnderPriced = await this.isTransactionUnderPriced(txGasLimit, nextMessageToClaim.fee);

      if (isTxUnderPriced) {
        this.logger.warn(`Fee underpriced found in this message: ${JSON.stringify(nextMessageToClaim)}`);
        await this.messageRepository.updateMessage(nextMessageToClaim.messageHash, nextMessageToClaim.direction, {
          status: MessageStatus.FEE_UNDERPRICED,
        });
        return;
      }

      if (await this.isRateLimitExceeded(nextMessageToClaim.fee, nextMessageToClaim.value)) {
        this.logger.warn(`Rate limit exceeded on L1 for this message: ${JSON.stringify(nextMessageToClaim)}`);

        await this.messageRepository.updateMessage(nextMessageToClaim.messageHash, nextMessageToClaim.direction, {
          claimGasEstimationThreshold: undefined,
        });

        await wait(this.pollingInterval);
        return;
      }

      await this.executeClaimTransaction(nextMessageToClaim, nonce, txGasLimit);
    } catch (e) {
      const parsedError = ErrorParser.parseErrorWithMitigation(e as ParsableError);
      if (parsedError?.mitigation && !parsedError.mitigation.shouldRetry) {
        if (nextMessageToClaim) {
          await this.messageRepository.updateMessage(nextMessageToClaim.messageHash, nextMessageToClaim.direction, {
            status: MessageStatus.NON_EXECUTABLE,
          });
        }
      }
      this.logger.error(
        `Error found in listenForReadyToBeClaimedMessages:\nFailed message: ${JSON.stringify(
          nextMessageToClaim,
        )}\nFounded error: ${JSON.stringify(e)}\nParsed error:${JSON.stringify(parsedError)}`,
      );
      await wait(interval);
    }
  }

  protected async getNonce(): Promise<number | null> {
    const lastTxNonce = await this.messageRepository.getLastTxNonce(this.direction);

    let nonce = await this.messageServiceContract.getCurrentNonce();
    if (lastTxNonce) {
      if (lastTxNonce - nonce > this.maxNonceDiff) {
        this.logger.warn(
          `Last recorded nonce in db (${lastTxNonce}) is higher than the latest nonce from blockchain (${nonce}) and exceeds the limit (${this.maxNonceDiff}), paused the claim message process now`,
        );
        return null;
      }
      nonce = Math.max(nonce, lastTxNonce + 1);
    }
    return nonce;
  }

  protected async calculateGasEstimationAndThresHold(
    message: MessageInDb,
  ): Promise<{ threshold: number; estimatedGasLimit: BigNumber }> {
    const gasEstimation = await this.messageServiceContract.estimateClaimGas(
      {
        messageSender: message.messageSender,
        destination: message.destination,
        fee: BigNumber.from(message.fee),
        value: BigNumber.from(message.value),
        calldata: message.calldata,
        messageNonce: BigNumber.from(message.messageNonce),
        feeRecipient: this.feeRecipient,
        messageHash: message.messageHash,
      },
      { ...(await this.messageServiceContract.get1559Fees()) },
    );

    return {
      threshold: parseFloat(message.fee) / gasEstimation.toNumber(),
      estimatedGasLimit: gasEstimation,
    };
  }

  protected getGasLimit(gasLimit: BigNumber): BigNumber | null {
    if (gasLimit.lte(this.maxClaimGasLimit)) {
      return gasLimit;
    } else {
      return null;
    }
  }

  protected async isTransactionUnderPriced(gasLimit: BigNumber, messageFee: string): Promise<boolean> {
    const maxFeePerGas = (await this.messageServiceContract.get1559Fees()).maxFeePerGas;

    if (
      gasLimit
        .mul(maxFeePerGas)
        .mul(Math.floor(this.profitMargin * PROFIT_MARGIN_MULTIPLIER))
        .gt(BigNumber.from(messageFee).mul(PROFIT_MARGIN_MULTIPLIER))
    ) {
      return true;
    }
    return false;
  }

  protected async executeClaimTransaction(message: MessageInDb, nonce: number, gasLimit: BigNumber) {
    await this.messageRepository.manager.transaction(async (entityManager) => {
      await entityManager.update(
        MessageEntity,
        { messageHash: message.messageHash, direction: message.direction },
        {
          claimTxCreationDate: new Date(),
          claimTxNonce: nonce,
          status: MessageStatus.PENDING,
          ...(message.status === MessageStatus.FEE_UNDERPRICED
            ? { claimNumberOfRetry: message.claimNumberOfRetry + 1, claimLastRetriedAt: new Date() }
            : {}),
        },
      );

      const tx = await this.messageServiceContract.claim(
        {
          messageSender: message.messageSender,
          destination: message.destination,
          fee: BigNumber.from(message.fee),
          value: BigNumber.from(message.value),
          calldata: message.calldata,
          messageNonce: BigNumber.from(message.messageNonce),
          feeRecipient: this.feeRecipient,
          messageHash: message.messageHash,
        },
        { nonce, gasLimit, ...(await this.messageServiceContract.get1559Fees()) },
      );

      await entityManager.update(
        MessageEntity,
        { messageHash: message.messageHash, direction: message.direction },
        {
          claimTxGasLimit: tx.gasLimit.toNumber(),
          claimTxMaxFeePerGas: tx.maxFeePerGas?.toBigInt(),
          claimTxMaxPriorityFeePerGas: tx.maxPriorityFeePerGas?.toBigInt(),
          claimTxHash: tx.hash,
        },
      );
    });
  }

  protected abstract isRateLimitExceeded(messageFee: string, messageValue: string): Promise<boolean>;
}
