import { loadFixture, time } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { TestRateLimiter } from "../typechain-types";
import { DEFAULT_ADMIN_ROLE, ONE_DAY_IN_SECONDS, RATE_LIMIT_SETTER_ROLE } from "./utils/constants";
import { deployUpgradableFromFactory } from "./utils/deployment";

describe("Rate limiter", () => {
  let testRateLimiter: TestRateLimiter;
  let defaultAdmin: SignerWithAddress;
  let resetAdmin: SignerWithAddress;

  const ONE_HUNDRED_ETH = ethers.utils.parseUnits("100", 18);
  const NINE_HUNDRED_ETH = ethers.utils.parseUnits("900", 18);

  async function deployTestRateLimiterFixture(): Promise<TestRateLimiter> {
    return deployUpgradableFromFactory("TestRateLimiter", [86400, ethers.utils.parseEther("1000")], {
      initializer: "initialize(uint256,uint256)",
    }) as Promise<TestRateLimiter>;
  }

  before(async () => {
    [defaultAdmin, resetAdmin] = await ethers.getSigners();
  });

  beforeEach(async () => {
    testRateLimiter = await loadFixture(deployTestRateLimiterFixture);
    await testRateLimiter.grantRole(RATE_LIMIT_SETTER_ROLE, resetAdmin.address);
  });

  describe("Initialization checks", () => {
    it("Deployer has default admin role", async () => {
      expect(await testRateLimiter.hasRole(DEFAULT_ADMIN_ROLE, defaultAdmin.address)).to.be.true;
    });

    it("fails to initialise when not initialising", async () => {
      await expect(testRateLimiter.tryInitialize(86400, ONE_HUNDRED_ETH)).to.be.revertedWith(
        "Initializable: contract is not initializing",
      );
    });

    it("fails to initialise when the limit is zero", async () => {
      await expect(
        deployUpgradableFromFactory("TestRateLimiter", [86400, 0], { initializer: "initialize(uint256,uint256)" }),
      ).to.revertedWithCustomError(testRateLimiter, "LimitIsZero");
    });

    it("fails to initialise when the period is zero", async () => {
      await expect(
        deployUpgradableFromFactory("TestRateLimiter", [0, ethers.utils.parseEther("1000")], {
          initializer: "initialize(uint256,uint256)",
        }),
      ).to.revertedWithCustomError(testRateLimiter, "PeriodIsZero");
    });
  });

  describe("Rate limit values", () => {
    it("currentPeriodAmountInWei increases when amounts withdrawn", async () => {
      expect(await testRateLimiter.currentPeriodAmountInWei()).to.be.equal(0);

      await testRateLimiter.withdrawSomeAmount(ONE_HUNDRED_ETH);
      expect(await testRateLimiter.currentPeriodAmountInWei()).to.be.equal(ONE_HUNDRED_ETH);
    });

    it("currentPeriodAmountInWei increases to the limit", async () => {
      await testRateLimiter.withdrawSomeAmount(ONE_HUNDRED_ETH);
      await testRateLimiter.withdrawSomeAmount(NINE_HUNDRED_ETH);
      expect(await testRateLimiter.currentPeriodAmountInWei()).to.be.equal(ONE_HUNDRED_ETH.add(NINE_HUNDRED_ETH));
    });

    it("withdrawing beyond the limit fails", async () => {
      await testRateLimiter.withdrawSomeAmount(ONE_HUNDRED_ETH.add(NINE_HUNDRED_ETH));
      expect(await testRateLimiter.currentPeriodAmountInWei()).to.be.equal(ONE_HUNDRED_ETH.add(NINE_HUNDRED_ETH));
      await expect(testRateLimiter.withdrawSomeAmount(1)).to.revertedWithCustomError(
        testRateLimiter,
        "RateLimitExceeded",
      );
    });

    it("limit resets with time", async () => {
      await time.increase(ONE_DAY_IN_SECONDS);
      await testRateLimiter.withdrawSomeAmount(ONE_HUNDRED_ETH);
      expect(await testRateLimiter.currentPeriodAmountInWei()).to.be.equal(ONE_HUNDRED_ETH);
    });

    it("limit resets with time", async () => {
      await time.increase(ONE_DAY_IN_SECONDS); //one day in seconds
      await testRateLimiter.withdrawSomeAmount(ONE_HUNDRED_ETH);
      expect(await testRateLimiter.currentPeriodAmountInWei())
        .to.emit(testRateLimiter, "AmountUsedInPeriodReset")
        .withArgs(resetAdmin.address);
    });
  });

  describe("Limit amount resetting", () => {
    it("resetting currentPeriodAmountInWei fails if not admin", async () => {
      await testRateLimiter.withdrawSomeAmount(ONE_HUNDRED_ETH);
      expect(await testRateLimiter.currentPeriodAmountInWei()).to.be.equal(ONE_HUNDRED_ETH);

      await expect(testRateLimiter.resetAmountUsedInPeriod()).to.be.revertedWith(
        "AccessControl: account " + defaultAdmin.address.toLowerCase() + " is missing role " + RATE_LIMIT_SETTER_ROLE,
      );
      expect(await testRateLimiter.currentPeriodAmountInWei()).to.be.equal(ONE_HUNDRED_ETH);
    });

    it("resetting currentPeriodAmountInWei succeeds", async () => {
      expect(await testRateLimiter.currentPeriodAmountInWei()).to.be.equal(0);

      await testRateLimiter.withdrawSomeAmount(ONE_HUNDRED_ETH);
      expect(await testRateLimiter.currentPeriodAmountInWei()).to.be.equal(ONE_HUNDRED_ETH);

      await testRateLimiter.connect(resetAdmin).resetAmountUsedInPeriod();
      expect(await testRateLimiter.currentPeriodAmountInWei()).to.be.equal(0);
    });

    it("resetting limit amount fails if not admin", async () => {
      await testRateLimiter.withdrawSomeAmount(ONE_HUNDRED_ETH);
      expect(await testRateLimiter.currentPeriodAmountInWei()).to.be.equal(ONE_HUNDRED_ETH);

      await expect(testRateLimiter.resetRateLimitAmount(NINE_HUNDRED_ETH)).to.be.revertedWith(
        "AccessControl: account " + defaultAdmin.address.toLowerCase() + " is missing role " + RATE_LIMIT_SETTER_ROLE,
      );
      expect(await testRateLimiter.currentPeriodAmountInWei()).to.be.equal(ONE_HUNDRED_ETH);
    });

    it("resetting limit amount succeeds resetting amount used", async () => {
      await testRateLimiter.withdrawSomeAmount(NINE_HUNDRED_ETH);

      await expect(testRateLimiter.connect(resetAdmin).resetRateLimitAmount(ONE_HUNDRED_ETH))
        .to.emit(testRateLimiter, "LimitAmountChanged")
        .withArgs(resetAdmin.address, ONE_HUNDRED_ETH, true, false);

      expect(await testRateLimiter.currentPeriodAmountInWei()).to.be.equal(ONE_HUNDRED_ETH);
    });

    it("resetting limit amount succeeds without resetting amount used", async () => {
      await testRateLimiter.withdrawSomeAmount(ONE_HUNDRED_ETH);

      await expect(testRateLimiter.connect(resetAdmin).resetRateLimitAmount(ONE_HUNDRED_ETH))
        .to.emit(testRateLimiter, "LimitAmountChanged")
        .withArgs(resetAdmin.address, ONE_HUNDRED_ETH, false, false);

      expect(await testRateLimiter.currentPeriodAmountInWei()).to.be.equal(ONE_HUNDRED_ETH);
    });

    it("Resetting limits with period expired fires the correct event", async () => {
      await time.increase(ONE_DAY_IN_SECONDS);
      await expect(testRateLimiter.connect(resetAdmin).resetRateLimitAmount(ONE_HUNDRED_ETH))
        .to.emit(testRateLimiter, "LimitAmountChanged")
        .withArgs(resetAdmin.address, ONE_HUNDRED_ETH, false, true);
    });

    it("Resetting limits with period sets the used amount to zero", async () => {
      await testRateLimiter.withdrawSomeAmount(ONE_HUNDRED_ETH);
      expect(await testRateLimiter.currentPeriodAmountInWei()).to.be.equal(ONE_HUNDRED_ETH);
      await time.increase(ONE_DAY_IN_SECONDS);
      await testRateLimiter.connect(resetAdmin).resetRateLimitAmount(ONE_HUNDRED_ETH);
      expect(await testRateLimiter.currentPeriodAmountInWei()).to.be.equal(0);
    });
  });
});
