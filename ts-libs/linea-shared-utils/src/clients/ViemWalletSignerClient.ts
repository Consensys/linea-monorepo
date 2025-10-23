import {
  Account,
  Address,
  Chain,
  createWalletClient,
  Hex,
  http,
  parseTransaction,
  serializeSignature,
  TransactionSerializable,
  WalletClient,
} from "viem";
import { IContractSignerClient } from "../core/client/IContractSignerClient";
import { privateKeyToAccount, privateKeyToAddress } from "viem/accounts";

export class ViemWalletSignerClient implements IContractSignerClient {
  private readonly account: Account;
  private readonly address: Address;
  private readonly wallet: WalletClient;

  constructor(privateKey: Hex, chain: Chain) {
    this.account = privateKeyToAccount(privateKey);
    this.address = privateKeyToAddress(privateKey);
    this.wallet = createWalletClient({
      account: this.account,
      chain,
      transport: http(),
    });
  }

  async sign(tx: TransactionSerializable): Promise<Hex> {
    // Remove any signature fields if they exist on the object
    // 'as any' required to avoid enforcing strict structural validation
    // Fine because we are only removing fields, not depending on them existing
    // Practical way to strip off optional keys from a union type
    const { r: r_void, s: s_void, v: v_void, yParity: yParity_void, ...unsigned } = tx as any; // eslint-disable-line @typescript-eslint/no-explicit-any
    void r_void;
    void s_void;
    void v_void;
    void yParity_void;

    const serializedSignedTx = await this.wallet.signTransaction({ ...unsigned });
    const { r, s, yParity } = await parseTransaction(serializedSignedTx);
    // TODO - Better error handling
    if (!r || !s || yParity === undefined) throw new Error("sign error");
    const signatureHex = serializeSignature({
      r,
      s,
      yParity,
    });

    return signatureHex;
  }

  getAddress(): Address {
    return this.address;
  }
}
