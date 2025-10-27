import { MS_PER_SECOND } from "../core/constants/time";

export function getCurrentUnixTimestampSeconds(): number {
  return Math.floor(Date.now() / 1000);
}

/**
 * Converts milliseconds to whole seconds (rounded down).
 * @param ms - Milliseconds value
 * @returns Number of seconds, floored
 */
export function msToSeconds(ms: number): number {
  return Math.floor(ms / MS_PER_SECOND);
}
