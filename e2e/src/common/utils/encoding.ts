import {
  encodeAbiParameters,
  encodeFunctionData,
  encodePacked,
  keccak256,
  EncodeFunctionDataParameters,
  getAddress,
} from "viem";

export function encodeFunctionCall(params: EncodeFunctionDataParameters) {
  return encodeFunctionData(params);
}

export function generateKeccak256(types: string[], values: unknown[], packed?: boolean) {
  return keccak256(encodeData(types, values, packed));
}

export function encodeData(types: string[], values: unknown[], packed?: boolean) {
  if (packed) {
    return encodePacked(types, values);
  }
  const params = types.map((type) => ({ type }));
  return encodeAbiParameters(params, values);
}

export function normalizeAddress(address: string) {
  return getAddress(address);
}
