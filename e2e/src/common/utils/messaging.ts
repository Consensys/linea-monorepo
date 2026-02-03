import { encodeFunctionCall } from "./encoding";
import { estimateLineaGas } from "./gas";
import { etherToWei } from "./number";
import { DummyContractAbi, L2MessageServiceV1Abi } from "../../generated";
import { PrivateKeyAccount, toHex, TransactionReceipt, Hash } from "viem";
import { randomBytes } from "crypto";
import { config } from "../../config/tests-config/setup";
import { L2RpcEndpoint } from "../../config/tests-config/setup/clients/l2-client";

// Constants
const DEFAULT_TRANSACTION_TIMEOUT_MS = 30_000;
const CALLDATA_BYTE_SIZE = 100;
const DEFAULT_L2_DESTINATION_ADDRESS = "0x8D97689C9818892B700e27F316cc3E41e17fBeb9";
const DEFAULT_L1_DESTINATION_ADDRESS = "0x8D97689C9818892B700e27F316cc3E41e17fBeb9";

// Types
export type SendMessageParams = {
  account: PrivateKeyAccount;
  fee?: bigint;
  value?: bigint;
  withCalldata?: boolean;
  timeoutMs?: number;
};

export type SendMessageResult = {
  txHash: Hash;
  receipt: TransactionReceipt;
};

/**
 * Generates calldata for a dummy contract call if withCalldata is true, otherwise returns empty calldata.
 */
function generateCalldata(withCalldata: boolean): `0x${string}` {
  if (!withCalldata) {
    return "0x";
  }

  return encodeFunctionCall({
    abi: DummyContractAbi,
    functionName: "setPayload",
    args: [toHex(randomBytes(CALLDATA_BYTE_SIZE).toString("hex"))],
  });
}

/**
 * Sends a message from L1 to L2 via the Linea Rollup contract.
 * @param params - Message sending parameters
 * @returns Transaction hash and receipt
 * @throws Error if transaction fails or receipt is not received
 */
export async function sendL1ToL2Message(params: SendMessageParams): Promise<SendMessageResult> {
  const { account, fee = 0n, value = 0n, withCalldata = false, timeoutMs = DEFAULT_TRANSACTION_TIMEOUT_MS } = params;

  const dummyContract = config.l2PublicClient().getDummyContract();
  const lineaRollup = config.l1WalletClient({ account }).getLineaRollup();
  const l1PublicClient = config.l1PublicClient();

  const calldata = generateCalldata(withCalldata);
  const destinationAddress = withCalldata ? dummyContract.address : DEFAULT_L2_DESTINATION_ADDRESS;

  const { maxPriorityFeePerGas, maxFeePerGas } = await l1PublicClient.estimateFeesPerGas();

  logger.debug(`Fetched fee data. maxPriorityFeePerGas=${maxPriorityFeePerGas} maxFeePerGas=${maxFeePerGas}`);

  const txHash = await lineaRollup.write.sendMessage([destinationAddress, fee, calldata], {
    value,
    maxPriorityFeePerGas,
    maxFeePerGas,
  });

  logger.debug(`sendMessage transaction sent. transactionHash=${txHash}`);

  logger.debug(`Waiting for transaction to be mined... transactionHash=${txHash}`);
  const receipt = await l1PublicClient.waitForTransactionReceipt({ hash: txHash, timeout: timeoutMs });

  if (!receipt) {
    throw new Error(`Transaction receipt not received for hash: ${txHash}`);
  }

  logger.debug(`Transaction mined. transactionHash=${txHash} status=${receipt.status}`);

  return { txHash, receipt };
}

/**
 * Sends a message from L2 to L1 via the L2 Message Service contract.
 * @param params - Message sending parameters
 * @returns Transaction hash and receipt
 * @throws Error if transaction fails or receipt is not received
 */
export async function sendL2ToL1Message(params: SendMessageParams): Promise<SendMessageResult> {
  const {
    account,
    fee = etherToWei("0.0001"),
    value = etherToWei("0.001"),
    withCalldata = false,
    timeoutMs = DEFAULT_TRANSACTION_TIMEOUT_MS,
  } = params;

  const dummyContract = config.l1PublicClient().getDummyContract();
  const l2MessageService = config.l2WalletClient({ account }).getL2MessageServiceContract();
  const l2PublicClient = config.l2PublicClient({ type: L2RpcEndpoint.BesuNode });

  const calldata = generateCalldata(withCalldata);
  const destinationAddress = withCalldata ? dummyContract.address : DEFAULT_L1_DESTINATION_ADDRESS;

  const { maxPriorityFeePerGas, maxFeePerGas, gasLimit } = await estimateLineaGas(l2PublicClient, {
    account: account,
    to: l2MessageService.address,
    data: encodeFunctionCall({
      abi: L2MessageServiceV1Abi,
      functionName: "sendMessage",
      args: [destinationAddress, fee, calldata],
    }),
    value,
  });

  logger.debug(
    `Fetched fee data. maxPriorityFeePerGas=${maxPriorityFeePerGas} maxFeePerGas=${maxFeePerGas} gasLimit=${gasLimit}`,
  );

  const txHash = await l2MessageService.write.sendMessage([destinationAddress, fee, calldata], {
    value,
    maxPriorityFeePerGas,
    maxFeePerGas,
    gas: gasLimit,
  });

  logger.debug(`sendMessage transaction sent. transactionHash=${txHash}`);

  logger.debug(`Waiting for transaction to be mined... transactionHash=${txHash}`);
  const receipt = await l2PublicClient.waitForTransactionReceipt({ hash: txHash, timeout: timeoutMs });

  if (!receipt) {
    throw new Error(`Transaction receipt not received for hash: ${txHash}`);
  }

  logger.debug(`Transaction mined. transactionHash=${txHash} status=${receipt.status}`);

  return { txHash, receipt };
}
