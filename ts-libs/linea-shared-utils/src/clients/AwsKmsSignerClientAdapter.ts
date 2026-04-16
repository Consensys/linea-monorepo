import { GetPublicKeyCommand, KMSClient, SignCommand } from "@aws-sdk/client-kms";
import {
  Address,
  Hex,
  hexToBytes,
  keccak256,
  serializeSignature,
  serializeTransaction,
  TransactionSerializable,
} from "viem";
import { publicKeyToAddress } from "viem/accounts";

import { IContractSignerClient } from "../core/client/IContractSignerClient";
import { ILogger } from "../logging/ILogger";
import { extractUncompressedPublicKeyFromDer } from "../utils/ec-publickey";
import { bigintToHex32, decodeDerSignature, recoverYParity } from "../utils/ecdsa-signature";

/**
 * Adapter for AWS KMS that provides contract signing functionality using an asymmetric
 * ECC_SECG_P256K1 (secp256k1) key. The private key never leaves the KMS boundary.
 *
 * DER encoding/decoding is delegated to dedicated utilities:
 * - {@link extractUncompressedPublicKeyFromDer} for SubjectPublicKeyInfo (RFC 5280)
 * - {@link decodeDerSignature} for ECDSA-Sig-Value (RFC 3279) with EIP-2 normalisation
 * - {@link recoverYParity} for signature recovery parameter determination
 *
 * Requires an async {@link init} call (or use the {@link create} factory) before
 * {@link sign} or {@link getAddress} can be used, because the Ethereum address is
 * derived from the KMS public key via an async API call.
 */
export class AwsKmsSignerClientAdapter implements IContractSignerClient {
  private address: Address | undefined;
  private readonly kmsClient: KMSClient;

  /**
   * Creates a new AwsKmsSignerClientAdapter instance.
   * Call {@link init} before using {@link sign} or {@link getAddress}.
   *
   * @param {ILogger} logger - The logger instance for logging signing operations.
   * @param {string} kmsKeyId - The AWS KMS key ID, key ARN, alias name, or alias ARN
   *   for the ECC_SECG_P256K1 key. Use {@link createKey} to provision a new one.
   * @param {KMSClient} [kmsClient] - Optional pre-configured KMS client.
   *   A default client is created if omitted (uses standard AWS credential chain).
   */
  constructor(
    private readonly logger: ILogger,
    private readonly kmsKeyId: string,
    kmsClient?: KMSClient,
  ) {
    this.logger.info("Initialising AwsKmsSignerClientAdapter");
    this.kmsClient = kmsClient ?? new KMSClient();
  }

  /**
   * Factory that creates and initializes an adapter in one step.
   *
   * @param {ILogger} logger - The logger instance.
   * @param {string} kmsKeyId - The AWS KMS key ID or ARN.
   * @param {KMSClient} [kmsClient] - Optional pre-configured KMS client.
   * @returns {Promise<AwsKmsSignerClientAdapter>} An initialized adapter ready for signing.
   */
  static async create(logger: ILogger, kmsKeyId: string, kmsClient?: KMSClient): Promise<AwsKmsSignerClientAdapter> {
    const adapter = new AwsKmsSignerClientAdapter(logger, kmsKeyId, kmsClient);
    await adapter.init();
    return adapter;
  }

  /**
   * Fetches the public key from KMS and derives the Ethereum address.
   * Must be called once before {@link sign} or {@link getAddress}.
   */
  async init(): Promise<void> {
    const publicKeyDer = await this.fetchDerPublicKey();
    const uncompressedKey = extractUncompressedPublicKeyFromDer(publicKeyDer);
    this.address = publicKeyToAddress(uncompressedKey);
    this.logger.info(`AwsKmsSignerClientAdapter initialized address=${this.address}`);
  }

  /**
   * Signs a transaction by hashing it locally (keccak256) and sending the 32-byte
   * digest to AWS KMS for ECDSA signing. The private key never leaves KMS.
   *
   * @param {TransactionSerializable} tx - The transaction to sign.
   * @returns {Promise<Hex>} The serialized signature as a hex string.
   */
  async sign(tx: TransactionSerializable): Promise<Hex> {
    const address = this.getInitializedAddress();
    this.logger.debug("Signing transaction via AWS KMS");

    const serializedTx = serializeTransaction(tx);
    const txHash = keccak256(serializedTx);
    const hashBytes = hexToBytes(txHash);

    const signResult = await this.kmsClient.send(
      new SignCommand({
        KeyId: this.kmsKeyId,
        Message: hashBytes,
        MessageType: "DIGEST",
        SigningAlgorithm: "ECDSA_SHA_256",
      }),
    );

    if (!signResult.Signature) {
      throw new Error("AWS KMS returned empty signature");
    }

    const { r, s } = decodeDerSignature(signResult.Signature);

    const rHex = bigintToHex32(r);
    const sHex = bigintToHex32(s);
    const yParity = await recoverYParity(txHash, rHex, sHex, address);

    const signatureHex = serializeSignature({ r: rHex, s: sHex, yParity });
    this.logger.debug(`Signing successful signature=${signatureHex}`);
    return signatureHex;
  }

  /**
   * Gets the Ethereum address derived from the KMS public key.
   *
   * @returns {Address} The Ethereum address.
   * @throws {Error} If {@link init} has not been called.
   */
  getAddress(): Address {
    return this.getInitializedAddress();
  }

  private getInitializedAddress(): Address {
    if (!this.address) {
      throw new Error("AwsKmsSignerClientAdapter not initialized. Call init() first.");
    }
    return this.address;
  }

  private async fetchDerPublicKey(): Promise<Uint8Array> {
    const result = await this.kmsClient.send(new GetPublicKeyCommand({ KeyId: this.kmsKeyId }));
    if (!result.PublicKey) {
      throw new Error("AWS KMS returned empty public key");
    }
    return result.PublicKey;
  }
}
