import { describe, it, expect, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";

import { IProvider } from "../../../../../core/clients/blockchain/IProvider";
import { TestLogger } from "../../../../../utils/testing/helpers";
import { InMemoryNonceManager } from "../InMemoryNonceManager";

const SIGNER_ADDRESS = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa";

describe("InMemoryNonceManager", () => {
  let provider: ReturnType<typeof mock<IProvider>>;
  let logger: TestLogger;

  beforeEach(() => {
    provider = mock<IProvider>();
    logger = new TestLogger("InMemoryNonceManager");
  });

  afterEach(() => {
    jest.resetAllMocks();
  });

  function buildManager(maxNonceDiff = 5) {
    return new InMemoryNonceManager(provider, SIGNER_ADDRESS, maxNonceDiff, logger);
  }

  describe("initialize", () => {
    it("sets nextNonce from on-chain nonce", async () => {
      provider.getTransactionCount.mockResolvedValue(7);
      const manager = buildManager();
      await manager.initialize();

      provider.getTransactionCount.mockResolvedValue(7);
      const nonce = await manager.acquireNonce();
      expect(nonce).toBe(7);
    });

    it("queries the correct signer address with pending block tag", async () => {
      provider.getTransactionCount.mockResolvedValue(0);
      const manager = buildManager();
      await manager.initialize();

      expect(provider.getTransactionCount).toHaveBeenCalledWith(SIGNER_ADDRESS, "pending");
    });
  });

  describe("acquireNonce", () => {
    it("returns sequential nonces", async () => {
      provider.getTransactionCount.mockResolvedValue(10);
      const manager = buildManager();
      await manager.initialize();

      const n1 = await manager.acquireNonce();
      const n2 = await manager.acquireNonce();
      const n3 = await manager.acquireNonce();

      expect(n1).toBe(10);
      expect(n2).toBe(11);
      expect(n3).toBe(12);
    });

    it("resynchronizes when drift exceeds maxNonceDiff", async () => {
      provider.getTransactionCount.mockResolvedValue(5);
      const manager = buildManager(2);
      await manager.initialize();

      // Acquire 5 nonces without committing → nonces 5,6,7,8,9
      await manager.acquireNonce();
      await manager.acquireNonce();
      await manager.acquireNonce();
      manager.commitNonce(5);
      manager.commitNonce(6);
      manager.commitNonce(7);

      // On-chain hasn't moved, drift = nextNonce(8) - onChain(5) = 3 > maxNonceDiff(2)
      const warnSpy = jest.spyOn(logger, "warn");
      const nonce = await manager.acquireNonce();

      expect(warnSpy).toHaveBeenCalledTimes(1);
      expect(nonce).toBe(5); // resynchronized to on-chain
    });

    it("does not resynchronize when reusable nonces are available", async () => {
      provider.getTransactionCount.mockResolvedValue(5);
      const manager = buildManager(0);
      await manager.initialize();

      const n1 = await manager.acquireNonce(); // 5
      await manager.acquireNonce(); // 6
      manager.rollbackNonce(n1); // 5 goes back to reusable

      // Even though drift is high, reusable nonce should be used
      const nonce = await manager.acquireNonce();
      expect(nonce).toBe(5);
    });
  });

  describe("commitNonce", () => {
    it("removes nonce from pending set", async () => {
      provider.getTransactionCount.mockResolvedValue(10);
      const manager = buildManager();
      await manager.initialize();

      const nonce = await manager.acquireNonce();
      manager.commitNonce(nonce);

      // No error thrown, committed nonces are simply removed
      expect(nonce).toBe(10);
    });
  });

  describe("rollbackNonce", () => {
    it("makes nonce available for reuse", async () => {
      provider.getTransactionCount.mockResolvedValue(10);
      const manager = buildManager();
      await manager.initialize();

      const n1 = await manager.acquireNonce(); // 10
      await manager.acquireNonce(); // 11
      manager.rollbackNonce(n1);

      const reused = await manager.acquireNonce();
      expect(reused).toBe(10); // reused the rolled-back nonce
    });

    it("reuses lowest rolled-back nonce first", async () => {
      provider.getTransactionCount.mockResolvedValue(10);
      const manager = buildManager();
      await manager.initialize();

      const n1 = await manager.acquireNonce(); // 10
      const n2 = await manager.acquireNonce(); // 11
      await manager.acquireNonce(); // 12

      manager.rollbackNonce(n2); // 11
      manager.rollbackNonce(n1); // 10

      const r1 = await manager.acquireNonce();
      const r2 = await manager.acquireNonce();
      expect(r1).toBe(10);
      expect(r2).toBe(11);
    });
  });

  describe("concurrency", () => {
    it("serializes concurrent acquireNonce calls", async () => {
      provider.getTransactionCount.mockResolvedValue(0);
      const manager = buildManager();
      await manager.initialize();

      const results = await Promise.all([manager.acquireNonce(), manager.acquireNonce(), manager.acquireNonce()]);

      // All nonces should be unique
      expect(new Set(results).size).toBe(3);
      expect(results.sort()).toEqual([0, 1, 2]);
    });
  });
});
