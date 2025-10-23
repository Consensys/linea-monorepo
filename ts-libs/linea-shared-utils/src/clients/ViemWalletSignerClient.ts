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
    // Remove any signature fields if they exist on the object
    // 'as any' required to avoid enforcing strict structural validation
    // Fine because we are only removing fields, not depending on them existing
    // Practical way to strip off optional keys from a union type
    const { r, s, v, yParity, ...unsigned } = tx as any; // eslint-disable-line @typescript-eslint/no-explicit-any
    void r;
    void s;
    void v;
    void yParity;

    return this.wallet.signTransaction({ ...unsigned, chainId: await this.publicClient.getChainId() });
  }

  getAddress(): Address {
    return this.address;
  }
}
