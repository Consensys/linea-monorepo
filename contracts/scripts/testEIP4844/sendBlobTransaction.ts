import { BytesLike, Transaction, Wallet, ethers } from "ethers";
import { commitmentsToVersionedHashes } from "@ethereumjs/util";
import * as kzg from "c-kzg";
import submissionDataJson from "./blocks-1-46.json";
import submissionDataJson2 from "./blocks-47-81.json";
import submissionDataJson3 from "./blocks-82-119.json";
import submissionDataJson4 from "./blocks-120-153.json";
import aggregateProof1to81 from "./aggregatedProof-1-81.json";
import aggregateProof82to153 from "./aggregatedProof-82-153.json";
import { DataHexString } from "ethers/lib.commonjs/utils/data";

type SubmissionData = {
  parentStateRootHash: string;
  dataParentHash: string;
  finalStateRootHash: string;
  firstBlockInData: bigint;
  finalBlockInData: bigint;
  snarkHash: string;
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

  kzg.loadTrustedSetup(`${__dirname}/trusted_setup.txt`);

  const { submissionData, blob } = generateSubmissionDataFromJSON(1, 46, submissionDataJson);
  const { submissionData: submissionData2, blob: blob2 } = generateSubmissionDataFromJSON(47, 81, submissionDataJson2);
  const { submissionData: submissionData3, blob: blob3 } = generateSubmissionDataFromJSON(82, 119, submissionDataJson3);
  const { submissionData: submissionData4, blob: blob4 } = generateSubmissionDataFromJSON(
    120,
    153,
    submissionDataJson4,
  );

  const commitments = kzg.blobToKzgCommitment(blob);
  const commitments2 = kzg.blobToKzgCommitment(blob2);
  const commitments3 = kzg.blobToKzgCommitment(blob3);
  const commitments4 = kzg.blobToKzgCommitment(blob4);

  const [versionedHashes, versionedHashes2, versionedHashes3, versionedHashes4] = kzgCommitmentsToVersionedHashes([
    commitments,
    commitments2,
    commitments3,
    commitments4,
  ]);

  let encodedCall = encodeCall(submissionData, [commitments], submissionDataJson);
  await submitBlob(provider, wallet, encodedCall, destinationAddress, [versionedHashes], blob);

  encodedCall = encodeCall(submissionData2, [commitments2], submissionDataJson2);
  await submitBlob(provider, wallet, encodedCall, destinationAddress, [versionedHashes2], blob2);

  encodedCall = encodeCall(submissionData3, [commitments3], submissionDataJson3);
  await submitBlob(provider, wallet, encodedCall, destinationAddress, [versionedHashes3], blob3);

  encodedCall = encodeCall(submissionData4, [commitments4], submissionDataJson4);
  await submitBlob(provider, wallet, encodedCall, destinationAddress, [versionedHashes4], blob4);

  await sendProof(aggregateProof1to81);
  await sendProof(aggregateProof82to153);
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function encodeCall(submissionData: SubmissionData, commitments: BytesLike[], submissionDataJson: any): DataHexString {
  const encodedCall = ethers.concat([
    "0x2d3c12e5",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["tuple(bytes32,bytes32,bytes32,uint256,uint256,bytes32)", "uint256", "bytes", "bytes"],
      [
        [
          submissionData.parentStateRootHash,
          submissionData.dataParentHash,
          submissionData.finalStateRootHash,
          submissionData.firstBlockInData,
          submissionData.finalBlockInData,
          submissionData.snarkHash,
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
    gasLimit: 5_000_000,
    kzg,
    blobs: [fullblob],
    blobVersionedHashes: versionedHashes,
    maxFeePerBlobGas: maxFeePerGas!,
  });

  const tx = await wallet.sendTransaction(transaction);
  const receipt = await tx.wait();
  console.log(receipt);
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
async function sendProof(proofFile: any) {
  const rpcUrl = requireEnv("RPC_URL");
  const privateKey = requireEnv("PRIVATE_KEY");
  const destinationAddress = requireEnv("DESTINATION_ADDRESS");

  const provider = new ethers.JsonRpcProvider(rpcUrl);
  const wallet = new Wallet(privateKey, provider);

  const encodedCall = ethers.concat([
    "0xd630280f",
    ethers.AbiCoder.defaultAbiCoder().encode(
      [
        "bytes",
        "uint256",
        "tuple(bytes32,bytes32[],bytes32,uint256,uint256,uint256,bytes32,uint256,bytes32[],uint256,bytes)",
      ],
      [
        proofFile.aggregatedProof,
        0,
        [
          proofFile.parentStateRootHash,
          proofFile.dataHashes,
          proofFile.dataParentHash,
          proofFile.finalBlockNumber,
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
  console.log({ transaction: tx, receipt: receipt });
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
