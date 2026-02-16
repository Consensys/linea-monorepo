export type DecodedCalldata = {
  name: string;
  args: Record<string, unknown>;
};

export interface ICalldataDecoder {
  decode(calldata: string): DecodedCalldata | null;
}
