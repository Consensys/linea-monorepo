import * as fs from "fs";
import assert from "assert";
import { BaseContract, BlockTag, ContractTransactionReceipt, TransactionReceipt, Wallet, ethers } from "ethers";
import path from "path";
import { exec } from "child_process";
import { L2MessageService, LineaRollup } from "../typechain";
import { PayableOverrides, TypedContractEvent, TypedDeferredTopicFilter, TypedEventLog } from "../typechain/common";
import { MessageEvent, SendMessageArgs } from "./types";
import { getAndIncreaseFeeData } from "./helpers";

export const wait = (timeout: number) => new Promise((resolve) => setTimeout(resolve, timeout));

export function increaseDate(currentDate: Date, seconds: number): Date {
  const newDate = new Date(currentDate.getTime());
  newDate.setSeconds(newDate.getSeconds() + seconds);
  return newDate;
}

export const subtractSecondsToDate = (date: Date, seconds: number): Date => {
  const dateCopy = new Date(date);
  dateCopy.setSeconds(date.getSeconds() - seconds);
  return dateCopy;
};

export function getWallet(privateKey: string, provider: ethers.JsonRpcProvider) {
  return new ethers.Wallet(privateKey, provider);
}

export function encodeFunctionCall(contractInterface: ethers.Interface, functionName: string, args: unknown[]) {
  return contractInterface.encodeFunctionData(functionName, args);
}

export const generateKeccak256 = (types: string[], values: unknown[], packed?: boolean) =>
  ethers.keccak256(encodeData(types, values, packed));

export const encodeData = (types: string[], values: unknown[], packed?: boolean) => {
  if (packed) {
    return ethers.solidityPacked(types, values);
  }
  return ethers.AbiCoder.defaultAbiCoder().encode(types, values);
};

export class RollupGetZkEVMBlockNumberClient {
  private endpoint: URL;
  private request = {
    method: "post",
    body: JSON.stringify({
      jsonrpc: "2.0",
      method: "rollup_getZkEVMBlockNumber",
      params: [],
      id: 1,
    }),
  };

  public constructor(endpoint: URL) {
    this.endpoint = endpoint;
  }

  public async rollupGetZkEVMBlockNumber(): Promise<number> {
    const response = await fetch(this.endpoint, this.request);
    const data = await response.json();
    assert("result" in data);
    return Number.parseInt(data.result);
  }
}

export async function getBlockByNumberOrBlockTag(rpcUrl: URL, blockTag: BlockTag): Promise<ethers.Block | null> {
  const provider = new ethers.JsonRpcProvider(rpcUrl.href);
  try {
    const blockNumber = await provider.getBlock(blockTag);
    return blockNumber;
  } catch (error) {
    return null;
  }
}

export async function getEvents<TContract extends LineaRollup | L2MessageService, TEvent extends TypedContractEvent>(
  contract: TContract,
  eventFilter: TypedDeferredTopicFilter<TEvent>,
  fromBlock?: BlockTag,
  toBlock?: BlockTag,
  criteria?: (events: TypedEventLog<TEvent>[]) => Promise<TypedEventLog<TEvent>[]>,
): Promise<Array<TypedEventLog<TEvent>>> {
  const events = await contract.queryFilter(
    eventFilter,
    fromBlock as string | number | undefined,
    toBlock as string | number | undefined,
  );

  if (criteria) {
    return await criteria(events);
  }

  return events;
}

export async function waitForEvents<
  TContract extends LineaRollup | L2MessageService,
  TEvent extends TypedContractEvent,
>(
  contract: TContract,
  eventFilter: TypedDeferredTopicFilter<TEvent>,
  pollingInterval: number = 500,
  fromBlock?: BlockTag,
  toBlock?: BlockTag,
  criteria?: (events: TypedEventLog<TEvent>[]) => Promise<TypedEventLog<TEvent>[]>,
): Promise<TypedEventLog<TEvent>[]> {
  let events = await getEvents(contract, eventFilter, fromBlock, toBlock, criteria);

  while (events.length === 0) {
    events = await getEvents(contract, eventFilter, fromBlock, toBlock, criteria);
    await wait(pollingInterval);
  }

  return events;
}

export function getFiles(directory: string, fileRegex: RegExp[]): string[] {
  const files = fs.readdirSync(directory, { withFileTypes: true });
  const filteredFiles = files.filter((file) => fileRegex.map((regex) => regex.test(file.name)).includes(true));
  return filteredFiles.map((file) => fs.readFileSync(path.join(directory, file.name), "utf-8"));
}

export async function waitForFile(
  directory: string,
  regex: RegExp,
  pollingInterval: number,
  timeout: number,
  criteria?: (fileName: string) => boolean,
): Promise<string> {
  const endTime = Date.now() + timeout;

  while (Date.now() < endTime) {
    try {
      const files = fs.readdirSync(directory);

      for (const file of files) {
        if (regex.test(file) && (!criteria || criteria(file))) {
          const filePath = path.join(directory, file);
          const content = fs.readFileSync(filePath, "utf-8");
          return content;
        }
      }
    } catch (err) {
      throw new Error(`Error reading directory: ${(err as Error).message}`);
    }

    await new Promise((resolve) => setTimeout(resolve, pollingInterval));
  }

  throw new Error("File check timed out");
}

export async function sendXTransactions(
  signer: Wallet,
  transactionRequest: ethers.TransactionRequest,
  numberOfTransactions: number,
) {
  for (let i = 0; i < numberOfTransactions; i++) {
    const tx = await signer.sendTransaction(transactionRequest);
    await tx.wait();
  }
}

export async function sendTransactionsToGenerateTrafficWithInterval(signer: Wallet, pollingInterval: number = 1000) {
  const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(await signer.provider!.getFeeData());
  const transactionRequest = {
    to: signer.address,
    value: ethers.parseEther("0.000001"),
    maxPriorityFeePerGas: maxPriorityFeePerGas,
    maxFeePerGas: maxFeePerGas,
  };

  let timeoutId: NodeJS.Timeout | null = null;
  let isRunning = true;

  const sendTransaction = async () => {
    if (!isRunning) return;

    try {
      const tx = await signer.sendTransaction(transactionRequest);
      await tx.wait();

      if (isRunning) {
        timeoutId = setTimeout(sendTransaction, pollingInterval);
      }
    } catch (error) {
      console.error("Error sending transaction:", error);
      if (isRunning) {
        timeoutId = setTimeout(sendTransaction, pollingInterval);
      }
    }
  };

  const stop = () => {
    isRunning = false;
    if (timeoutId) {
      clearTimeout(timeoutId);
      timeoutId = null;
    }
    console.log("Transaction loop stopped.");
  };

  sendTransaction();

  return stop;
}

export function getMessageSentEventFromLogs<T extends BaseContract>(
  contract: T,
  receipts: TransactionReceipt[],
): MessageEvent[] {
  return receipts
    .flatMap((receipt) => receipt.logs)
    .filter((log) => log.topics[0] === "0xe856c2b8bd4eb0027ce32eeaf595c21b0b6b4644b326e5b7bd80a1cf8db72e6c")
    .map((log) => {
      const logDescription = contract.interface.parseLog(log);
      if (!logDescription) {
        throw new Error("Invalid log description");
      }
      const { args } = logDescription;
      return {
        from: args._from,
        to: args._to,
        fee: args._fee,
        value: args._value,
        messageNumber: args._nonce,
        calldata: args._calldata,
        messageHash: args._messageHash,
        blockNumber: log.blockNumber,
      };
    });
}

export async function waitForMessageAnchoring(
  contract: L2MessageService,
  messageHash: string,
  pollingInterval: number,
) {
  let messageStatus = await contract.inboxL1L2MessageStatus(messageHash);
  while (messageStatus === 0n) {
    messageStatus = await contract.inboxL1L2MessageStatus(messageHash);
    await wait(pollingInterval);
  }
}

export const sendMessage = async <T extends LineaRollup | L2MessageService>(
  signer: Wallet,
  contract: T,
  args: SendMessageArgs,
  overrides?: PayableOverrides,
): Promise<TransactionReceipt> => {
  const tx = await signer.sendTransaction({
    to: await contract.getAddress(),
    value: overrides?.value || 0n,
    data: contract.interface.encodeFunctionData("sendMessage", [args.to, args.fee, args.calldata]),
    ...overrides,
  });

  const receipt = await tx.wait();

  if (!receipt) {
    throw new Error("Transaction receipt is undefined");
  }
  return receipt;
};

export const sendMessagesForNSeconds = async <T extends LineaRollup | L2MessageService>(
  provider: ethers.JsonRpcProvider,
  signer: Wallet,
  contract: T,
  duration: number,
  args: SendMessageArgs,
  overrides?: PayableOverrides,
): Promise<ContractTransactionReceipt[]> => {
  let nonce = await provider.getTransactionCount(signer.address);

  const currentDate = new Date();
  const endDate = increaseDate(currentDate, duration);

  const sendMessagePromises: Promise<TransactionReceipt | null>[] = [];
  let currentTime = new Date().getTime();

  const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(await provider.getFeeData());
  while (currentTime < endDate.getTime()) {
    sendMessagePromises.push(
      sendMessage(signer, contract.connect(signer), args, {
        ...overrides,
        maxPriorityFeePerGas,
        maxFeePerGas,
        nonce,
      }).catch(() => {
        return null;
      }),
    );
    nonce++;

    if (sendMessagePromises.length % 10 === 0) {
      await wait(10_000);
    }
    currentTime = new Date().getTime();
  }

  const result = (await Promise.all(sendMessagePromises)).filter(
    (receipt) => receipt !== null,
  ) as ContractTransactionReceipt[];
  return result;
};

export async function execDockerCommand(command: string, containerName: string): Promise<string> {
  const dockerCommand = `docker ${command} ${containerName}`;
  console.log(`Executing: ${dockerCommand}...`);
  return new Promise((resolve, reject) => {
    exec(dockerCommand, (error, stdout, stderr) => {
      if (error) {
        console.error(`Error executing (${dockerCommand}): ${stderr}`);
        reject(error);
      }
      console.log(`Execution success (${dockerCommand}): ${stdout}`);
      resolve(stdout);
    });
  });
}