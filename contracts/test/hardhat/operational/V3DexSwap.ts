import { ethers, network } from "hardhat";
import { toChecksumAddress } from "@ethereumjs/util";
import { expect } from "chai";
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { loadFixture, time } from "@nomicfoundation/hardhat-network-helpers";
import { deployWETH9Fixture } from "./helpers/deploy";
import { deployFromFactory } from "../common/deployment";
import { V3DexSwap, TestDexRouter, TestERC20 } from "../../../typechain-types";
import { expectRevertWithCustomError, generateRandomBytes } from "../common/helpers";
import { ADDRESS_ZERO, ONE_MINUTE_IN_SECONDS } from "../common/constants";

describe("V3DexSwap", () => {
  let dexSwap: V3DexSwap;
  let admin: SignerWithAddress;
  let rollupRevenueVault: SignerWithAddress;
  let lineaToken: TestERC20;
  let testWETH9Address: string;
  let router: TestDexRouter;

  async function deployDexSwapContractFixture() {
    const [, rollupRevenueVault] = await ethers.getSigners();

    const lineaToken = (await deployFromFactory(
      "TestERC20",
      "TestERC20",
      "TEST",
      ethers.parseUnits("1000000000", 18),
    )) as TestERC20;

    const testWETH9 = await deployWETH9Fixture();

    const router = (await deployFromFactory("TestDexRouter")) as TestDexRouter;

    const dexSwap = (await deployFromFactory(
      "V3DexSwap",
      await router.getAddress(),
      testWETH9,
      await lineaToken.getAddress(),
      rollupRevenueVault.address,
      50,
    )) as V3DexSwap;

    return { dexSwap, lineaToken, testWETH9, router };
  }

  before(async () => {
    await network.provider.send("hardhat_reset");
    [admin, rollupRevenueVault] = await ethers.getSigners();
  });

  beforeEach(async () => {
    ({ dexSwap, lineaToken, testWETH9: testWETH9Address, router } = await loadFixture(deployDexSwapContractFixture));
  });

  describe("construtor", () => {
    it("Should revert when rollupRevenueVault address is zero", async () => {
      const randomnAddress = toChecksumAddress(generateRandomBytes(20));
      await expectRevertWithCustomError(
        dexSwap,
        deployFromFactory("V3DexSwap", ADDRESS_ZERO, randomnAddress, randomnAddress, randomnAddress, 50),
        "ZeroAddressNotAllowed",
      );
    });

    it("Should revert when WETH token address is zero", async () => {
      const randomnAddress = toChecksumAddress(generateRandomBytes(20));
      await expectRevertWithCustomError(
        dexSwap,
        deployFromFactory("V3DexSwap", randomnAddress, ADDRESS_ZERO, randomnAddress, randomnAddress, 50),
        "ZeroAddressNotAllowed",
      );
    });

    it("Should revert when LINEA token address is zero", async () => {
      const randomnAddress = toChecksumAddress(generateRandomBytes(20));
      await expectRevertWithCustomError(
        dexSwap,
        deployFromFactory("V3DexSwap", randomnAddress, randomnAddress, ADDRESS_ZERO, randomnAddress, 50),
        "ZeroAddressNotAllowed",
      );
    });

    it("Should revert when Dex router address is zero", async () => {
      const randomnAddress = toChecksumAddress(generateRandomBytes(20));
      await expectRevertWithCustomError(
        dexSwap,
        deployFromFactory("V3DexSwap", randomnAddress, randomnAddress, randomnAddress, ADDRESS_ZERO, 50),
        "ZeroAddressNotAllowed",
      );
    });

    it("Should revert when tick spacing is zero", async () => {
      const randomnAddress = toChecksumAddress(generateRandomBytes(20));
      await expectRevertWithCustomError(
        dexSwap,
        deployFromFactory("V3DexSwap", randomnAddress, randomnAddress, randomnAddress, randomnAddress, 0),
        "ZeroTickSpacingNotAllowed",
      );
    });

    it("Should set the correct addresses", async () => {
      const lineaTokenAddress = await dexSwap.LINEA_TOKEN();
      const wethTokenAddress = await dexSwap.WETH_TOKEN();
      const rollupRevenueVaultAddress = await dexSwap.ROLLUP_REVENUE_VAULT();
      const routerAddress = await dexSwap.ROUTER();

      expect(lineaTokenAddress).to.equal(await lineaToken.getAddress());
      expect(wethTokenAddress).to.equal(testWETH9Address);
      expect(rollupRevenueVaultAddress).to.equal(rollupRevenueVault.address);
      expect(routerAddress).to.equal(await router.getAddress());
    });
  });

  describe("swap", () => {
    it("Should revert when msg.value == 0", async () => {
      const minLineaOut = 200n;
      const deadline = (await time.latest()) + ONE_MINUTE_IN_SECONDS;
      await expectRevertWithCustomError(
        dexSwap,
        dexSwap.connect(admin).swap(minLineaOut, deadline, 0n, { value: 0n }),
        "NoEthSend",
      );
    });

    it("Should swap ETH to LINEA tokens", async () => {
      const minLineaOut = ethers.parseUnits("2", 18);
      const deadline = (await time.latest()) + ONE_MINUTE_IN_SECONDS;

      const ethValueToSwap = ethers.parseEther("1");
      const rollupRevenueVaultLineaTokensBalanceBefore = await lineaToken.balanceOf(rollupRevenueVault.address);
      await dexSwap.connect(admin).swap(minLineaOut, deadline, 0n, { value: ethValueToSwap });

      const rollupRevenueVaultLineaTokensBalanceAfter = await lineaToken.balanceOf(rollupRevenueVault.address);
      expect(rollupRevenueVaultLineaTokensBalanceAfter).to.equal(
        rollupRevenueVaultLineaTokensBalanceBefore + ethValueToSwap * 2n, // 1 ETH = 2 LINEA in the TestDexRouter
      );
    });
  });
});
