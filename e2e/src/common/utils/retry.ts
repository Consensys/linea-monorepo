import { Abi, BaseError, Client, decodeErrorResult, Hash, RawContractError, TransactionReceipt } from "viem";
import {
  waitForTransactionReceipt,
  getTransactionReceipt,
  getTransaction,
  call,
  SendTransactionErrorType,
  WaitForTransactionReceiptErrorType,
} from "viem/actions";

import { createTestLogger } from "../../config/logger";

const DEFAULT_RECEIPT_TIMEOUT_MS = 30_000;
const DEFAULT_FEE_BUMP_FACTOR = 115n;
const DEFAULT_FEE_BUMP_STEP = 20n;
const MAX_FEE_MULTIPLIER = 10n;
const DEFAULT_MAX_RETRIES = 20;
const DEFAULT_OVERALL_TIMEOUT_MS = 3 * 60_000;
const DEFAULT_RETRY_ON_REVERT_DELAY_MS = 1_000;

export type FeeOverrides = {
  maxPriorityFeePerGas: bigint | undefined;
  maxFeePerGas: bigint | undefined;
};

export type SendTransactionWithRetryOptions = {
  receiptTimeoutMs?: number;
  feeBumpFactor?: bigint;
  maxRetries?: number;
  overallTimeoutMs?: number;
  rejectOnRevert?: boolean;
  abi?: Abi;
  retryOnRevert?: boolean;
  retryOnRevertDelayMs?: number;
  beforeRetry?: () => Promise<void>;
};

export type TransactionResult = {
  hash: Hash;
  receipt: TransactionReceipt;
};

export type GasPriceFeeOverrides = {
  gasPrice: bigint;
};

export type SendTransactionWithGasPriceRetryOptions = SendTransactionWithRetryOptions;

const logger = createTestLogger();

/* ---------------- helpers ---------------- */

function isReceiptTimeout(error: unknown): boolean {
  return (error as WaitForTransactionReceiptErrorType)?.name === "WaitForTransactionReceiptTimeoutError";
}

function isNonceTooLow(error: unknown): boolean {
  const e = error as SendTransactionErrorType;
  return e?.name === "TransactionExecutionError" && e?.cause?.name === "NonceTooLowError";
}

function isContractRevert(error: unknown): boolean {
  const e = error as { name?: string; cause?: { name?: string } };
  return (
    e?.name === "TransactionExecutionError" &&
    (e?.cause?.name === "ContractFunctionRevertedError" || e?.cause?.name === "CallExecutionError")
  );
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function safeGetReceipt(client: Client, hash: Hash): Promise<TransactionReceipt | undefined> {
  try {
    return await getTransactionReceipt(client, { hash });
  } catch {
    return undefined;
  }
}

function bumpFeesFromTx(
  tx: { maxFeePerGas: bigint | undefined; maxPriorityFeePerGas: bigint | undefined },
  bumpFactor: bigint,
): FeeOverrides {
  if (tx.maxFeePerGas === undefined || tx.maxPriorityFeePerGas === undefined) {
    throw new Error("sendTransactionWithRetry: non EIP-1559 transaction");
  }

  return {
    maxFeePerGas: (tx.maxFeePerGas * bumpFactor) / 100n,
    maxPriorityFeePerGas: (tx.maxPriorityFeePerGas * bumpFactor) / 100n,
  };
}

function capBumpedFees(fees: FeeOverrides, baseFees: FeeOverrides): FeeOverrides {
  if (
    fees.maxFeePerGas === undefined ||
    fees.maxPriorityFeePerGas === undefined ||
    baseFees.maxFeePerGas === undefined ||
    baseFees.maxPriorityFeePerGas === undefined
  ) {
    throw new Error("sendTransactionWithRetry: non EIP-1559 transaction");
  }

  const maxAllowedFeePerGas = baseFees.maxFeePerGas * MAX_FEE_MULTIPLIER;
  const maxAllowedPriorityFeePerGas = baseFees.maxPriorityFeePerGas * MAX_FEE_MULTIPLIER;

  return {
    maxFeePerGas: fees.maxFeePerGas > maxAllowedFeePerGas ? maxAllowedFeePerGas : fees.maxFeePerGas,
    maxPriorityFeePerGas:
      fees.maxPriorityFeePerGas > maxAllowedPriorityFeePerGas ? maxAllowedPriorityFeePerGas : fees.maxPriorityFeePerGas,
  };
}

function bumpGasPriceFromTx(tx: { gasPrice: bigint | undefined }, bumpFactor: bigint): GasPriceFeeOverrides {
  if (tx.gasPrice === undefined) {
    throw new Error("sendTransactionWithGasPriceRetry: transaction has no gasPrice");
  }

  return { gasPrice: (tx.gasPrice * bumpFactor) / 100n };
}

function capBumpedGasPrice(fees: GasPriceFeeOverrides, baseGasPrice: bigint): GasPriceFeeOverrides {
  const maxAllowedGasPrice = baseGasPrice * MAX_FEE_MULTIPLIER;
  return { gasPrice: fees.gasPrice > maxAllowedGasPrice ? maxAllowedGasPrice : fees.gasPrice };
}

async function getRevertReason(
  client: Client,
  hash: Hash,
  receipt: TransactionReceipt,
  abi?: Abi,
): Promise<string | undefined> {
  try {
    const tx = await getTransaction(client, { hash });
    await call(client, {
      account: tx.from,
      to: tx.to,
      data: tx.input,
      value: tx.value,
      gas: tx.gas,
      blockNumber: receipt.blockNumber,
    });
    return "unknown (eth_call did not revert on replay)";
  } catch (err) {
    if (!(err instanceof BaseError)) return undefined;

    const rawError = err.walk() as RawContractError;
    if (rawError.data && abi) {
      const data = typeof rawError.data === "object" ? rawError.data?.data : rawError.data;
      if (!data) return undefined;

      try {
        const decoded = decodeErrorResult({ abi, data });
        const args = decoded.args?.map((a) => (typeof a === "bigint" ? a.toString() : JSON.stringify(a)));
        return `${decoded.errorName}(${args?.join(", ") ?? ""})`;
      } catch {
        return `raw revert data: ${data}`;
      }
    }

    return (err as Error).message ?? String(err);
  }
}

async function assertReceiptSuccess(
  client: Client,
  hash: Hash,
  receipt: TransactionReceipt,
  rejectOnRevert: boolean,
  abi?: Abi,
): Promise<void> {
  if (rejectOnRevert && receipt.status === "reverted") {
    const reason = await getRevertReason(client, hash, receipt, abi);
    throw new Error(`Transaction reverted: hash=${hash} blockNumber=${receipt.blockNumber}\nRevert reason: ${reason}`);
  }
}

/* ---------------- final wrapper ---------------- */

export async function sendTransactionWithRetry(
  client: Client,
  sendFn: (fees?: FeeOverrides) => Promise<Hash>,
  options: SendTransactionWithRetryOptions = {},
): Promise<TransactionResult> {
  const receiptTimeoutMs = options.receiptTimeoutMs ?? DEFAULT_RECEIPT_TIMEOUT_MS;
  const feeBumpFactor = options.feeBumpFactor ?? DEFAULT_FEE_BUMP_FACTOR;
  const maxRetries = options.maxRetries ?? DEFAULT_MAX_RETRIES;
  const overallTimeoutMs = options.overallTimeoutMs ?? DEFAULT_OVERALL_TIMEOUT_MS;
  const rejectOnRevert = options.rejectOnRevert ?? true;
  const abi = options.abi;
  const retryOnRevert = options.retryOnRevert ?? false;
  const retryOnRevertDelayMs = options.retryOnRevertDelayMs ?? DEFAULT_RETRY_ON_REVERT_DELAY_MS;
  const beforeRetry = options.beforeRetry;

  const startedAt = Date.now();
  let lastHash!: Hash;
  let fees!: FeeOverrides;
  let baseFees!: FeeOverrides;
  let attempt = 0;
  let needsSend = true;

  function withinLimits(): boolean {
    return attempt < maxRetries && Date.now() - startedAt <= overallTimeoutMs;
  }

  async function freshSend(): Promise<void> {
    lastHash = await sendFn();
    const tx = await getTransaction(client, { hash: lastHash });
    fees = { maxPriorityFeePerGas: tx.maxPriorityFeePerGas, maxFeePerGas: tx.maxFeePerGas };
    baseFees = fees;
    logger.debug(
      `tx sent hash=${lastHash} maxFeePerGas=${fees.maxFeePerGas} maxPriorityFeePerGas=${fees.maxPriorityFeePerGas}`,
    );
  }

  async function handleRevertRetry(hash: Hash, receipt: TransactionReceipt): Promise<boolean> {
    if (receipt.status !== "reverted" || !retryOnRevert || !withinLimits()) return false;

    const reason = await getRevertReason(client, hash, receipt, abi);
    logger.debug(`tx reverted, will retry: hash=${hash} attempt=${attempt} reason=${reason}`);

    await beforeRetry?.();
    await sleep(retryOnRevertDelayMs);
    needsSend = true;
    attempt++;
    return true;
  }

  while (attempt <= maxRetries) {
    /* ---------- (re)send ---------- */
    if (needsSend) {
      needsSend = false;
      try {
        await freshSend();
      } catch (err) {
        if (retryOnRevert && isContractRevert(err) && withinLimits()) {
          logger.debug(`sendFn reverted at simulation, retrying: attempt=${attempt} error=${(err as Error).message}`);
          await beforeRetry?.();
          await sleep(retryOnRevertDelayMs);
          needsSend = true;
          attempt++;
          continue;
        }
        throw err;
      }
    }

    /* ---------- hard deadline ---------- */
    if (Date.now() - startedAt > overallTimeoutMs) {
      logger.debug(`overall timeout exceeded hash=${lastHash} attempt=${attempt}; probing receipt`);

      const txReceipt = await safeGetReceipt(client, lastHash);
      if (txReceipt) {
        logger.debug(`tx confirmed during final probe hash=${lastHash} blockNumber=${txReceipt.blockNumber}`);
        await assertReceiptSuccess(client, lastHash, txReceipt, rejectOnRevert, abi);
        return { hash: lastHash, receipt: txReceipt };
      }

      throw new Error("sendTransactionWithRetry: overall timeout exceeded");
    }

    /* ---------- primary wait ---------- */
    try {
      logger.debug(`waiting for receipt hash=${lastHash} attempt=${attempt} timeoutMs=${receiptTimeoutMs}`);

      const receipt = await waitForTransactionReceipt(client, {
        hash: lastHash,
        timeout: receiptTimeoutMs,
      });

      logger.debug(`tx confirmed hash=${lastHash} blockNumber=${receipt.blockNumber} status=${receipt.status}`);

      if (await handleRevertRetry(lastHash, receipt)) continue;
      await assertReceiptSuccess(client, lastHash, receipt, rejectOnRevert, abi);
      return { hash: lastHash, receipt };
    } catch (err) {
      if (!isReceiptTimeout(err)) throw err;

      logger.debug(`receipt timeout for hash=${lastHash} attempt=${attempt}`);
    }

    /* ---------- race guard ---------- */
    const raceConditionReceipt = await safeGetReceipt(client, lastHash);
    if (raceConditionReceipt) {
      logger.debug(`tx mined during timeout race hash=${lastHash} blockNumber=${raceConditionReceipt.blockNumber}`);

      if (await handleRevertRetry(lastHash, raceConditionReceipt)) continue;
      await assertReceiptSuccess(client, lastHash, raceConditionReceipt, rejectOnRevert, abi);
      return { hash: lastHash, receipt: raceConditionReceipt };
    }

    if (attempt === maxRetries) {
      logger.debug(`max retries reached hash=${lastHash} attempt=${attempt}; probing receipt`);

      const receipt = await safeGetReceipt(client, lastHash);
      if (receipt) {
        await assertReceiptSuccess(client, lastHash, receipt, rejectOnRevert, abi);
        return { hash: lastHash, receipt: receipt };
      }

      throw new Error("sendTransactionWithRetry: max retries exceeded");
    }

    /* ---------- bump from actual tx ---------- */
    const bumpFactor = feeBumpFactor + BigInt(attempt) * DEFAULT_FEE_BUMP_STEP;
    const nextFees = bumpFeesFromTx(fees, bumpFactor);
    const cappedFees = capBumpedFees(nextFees, baseFees);
    const wasCapped =
      cappedFees.maxFeePerGas !== nextFees.maxFeePerGas ||
      cappedFees.maxPriorityFeePerGas !== nextFees.maxPriorityFeePerGas;

    logger.debug(
      `bumping fees hash=${lastHash} attempt=${attempt + 1} bumpFactor=${bumpFactor} ` +
        `prevMaxFeePerGas=${fees.maxFeePerGas} nextMaxFeePerGas=${cappedFees.maxFeePerGas} ` +
        `prevMaxPriorityFeePerGas=${fees.maxPriorityFeePerGas} nextMaxPriorityFeePerGas=${cappedFees.maxPriorityFeePerGas}`,
    );

    if (wasCapped) {
      logger.debug(
        `fee cap reached hash=${lastHash} maxFeePerGasCap=${baseFees.maxFeePerGas! * MAX_FEE_MULTIPLIER} ` +
          `maxPriorityFeePerGasCap=${baseFees.maxPriorityFeePerGas! * MAX_FEE_MULTIPLIER}`,
      );
    }

    fees = cappedFees;
    attempt++;

    try {
      lastHash = await sendFn(fees);

      logger.debug(`replacement tx sent hash=${lastHash} attempt=${attempt}`);
    } catch (sendError) {
      if (isNonceTooLow(sendError)) {
        logger.debug(`nonce too low while retrying hash=${lastHash}; original tx likely mined`);

        const receipt = await safeGetReceipt(client, lastHash);
        if (receipt) {
          if (await handleRevertRetry(lastHash, receipt)) continue;
          await assertReceiptSuccess(client, lastHash, receipt, rejectOnRevert, abi);
          return { hash: lastHash, receipt: receipt };
        }
        continue;
      }

      if (retryOnRevert && isContractRevert(sendError) && withinLimits()) {
        logger.debug(
          `sendFn reverted at simulation, retrying: attempt=${attempt} error=${(sendError as Error).message}`,
        );
        await beforeRetry?.();
        await sleep(retryOnRevertDelayMs);
        needsSend = true;
        attempt++;
        continue;
      }

      throw sendError;
    }
  }

  throw new Error("sendTransactionWithRetry: unreachable");
}

export async function sendTransactionWithGasPriceRetry(
  client: Client,
  sendFn: (fees?: GasPriceFeeOverrides) => Promise<Hash>,
  options: SendTransactionWithGasPriceRetryOptions = {},
): Promise<TransactionResult> {
  const receiptTimeoutMs = options.receiptTimeoutMs ?? DEFAULT_RECEIPT_TIMEOUT_MS;
  const feeBumpFactor = options.feeBumpFactor ?? DEFAULT_FEE_BUMP_FACTOR;
  const maxRetries = options.maxRetries ?? DEFAULT_MAX_RETRIES;
  const overallTimeoutMs = options.overallTimeoutMs ?? DEFAULT_OVERALL_TIMEOUT_MS;
  const rejectOnRevert = options.rejectOnRevert ?? true;
  const abi = options.abi;
  const retryOnRevert = options.retryOnRevert ?? false;
  const retryOnRevertDelayMs = options.retryOnRevertDelayMs ?? DEFAULT_RETRY_ON_REVERT_DELAY_MS;
  const beforeRetry = options.beforeRetry;

  const startedAt = Date.now();
  let lastHash!: Hash;
  let gasPriceFees!: GasPriceFeeOverrides;
  let baseGasPrice!: bigint;
  let attempt = 0;
  let needsSend = true;

  function withinLimits(): boolean {
    return attempt < maxRetries && Date.now() - startedAt <= overallTimeoutMs;
  }

  async function freshSendGasPrice(): Promise<void> {
    lastHash = await sendFn();
    const tx = await getTransaction(client, { hash: lastHash });
    if (tx.gasPrice === undefined) {
      throw new Error("sendTransactionWithGasPriceRetry: transaction has no gasPrice");
    }
    gasPriceFees = { gasPrice: tx.gasPrice };
    baseGasPrice = tx.gasPrice;
    logger.debug(`tx sent hash=${lastHash} gasPrice=${gasPriceFees.gasPrice}`);
  }

  async function handleRevertRetryGasPrice(hash: Hash, receipt: TransactionReceipt): Promise<boolean> {
    if (receipt.status !== "reverted" || !retryOnRevert || !withinLimits()) return false;

    const reason = await getRevertReason(client, hash, receipt, abi);
    logger.debug(`tx reverted, will retry: hash=${hash} attempt=${attempt} reason=${reason}`);

    await beforeRetry?.();
    await sleep(retryOnRevertDelayMs);
    needsSend = true;
    attempt++;
    return true;
  }

  while (attempt <= maxRetries) {
    if (needsSend) {
      needsSend = false;
      try {
        await freshSendGasPrice();
      } catch (err) {
        if (retryOnRevert && isContractRevert(err) && withinLimits()) {
          logger.debug(`sendFn reverted at simulation, retrying: attempt=${attempt} error=${(err as Error).message}`);
          await beforeRetry?.();
          await sleep(retryOnRevertDelayMs);
          needsSend = true;
          attempt++;
          continue;
        }
        throw err;
      }
    }

    if (Date.now() - startedAt > overallTimeoutMs) {
      logger.debug(`overall timeout exceeded hash=${lastHash} attempt=${attempt}; probing receipt`);

      const txReceipt = await safeGetReceipt(client, lastHash);
      if (txReceipt) {
        logger.debug(`tx confirmed during final probe hash=${lastHash} blockNumber=${txReceipt.blockNumber}`);
        await assertReceiptSuccess(client, lastHash, txReceipt, rejectOnRevert, abi);
        return { hash: lastHash, receipt: txReceipt };
      }

      throw new Error("sendTransactionWithGasPriceRetry: overall timeout exceeded");
    }

    try {
      logger.debug(`waiting for receipt hash=${lastHash} attempt=${attempt} timeoutMs=${receiptTimeoutMs}`);

      const receipt = await waitForTransactionReceipt(client, {
        hash: lastHash,
        timeout: receiptTimeoutMs,
      });

      logger.debug(`tx confirmed hash=${lastHash} blockNumber=${receipt.blockNumber} status=${receipt.status}`);

      if (await handleRevertRetryGasPrice(lastHash, receipt)) continue;
      await assertReceiptSuccess(client, lastHash, receipt, rejectOnRevert, abi);
      return { hash: lastHash, receipt };
    } catch (err) {
      if (!isReceiptTimeout(err)) throw err;

      logger.debug(`receipt timeout for hash=${lastHash} attempt=${attempt}`);
    }

    const raceConditionReceipt = await safeGetReceipt(client, lastHash);
    if (raceConditionReceipt) {
      logger.debug(`tx mined during timeout race hash=${lastHash} blockNumber=${raceConditionReceipt.blockNumber}`);

      if (await handleRevertRetryGasPrice(lastHash, raceConditionReceipt)) continue;
      await assertReceiptSuccess(client, lastHash, raceConditionReceipt, rejectOnRevert, abi);
      return { hash: lastHash, receipt: raceConditionReceipt };
    }

    if (attempt === maxRetries) {
      logger.debug(`max retries reached hash=${lastHash} attempt=${attempt}; probing receipt`);

      const receipt = await safeGetReceipt(client, lastHash);
      if (receipt) {
        await assertReceiptSuccess(client, lastHash, receipt, rejectOnRevert, abi);
        return { hash: lastHash, receipt };
      }

      throw new Error("sendTransactionWithGasPriceRetry: max retries exceeded");
    }

    const bumpFactor = feeBumpFactor + BigInt(attempt) * DEFAULT_FEE_BUMP_STEP;
    const nextFees = bumpGasPriceFromTx(gasPriceFees, bumpFactor);
    const cappedFees = capBumpedGasPrice(nextFees, baseGasPrice);
    const wasCapped = cappedFees.gasPrice !== nextFees.gasPrice;

    logger.debug(
      `bumping gasPrice hash=${lastHash} attempt=${attempt + 1} bumpFactor=${bumpFactor} ` +
        `prevGasPrice=${gasPriceFees.gasPrice} nextGasPrice=${cappedFees.gasPrice}`,
    );

    if (wasCapped) {
      logger.debug(`gasPrice cap reached hash=${lastHash} cap=${baseGasPrice * MAX_FEE_MULTIPLIER}`);
    }

    gasPriceFees = cappedFees;
    attempt++;

    try {
      lastHash = await sendFn(gasPriceFees);

      logger.debug(`replacement tx sent hash=${lastHash} attempt=${attempt}`);
    } catch (sendError) {
      if (isNonceTooLow(sendError)) {
        logger.debug(`nonce too low while retrying hash=${lastHash}; original tx likely mined`);

        const receipt = await safeGetReceipt(client, lastHash);
        if (receipt) {
          if (await handleRevertRetryGasPrice(lastHash, receipt)) continue;
          await assertReceiptSuccess(client, lastHash, receipt, rejectOnRevert, abi);
          return { hash: lastHash, receipt };
        }
        continue;
      }

      if (retryOnRevert && isContractRevert(sendError) && withinLimits()) {
        logger.debug(
          `sendFn reverted at simulation, retrying: attempt=${attempt} error=${(sendError as Error).message}`,
        );
        await beforeRetry?.();
        await sleep(retryOnRevertDelayMs);
        needsSend = true;
        attempt++;
        continue;
      }

      throw sendError;
    }
  }

  throw new Error("sendTransactionWithGasPriceRetry: unreachable");
}
