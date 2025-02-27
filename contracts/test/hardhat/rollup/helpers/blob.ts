import * as kzg from "c-kzg";
import { BaseContract, Contract, Transaction } from "ethers";
import { ethers } from "hardhat";
import path from "path";

import { getWalletForIndex } from "./";
import {
  expectEventDirectFromReceiptData,
  generateBlobDataSubmission,
  generateBlobDataSubmissionFromFile,
} from "../../common/helpers";
import { TestLineaRollup } from "../../../../typechain-types";

export async function sendBlobTransaction(
  lineaRollup: TestLineaRollup,
  startIndex: number,
  finalIndex: number,
  isMultiple: boolean = false,
) {
  const operatorHDSigner = getWalletForIndex(2);
  const lineaRollupAddress = await lineaRollup.getAddress();

  const {
    blobDataSubmission: blobSubmission,
    compressedBlobs: compressedBlobs,
    parentShnarf: parentShnarf,
    finalShnarf: finalShnarf,
  } = generateBlobDataSubmission(startIndex, finalIndex, isMultiple);

  const encodedCall = lineaRollup.interface.encodeFunctionData("submitBlobs", [
    blobSubmission,
    parentShnarf,
    finalShnarf,
  ]);

  const { maxFeePerGas, maxPriorityFeePerGas } = await ethers.provider.getFeeData();
  const nonce = await operatorHDSigner.getNonce();

  const transaction = Transaction.from({
    data: encodedCall,
    maxPriorityFeePerGas: maxPriorityFeePerGas!,
    maxFeePerGas: maxFeePerGas!,
    to: lineaRollupAddress,
    chainId: (await ethers.provider.getNetwork()).chainId,
    type: 3,
    nonce: nonce,
    value: 0,
    gasLimit: 5_000_000,
    kzg,
    maxFeePerBlobGas: 1n,
    blobs: compressedBlobs,
  });

  const signedTx = await operatorHDSigner.signTransaction(transaction);
  const txResponse = await ethers.provider.broadcastTransaction(signedTx);

  const receipt = await ethers.provider.getTransactionReceipt(txResponse.hash);

  const expectedEventArgs = [parentShnarf, finalShnarf, blobSubmission[blobSubmission.length - 1].finalStateRootHash];

  expectEventDirectFromReceiptData(lineaRollup as BaseContract, receipt!, "DataSubmittedV3", expectedEventArgs);
}

export async function sendBlobTransactionFromFile(
  lineaRollup: TestLineaRollup,
  filePath: string,
  betaV1LineaRollup: TestLineaRollup,
) {
  const operatorHDSigner = getWalletForIndex(2);
  const lineaRollupAddress = await betaV1LineaRollup.getAddress();

  const {
    blobDataSubmission: blobSubmission,
    compressedBlobs: compressedBlobs,
    parentShnarf: parentShnarf,
    finalShnarf: finalShnarf,
  } = generateBlobDataSubmissionFromFile(path.resolve(__dirname, "../../_testData/betaV1", filePath));

  const encodedCall = lineaRollup.interface.encodeFunctionData("submitBlobs", [
    blobSubmission,
    parentShnarf,
    finalShnarf,
  ]);

  const { maxFeePerGas, maxPriorityFeePerGas } = await ethers.provider.getFeeData();
  const nonce = await operatorHDSigner.getNonce();

  const transaction = Transaction.from({
    data: encodedCall,
    maxPriorityFeePerGas: maxPriorityFeePerGas!,
    maxFeePerGas: maxFeePerGas!,
    to: lineaRollupAddress,
    chainId: (await ethers.provider.getNetwork()).chainId,
    type: 3,
    nonce: nonce,
    value: 0,
    gasLimit: 5_000_000,
    kzg,
    maxFeePerBlobGas: 1n,
    blobs: compressedBlobs,
  });

  const signedTx = await operatorHDSigner.signTransaction(transaction);
  const txResponse = await ethers.provider.broadcastTransaction(signedTx);
  const receipt = await ethers.provider.getTransactionReceipt(txResponse.hash);
  const expectedEventArgs = [parentShnarf, finalShnarf, blobSubmission[blobSubmission.length - 1].finalStateRootHash];

  expectEventDirectFromReceiptData(lineaRollup as BaseContract, receipt!, "DataSubmittedV3", expectedEventArgs);
}

export async function sendBlobTransactionViaCallForwarder(
  lineaRollupUpgraded: Contract,
  startIndex: number,
  finalIndex: number,
  callforwarderAddress: string,
) {
  const operatorHDSigner = getWalletForIndex(2);

  const {
    blobDataSubmission: blobSubmission,
    compressedBlobs: compressedBlobs,
    parentShnarf: parentShnarf,
    finalShnarf: finalShnarf,
  } = generateBlobDataSubmission(startIndex, finalIndex, false);

  const encodedCall = lineaRollupUpgraded.interface.encodeFunctionData("submitBlobs", [
    blobSubmission,
    parentShnarf,
    finalShnarf,
  ]);

  const { maxFeePerGas, maxPriorityFeePerGas } = await ethers.provider.getFeeData();
  const nonce = await operatorHDSigner.getNonce();

  const transaction = Transaction.from({
    data: encodedCall,
    maxPriorityFeePerGas: maxPriorityFeePerGas!,
    maxFeePerGas: maxFeePerGas!,
    to: callforwarderAddress,
    chainId: (await ethers.provider.getNetwork()).chainId,
    type: 3,
    nonce: nonce,
    value: 0,
    gasLimit: 5_000_000,
    kzg,
    maxFeePerBlobGas: 1n,
    blobs: compressedBlobs,
  });

  const signedTx = await operatorHDSigner.signTransaction(transaction);
  const txResponse = await ethers.provider.broadcastTransaction(signedTx);
  const receipt = await ethers.provider.getTransactionReceipt(txResponse.hash);

  const expectedEventArgs = [parentShnarf, finalShnarf, blobSubmission[blobSubmission.length - 1].finalStateRootHash];

  expectEventDirectFromReceiptData(lineaRollupUpgraded as BaseContract, receipt!, "DataSubmittedV3", expectedEventArgs);
}
