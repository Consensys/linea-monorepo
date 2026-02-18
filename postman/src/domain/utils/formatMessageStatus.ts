import { OnChainMessageStatus } from "../types/enums";

const MESSAGE_ANCHORED_STATUS = 1;
const MESSAGE_CLAIMED_STATUS = 2;

export function formatMessageStatus(status: bigint): OnChainMessageStatus {
  if (status === BigInt(MESSAGE_CLAIMED_STATUS)) {
    return OnChainMessageStatus.CLAIMED;
  }

  if (status === BigInt(MESSAGE_ANCHORED_STATUS)) {
    return OnChainMessageStatus.CLAIMABLE;
  }

  return OnChainMessageStatus.UNKNOWN;
}
