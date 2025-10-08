import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { getAccountsFixture } from "../../common/helpers";
import { deployLidoStVaultYieldProviderFactory } from "../helpers";
import {
  LidoStVaultYieldProviderFactory,
  MockLineaRollup,
  MockSTETH,
  MockVaultHub,
  TestYieldManager,
} from "contracts/typechain-types";
import { expect } from "chai";
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";

describe("YieldManager contract - control operations", () => {
  let lidoStVaultYieldProviderFactory: LidoStVaultYieldProviderFactory;
  let mockVaultHub: MockVaultHub;
  let mockSTETH: MockSTETH;
  let mockLineaRollup: MockLineaRollup;
  let yieldManager: TestYieldManager;

  let nativeYieldOperator: SignerWithAddress;

  before(async () => {
    ({ nativeYieldOperator } = await loadFixture(getAccountsFixture));
  });

  beforeEach(async () => {
    ({ lidoStVaultYieldProviderFactory, mockVaultHub, mockLineaRollup, yieldManager, mockSTETH } = await loadFixture(
      deployLidoStVaultYieldProviderFactory,
    ));
  });

  describe("Immutables", () => {
    it("Should deploy with correct VaultHub address", async () => {
      expect(await lidoStVaultYieldProviderFactory.VAULT_HUB()).eq(await mockVaultHub.getAddress());
    });
    it("Should deploy with correct STETH address", async () => {
      expect(await lidoStVaultYieldProviderFactory.STETH()).eq(await mockSTETH.getAddress());
    });
    it("Should deploy with correct L1MessageService address", async () => {
      expect(await lidoStVaultYieldProviderFactory.L1_MESSAGE_SERVICE()).eq(await mockLineaRollup.getAddress());
    });
    it("Should deploy with correct YieldManager address", async () => {
      expect(await lidoStVaultYieldProviderFactory.YIELD_MANAGER()).eq(await yieldManager.getAddress());
    });
  });

  describe("createLidoStVaultYieldProvider", () => {
    it("Should successfully create new LidoStVaultYieldProvider and emit expected event", async () => {
      const yieldProviderAddress = await lidoStVaultYieldProviderFactory.createLidoStVaultYieldProvider.staticCall();
      const call = lidoStVaultYieldProviderFactory.connect(nativeYieldOperator).createLidoStVaultYieldProvider();
      await expect(call)
        .to.emit(lidoStVaultYieldProviderFactory, "LidoStVaultYieldProviderCreated")
        .withArgs(yieldProviderAddress);
    });
  });
});
