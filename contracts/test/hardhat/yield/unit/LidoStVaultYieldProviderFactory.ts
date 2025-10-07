import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { getAccountsFixture } from "../../common/helpers";
import { deployLidoStVaultYieldProviderFactory } from "../helpers";
import { LidoStVaultYieldProviderFactory } from "contracts/typechain-types";
import { expect } from "chai";
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";

describe("YieldManager contract - control operations", () => {
  let lidoStVaultYieldProviderFactory: LidoStVaultYieldProviderFactory;
  let beaconAddress: string;

  let nativeYieldOperator: SignerWithAddress;

  before(async () => {
    ({ nativeYieldOperator } = await loadFixture(getAccountsFixture));
  });

  beforeEach(async () => {
    ({ beaconAddress, lidoStVaultYieldProviderFactory } = await loadFixture(deployLidoStVaultYieldProviderFactory));
  });

  describe("Deployment", () => {
    it("Should deploy with correct beacon", async () => {
      expect(await lidoStVaultYieldProviderFactory.BEACON()).eq(beaconAddress);
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
