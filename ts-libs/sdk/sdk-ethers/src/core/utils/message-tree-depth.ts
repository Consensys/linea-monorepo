import { MAX_L2_MESSAGE_TREE_DEPTH } from "../constants";

export function validateL2MessageTreeDepth(depth: number, label = "l2MessageTreeDepth"): number {
  if (!Number.isFinite(depth) || !Number.isSafeInteger(depth)) {
    throw new Error(`${label} must be a finite safe integer`);
  }

  if (depth <= 1) {
    throw new Error(`${label} must be greater than 1`);
  }

  if (depth > MAX_L2_MESSAGE_TREE_DEPTH) {
    throw new Error(`${label} must be less than or equal to ${MAX_L2_MESSAGE_TREE_DEPTH}`);
  }

  return depth;
}

export function validateL2MessageTreeDepthFromLog(treeDepth: bigint, expectedDepth: number): number {
  const validExpectedDepth = validateL2MessageTreeDepth(expectedDepth);

  if (treeDepth !== BigInt(validExpectedDepth)) {
    throw new Error(
      `Finalization treeDepth must equal configured l2MessageTreeDepth. Expected ${validExpectedDepth}, received ${treeDepth.toString()}`,
    );
  }

  return validExpectedDepth;
}
