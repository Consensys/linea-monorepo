import { GoNativeCompressor } from "@consensys/linea-native-libs";
import { type Hex, type Address, serializeTransaction, toBytes } from "viem";

import { BaseError } from "../../../domain/errors/BaseError";
import { serialize } from "../../../domain/utils/serialize";

import type { ViemL2ContractClient } from "./ViemL2ContractClient";
import type { MessageProps } from "../../../domain/message/Message";
import type { IL2ClaimTransactionSizeCalculator } from "../../../domain/ports/IL2ClaimTransactionSizeCalculator";
import type { LineaGasFees } from "../../../domain/types/blockchain";

export class ViemTransactionSizeCalculator implements IL2ClaimTransactionSizeCalculator {
  private compressor: GoNativeCompressor;

  constructor(private readonly l2ContractClient: ViemL2ContractClient) {
    this.compressor = new GoNativeCompressor(800_000);
  }

  public async calculateTransactionSize(
    message: MessageProps & { feeRecipient?: string },
    fees: LineaGasFees,
  ): Promise<number> {
    try {
      const transactionData = this.l2ContractClient.encodeClaimMessageTransactionData(message);
      const signerAddress = this.l2ContractClient.getSigner();
      const destinationAddress = this.l2ContractClient.getContractAddress();
      const { gasLimit, maxFeePerGas, maxPriorityFeePerGas } = fees;

      if (!signerAddress) {
        throw new BaseError("Signer is undefined.");
      }

      const serializedTx = serializeTransaction({
        to: destinationAddress as Address,
        value: 0n,
        data: transactionData as Hex,
        maxPriorityFeePerGas,
        maxFeePerGas,
        gas: gasLimit,
        type: "eip1559" as const,
        chainId: 1,
      });

      const txBytes = toBytes(serializedTx);
      return this.compressor.getCompressedTxSize(txBytes);
    } catch (error) {
      throw new BaseError(`Transaction size calculation error: ${serialize(error)}`);
    }
  }
}
