import { fromUnixTime } from "date-fns/fromUnixTime";
import { compareAsc } from "date-fns/compareAsc";
import { subDays } from "date-fns/subDays";

// Transactions with an age > TOO_OLD_IN_DAYS_THRESHOLD will be considered stale
const TOO_OLD_IN_DAYS_THRESHOLD = 90;

export function isTimestampTooOld(timestamp: bigint): boolean {
  const currentTimestamp = new Date();
  return (
    compareAsc(fromUnixTime(Number(timestamp.toString())), subDays(currentTimestamp, TOO_OLD_IN_DAYS_THRESHOLD)) === -1
  );
}
