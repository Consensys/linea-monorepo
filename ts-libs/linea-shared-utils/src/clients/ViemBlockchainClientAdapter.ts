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
  RpcRequestError,
} from "viem";
import { sendRawTransaction, waitForTransactionReceipt } from "viem/actions";
import { IContractSignerClient } from "../core/client/IContractSignerClient";
import { ILogger } from "../logging/ILogger";
import { MAX_BPS } from "../core/constants/maths";

/**
 * Adapter that wraps viem's PublicClient to provide blockchain interaction functionality.
 * Implements transaction sending with retry logic, gas fee estimation, and connection pooling.
 * Uses a single PublicClient instance for better connection pooling and memory efficiency.
 */
export class ViemBlockchainClientAdapter implements IBlockchainClient<PublicClient, TransactionReceipt> {
  blockchainClient: PublicClient;

  /**
   * Creates a new ViemBlockchainClientAdapter instance.
   *
   * @param {ILogger} logger - The logger instance for logging blockchain operations.
   * @param {string} rpcUrl - The RPC URL for the blockchain network.
   * @param {Chain} chain - The blockchain chain configuration.
   * @param {IContractSignerClient} contractSignerClient - The client for signing transactions.
   * @param {number} [sendTransactionsMaxRetries=3] - Maximum number of retry attempts for sending transactions (must be at least 1).
   * @param {bigint} [gasRetryBumpBps=1000n] - Gas price bump in basis points per retry (e.g., 1000n = +10% per retry).
   * @param {number} [sendTransactionAttemptTimeoutMs=300000] - Timeout in milliseconds for each transaction attempt (default: 5 minutes).
   * @throws {Error} If sendTransactionsMaxRetries is less than 1.
   */
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
        onFetchRequest: async (request) => {
          const cloned = request.clone(); // clone before reading body
          try {
            const bodyText = await cloned.text();
            this.logger.debug("onFetchRequest", {
              method: request.method,
              url: request.url,
              body: bodyText,
            });
          } catch (err) {
            this.logger.warn("Failed to read request body", { err });
          }
        },
        onFetchResponse: async (resp) => {
          const cloned = resp.clone(); // clone before reading body
          try {
            const bodyText = await cloned.text();
            this.logger.debug("onFetchResponse", {
              status: resp.status,
              statusText: resp.statusText,
              body: bodyText,
            });
          } catch (err) {
            this.logger.warn("Failed to read response body", { err });
          }
        },
      }),
      batch: {
        // Not sure if this will help or not, need to experiment in testnet
        multicall: true,
      },
    });
  }

  /**
   * Gets the underlying viem PublicClient instance.
   *
   * @returns {PublicClient} The viem PublicClient instance used for blockchain interactions.
   */
  getBlockchainClient(): PublicClient {
    return this.blockchainClient;
  }

  /**
   * Gets the chain ID of the connected blockchain network.
   *
   * @returns {Promise<number>} The chain ID of the blockchain network.
   */
  async getChainId(): Promise<number> {
    return await this.blockchainClient.getChainId();
  }

  /**
   * Gets the balance of an Ethereum address in wei.
   *
   * @param {Address} address - The Ethereum address to query the balance for.
   * @returns {Promise<bigint>} The balance in wei.
   */
  async getBalance(address: Address): Promise<bigint> {
    return await this.blockchainClient.getBalance({
      address,
    });
  }

  /**
   * Estimates the current gas fees for EIP-1559 transactions.
   *
   * @returns {Promise<{ maxFeePerGas: bigint; maxPriorityFeePerGas: bigint }>} An object containing the estimated maxFeePerGas and maxPriorityFeePerGas values.
   */
  async estimateGasFees(): Promise<{ maxFeePerGas: bigint; maxPriorityFeePerGas: bigint }> {
    const { maxFeePerGas, maxPriorityFeePerGas } = await this.blockchainClient.estimateFeesPerGas();
    return { maxFeePerGas, maxPriorityFeePerGas };
  }

  /**
   * Attempts to send a signed transaction with retry-on-timeout semantics.
   * On each retry, bumps gas fees by `gasRetryBumpBps` basis points.
   * Uses a single nonce for all retry attempts to prevent nonce conflicts.
   * Does not retry on contract revert errors or non-timeout errors, only on timeout.
   *
   * @param {Address} contractAddress - The address of the contract to interact with.
   * @param {Hex} calldata - The encoded function call data.
   * @param {bigint} [value=0n] - The amount of ether to send with the transaction (default: 0).
   * @returns {Promise<TransactionReceipt>} The transaction receipt if successful.
   * @throws {ContractFunctionRevertedError} If the contract call reverts (not retried).
   * @throws {Error} If retry attempts are exhausted or a non-timeout error occurs.
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
        // We don't want to retry for Solidity revert, "-32015" -> VM execution error - https://www.quicknode.com/docs/ethereum/error-references
          if (error instanceof RpcRequestError && error.code === -32015) {
            this.logger.error('Transaction execution failed:', {
              code: error.code,
              message: error.message,
              data: error.data,
            });
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
          error instanceof TimeoutError || (error instanceof BaseError && error.name === "TimeoutError");
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

  /**
   * Internal method that sends a signed EIP-1559 transaction to the blockchain.
   * Signs the transaction using the contract signer client, serializes it, and broadcasts it.
   * Waits for the transaction receipt before returning.
   *
   * @param {Address} contractAddress - The address of the contract to interact with.
   * @param {Hex} calldata - The encoded function call data.
   * @param {bigint} value - The amount of ether to send with the transaction.
   * @param {number} nonce - The transaction nonce.
   * @param {bigint} gasLimit - The estimated gas limit for the transaction.
   * @param {number} chainId - The chain ID of the blockchain network.
   * @param {bigint} maxFeePerGas - The maximum fee per gas (EIP-1559).
   * @param {bigint} maxPriorityFeePerGas - The maximum priority fee per gas (EIP-1559).
   * @param {bigint} gasMultiplierBps - Gas multiplier in basis points to apply to gas values.
   * @returns {Promise<TransactionReceipt>} The transaction receipt after the transaction is mined.
   */
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
