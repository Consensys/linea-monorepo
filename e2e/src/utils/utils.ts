import { BlockTag } from "@ethersproject/providers";
import * as fs from "fs";
import assert from "assert";
import { Contract, ContractReceipt, PayableOverrides, Wallet, ethers } from "ethers";
import path from "path";
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

export function getWallet(privateKey: string, provider: ethers.providers.JsonRpcProvider) {
  return new ethers.Wallet(privateKey, provider);
}

export function encodeFunctionCall(contractInterface: ethers.utils.Interface, functionName: string, args: unknown[]) {
  return contractInterface.encodeFunctionData(functionName, args);
}

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

export async function getEvents<TContract extends LineaRollup | L2MessageService, TEvent extends TypedEvent>(
  contract: TContract,
  eventFilter: TypedEventFilter<TEvent>,
  fromBlock?: BlockTag,
  toBlock?: BlockTag,
): Promise<Array<TEvent>> {
  return contract.queryFilter(eventFilter, fromBlock, toBlock);
}

export async function waitForEvents<TContract extends LineaRollup | L2MessageService, TEvent extends TypedEvent>(
  contract: TContract,
  eventFilter: TypedEventFilter<TEvent>,
  pollingInterval: number,
  fromBlock?: BlockTag,
  toBlock?: BlockTag,
): Promise<TEvent[]> {
  let events = await getEvents(contract, eventFilter, fromBlock, toBlock);

  while (events.length === 0) {
    events = await getEvents(contract, eventFilter, fromBlock, toBlock);
    await wait(pollingInterval);
  }

  return events;
}

export async function waitForFile(
  directory: string,
  regex: RegExp,
  pollingInterval: number,
  timeout: number,
): Promise<string> {
  return new Promise((resolve, reject) => {
    const interval = setInterval(() => {
      fs.readdir(directory, (err, files) => {
        if (err) {
          clearInterval(interval);
          reject(new Error(`Error reading directory: ${err}`));
          return;
        }

        for (const file of files) {
          if (regex.test(file)) {
            clearInterval(interval);
            resolve(fs.readFileSync(path.join(directory, file), "utf-8"));
            return;
          }
        }
      });
    }, pollingInterval);

    setTimeout(() => {
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
