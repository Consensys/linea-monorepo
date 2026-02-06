import { Account, Chain, Client, Transport } from "viem";
import { sendTransaction, SendTransactionParameters, waitForTransactionReceipt } from "viem/actions";
import { EstimateGasParameters } from "viem/linea";

import { estimateLineaGas } from "./gas";
import { serialize } from "./misc";
import { etherToWei } from "./number";
import { createTestLogger } from "../../config/logger";

const logger = createTestLogger();

export async function sendTransactionsToGenerateTrafficWithInterval<
  chain extends Chain | undefined,
  account extends Account | undefined,
>(
  walletClient: Client<Transport, chain, account>,
  publicClient: Client<Transport, chain, account>,
  params: {
    pollingInterval?: number;
  },
) {
  const account_ = walletClient.account;
  if (!account_) {
    throw new Error("Wallet client does not have an associated account");
  }

  const { pollingInterval = 1_000 } = params;

  let timeoutId: NodeJS.Timeout | null = null;
  let isRunning = true;

  const innerSendTransaction = async () => {
    // Atomic check: if stopped, don't proceed
    if (!isRunning) {
      return;
    }

    try {
      const { maxPriorityFeePerGas, maxFeePerGas } = await estimateLineaGas(publicClient, {
        account: account_.address,
        to: account_.address,
        value: etherToWei("0.000001"),
      } as EstimateGasParameters);

      // Final check before sending transaction
      if (!isRunning) {
        return;
      }

      const txHash = await sendTransaction(walletClient, {
        to: account_.address,
        value: etherToWei("0.000001"),
        type: "eip1559",
        maxPriorityFeePerGas,
        maxFeePerGas,
      } as SendTransactionParameters);

      // Check before waiting for receipt
      if (!isRunning) {
        logger.debug(`Stopped before waiting for receipt. hash=${txHash}`);
        return;
      }

      await waitForTransactionReceipt(publicClient, { hash: txHash, timeout: 30_000 });
      logger.debug(
        `Transaction sent successfully. hash=${txHash} maxPriorityFeePerGas=${maxPriorityFeePerGas} maxFeePerGas=${maxFeePerGas}`,
      );
    } catch (error) {
      logger.error(`Error sending transaction. error=${serialize(error)}`);
    } finally {
      // Only schedule next iteration if still running
      if (isRunning) {
        timeoutId = setTimeout(innerSendTransaction, pollingInterval);
      }
    }
  };

  const stop = () => {
    isRunning = false;
    if (timeoutId) {
      clearTimeout(timeoutId);
      timeoutId = null;
    }
    logger.debug("Stopped generating traffic on L2");
  };

  innerSendTransaction();

  return stop;
}
