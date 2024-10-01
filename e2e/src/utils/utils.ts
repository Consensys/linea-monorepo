import { BlockTag } from "@ethersproject/providers";
import * as fs from "fs";
import assert from "assert";
import {Contract, ContractReceipt, PayableOverrides, Wallet, ethers} from "ethers";
import path from "path";
import { exec } from "child_process";
import { L2MessageService, LineaRollup } from "../typechain";
import { TypedEvent, TypedEventFilter } from "../typechain/common";
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

export function getWallet(privateKey: string, provider: ethers.providers.JsonRpcProvider) {
  return new ethers.Wallet(privateKey, provider);
}

export function encodeFunctionCall(contractInterface: ethers.utils.Interface, functionName: string, args: unknown[]) {
  return contractInterface.encodeFunctionData(functionName, args);
}

export const generateKeccak256 = (types: string[], values: unknown[], packed?: boolean) =>
  ethers.utils.keccak256(encodeData(types, values, packed));

export const encodeData = (types: string[], values: unknown[], packed?: boolean) => {
  if (packed) {
    return ethers.utils.solidityPack(types, values);
  }
  return ethers.utils.defaultAbiCoder.encode(types, values);
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

export async function getBlockByNumberOrBlockTag(
  rpcUrl: URL,
  blockTag: BlockTag
): Promise<ethers.providers.Block> {
  const provider = new ethers.providers.JsonRpcProvider(rpcUrl.href);
  return provider.getBlock(blockTag)
}

export async function getEvents<TContract extends LineaRollup | L2MessageService, TEvent extends TypedEvent>(
  contract: TContract,
  eventFilter: TypedEventFilter<TEvent>,
  fromBlock?: BlockTag,
  toBlock?: BlockTag,
  criteria?: (events: TEvent[]) => Promise<TEvent[]>,
): Promise<Array<TEvent>> {
  const events = await contract.queryFilter(eventFilter, fromBlock, toBlock);

  if (criteria) {
    return await criteria(events);
  }

  return events;
}

export async function waitForEvents<TContract extends LineaRollup | L2MessageService, TEvent extends TypedEvent>(
  contract: TContract,
  eventFilter: TypedEventFilter<TEvent>,
  pollingInterval: number = 500,
  fromBlock?: BlockTag,
  toBlock?: BlockTag,
  criteria?: (events: TEvent[]) => Promise<TEvent[]>,
): Promise<TEvent[]> {
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
  return new Promise((resolve, reject) => {
    const interval = setInterval(() => {
      fs.readdir(directory, (err, files) => {
        if (err) {
          clearInterval(interval);
          clearTimeout(timeoutId);
          reject(new Error(`Error reading directory: ${err}`));
          return;
        }

        for (const file of files) {
          if (regex.test(file) && (!criteria || criteria(file))) {
            clearInterval(interval);
            clearTimeout(timeoutId);
            resolve(fs.readFileSync(path.join(directory, file), "utf-8"));
            return;
          }
        }
      });
    }, pollingInterval);

    const timeoutId = setTimeout(() => {
      clearInterval(interval);
      reject(new Error("File check timed out"));
    }, timeout);
  });
}

export function sendTransactionsWithInterval(
  signer: Wallet,
  transactionRequest: ethers.providers.TransactionRequest,
  pollingInterval: number,
) {
  return setInterval(async function () {
    const tx = await signer.sendTransaction(transactionRequest);
    await tx.wait();
  }, pollingInterval);
}

export async function sendXTransactions(
  signer: Wallet,
  transactionRequest: ethers.providers.TransactionRequest,
  numberOfTransactions: number,
) {
  for (let i = 0; i < numberOfTransactions; i++) {
    const tx = await signer.sendTransaction(transactionRequest);
    await tx.wait();
  }
}

export async function sendTransactionsToGenerateTrafficWithInterval(
  signer: Wallet,
  pollingInterval: number = 1000,
) {
  const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(await signer.provider.getFeeData());
  const transactionRequest = {
    to: signer.address,
    value: ethers.utils.parseEther("0.000001"),
    maxPriorityFeePerGas: maxPriorityFeePerGas,
    maxFeePerGas: maxFeePerGas,
  }

  return setInterval(async function () {
    const tx = await signer.sendTransaction(transactionRequest);
    await tx.wait();
  }, pollingInterval);
}

export function getMessageSentEventFromLogs<T extends Contract>(
  contract: T,
  receipts: ContractReceipt[],
): MessageEvent[] {
  return receipts
    .flatMap((receipt) => receipt.logs)
    .filter((log) => log.topics[0] === "0xe856c2b8bd4eb0027ce32eeaf595c21b0b6b4644b326e5b7bd80a1cf8db72e6c")
    .map((log) => {
      const { args } = contract.interface.parseLog(log);

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
  while (messageStatus.toNumber() === 0) {
    messageStatus = await contract.inboxL1L2MessageStatus(messageHash);
    await wait(pollingInterval);
  }
}

export const sendMessage = async <T extends Contract>(
  contract: T,
  args: SendMessageArgs,
  overrides?: PayableOverrides,
): Promise<ContractReceipt> => {
  const tx = await contract.sendMessage(args.to, args.fee, args.calldata, overrides);
  return await tx.wait();
};

export const sendMessagesForNSeconds = async <T extends Contract>(
  provider: ethers.providers.JsonRpcProvider,
  signer: Wallet,
  contract: T,
  duration: number,
  args: SendMessageArgs,
  overrides?: PayableOverrides,
): Promise<ContractReceipt[]> => {
  let nonce = await provider.getTransactionCount(signer.address);

  const currentDate = new Date();
  const endDate = increaseDate(currentDate, duration);

  const sendMessagePromises: Promise<ContractReceipt | null>[] = [];
  let currentTime = new Date().getTime();

  const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(await provider.getFeeData());
  while (currentTime < endDate.getTime()) {
    sendMessagePromises.push(
      sendMessage(contract.connect(signer), args, {
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

  const result = (await Promise.all(sendMessagePromises)).filter((receipt) => receipt !== null) as ContractReceipt[];
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
