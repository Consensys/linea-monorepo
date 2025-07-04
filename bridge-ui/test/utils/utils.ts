import { localL2Network } from "@/constants";
import { Client, parseEther } from "viem";
import { getAddresses, sendTransaction, waitForTransactionReceipt } from "viem/actions";
import { estimateGas } from "viem/linea";

export async function sendTransactionsToGenerateTrafficWithInterval(client: Client, pollingInterval: number = 1_000) {
  let timeoutId: NodeJS.Timeout | null = null;
  let isRunning = true;

  const [accountAddress] = await getAddresses(client);
  const sendTx = async () => {
    if (!isRunning) return;

    try {
      const { priorityFeePerGas, baseFeePerGas, gasLimit } = await estimateGas(client, {
        account: client.account!,
        to: accountAddress,
        value: parseEther("0.000001"),
      });
      const tx = await sendTransaction(client, {
        to: accountAddress,
        value: parseEther("0.000001"),
        account: client.account!,
        chain: localL2Network,
        gas: gasLimit,
        maxFeePerGas: baseFeePerGas + priorityFeePerGas,
        maxPriorityFeePerGas: priorityFeePerGas,
      });

      await waitForTransactionReceipt(client, { hash: tx, confirmations: 1 });
    } catch (error) {
      console.error(`Error sending transaction. error=${JSON.stringify(error)}`);
    } finally {
      if (isRunning) {
        timeoutId = setTimeout(sendTx, pollingInterval);
      }
    }
  };

  const stop = () => {
    isRunning = false;
    if (timeoutId) {
      clearTimeout(timeoutId);
      timeoutId = null;
    }
    console.log("Stopped generating traffic on L2");
  };

  sendTx();

  return stop;
}
