import { Wallet, ethers } from "ethers";
import { JsonRpcProvider } from "@ethersproject/providers";

export default class ProviderService {
  public provider: JsonRpcProvider;

  constructor(rpcUrl: string) {
    this.provider = new JsonRpcProvider(rpcUrl);
  }

  public getSigner(privateKey: string): ethers.Signer {
    try {
      return new Wallet(privateKey, this.provider);
    } catch (e) {
      throw new Error(
        "Something went wrong when trying to generate Wallet. Please check your private key and the provider url.",
      );
    }
  }
}
