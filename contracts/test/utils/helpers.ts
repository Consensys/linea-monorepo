import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import fs from "fs";
import { ethers } from "hardhat";
import path from "path";
import { DEFAULT_MESSAGE_NONCE, MESSAGE_FEE, MESSAGE_VALUE_1ETH } from "./constants";
import { DebugData, FormattedBlockData, RawBlockData } from "./types";

export const generateKeccak256Hash = (str: string) => ethers.utils.keccak256(ethers.utils.toUtf8Bytes(str));

export const generateNKeccak256Hashes = (str: string, numberOfHashToGenerate: number): string[] => {
  let arr: string[] = [];
  for (let i = 1; i < numberOfHashToGenerate + 1; i++) {
    arr = [...arr, ethers.utils.keccak256(ethers.utils.toUtf8Bytes(`${str}${i}`))];
  }
  return arr;
};

export async function encodeSendMessageLog(
  sender: SignerWithAddress,
  receiver: SignerWithAddress,
  messageHash: string,
  calldata: string,
) {
  const abiCoder = new ethers.utils.AbiCoder();
  const topic = ethers.utils.id("MessageSent(address,address,uint256,uint256,uint256,bytes,bytes32)");
  const data = abiCoder.encode(
    ["address", "address", "uint256", "uint256", "uint256", "bytes", "bytes32"],
    [sender.address, receiver.address, MESSAGE_FEE, MESSAGE_VALUE_1ETH, DEFAULT_MESSAGE_NONCE, calldata, messageHash],
  );

  return {
    topic,
    data,
  };
}

export async function encodeSendMessage(
  sender: string,
  receiver: string,
  fee: BigNumber,
  amount: BigNumber,
  salt: BigNumber,
  calldata: string,
) {
  const abiCoder = ethers.utils.defaultAbiCoder;
  const data = abiCoder.encode(
    ["address", "address", "uint256", "uint256", "uint256", "bytes"],
    [sender, receiver, fee, amount, salt, calldata],
  );

  return data;
}

export function getProverTestData(
  filename: string,
  folder: string,
): {
  blocks: FormattedBlockData[];
  proverMode: string;
  parentStateRootHash: string;
  version: string;
  firstBlockNumber: number;
  proof: string;
  debugData: DebugData;
} {
  const testFilePath = path.resolve(__dirname, "..", "testData", filename, folder);
  const testData = JSON.parse(fs.readFileSync(testFilePath, "utf8"));

  return {
    blocks: testData.blocksData.map((block: RawBlockData) => ({
      blockRootHash: block.rootHash,
      transactions: block.rlpEncodedTransactions,
      l2BlockTimestamp: block.timestamp,
      l2ToL1MsgHashes: block.l2ToL1MsgHashes,
      fromAddresses: block.fromAddresses,
      batchReceptionIndices: block.batchReceptionIndices,
    })),
    proverMode: testData.proverMode,
    parentStateRootHash: testData.parentStateRootHash,
    version: testData.version,
    firstBlockNumber: testData.firstBlockNumber,
    proof: testData.proof,
    debugData: testData.DebugData,
  };
}

export function getRLPEncodeTransactions(filename: string): {
  shortEip1559Transaction: string;
  eip1559Transaction: string;
  eip1559TransactionHashes: string[];
  legacyTransaction: string;
  legacyTransactionHashes: string[];
  eip2930Transaction: string;
  eip2930TransactionHashes: string[];
} {
  const testFilePath = path.resolve(__dirname, "..", "testData", filename);
  const testData = JSON.parse(fs.readFileSync(testFilePath, "utf8"));

  return {
    shortEip1559Transaction: testData.shortEip1559Transaction,
    eip1559Transaction: testData.eip1559Transaction,
    eip1559TransactionHashes: testData.eip1559TransactionHashes,
    legacyTransaction: testData.legacyTransaction,
    legacyTransactionHashes: testData.legacyTransactionHashes,
    eip2930Transaction: testData.eip2930Transaction,
    eip2930TransactionHashes: testData.eip2930TransactionHashes,
  };
}

export function getTransactionsToBeDecoded(blocks: FormattedBlockData[]): string[] {
  const txsToBeDecoded = [];
  for (let i = 0; i < blocks.length; i++) {
    for (let j = 0; j < blocks[i].batchReceptionIndices.length; j++) {
      const txIndex = blocks[i].batchReceptionIndices[j];
      txsToBeDecoded.push(blocks[i].transactions[txIndex]);
    }
  }
  return txsToBeDecoded;
}
