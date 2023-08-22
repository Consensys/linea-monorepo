export const wait = (timeout: number) => new Promise((resolve) => setTimeout(resolve, timeout));

export const subtractSeconds = (date: Date, seconds: number): Date => {
  const dateCopy = new Date(date);
  dateCopy.setSeconds(date.getSeconds() - seconds);
  return dateCopy;
};

export function isEmptyBytes(bytes: string): boolean {
  if (!bytes || bytes === "0x" || bytes === "" || bytes === "0") {
    return true;
  }

  const hexString = bytes.replace(/^0x/, "");
  return /^00*$/.test(hexString);
}
