export const getExpectedL2BlockNumberForForcedTx = (params: {
  blockTimestamp: bigint;
  lastFinalizedBlockTimestamp: bigint;
  currentFinalizedL2BlockNumber: bigint;
  l2BlockBuffer: bigint;
}) => {
  const { blockTimestamp, lastFinalizedBlockTimestamp, currentFinalizedL2BlockNumber, l2BlockBuffer } = params;
  return currentFinalizedL2BlockNumber + blockTimestamp - lastFinalizedBlockTimestamp + l2BlockBuffer;
};
