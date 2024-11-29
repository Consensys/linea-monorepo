import { ethers, Signer } from "ethers";
import { BaseError } from "../core/errors";

export class Wallet<T extends ethers.Provider> extends ethers.Wallet {
  constructor(privateKey: string, provider?: T) {
    try {
      super(privateKey, provider);
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (e: any) {
      if (e.message.includes("invalid private key")) {
        throw new BaseError(
          "Something went wrong when trying to generate Wallet. Please check your private key and the provider url.",
        );
      }
      throw e;
    }
  }

  public static getWallet<T extends ethers.Provider>(privateKeyOrWallet: string | Wallet<T>): Signer {
    try {
      return privateKeyOrWallet instanceof Wallet ? privateKeyOrWallet : new Wallet(privateKeyOrWallet);
    } catch (e) {
      throw new BaseError(
        "Something went wrong when trying to generate Wallet. Please check your private key and the provider url.",
      );
    }
  }
}
