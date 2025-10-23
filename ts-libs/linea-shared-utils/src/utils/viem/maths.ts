export function safeSub(a: bigint, b: bigint): bigint {
  return a > b ? a - b : 0n;
}

export function min(a: bigint, b: bigint): bigint {
  return a < b ? a : b;
}
