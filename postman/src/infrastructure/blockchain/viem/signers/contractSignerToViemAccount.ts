import { IContractSignerClient } from "@consensys/linea-shared-utils";
import { parseSignature, serializeTransaction } from "viem";
import { toAccount } from "viem/accounts";

import type { LocalAccount } from "viem";

/**
 * Bridges an IContractSignerClient into a viem LocalAccount so it can be used
 * to create a WalletClient for transaction submission.
 *
 * The signing backend (ViemWalletSignerClientAdapter or Web3SignerClientAdapter)
 * is fully transparent to callers of the returned account.
 */
export function contractSignerToViemAccount(signer: IContractSignerClient): LocalAccount {
  return toAccount({
    address: signer.getAddress(),
    async signTransaction(transaction) {
      const signatureHex = await signer.sign(transaction);
      const signature = parseSignature(signatureHex);
      return serializeTransaction(transaction, signature);
    },
    async signMessage() {
      throw new Error("signMessage is not supported by IContractSignerClient");
    },
    async signTypedData() {
      throw new Error("signTypedData is not supported by IContractSignerClient");
    },
  });
}
