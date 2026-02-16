export function isEmptyBytes(bytes: string): boolean {
  if (!bytes || bytes === "0x" || bytes === "" || bytes === "0") {
    return true;
  }

  const hexString = bytes.replace(/^0x/, "");
  return /^00*$/.test(hexString);
}
