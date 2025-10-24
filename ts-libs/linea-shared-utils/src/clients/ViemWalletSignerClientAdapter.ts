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
import { ILogger } from "../logging/ILogger";

export class ViemWalletSignerClientAdapter implements IContractSignerClient {
  private readonly account: Account;
  private readonly address: Address;
  private readonly wallet: WalletClient;

  constructor(
    private readonly logger: ILogger,
    rpcUrl: string,
    privateKey: Hex,
    chain: Chain,
  ) {
    this.account = privateKeyToAccount(privateKey);
    this.address = privateKeyToAddress(privateKey);
    this.wallet = createWalletClient({
      account: this.account,
      chain,
      transport: http(rpcUrl),
    });
  }

  async sign(tx: TransactionSerializable): Promise<Hex> {
    this.logger.debug("sign started...", { tx });
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
    const parsedTx = parseTransaction(serializedSignedTx);
    this.logger.debug("sign", { parsedTx });
    const { r, s, yParity } = parsedTx;
    // TODO - Better error handling
    if (!r || !s || yParity === undefined) {
      this.logger.error("sign - r, s or yParity missing");
      throw new Error("sign - r, s or yParity missing");
    }

    const signatureHex = serializeSignature({
      r,
      s,
      yParity,
    });

    this.logger.debug(`sign completed signatureHex=${signatureHex}`);
    return signatureHex;
  }

  getAddress(): Address {
    return this.address;
  }
}
