import { Eip1559Transaction } from "../types";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function buildEip1559Transaction(data: any): Eip1559Transaction {
  return {
    nonce: data.nonce,
    maxPriorityFeePerGas: data.maxPriorityFeePerGas,
    maxFeePerGas: data.maxFeePerGas,
    gasLimit: data.gas,
    to: data.to,
    value: data.value,
    input: data.input,
    yParity: data.yParity,
    r: data.r,
    s: data.s,
  };
}
