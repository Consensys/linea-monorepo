import { IBlockchainClient } from "../core/client/IBlockchainClient";
import {
  Hex,
  createPublicClient,
  http,
  PublicClient,
  TransactionReceipt,
  Address,
  TransactionSerializableEIP1559,
  serializeTransaction,
  parseSignature,
  Chain,
  ContractFunctionRevertedError,
  withTimeout,
  TimeoutError,
  BaseError,
} from "viem";
import { sendRawTransaction, waitForTransactionReceipt } from "viem/actions";
import { IContractSignerClient } from "../core/client/IContractSignerClient";
import { ILogger } from "../logging/ILogger";
import { MAX_BPS } from "../core/constants/maths";

// TODO - Ponder the edge cases for retry/timeout/throw logic in sendSignedTransaction before writing unit tests
// Re-use via composition in ContractClients
// Hope that using strategy pattern like this makes us more 'viem-agnostic'
export class ViemBlockchainClientAdapter implements IBlockchainClient<PublicClient, TransactionReceipt> {
  blockchainClient: PublicClient;

  constructor(
    private readonly logger: ILogger,
    rpcUrl: string,
    chain: Chain,
    private readonly contractSignerClient: IContractSignerClient,
    private readonly sendTransactionsMaxRetries = 3,
    private readonly gasRetryBumpBps: bigint = 1000n, // +10% per retry
    private readonly sendTransactionAttemptTimeoutMs = 300_000, // 5m
  ) {
    if (sendTransactionsMaxRetries < 1) {
      throw new Error("sendTransactionsMaxRetries must be at least 1");
    }
    // Aim re-use single blockchain client for
    // i.) Better connection pooling
    // ii.) Memory efficient
    // iii.) Single point of configuration
    this.blockchainClient = createPublicClient({
      chain,
      transport: http(rpcUrl, {
        batch: true,
        // TODO - How does this interact with our custom retry logic in sendSignedTransaction?
        // Hypothesis - Default Viem timeout of 10s will kick in first. It should still retry because we are using the native Viem Timeout error.
        retryCount: 3,
        onFetchRequest: (request) => {
          this.logger.debug("onFetchRequest", request);
        },
        onFetchResponse: (resp) => {
          this.logger.debug("onFetchResponse", resp);
        },
      }),
      batch: {
        // Not sure if this will help or not, need to experiment in testnet
        multicall: true,
      },
    });
  }

  getBlockchainClient(): PublicClient {
    return this.blockchainClient;
  }

  async getChainId(): Promise<number> {
    return await this.blockchainClient.getChainId();
  }

  async getBalance(address: Address): Promise<bigint> {
    return await this.blockchainClient.getBalance({
      address,
    });
  }

  async estimateGasFees(): Promise<{ maxFeePerGas: bigint; maxPriorityFeePerGas: bigint }> {
    const { maxFeePerGas, maxPriorityFeePerGas } = await this.blockchainClient.estimateFeesPerGas();
    return { maxFeePerGas, maxPriorityFeePerGas };
  }

  /**
   * Attempts to send a signed tx with retry-on-timeout semantics.
   * On each retry, bumps gas by `gasRetryBumpBps`.
   * Does not retry on errors, only on timeout.
   */
  async sendSignedTransaction(
    contractAddress: Address,
    calldata: Hex,
    value: bigint = 0n,
  ): Promise<TransactionReceipt> {
    this.logger.debug("sendSignedTransaction started");
    let gasMultiplierBps = MAX_BPS; // Start at 100%
    let maxFeePerGas = 0n;
    let maxPriorityFeePerGas = 0n;
    let lastError: unknown;

    // Use a single nonce for all retries.
    const [nonce, gasLimit, chainId] = await Promise.all([
      this.blockchainClient.getTransactionCount({ address: this.contractSignerClient.getAddress() }),
      this.blockchainClient.estimateGas({
        account: this.contractSignerClient.getAddress(),
        to: contractAddress,
        data: calldata,
        value,
      }),
      this.getChainId(),
    ]);

    for (let attempt = 1; attempt <= this.sendTransactionsMaxRetries; attempt += 1) {
      // Try to send tx with a timeout
      try {
        // Get new fee estimate each time
        const { maxFeePerGas: maxFeePerGasCandidate, maxPriorityFeePerGas: maxPriorityFeePerGasCandidate } =
          await this.estimateGasFees();

        // Use the highest fee estimate retrieved
        maxFeePerGas = maxFeePerGasCandidate > maxFeePerGas ? maxFeePerGasCandidate : maxFeePerGas;
        maxPriorityFeePerGas =
          maxPriorityFeePerGasCandidate > maxPriorityFeePerGas ? maxPriorityFeePerGasCandidate : maxPriorityFeePerGas;

        const receipt = await withTimeout(
          () =>
            this._sendSignedEIP1559Transaction(
              contractAddress,
              calldata,
              value,
              nonce,
              gasLimit,
              chainId,
              maxFeePerGas,
              maxPriorityFeePerGas,
              gasMultiplierBps,
            ),
          {
            timeout: this.sendTransactionAttemptTimeoutMs,
            signal: false, // don’t try to abort, just reject
            errorInstance: new TimeoutError({
              body: { message: "sendSignedTransaction attempt timed out" },
              url: "local:sendSignedTransaction",
            }),
          },
        );
        this.logger.debug(`sendSignedTransaction succeeded`, { receipt });
        return receipt;
      } catch (error) {
        // We don't want to retry for ContractFunctionRevertedError
        if (error instanceof ContractFunctionRevertedError) {
          this.logger.error("❌ sendSignedTransaction contract call reverted:", {
            shortMessage: error.shortMessage,
          });
          this.logger.error("Reason:", { reason: error.data?.errorName || error.message });
          throw error;
        }
        if (attempt >= this.sendTransactionsMaxRetries) {
          this.logger.error(
            `sendSignedTransaction retry attempts exhausted sendTransactionsMaxRetries=${this.sendTransactionsMaxRetries}`,
            { error },
          );
          throw error;
        }
        const isTimeout =
          error instanceof TimeoutError ||
          (error instanceof BaseError && error.name === "TimeoutError") ||
          (error as any)?.name === "TimeoutError";
        if (!isTimeout) {
          this.logger.error(`sendSignedTransaction error`, { error });
          throw error; // not a timeout → bail
        }

        this.logger.warn(
          `sendSignedTransaction retry attempt failed attempt=${attempt} sendTransactionsMaxRetries=${this.sendTransactionsMaxRetries}`,
          { error },
        );

        lastError = error;
        // Compound gas for next retry
        gasMultiplierBps = (gasMultiplierBps * (MAX_BPS + this.gasRetryBumpBps)) / MAX_BPS;
      }
    }

    // Unreachable but required to simplify TypeScript return type
    throw lastError;
  }

  private async _sendSignedEIP1559Transaction(
    contractAddress: Address,
    calldata: Hex,
    value: bigint,
    nonce: number,
    gasLimit: bigint,
    chainId: number,
    maxFeePerGas: bigint,
    maxPriorityFeePerGas: bigint,
    gasMultiplierBps: bigint,
  ): Promise<TransactionReceipt> {
    const tx: TransactionSerializableEIP1559 = {
      to: contractAddress,
      type: "eip1559",
      data: calldata,
      chainId: chainId,
      gas: (gasLimit * gasMultiplierBps) / MAX_BPS,
      maxFeePerGas: (maxFeePerGas * gasMultiplierBps) / MAX_BPS,
      maxPriorityFeePerGas: (maxPriorityFeePerGas * gasMultiplierBps) / MAX_BPS,
      nonce,
      value,
    };
    this.logger.debug("_sendSignedTransaction tx for signing", { tx });
    const signature = await this.contractSignerClient.sign(tx);
    const serializedTransaction = serializeTransaction(tx, parseSignature(signature));

    this.logger.debug(
      `_sendSignedTransaction - sending raw transaction serializedTransaction=${serializedTransaction}`,
    );
    const txHash = await sendRawTransaction(this.blockchainClient, { serializedTransaction });
    const receipt = await waitForTransactionReceipt(this.blockchainClient, { hash: txHash });
    return receipt;
  }
}
