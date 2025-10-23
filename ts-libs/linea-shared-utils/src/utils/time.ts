export function getCurrentUnixTimestampSeconds(): number {
  return Math.floor(Date.now() / 1000);
}
