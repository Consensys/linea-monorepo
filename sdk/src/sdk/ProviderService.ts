import { Wallet, Signer, JsonRpcProvider } from "ethers";
import { BaseError } from "../core/errors/Base";

export default class ProviderService {
  public provider: JsonRpcProvider;

  /**
   * Constructs a new `ProviderService` instance.
   *
   * @param {string | JsonRpcProvider} rpcUrlOrProvider - The Ethereum JSON RPC URL or an existing `JsonRpcProvider` instance.
   */
  constructor(rpcUrlOrProvider: string | JsonRpcProvider) {
    if (rpcUrlOrProvider instanceof JsonRpcProvider) {
      this.provider = rpcUrlOrProvider;
    } else {
      this.provider = new JsonRpcProvider(rpcUrlOrProvider);
    }
  }

  /**
   * Retrieves a signer for transaction signing. This method supports both Wallet instances and private keys.
   *
   * @param {string | Wallet} privateKeyOrWallet - A private key as a string or an existing Wallet instance.
   * @returns {Signer} A signer, represented by a Wallet instance, associated with the current provider.
   */
  public getSigner(privateKeyOrWallet: string | Wallet): Signer {
    try {
      return privateKeyOrWallet instanceof Wallet ? privateKeyOrWallet : new Wallet(privateKeyOrWallet, this.provider);
    } catch (e) {
      throw new BaseError(
        "Something went wrong when trying to generate Wallet. Please check your private key and the provider url.",
      );
    }
  }
}
