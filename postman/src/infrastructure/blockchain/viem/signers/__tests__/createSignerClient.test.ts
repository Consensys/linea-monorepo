import { ViemWalletSignerClientAdapter, Web3SignerClientAdapter } from "@consensys/linea-shared-utils";
import { describe, it, expect } from "@jest/globals";
import { mainnet } from "viem/chains";

import { TestLogger } from "../../../../../utils/testing/helpers";
import { createSignerClient } from "../createSignerClient";

import type { SignerConfig } from "../SignerConfig";

jest.mock("@consensys/linea-shared-utils", () => ({
  ViemWalletSignerClientAdapter: jest.fn().mockImplementation(() => ({ type: "viem-wallet" })),
  Web3SignerClientAdapter: jest.fn().mockImplementation(() => ({ type: "web3signer" })),
}));

describe("createSignerClient", () => {
  const logger = new TestLogger("createSignerClient");

  afterEach(() => {
    jest.clearAllMocks();
  });

  it("creates a ViemWalletSignerClientAdapter for private-key config", () => {
    const config: SignerConfig = {
      type: "private-key",
      privateKey: "0x0000000000000000000000000000000000000000000000000000000000000001",
    };

    const client = createSignerClient(config, logger, "http://localhost:8545", mainnet);

    expect(ViemWalletSignerClientAdapter).toHaveBeenCalledWith(
      logger,
      "http://localhost:8545",
      config.privateKey,
      mainnet,
    );
    expect(client).toBeDefined();
  });

  it("creates a Web3SignerClientAdapter for web3signer config without TLS", () => {
    const config: SignerConfig = {
      type: "web3signer",
      endpoint: "http://web3signer:9000",
      publicKey: "0xaabbccdd",
    };

    const client = createSignerClient(config, logger, "http://localhost:8545", mainnet);

    expect(Web3SignerClientAdapter).toHaveBeenCalledWith(logger, config.endpoint, config.publicKey, "", "", "", "");
    expect(client).toBeDefined();
  });

  it("creates a Web3SignerClientAdapter for web3signer config with TLS", () => {
    const config: SignerConfig = {
      type: "web3signer",
      endpoint: "https://web3signer:9000",
      publicKey: "0xaabbccdd",
      tls: {
        keyStorePath: "/certs/keystore.p12",
        keyStorePassword: "keystorepass",
        trustStorePath: "/certs/truststore.p12",
        trustStorePassword: "truststorepass",
      },
    };

    createSignerClient(config, logger, "http://localhost:8545", mainnet);

    expect(Web3SignerClientAdapter).toHaveBeenCalledWith(
      logger,
      config.endpoint,
      config.publicKey,
      config.tls!.keyStorePath,
      config.tls!.keyStorePassword,
      config.tls!.trustStorePath,
      config.tls!.trustStorePassword,
    );
  });
});
