export function bigintReplacer(key: string, value: unknown) {
  return typeof value === "bigint" ? value.toString() : value;
}
