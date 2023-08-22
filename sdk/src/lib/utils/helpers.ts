import { BigNumber } from "ethers";
import { MESSAGE_ANCHORED_STATUS, MESSAGE_CLAIMED_STATUS } from "./constants";
import { OnChainMessageStatus } from "./enum";

export function formatMessageStatus(status: BigNumber): OnChainMessageStatus {
  if (status.eq(MESSAGE_CLAIMED_STATUS)) {
    return OnChainMessageStatus.CLAIMED;
  }

  if (status.eq(MESSAGE_ANCHORED_STATUS)) {
    return OnChainMessageStatus.CLAIMABLE;
  }

  return OnChainMessageStatus.UNKNOWN;
}
