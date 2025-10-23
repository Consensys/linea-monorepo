import {
  Account,
  Address,
  createWalletClient,
  Hex,
  http,
  PublicClient,
  TransactionSerializable,
  WalletClient,
} from "viem";
import { IContractSignerClient } from "../core/client/IContractSignerClient";
import { privateKeyToAccount, privateKeyToAddress } from "viem/accounts";

export class ViemWalletSignerClient implements IContractSignerClient {
  private readonly account: Account;
  private readonly address: Address;
  private readonly wallet: WalletClient;

  constructor(
    private readonly publicClient: PublicClient,
    privateKey: Hex,
  ) {
    this.account = privateKeyToAccount(privateKey);
    this.address = privateKeyToAddress(privateKey);
    this.wallet = createWalletClient({
      account: this.account,
      chain: this.publicClient.chain,
      transport: http(),
    });
  }

  async sign(tx: TransactionSerializable): Promise<Hex> {
    return await this.wallet.signTransaction({ ...tx, chainId: await this.publicClient.getChainId() });
  }

  getAddress(): Address {
    return this.address;
  }
}
