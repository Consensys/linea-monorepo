import { AbiCoder, BytesLike, Transaction, Wallet, ethers } from "ethers";
import { commitmentsToVersionedHashes } from "@ethereumjs/util";
import * as kzg from "c-kzg";
import submissionDataJson2 from "../SixInOne/blocks-1-46.json";
import submissionDataJson from "./blocks-47-81.json";
import submissionDataJson3 from "./blocks-82-114.json";
import aggregateProof1to114 from "./aggregatedProof-1-114.json";
import { DataHexString } from "ethers/lib.commonjs/utils/data";

export function generateKeccak256(types: string[], values: unknown[], packed?: boolean) {
  return ethers.keccak256(encodeData(types, values, packed));
}

export const encodeData = (types: string[], values: unknown[], packed?: boolean) => {
  if (packed) {
    return ethers.solidityPacked(types, values);
  }
  return AbiCoder.defaultAbiCoder().encode(types, values);
};

type SubmissionData = {
  parentStateRootHash: string;
  dataParentHash: string;
  finalStateRootHash: string;
  firstBlockInData: bigint;
  finalBlockInData: bigint;
  snarkHash: string;
};

export type ParentSubmissionData = {
  finalStateRootHash: string;
  firstBlockInData: bigint;
  finalBlockInData: bigint;
  shnarf: string;
  dataParentHash: string;
};

export function generateSubmissionDataFromJSON(
  startingBlockNumber: number,
  endingBlockNumber: number,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  parsedJSONData: any,
): { submissionData: SubmissionData; blob: Uint8Array } {
  const returnData = {
    parentStateRootHash: parsedJSONData.parentStateRootHash,
    dataParentHash: parsedJSONData.parentDataHash,
    finalStateRootHash: parsedJSONData.finalStateRootHash,
    firstBlockInData: BigInt(startingBlockNumber),
    finalBlockInData: BigInt(endingBlockNumber),
    snarkHash: parsedJSONData.snarkHash,
  };

  return {
    submissionData: returnData,
    blob: ethers.decodeBase64(parsedJSONData.compressedData),
  };
}

export function generateParentSubmissionData(finalStateRootHash: string, parentDataHash: string): ParentSubmissionData {
  return {
    finalStateRootHash: finalStateRootHash,
    firstBlockInData: BigInt(0),
    finalBlockInData: BigInt(0),
    shnarf: ethers.keccak256(ethers.toUtf8Bytes(ethers.ZeroHash)),
    dataParentHash: parentDataHash,
  };
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function generateParentSubmissionDataFromJson(parsedJSONData: any): ParentSubmissionData {
  return {
    finalStateRootHash: parsedJSONData.finalStateRootHash,
    firstBlockInData: BigInt(parsedJSONData.conflationOrder.startingBlockNumber),
    finalBlockInData: BigInt(parsedJSONData.conflationOrder.upperBoundaries.slice(-1)[0]),
    shnarf: parsedJSONData.expectedShnarf,
    dataParentHash: parsedJSONData.parentDataHash,
  };
}

export function generateSubmissionCallDataFromJSON(
  startingBlockNumber: number,
  endingBlockNumber: number,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  parsedJSONData: any,
): { submissionData: SubmissionData } {
  const returnData = {
    parentStateRootHash: parsedJSONData.parentStateRootHash,
    dataParentHash: parsedJSONData.parentDataHash,
    finalStateRootHash: parsedJSONData.finalStateRootHash,
    firstBlockInData: BigInt(startingBlockNumber),
    finalBlockInData: BigInt(endingBlockNumber),
    snarkHash: parsedJSONData.snarkHash,
    compressedData: ethers.hexlify(ethers.decodeBase64(parsedJSONData.compressedData)),
  };

  return {
    submissionData: returnData,
  };
}

function getPadded(data: Uint8Array): Uint8Array {
  const pdata = new Uint8Array(131072).fill(0);
  pdata.set(data);
  return pdata;
}

function requireEnv(name: string): string {
  const envVariable = process.env[name];
  if (!envVariable) {
    throw new Error(`Missing ${name} environment variable`);
  }

  return envVariable;
}

function kzgCommitmentsToVersionedHashes(commitments: Uint8Array[]): string[] {
  const versionedHashes = commitmentsToVersionedHashes(commitments);
  return versionedHashes.map((versionedHash) => ethers.hexlify(versionedHash));
}

async function main() {
  const rpcUrl = requireEnv("RPC_URL");
  const privateKey = requireEnv("PRIVATE_KEY");
  const destinationAddress = requireEnv("DESTINATION_ADDRESS");

  const provider = new ethers.JsonRpcProvider(rpcUrl);
  const wallet = new Wallet(privateKey, provider);

  kzg.loadTrustedSetup(`${__dirname}/../trusted_setup.txt`);

  const { submissionData: submissionData2, blob: blob2 } = generateSubmissionDataFromJSON(1, 46, submissionDataJson2);
  const parentSubmissionData1 = generateParentSubmissionData(
    "0x072ead6777750dc20232d1cee8dc9a395c2d350df4bbaa5096c6f59b214dcecd",
    ethers.ZeroHash,
  );

  const { submissionData } = generateSubmissionCallDataFromJSON(47, 81, submissionDataJson);
  const parentSubmissionData2 = generateParentSubmissionDataFromJson(submissionDataJson2);

  const { submissionData: submissionData3, blob: blob3 } = generateSubmissionDataFromJSON(82, 114, submissionDataJson3);
  const parentSubmissionData3 = generateParentSubmissionDataFromJson(submissionDataJson);

  const finalSubmissionData = generateParentSubmissionDataFromJson(submissionDataJson3);

  const fullblob2 = getPadded(blob2);
  const fullblob3 = getPadded(blob3);

  const commitment2 = kzg.blobToKzgCommitment(fullblob2);
  const commitment3 = kzg.blobToKzgCommitment(fullblob3);

  const [versionedHash2, versionedHash3] = kzgCommitmentsToVersionedHashes([commitment2, commitment3]);

  let encodedCall = encodeCall(submissionData2, parentSubmissionData1, [commitment2], submissionDataJson2);
  await submitBlob(provider, wallet, encodedCall, destinationAddress, [versionedHash2], fullblob2);

  await submitCalldata(submissionData, parentSubmissionData2);

  encodedCall = encodeCall(submissionData3, parentSubmissionData3, [commitment3], submissionDataJson3);

  await submitBlob(provider, wallet, encodedCall, destinationAddress, [versionedHash3], fullblob3);

  await sendProof(aggregateProof1to114, parentSubmissionData1, finalSubmissionData);
}

function encodeCall(
  submissionData: SubmissionData,
  parentSubmissionData: ParentSubmissionData,
  commitments: BytesLike[],
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  submissionDataJson: any,
): DataHexString {
  const encodedCall = ethers.concat([
    "0xf82c988c",
    ethers.AbiCoder.defaultAbiCoder().encode(
      [
        "tuple(bytes32,bytes32,bytes32,uint256,uint256,bytes32)",
        "tuple(bytes32,uint256,uint256,bytes32,bytes32)",
        "uint256",
        "bytes",
        "bytes",
      ],
      [
        [
          submissionData.parentStateRootHash,
          submissionData.dataParentHash,
          submissionData.finalStateRootHash,
          submissionData.firstBlockInData,
          submissionData.finalBlockInData,
          submissionData.snarkHash,
        ],
        [
          parentSubmissionData.finalStateRootHash,
          parentSubmissionData.firstBlockInData,
          parentSubmissionData.finalBlockInData,
          parentSubmissionData.shnarf,
          parentSubmissionData.dataParentHash,
        ],
        submissionDataJson.expectedY,
        commitments[0],
        submissionDataJson.kzgProofContract,
      ],
    ),
  ]);

  return encodedCall;
}

async function submitBlob(
  provider: ethers.JsonRpcProvider,
  wallet: Wallet,
  encodedCall: string,
  destinationAddress: string,
  versionedHashes: string[],
  fullblob: Uint8Array,
) {
  const { maxFeePerGas, maxPriorityFeePerGas } = await provider.getFeeData();
  const nonce = await provider.getTransactionCount(wallet.address);

  const transaction = Transaction.from({
    data: encodedCall,
    maxPriorityFeePerGas: maxPriorityFeePerGas!,
    maxFeePerGas: maxFeePerGas!,
    to: destinationAddress,
    chainId: 31648428,
    type: 3,
    nonce,
    value: 0,
    kzg,
    blobs: [fullblob],
    gasLimit: 5_000_000,
    blobVersionedHashes: versionedHashes,
    maxFeePerBlobGas: maxFeePerGas!,
  });

  const tx = await wallet.sendTransaction(transaction);
  const receipt = await tx.wait();
  console.log({ transaction: tx, receipt });
}

async function sendProof(
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  proofFile: any,
  submissionData: ParentSubmissionData,
  finalSubmissionData: ParentSubmissionData,
) {
  console.log("proof");

  const rpcUrl = requireEnv("RPC_URL");
  const privateKey = requireEnv("PRIVATE_KEY");
  const destinationAddress = requireEnv("DESTINATION_ADDRESS");

  const provider = new ethers.JsonRpcProvider(rpcUrl);
  const wallet = new Wallet(privateKey, provider);

  const xxx = [
    proofFile.aggregatedProof,
    0,
    [
      proofFile.parentStateRootHash,
      submissionData.shnarf,
      proofFile.dataHashes.slice(-1)[0],
      [
        finalSubmissionData.finalStateRootHash,
        finalSubmissionData.firstBlockInData,
        finalSubmissionData.finalBlockInData,
        finalSubmissionData.shnarf,
        finalSubmissionData.dataParentHash,
      ],
      proofFile.parentAggregationLastBlockTimestamp,
      proofFile.finalTimestamp,
      proofFile.l1RollingHash,
      proofFile.l1RollingHashMessageNumber,
      proofFile.l2MerkleRoots,
      proofFile.l2MerkleTreesDepth,
      proofFile.l2MessagingBlocksOffsets,
    ],
  ];

  console.log(xxx);

  const encodedCall = ethers.concat([
    "0x727073c8",
    ethers.AbiCoder.defaultAbiCoder().encode(
      [
        "bytes",
        "uint256",
        "tuple(bytes32,bytes32,bytes32,tuple(bytes32,uint256,uint256,bytes32,bytes32),uint256,uint256,bytes32,uint256,bytes32[],uint256,bytes)",
      ],
      [
        proofFile.aggregatedProof,
        0,
        [
          proofFile.parentStateRootHash,
          submissionData.shnarf,
          proofFile.dataHashes.slice(-1)[0],
          [
            finalSubmissionData.finalStateRootHash,
            finalSubmissionData.firstBlockInData,
            finalSubmissionData.finalBlockInData,
            finalSubmissionData.shnarf,
            finalSubmissionData.dataParentHash,
          ],
          proofFile.parentAggregationLastBlockTimestamp,
          proofFile.finalTimestamp,
          proofFile.l1RollingHash,
          proofFile.l1RollingHashMessageNumber,
          proofFile.l2MerkleRoots,
          proofFile.l2MerkleTreesDepth,
          proofFile.l2MessagingBlocksOffsets,
        ],
      ],
    ),
  ]);

  const { maxFeePerGas, maxPriorityFeePerGas } = await provider.getFeeData();
  const nonce = await provider.getTransactionCount(wallet.address);

  const transaction = Transaction.from({
    data: encodedCall,
    maxPriorityFeePerGas: maxPriorityFeePerGas!,
    maxFeePerGas: maxFeePerGas!,
    to: destinationAddress,
    chainId: 31648428,
    nonce,
    value: 0,
    gasLimit: 5_000_000,
  });

  const tx = await wallet.sendTransaction(transaction);
  const receipt = await tx.wait();
  console.log({ transaction: tx, receipt });
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
async function submitCalldata(calldata: any, parentSubmissionData: ParentSubmissionData) {
  const rpcUrl = requireEnv("RPC_URL");
  const privateKey = requireEnv("PRIVATE_KEY");
  const destinationAddress = requireEnv("DESTINATION_ADDRESS");

  const provider = new ethers.JsonRpcProvider(rpcUrl);
  const wallet = new Wallet(privateKey, provider);

  const encodedCall = ethers.concat([
    "0xc6251c58",
    ethers.AbiCoder.defaultAbiCoder().encode(
      [
        "tuple(bytes32,bytes32,bytes32,uint256,uint256,bytes32,bytes)",
        "tuple(bytes32,uint256,uint256,bytes32,bytes32)",
      ],
      [
        [
          calldata.parentStateRootHash,
          calldata.dataParentHash,
          calldata.finalStateRootHash,
          calldata.firstBlockInData,
          calldata.finalBlockInData,
          calldata.snarkHash,
          calldata.compressedData,
        ],
        [
          parentSubmissionData.finalStateRootHash,
          parentSubmissionData.firstBlockInData,
          parentSubmissionData.finalBlockInData,
          parentSubmissionData.shnarf,
          parentSubmissionData.dataParentHash,
        ],
      ],
    ),
  ]);

  const { maxFeePerGas, maxPriorityFeePerGas } = await provider.getFeeData();
  const nonce = await provider.getTransactionCount(wallet.address);

  const transaction = Transaction.from({
    data: encodedCall,
    maxPriorityFeePerGas: maxPriorityFeePerGas!,
    maxFeePerGas: maxFeePerGas!,
    to: destinationAddress,
    chainId: 31648428,
    nonce,
    value: 0,
    gasLimit: 5_000_000,
  });

  const tx = await wallet.sendTransaction(transaction);
  const receipt = await tx.wait();
  console.log({ transaction: tx, receipt });
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
