export function serialize(value: unknown): string {
  return JSON.stringify(value, (_, v: unknown) => (typeof v === "bigint" ? v.toString() : v));
}
