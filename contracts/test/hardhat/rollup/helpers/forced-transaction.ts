import { time as networkTime } from "@nomicfoundation/hardhat-network-helpers";
import { LineaRollup, Mimc, TestEip1559RlpEncoder } from "contracts/typechain-types";
import { encodeData, generateKeccak256 } from "contracts/common/helpers";
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
  const mostSignificantBytesHashedPayload = "0x" + hashedPayload.slice(2, 32 + 2).padStart(64, "0");
  const leastSignificantBytesHashedPayload = "0x" + hashedPayload.slice(32 + 2, 64 + 2).padStart(64, "0");
  const mimcPayload = encodeData(
    ["bytes32", "bytes32", "bytes32", "uint256", "address"],
    [
      previousRollingHash,
      mostSignificantBytesHashedPayload,
      leastSignificantBytesHashedPayload,
      expectedBlockNumber,
      from,
    ],
  );
  return await mimcLibrary.hash(mimcPayload);
};

export const getForcedTransactionRollingHash = async (
  mimcLibrary: Mimc,
  lineaRollup: LineaRollup,
  eip1559RlpEncoder: TestEip1559RlpEncoder,
  eip1559Tx: Eip1559Transaction,
  expectedBlockNumber: bigint,
  from: string,
): Promise<string> => {
  const { previousForcedTransactionRollingHash } = await lineaRollup.getRequiredForcedTransactionFields();
  const { rlpEncodedTransaction } = await eip1559RlpEncoder.encodeEip1559Transaction(eip1559Tx);
  // Strip `yParity`, `r` and `s` from the rlpEncodedTransaction
  const rlpEncodedTransactionWithoutSignature = rlpEncodedTransaction.slice(0, -134);
  // Get length byte and subtract by 67 (bytes used by signature)
  const lengthByte = rlpEncodedTransactionWithoutSignature.slice(6, 8);
  const lengthWithSignature = parseInt(lengthByte, 16);
  const lengthWithoutSignature = lengthWithSignature - 67;
  const lengthByteWithoutSignature = lengthWithoutSignature.toString(16);
  const rlpEncodedUnsignedTransaction =
    rlpEncodedTransactionWithoutSignature.slice(0, 6) +
    lengthByteWithoutSignature +
    rlpEncodedTransactionWithoutSignature.slice(8);
  const hashedPayload = generateKeccak256(["bytes"], [rlpEncodedUnsignedTransaction], { encodePacked: true });
  return await _computeForcedTransactionRollingHash(
    mimcLibrary,
    previousForcedTransactionRollingHash,
    hashedPayload,
    expectedBlockNumber,
    from,
  );
};
