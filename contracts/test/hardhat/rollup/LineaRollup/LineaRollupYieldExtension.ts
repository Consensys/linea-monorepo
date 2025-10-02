// TODO rename to LineaRollupYieldExtension
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";

import { TestLineaRollup } from "contracts/typechain-types";
import { deployLineaRollupFixture, getAccountsFixture } from "../helpers";
import {
  ADDRESS_ZERO,
  FUNDER_ROLE,
  GENERAL_PAUSE_TYPE,
  L1_L2_PAUSE_TYPE,
  NATIVE_YIELD_STAKING_PAUSE_TYPE,
  RESERVE_OPERATOR_ROLE,
  SET_YIELD_MANAGER_ROLE,
} from "../../common/constants";
import {
  expectEvent,
  buildAccessErrorMessage,
  expectRevertWithCustomError,
  expectRevertWithReason,
} from "../../common/helpers";

describe("Linea Rollup contract", () => {
  let lineaRollup: TestLineaRollup;
  let mockYieldManager: string;
  let operator: SignerWithAddress;

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  let admin: SignerWithAddress;
  let securityCouncil: SignerWithAddress;
  let nonAuthorizedAccount: SignerWithAddress;

  before(async () => {
    ({ admin, securityCouncil, operator, nonAuthorizedAccount } = await loadFixture(getAccountsFixture));
  });

  beforeEach(async () => {
    ({ mockYieldManager, lineaRollup } = await loadFixture(deployLineaRollupFixture));
  });

  describe("Change yield manager address", () => {
    it("Should revert if the caller does not the SET_YIELD_MANAGER_ROLE", async () => {
      const newYieldManager = ethers.Wallet.createRandom().address;
      const setYieldManagerCall = lineaRollup.connect(nonAuthorizedAccount).setYieldManager(newYieldManager);

      await expectRevertWithReason(
        setYieldManagerCall,
        buildAccessErrorMessage(nonAuthorizedAccount, SET_YIELD_MANAGER_ROLE),
      );
    });

    it("Should revert if the address being set is the zero address", async () => {
      const setYieldManagerCall = lineaRollup.connect(securityCouncil).setYieldManager(ADDRESS_ZERO);
      await expectRevertWithCustomError(lineaRollup, setYieldManagerCall, "ZeroAddressNotAllowed");
    });

    it("Should set the new yield manager address", async () => {
      const newYieldManager = ethers.Wallet.createRandom().address;
      await lineaRollup.connect(securityCouncil).setYieldManager(newYieldManager);
      expect(await lineaRollup.yieldManager()).to.be.equal(newYieldManager);
    });

    it("Should emit the correct event", async () => {
      const oldYieldManagerAddress = await lineaRollup.yieldManager();
      const newYieldManagerAddress = ethers.Wallet.createRandom().address;
      const setYieldManagerCall = lineaRollup.connect(securityCouncil).setYieldManager(newYieldManagerAddress);

      await expectEvent(lineaRollup, setYieldManagerCall, "YieldManagerChanged", [
        oldYieldManagerAddress,
        newYieldManagerAddress,
        securityCouncil.address,
      ]);
    });
  });

  describe("IS_WITHDRAW_LST_ALLOWED toggle", () => {
    it("isWithdrawLSTAllowed should return false", async () => {
      expect(await lineaRollup.isWithdrawLSTAllowed()).to.be.false;
    });
  });

  describe("fund() to receive funding", () => {
    it("Should revert if the caller does not have the FUNDER_ROLE", async () => {
      const fundCall = lineaRollup.connect(nonAuthorizedAccount).fund({ value: ethers.parseEther("1") });

      await expectRevertWithReason(fundCall, buildAccessErrorMessage(nonAuthorizedAccount, FUNDER_ROLE));
    });

    it("Should succeed if the caller has the FUNDER_ROLE, and emit the correct event", async () => {
      const amount = ethers.parseEther("1");
      const fundCall = lineaRollup.connect(securityCouncil).fund({ value: amount });

      await expectEvent(lineaRollup, fundCall, "FundingReceived", [securityCouncil.address, amount]);

      const lineaRollupBalance = await ethers.provider.getBalance(await lineaRollup.getAddress());
      expect(lineaRollupBalance).to.equal(amount);
    });
  });

  describe("transferFundsForNativeYield() to transfer funds to YieldManager", () => {
    it("Should revert if the caller does not the RESERVE_OPERATOR_ROLE", async () => {
      const transferCall = lineaRollup
        .connect(nonAuthorizedAccount)
        .transferFundsForNativeYield(ethers.parseEther("1"));

      await expectRevertWithReason(transferCall, buildAccessErrorMessage(nonAuthorizedAccount, RESERVE_OPERATOR_ROLE));
    });

    it("Should revert if GENERAL_PAUSE_TYPE is enabled", async () => {
      await lineaRollup.connect(securityCouncil).pauseByType(GENERAL_PAUSE_TYPE);

      const transferCall = lineaRollup.connect(securityCouncil).transferFundsForNativeYield(0n);

      await expectRevertWithCustomError(lineaRollup, transferCall, "IsPaused", [GENERAL_PAUSE_TYPE]);
    });

    it("Should revert if NATIVE_YIELD_STAKING pause type is enabled", async () => {
      await lineaRollup.connect(securityCouncil).pauseByType(NATIVE_YIELD_STAKING_PAUSE_TYPE);

      const transferCall = lineaRollup.connect(securityCouncil).transferFundsForNativeYield(0n);

      await expectRevertWithCustomError(lineaRollup, transferCall, "IsPaused", [NATIVE_YIELD_STAKING_PAUSE_TYPE]);
    });

    it("Should revert if LineaRollup has balance < _amount", async () => {
      const amount = ethers.parseEther("1");
      const transferCall = lineaRollup.connect(securityCouncil).transferFundsForNativeYield(amount);

      await expect(transferCall).to.be.reverted;
    });

    it("Should successfully transfer ETH to the YieldManager", async () => {
      const amount = ethers.parseEther("1");
      await lineaRollup.connect(securityCouncil).fund({ value: amount });

      const lineaRollupAddress = await lineaRollup.getAddress();
      const initialContractBalance = await ethers.provider.getBalance(lineaRollupAddress);
      const initialYieldManagerBalance = await ethers.provider.getBalance(mockYieldManager);

      await lineaRollup.connect(securityCouncil).transferFundsForNativeYield(amount);

      const finalContractBalance = await ethers.provider.getBalance(lineaRollupAddress);
      const finalYieldManagerBalance = await ethers.provider.getBalance(mockYieldManager);

      expect(finalContractBalance).to.equal(initialContractBalance - amount);
      expect(finalYieldManagerBalance).to.equal(initialYieldManagerBalance + amount);
    });
  });

  describe("Yield reporting", () => {
    async function impersonateYieldManager() {
      await ethers.provider.send("hardhat_impersonateAccount", [mockYieldManager]);
      await ethers.provider.send("hardhat_setBalance", [mockYieldManager, ethers.toBeHex(ethers.parseEther("1"))]);
      return await ethers.getSigner(mockYieldManager);
    }

    async function stopYieldManagerImpersonation() {
      await ethers.provider.send("hardhat_stopImpersonatingAccount", [mockYieldManager]);
    }

    it("Should revert if caller is not the YieldManager", async () => {
      const reportCall = lineaRollup.connect(securityCouncil).reportNativeYield(1n, operator.address);

      await expectRevertWithCustomError(lineaRollup, reportCall, "CallerIsNotYieldManager");
    });

    it("Should revert if GENERAL_PAUSE_TYPE is enabled", async () => {
      const yieldManagerSigner = await impersonateYieldManager();
      await lineaRollup.connect(securityCouncil).pauseByType(GENERAL_PAUSE_TYPE);

      const reportCall = lineaRollup.connect(yieldManagerSigner).reportNativeYield(1n, operator.address);

      await expectRevertWithCustomError(lineaRollup, reportCall, "IsPaused", [GENERAL_PAUSE_TYPE]);

      await stopYieldManagerImpersonation();
    });

    it("Should revert if L1_L2 pause type is enabled", async () => {
      const yieldManagerSigner = await impersonateYieldManager();
      await lineaRollup.connect(securityCouncil).pauseByType(L1_L2_PAUSE_TYPE);

      const reportCall = lineaRollup.connect(yieldManagerSigner).reportNativeYield(1n, operator.address);

      await expectRevertWithCustomError(lineaRollup, reportCall, "IsPaused", [L1_L2_PAUSE_TYPE]);

      await stopYieldManagerImpersonation();
    });

    it("Should revert if l2YieldRecipient is the 0 address", async () => {
      const yieldManagerSigner = await impersonateYieldManager();

      const reportCall = lineaRollup.connect(yieldManagerSigner).reportNativeYield(1n, ADDRESS_ZERO);

      await expectRevertWithCustomError(lineaRollup, reportCall, "ZeroAddressNotAllowed");

      await stopYieldManagerImpersonation();
    });

    it("Should successfully emit a synthetic MessageSent event with valid parameters", async () => {
      const yieldManagerSigner = await impersonateYieldManager();
      const amount = ethers.parseEther("1");
      const lineaRollupAddress = await lineaRollup.getAddress();
      const nextMessageNumberBefore = await lineaRollup.nextMessageNumber();
      const l2YieldRecipient = ethers.Wallet.createRandom().address;

      const expectedMessageHash = ethers.keccak256(
        ethers.AbiCoder.defaultAbiCoder().encode(
          ["address", "address", "uint256", "uint256", "uint256", "bytes"],
          [lineaRollupAddress, l2YieldRecipient, 0, amount, nextMessageNumberBefore, "0x"],
        ),
      );

      const reportCall = lineaRollup.connect(yieldManagerSigner).reportNativeYield(amount, l2YieldRecipient);

      await expectEvent(lineaRollup, reportCall, "MessageSent", [
        // _from should be from YieldManager
        mockYieldManager,
        // _to should be from _l2YieldRecipient function param
        l2YieldRecipient,
        // _fee should be 0
        0n,
        // _value should be _amount function param
        amount,
        // _nonce should be nextMessageNumberBefore
        nextMessageNumberBefore,
        // _calldata should be empty bytes
        "0x",
        expectedMessageHash,
      ]);

      expect(await lineaRollup.nextMessageNumber()).to.equal(nextMessageNumberBefore + 1n);

      await stopYieldManagerImpersonation();
    });
  });
});
