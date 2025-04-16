// import { LineaRollup, Mimc, TestEip1559RlpEncoder } from "contracts/typechain-types";
import { Mimc } from "contracts/typechain-types";
import { encodeData } from "contracts/common/helpers";
// import { Eip1559Transaction } from "../../common/types";

export const getExpectedL2BlockNumberForForcedTx = (params: {
  blockTimestamp: bigint;
  lastFinalizedBlockTimestamp: bigint;
  currentFinalizedL2BlockNumber: bigint;
  l2BlockBuffer: bigint;
}) => {
  const { blockTimestamp, lastFinalizedBlockTimestamp, currentFinalizedL2BlockNumber, l2BlockBuffer } = params;
  return currentFinalizedL2BlockNumber + blockTimestamp - lastFinalizedBlockTimestamp + l2BlockBuffer;
};

// TODO
// export const getForcedTransactionRollingHash = async (
//   mimcLibrary: Mimc,
//   lineaRollup: LineaRollup,
//   eip1559RlpEncoder: TestEip1559RlpEncoder,
//   eip1559Tx: Eip1559Transaction,
//   expectedBlockNumber: bigint,
//   from: string,
// ): Promise<string> => {
//   const { previousForcedTransactionRollingHash } = await lineaRollup.getNextForcedTransactionFields();
//   const { rlpEncodedTransaction } = await eip1559RlpEncoder.encodeEip1559Transaction(eip1559Tx);
//   // Strip `yParity`, `r` and `s` from the rlpEncodedTransaction
//   const rlpEncodedUnsignedTx = rlpEncodedTransaction.slice(0, -134);
// };

export const _computeForcedTransactionRollingHash = async (
  mimcLibrary: Mimc,
  previousRollingHash: string,
  hashedPayload: string,
  expectedBlockNumber: bigint,
  from: string,
): Promise<string> => {
  const mostSignificantBytesHashedPayload = hashedPayload.slice(0, 32 + 2);
  const leastSignificantBytesHashedPayload = `0x${hashedPayload.slice(32 + 2, 64 + 2)}`;
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
