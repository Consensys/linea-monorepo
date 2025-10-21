import { Address, Hex, TransactionSerializable } from "viem";

export interface IContractSignerService {
  sign(tx: TransactionSerializable): Promise<Hex>;
  getAddress(): Address;
}
