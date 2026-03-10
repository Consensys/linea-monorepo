import { describe, it, expect, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";

import { Direction } from "../../../../../core/enums";
import { IMessageRepository } from "../../../../../core/persistence/IMessageRepository";
import { TestLogger } from "../../../../../utils/testing/helpers";
import { InlineNonceManager } from "../InlineNonceManager";

import type { PublicClient } from "viem";

const SIGNER_ADDRESS = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa";

describe("InlineNonceManager", () => {
  let publicClient: ReturnType<typeof mock<PublicClient>>;
  let databaseService: ReturnType<typeof mock<IMessageRepository>>;
  let logger: TestLogger;

  beforeEach(() => {
    publicClient = mock<PublicClient>();
    databaseService = mock<IMessageRepository>();
    logger = new TestLogger("InlineNonceManager");
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  function buildManager(maxNonceDiff = 5) {
    return new InlineNonceManager(
      databaseService,
      publicClient,
      SIGNER_ADDRESS,
      maxNonceDiff,
      Direction.L1_TO_L2,
      logger,
    );
  }

  describe("acquireNonce", () => {
    it("returns on-chain nonce when DB has no last nonce", async () => {
      databaseService.getLastClaimTxNonce.mockResolvedValue(null);
      publicClient.getTransactionCount.mockResolvedValue(7);

      const nonce = await buildManager().acquireNonce();
      expect(nonce).toBe(7);
    });

    it("returns max(onChainNonce, lastTxNonce+1) when lastTxNonce < onChainNonce", async () => {
      databaseService.getLastClaimTxNonce.mockResolvedValue(3);
      publicClient.getTransactionCount.mockResolvedValue(10);

      const nonce = await buildManager().acquireNonce();
      expect(nonce).toBe(10); // max(10, 4)
    });

    it("returns lastTxNonce+1 when it is greater than onChainNonce", async () => {
      databaseService.getLastClaimTxNonce.mockResolvedValue(12);
      publicClient.getTransactionCount.mockResolvedValue(10);

      const nonce = await buildManager().acquireNonce();
      expect(nonce).toBe(13); // max(10, 13)
    });

    it("returns null when nonce diff exceeds maxNonceDiff", async () => {
      databaseService.getLastClaimTxNonce.mockResolvedValue(20);
      publicClient.getTransactionCount.mockResolvedValue(10);

      const warnSpy = jest.spyOn(logger, "warn");
      const nonce = await buildManager(5).acquireNonce();

      expect(nonce).toBeNull();
      expect(warnSpy).toHaveBeenCalledTimes(1);
    });

    it("does not pause when nonce diff exactly equals maxNonceDiff", async () => {
      databaseService.getLastClaimTxNonce.mockResolvedValue(15);
      publicClient.getTransactionCount.mockResolvedValue(10);

      const nonce = await buildManager(5).acquireNonce();
      expect(nonce).toBe(16); // diff is exactly 5, not > 5
    });

    it("queries the correct signer address on-chain", async () => {
      databaseService.getLastClaimTxNonce.mockResolvedValue(null);
      publicClient.getTransactionCount.mockResolvedValue(3);

      await buildManager().acquireNonce();

      expect(publicClient.getTransactionCount).toHaveBeenCalledWith({
        address: SIGNER_ADDRESS,
      });
    });
  });

  describe("releaseNonce / reportFailure", () => {
    it("releaseNonce is a no-op", () => {
      expect(() => buildManager().releaseNonce(5, "0xhash")).not.toThrow();
    });

    it("reportFailure is a no-op", () => {
      expect(() => buildManager().reportFailure(5)).not.toThrow();
    });
  });
});
