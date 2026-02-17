export type DecodedCalldata = {
  name: string;
  args: readonly unknown[] | undefined;
};

export interface ICalldataDecoder {
  decode(calldata: string): DecodedCalldata | null;
}
