import { estimateLineaGas } from "./gas";
import { etherToWei } from "./number";
import { sendTransaction, SendTransactionParameters, waitForTransactionReceipt } from "viem/actions";
import { EstimateGasParameters } from "viem/linea";
import { Account, Chain, Client, Transport } from "viem";
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
    if (!isRunning) return;

    try {
      const { maxPriorityFeePerGas, maxFeePerGas } = await estimateLineaGas(publicClient, {
        account: account_.address,
        to: account_.address,
        value: etherToWei("0.000001"),
      } as EstimateGasParameters);

      const txHash = await sendTransaction(walletClient, {
        to: account_.address,
        value: etherToWei("0.000001"),
        type: "eip1559",
        maxPriorityFeePerGas,
        maxFeePerGas,
      } as SendTransactionParameters);
      await waitForTransactionReceipt(publicClient, { hash: txHash });
      logger.debug(
        `Transaction sent successfully. hash=${txHash} maxPriorityFeePerGas=${maxPriorityFeePerGas} maxFeePerGas=${maxFeePerGas}`,
      );
    } catch (error) {
      logger.error(`Error sending transaction. error=${JSON.stringify(error)}`);
    } finally {
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
