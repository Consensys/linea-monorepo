import { time as networkTime } from "@nomicfoundation/hardhat-network-helpers";
import { LineaRollup, Mimc } from "contracts/typechain-types";
import { encodeData } from "contracts/common/helpers";
import { Eip1559Transaction } from "../../common/types";
import { THREE_DAYS_IN_SECONDS } from "../../common/constants";

const _getExpectedL2BlockNumberForForcedTx = (params: {
  blockTimestamp: bigint;
  lastFinalizedBlockTimestamp: bigint;
  currentFinalizedL2BlockNumber: bigint;
  l2BlockBuffer: bigint;
}) => {
  const { blockTimestamp, lastFinalizedBlockTimestamp, currentFinalizedL2BlockNumber, l2BlockBuffer } = params;
  return currentFinalizedL2BlockNumber + blockTimestamp - lastFinalizedBlockTimestamp + l2BlockBuffer;
};

export const setNextExpectedL2BlockNumberForForcedTx = async (
  lineaRollup: LineaRollup,
  nextNetworkTimestamp: bigint,
  lastFinalizedBlockTimestamp: bigint,
) => {
  await networkTime.setNextBlockTimestamp(nextNetworkTimestamp);
  const lastFinalizedBlock = await lineaRollup.currentL2BlockNumber();
  const expectedBlockNumber = _getExpectedL2BlockNumberForForcedTx({
    blockTimestamp: nextNetworkTimestamp,
    l2BlockBuffer: BigInt(THREE_DAYS_IN_SECONDS),
    currentFinalizedL2BlockNumber: lastFinalizedBlock,
    lastFinalizedBlockTimestamp: lastFinalizedBlockTimestamp,
  });
  return expectedBlockNumber;
};

const _computeForcedTransactionRollingHash = async (
  mimcLibrary: Mimc,
  previousRollingHash: string,
  hashedPayload: string,
  expectedBlockNumber: bigint,
  from: string,
): Promise<string> => {
  const mimcPayload = encodeData(
    ["bytes32", "bytes32", "uint256", "address"],
    [previousRollingHash, hashedPayload, expectedBlockNumber, from],
  );
  return await mimcLibrary.hash(mimcPayload);
};

const _computeMimcPayloadHash = async (
  mimcLibrary: Mimc,
  eip1559Tx: Eip1559Transaction,
  chainId: bigint,
): Promise<string> => {
  const types = [
    "tuple(" +
      "uint256," +
      "uint256," +
      "uint256," +
      "uint256," +
      "uint256," +
      "address," +
      "uint256," +
      "bytes," +
      "tuple(address,bytes32[])[]" +
      ")",
  ];

  const values = [
    [
      chainId,
      eip1559Tx.nonce,
      eip1559Tx.maxPriorityFeePerGas,
      eip1559Tx.maxFeePerGas,
      eip1559Tx.gasLimit,
      eip1559Tx.to,
      eip1559Tx.value,
      eip1559Tx.input,
      eip1559Tx.accessList,
    ],
  ];

  const mimcPayload = encodeData(types, values);

  return await mimcLibrary.hash("0x" + mimcPayload.slice(66)); // stripped out the first offset
};

export const getForcedTransactionRollingHash = async (
  mimcLibrary: Mimc,
  lineaRollup: LineaRollup,
  eip1559Tx: Eip1559Transaction,
  expectedBlockNumber: bigint,
  from: string,
  chainId: bigint,
): Promise<string> => {
  const { previousForcedTransactionRollingHash } = await lineaRollup.getRequiredForcedTransactionFields();

  const hashedPayload = await _computeMimcPayloadHash(mimcLibrary, eip1559Tx, chainId);

  return await _computeForcedTransactionRollingHash(
    mimcLibrary,
    previousForcedTransactionRollingHash,
    hashedPayload,
    expectedBlockNumber,
    from,
  );
};
