import { describe, it, expect } from "@jest/globals";
import { createPublicClient, createWalletClient } from "viem";

import { TestLogger } from "../../../../../utils/testing/helpers";
import { contractSignerToViemAccount } from "../contractSignerToViemAccount";
import { createChainContext } from "../createChainContext";
import { createSignerClient } from "../createSignerClient";

import type { SignerConfig } from "../SignerConfig";

const MOCK_CHAIN_ID = 59144;
const MOCK_ADDRESS = "0x1234567890abcdef1234567890abcdef12345678" as `0x${string}`;

jest.mock("viem", () => {
  const actual = jest.requireActual("viem");
  return {
    ...actual,
    createPublicClient: jest.fn(() => ({
      getChainId: jest.fn().mockResolvedValue(MOCK_CHAIN_ID),
    })),
    createWalletClient: jest.fn(() => ({ type: "walletClient" })),
  };
});

jest.mock("../createSignerClient", () => ({
  createSignerClient: jest.fn(() => ({
    getAddress: jest.fn(() => MOCK_ADDRESS),
    sign: jest.fn(),
  })),
}));

jest.mock("../contractSignerToViemAccount", () => ({
  contractSignerToViemAccount: jest.fn(() => ({
    address: MOCK_ADDRESS,
    type: "local",
  })),
}));

describe("createChainContext", () => {
  const logger = new TestLogger("createChainContext");
  const rpcUrl = "http://localhost:8545";
  const signerConfig: SignerConfig = {
    type: "private-key",
    privateKey: "0x0000000000000000000000000000000000000000000000000000000000000001",
  };

  afterEach(() => {
    jest.clearAllMocks();
  });

  it("returns a complete ChainContext with the correct chainId", async () => {
    const ctx = await createChainContext(rpcUrl, signerConfig, logger);

    expect(ctx.chainId).toBe(MOCK_CHAIN_ID);
    expect(ctx.chain.id).toBe(MOCK_CHAIN_ID);
    expect(ctx.publicClient).toBeDefined();
    expect(ctx.walletClient).toBeDefined();
    expect(ctx.account).toBeDefined();
    expect(ctx.account.address).toBe(MOCK_ADDRESS);
    expect(ctx.signer).toBeDefined();
  });

  it("creates a public client and fetches chain ID", async () => {
    const ctx = await createChainContext(rpcUrl, signerConfig, logger);

    expect(createPublicClient).toHaveBeenCalledWith({ transport: expect.anything() });
    expect(ctx.publicClient.getChainId).toHaveBeenCalled();
  });

  it("passes signer config to createSignerClient", async () => {
    await createChainContext(rpcUrl, signerConfig, logger);

    expect(createSignerClient).toHaveBeenCalledWith(
      signerConfig,
      logger,
      rpcUrl,
      expect.objectContaining({ id: MOCK_CHAIN_ID }),
    );
  });

  it("converts the signer to a viem account", async () => {
    await createChainContext(rpcUrl, signerConfig, logger);

    expect(contractSignerToViemAccount).toHaveBeenCalledWith(
      expect.objectContaining({ getAddress: expect.any(Function) }),
    );
  });

  it("creates a wallet client with the account and chain", async () => {
    await createChainContext(rpcUrl, signerConfig, logger);

    expect(createWalletClient).toHaveBeenCalledWith(
      expect.objectContaining({
        account: expect.objectContaining({ address: MOCK_ADDRESS }),
        chain: expect.objectContaining({ id: MOCK_CHAIN_ID }),
      }),
    );
  });
});
