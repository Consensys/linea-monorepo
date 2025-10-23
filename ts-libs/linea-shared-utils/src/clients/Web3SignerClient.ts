import axios from "axios";
import { Agent } from "https";
import { Address, Hex, serializeTransaction, TransactionSerializable } from "viem";
import { IContractSignerClient } from "../core/client/IContractSignerClient";
import { publicKeyToAddress } from "viem/accounts";
import forge from "node-forge";
import { readFileSync } from "fs";
import path from "path";

export class Web3SignerClient implements IContractSignerClient {
  private readonly agent: Agent;
  constructor(
    private readonly web3SignerUrl: string,
    private readonly web3SignerPublicKey: Hex,
    web3SignerKeystorePath: string,
    web3SignerKeystorePassphrase: string,
    web3SignerTrustedStorePath: string,
    web3SignerTrustedStorePassphrase: string,
  ) {
    this.agent = this.getHttpsAgent(
      web3SignerKeystorePath,
      web3SignerKeystorePassphrase,
      web3SignerTrustedStorePath,
      web3SignerTrustedStorePassphrase,
    );
  }

  async sign(tx: TransactionSerializable): Promise<Hex> {
    const { data } = await axios.post(
      `${this.web3SignerUrl}/api/v1/eth1/sign/${this.web3SignerPublicKey}`,
      {
        data: serializeTransaction(tx),
      },
      { httpsAgent: this.agent },
    );
    return data;
  }

  getAddress(): Address {
    return publicKeyToAddress(this.web3SignerPublicKey);
  }

  private convertToPem(p12base64: string | forge.util.ByteStringBuffer, password: string) {
    const p12Asn1 = forge.asn1.fromDer(p12base64);
    const p12 = forge.pkcs12.pkcs12FromAsn1(p12Asn1, false, password);

    return this.getCertificateFromP12(p12);
  }

  private getCertificateFromP12(p12: forge.pkcs12.Pkcs12Pfx) {
    const certData = p12.getBags({ bagType: forge.pki.oids.certBag });
    const certificate = certData[forge.pki.oids.certBag]?.[0];

    if (!certificate?.cert) {
      throw new Error("Certificate not found in P12");
    }

    const pemCertificate = forge.pki.certificateToPem(certificate.cert);
    return { pemCertificate };
  }

  private getHttpsAgent(
    keystorePath: string,
    keystorePassphrase: string,
    trustedStorePath: string,
    trustedStorePassphrase: string,
  ): Agent {
    const trustedStoreFile = readFileSync(path.resolve(__dirname, trustedStorePath), { encoding: "binary" });

    const { pemCertificate } = this.convertToPem(trustedStoreFile, trustedStorePassphrase);

    return new Agent({
      pfx: readFileSync(path.resolve(__dirname, keystorePath)),
      passphrase: keystorePassphrase,
      ca: pemCertificate,
    });
  }
}
