import { DataSource } from "typeorm";
import { MessageRepository } from "../repositories/MessageRepository";
import { MessageInDb, L1NetworkConfig, L2NetworkConfig } from "../utils/types";
import { LineaLogger } from "../../logger";
import { Direction, MessageStatus } from "../utils/enums";
import { L1MessageServiceContract, L2MessageServiceContract } from "../../contracts";
import { TransactionReceipt } from "@ethersproject/providers";
import { ErrorParser, ParsableError } from "../../errorHandlers";
import { DEFAULT_LISTENER_INTERVAL, DEFAULT_MESSAGE_SUBMISSION_TIMEOUT } from "../../utils/constants";
import { wait } from "../utils/helpers";

export abstract class ClaimStatusWatcher<
  TMessageServiceContract extends L1MessageServiceContract | L2MessageServiceContract,
> {
  protected logger: LineaLogger;
  protected shouldStopListening: boolean;
  protected messageRepository: MessageRepository;
  private submissionTimeout: number;
  protected pollingInterval: number;

  constructor(
    private dataSource: DataSource,
    private readonly messageServiceContract: TMessageServiceContract,
    config: L1NetworkConfig | L2NetworkConfig,
    protected readonly direction: Direction,
  ) {
    this.submissionTimeout = config.claiming.messageSubmissionTimeout ?? DEFAULT_MESSAGE_SUBMISSION_TIMEOUT;
    this.shouldStopListening = false;
    this.messageRepository = new MessageRepository(this.dataSource);
    this.pollingInterval = config.listener.pollingInterval ?? DEFAULT_LISTENER_INTERVAL;
  }

  public async start() {
    while (!this.shouldStopListening) {
      await this.waitForReceipt(this.pollingInterval);
    }
  }

  public stop() {
    this.shouldStopListening = true;
  }

  protected async waitForReceipt(interval: number) {
    let message: MessageInDb | null = null;
    try {
      message = await this.messageRepository.getFirstPendingMessage(this.direction);
      if (!message || !message.claimTxHash) {
        await wait(interval);
        return;
      }

      const receipt = await this.messageServiceContract.provider.getTransactionReceipt(message.claimTxHash);
      if (!receipt) {
        if (message.updatedAt && new Date().getTime() - message.updatedAt?.getTime() > this.submissionTimeout) {
          this.logger.warn(`Retring to claim:\nmessage: ${JSON.stringify(message)}`);
          let tx;
          try {
            tx = await this.messageServiceContract.retryTransactionWithHigherFee(
              message.claimTxHash,
              await this.messageServiceContract.get1559Fees(),
            );
            await this.messageRepository.updateMessage(message.messageHash, message.direction, {
              claimTxGasLimit: tx?.gasLimit.toNumber(),
              claimTxMaxFeePerGas: tx?.maxFeePerGas?.toBigInt(),
              claimTxMaxPriorityFeePerGas: tx?.maxPriorityFeePerGas?.toBigInt(),
              claimTxHash: tx?.hash,
            });
            this.logger.warn(`Retried to claim done:\nmessage: ${JSON.stringify(message)}\ntx: ${JSON.stringify(tx)}`);
          } catch (error) {
            this.logger.error(
              `Error found in retryTransactionWithHigherFee:\nFailed message: ${JSON.stringify(
                message,
              )}\nFounded error: ${JSON.stringify(error)}`,
            );
            await this.messageRepository.updateMessage(message.messageHash, message.direction, {
              status: MessageStatus.NON_EXECUTABLE,
            });
            await wait(interval);
            return;
          }
        }

        await wait(interval);
        return;
      }

      await this.updateReceiptStatus(receipt);
    } catch (e) {
      const parsedError = ErrorParser.parseErrorWithMitigation(e as ParsableError);
      if (parsedError?.mitigation && !parsedError.mitigation.shouldRetry) {
        if (message) {
          await this.messageRepository.updateMessage(message.messageHash, message.direction, {
            status: MessageStatus.NON_EXECUTABLE,
          });
        }
      }
      this.logger.error(
        `Error found in waitForReceipt:\nFailed message: ${JSON.stringify(message)}\nFounded error: ${JSON.stringify(
          e,
        )}\nParsed error: ${JSON.stringify(parsedError)}`,
      );
      await wait(interval);
    }
  }

  protected async updateReceiptStatus(receipt: TransactionReceipt): Promise<void> {
    if (receipt.status === 0) {
      const isRateLimitExceeded = await this.isRateLimitExceededError(receipt.transactionHash);

      if (isRateLimitExceeded) {
        await this.messageRepository.updateMessageByTransactionHash(receipt.transactionHash, this.direction, {
          status: MessageStatus.SENT,
          claimGasEstimationThreshold: undefined,
        });
        return;
      }

      await this.messageRepository.updateMessageByTransactionHash(receipt.transactionHash, this.direction, {
        status: MessageStatus.CLAIMED_REVERTED,
      });
      this.logger.warn(`CLAIMED_REVERTED: Message with tx hash ${receipt.transactionHash} has been reverted.`);

      return;
    }

    await this.messageRepository.updateMessageByTransactionHash(receipt.transactionHash, this.direction, {
      status: MessageStatus.CLAIMED_SUCCESS,
    });

    this.logger.info(`CLAIMED_SUCCESS: Message with tx hash ${receipt.transactionHash} has been claimed.`);
  }

  protected abstract isRateLimitExceededError(transactionHash: string): Promise<boolean>;
}
