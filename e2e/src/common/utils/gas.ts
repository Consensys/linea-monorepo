import { Account, Chain, Client, Transport } from "viem";
import { estimateGas, EstimateGasParameters } from "viem/linea";

export async function estimateLineaGas<
  chain extends Chain | undefined = Chain | undefined,
  account extends Account | undefined = Account | undefined,
>(client: Client<Transport, chain, account>, params: EstimateGasParameters<chain>) {
  const BASE_FEE_MULTIPLIER = 1.35;
  const PRIORITY_FEE_MULTIPLIER = 1.05;
  const result = await estimateGas(client, params);

  const baseFeePerGas = (result.baseFeePerGas * BigInt(BASE_FEE_MULTIPLIER * 100)) / 100n;
  const maxPriorityFeePerGas = (result.priorityFeePerGas * BigInt(PRIORITY_FEE_MULTIPLIER * 100)) / 100n;

  return {
    maxFeePerGas: baseFeePerGas + maxPriorityFeePerGas,
    maxPriorityFeePerGas,
    gasLimit: result.gasLimit,
  };
}
