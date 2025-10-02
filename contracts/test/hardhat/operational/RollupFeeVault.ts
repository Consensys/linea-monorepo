/* eslint-disable @typescript-eslint/no-unused-vars */
import { loadFixture, time } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { L2MessageService, RollupFeeVault, TestDexSwap, TestERC20, TokenBridge } from "../../../typechain-types";
import { getRollupFeeVaultAccountsFixture } from "./helpers/before";
import { deployRollupFeeVaultFixture } from "./helpers/deploy";
import { ethers } from "hardhat";
import { EMPTY_CALLDATA } from "../common/constants";
import { expect } from "chai";
import {
  expectEvent,
  expectRevertWithCustomError,
  expectRevertWithReason,
  generateRandomBytes,
} from "../common/helpers";
import { deployUpgradableFromFactory } from "../common/deployment";
import { ZeroAddress } from "ethers";
import { ROLLUP_FEE_VAULT_INITIALIZE_SIGNATURE } from "./constants";
import { toChecksumAddress } from "@ethereumjs/util";

describe("RollupFeeVault", () => {
  let rollupFeeVault: RollupFeeVault;
  let l2LineaToken: TestERC20;
  let tokenBridge: TokenBridge;
  let messageService: L2MessageService;
  let dex: TestDexSwap;

  let admin: SignerWithAddress;
  let invoiceSetter: SignerWithAddress;
  let burner: SignerWithAddress;
  let operatingCostsReceiver: SignerWithAddress;
  let l1BurnerContract: SignerWithAddress;
  let nonAuthorizedAccount: SignerWithAddress;

  before(async () => {
    ({ admin, invoiceSetter, burner, operatingCostsReceiver, l1BurnerContract, nonAuthorizedAccount } =
      await loadFixture(getRollupFeeVaultAccountsFixture));
  });

  beforeEach(async () => {
    ({ rollupFeeVault, l2LineaToken, tokenBridge, messageService, dex } =
      await loadFixture(deployRollupFeeVaultFixture));
  });

  describe("Fallback/Receive", () => {
    const sendEthToContract = async (value: bigint, data: string) => {
      return admin.sendTransaction({ to: await rollupFeeVault.getAddress(), value, data });
    };

    it("Should send eth to the rollupFeeVault contract through the receive", async () => {
      const value = ethers.parseEther("1");
      await expectEvent(rollupFeeVault, sendEthToContract(value, EMPTY_CALLDATA), "EthReceived", [value]);
    });

    it("Should fail to send eth to the rollupFeeVault contract through the fallback function", async () => {
      const value = ethers.parseEther("1");
      await expectEvent(rollupFeeVault, sendEthToContract(value, "0x1234"), "EthReceived", [value]);
    });
  });

  describe("Initialization", () => {
    it("Should revert if defaultAdmin address is zero address", async () => {
      const randomAddress = generateRandomBytes(20);
      const deployCall = deployUpgradableFromFactory(
        "RollupFeeVault",
        [
          ZeroAddress,
          invoiceSetter.address,
          burner.address,
          operatingCostsReceiver.address,
          randomAddress,
          randomAddress,
          randomAddress,
          randomAddress,
          randomAddress,
        ],
        {
          initializer: ROLLUP_FEE_VAULT_INITIALIZE_SIGNATURE,
          unsafeAllow: ["constructor"],
        },
      );
      await expectRevertWithCustomError(rollupFeeVault, deployCall, "ZeroAddressNotAllowed");
    });

    it("Should revert if invoiceSetter address is zero address", async () => {
      const randomAddress = generateRandomBytes(20);
      const deployCall = deployUpgradableFromFactory(
        "RollupFeeVault",
        [
          admin.address,
          ZeroAddress,
          burner.address,
          operatingCostsReceiver.address,
          randomAddress,
          randomAddress,
          randomAddress,
          randomAddress,
          randomAddress,
        ],
        {
          initializer: ROLLUP_FEE_VAULT_INITIALIZE_SIGNATURE,
          unsafeAllow: ["constructor"],
        },
      );
      await expectRevertWithCustomError(rollupFeeVault, deployCall, "ZeroAddressNotAllowed");
    });

    it("Should revert if burner address is zero address", async () => {
      const randomAddress = generateRandomBytes(20);
      const deployCall = deployUpgradableFromFactory(
        "RollupFeeVault",
        [
          admin.address,
          invoiceSetter.address,
          ZeroAddress,
          operatingCostsReceiver.address,
          randomAddress,
          randomAddress,
          randomAddress,
          randomAddress,
          randomAddress,
        ],
        {
          initializer: ROLLUP_FEE_VAULT_INITIALIZE_SIGNATURE,
          unsafeAllow: ["constructor"],
        },
      );
      await expectRevertWithCustomError(rollupFeeVault, deployCall, "ZeroAddressNotAllowed");
    });

    it("Should revert if operatingCostsReceiver address is zero address", async () => {
      const randomAddress = generateRandomBytes(20);
      const deployCall = deployUpgradableFromFactory(
        "RollupFeeVault",
        [
          admin.address,
          invoiceSetter.address,
          burner.address,
          ZeroAddress,
          randomAddress,
          randomAddress,
          randomAddress,
          randomAddress,
          randomAddress,
        ],
        {
          initializer: ROLLUP_FEE_VAULT_INITIALIZE_SIGNATURE,
          unsafeAllow: ["constructor"],
        },
      );
      await expectRevertWithCustomError(rollupFeeVault, deployCall, "ZeroAddressNotAllowed");
    });

    it("Should revert if tokenBridge address is zero address", async () => {
      const randomAddress = generateRandomBytes(20);
      const deployCall = deployUpgradableFromFactory(
        "RollupFeeVault",
        [
          admin.address,
          invoiceSetter.address,
          burner.address,
          operatingCostsReceiver.address,
          ZeroAddress,
          randomAddress,
          randomAddress,
          randomAddress,
          randomAddress,
        ],
        {
          initializer: ROLLUP_FEE_VAULT_INITIALIZE_SIGNATURE,
          unsafeAllow: ["constructor"],
        },
      );
      await expectRevertWithCustomError(rollupFeeVault, deployCall, "ZeroAddressNotAllowed");
    });

    it("Should revert if messageService address is zero address", async () => {
      const randomAddress = generateRandomBytes(20);
      const deployCall = deployUpgradableFromFactory(
        "RollupFeeVault",
        [
          admin.address,
          invoiceSetter.address,
          burner.address,
          operatingCostsReceiver.address,
          randomAddress,
          ZeroAddress,
          randomAddress,
          randomAddress,
          randomAddress,
        ],
        {
          initializer: ROLLUP_FEE_VAULT_INITIALIZE_SIGNATURE,
          unsafeAllow: ["constructor"],
        },
      );
      await expectRevertWithCustomError(rollupFeeVault, deployCall, "ZeroAddressNotAllowed");
    });

    it("Should revert if l1BurnerContract address is zero address", async () => {
      const randomAddress = generateRandomBytes(20);
      const deployCall = deployUpgradableFromFactory(
        "RollupFeeVault",
        [
          admin.address,
          invoiceSetter.address,
          burner.address,
          operatingCostsReceiver.address,
          randomAddress,
          randomAddress,
          ZeroAddress,
          randomAddress,
          randomAddress,
        ],
        {
          initializer: ROLLUP_FEE_VAULT_INITIALIZE_SIGNATURE,
          unsafeAllow: ["constructor"],
        },
      );
      await expectRevertWithCustomError(rollupFeeVault, deployCall, "ZeroAddressNotAllowed");
    });

    it("Should revert if l2LineaToken address is zero address", async () => {
      const randomAddress = generateRandomBytes(20);
      const deployCall = deployUpgradableFromFactory(
        "RollupFeeVault",
        [
          admin.address,
          invoiceSetter.address,
          burner.address,
          operatingCostsReceiver.address,
          randomAddress,
          randomAddress,
          randomAddress,
          ZeroAddress,
          randomAddress,
        ],
        {
          initializer: ROLLUP_FEE_VAULT_INITIALIZE_SIGNATURE,
          unsafeAllow: ["constructor"],
        },
      );
      await expectRevertWithCustomError(rollupFeeVault, deployCall, "ZeroAddressNotAllowed");
    });

    it("Should revert if DexSwap contract address is zero address", async () => {
      const randomAddress = generateRandomBytes(20);
      const deployCall = deployUpgradableFromFactory(
        "RollupFeeVault",
        [
          admin.address,
          invoiceSetter.address,
          burner.address,
          operatingCostsReceiver.address,
          randomAddress,
          randomAddress,
          randomAddress,
          randomAddress,
          ZeroAddress,
        ],
        {
          initializer: ROLLUP_FEE_VAULT_INITIALIZE_SIGNATURE,
          unsafeAllow: ["constructor"],
        },
      );
      await expectRevertWithCustomError(rollupFeeVault, deployCall, "ZeroAddressNotAllowed");
    });

    it("Should initialize correctly the contract", async () => {
      expect(await rollupFeeVault.hasRole(await rollupFeeVault.DEFAULT_ADMIN_ROLE(), admin.address)).to.equal(true);
      expect(await rollupFeeVault.hasRole(await rollupFeeVault.INVOICE_SETTER_ROLE(), invoiceSetter.address)).to.equal(
        true,
      );
      expect(await rollupFeeVault.hasRole(await rollupFeeVault.BURNER_ROLE(), burner.address)).to.equal(true);
      expect(await rollupFeeVault.operatingCostsReceiver()).to.equal(operatingCostsReceiver.address);
      expect(await rollupFeeVault.tokenBridge()).to.equal(await tokenBridge.getAddress());
      expect(await rollupFeeVault.messageService()).to.equal(await messageService.getAddress());
      expect(await rollupFeeVault.l1BurnerContract()).to.equal(l1BurnerContract.address);
      expect(await rollupFeeVault.lineaToken()).to.equal(await l2LineaToken.getAddress());
      expect(await rollupFeeVault.v3Dex()).to.equal(await dex.getAddress());
    });
  });

  describe("sendOperatingCosts", () => {
    it("Should revert if caller is not invoiceSetter", async () => {
      const startTimestamp = (await time.latest()) + 1;
      const endTimestamp = startTimestamp + 86400;
      await expectRevertWithReason(
        rollupFeeVault.connect(nonAuthorizedAccount).sendOperatingCosts(startTimestamp, endTimestamp, 100n),
        "AccessControl: account " +
          nonAuthorizedAccount.address.toLowerCase() +
          " is missing role " +
          (await rollupFeeVault.INVOICE_SETTER_ROLE()).toLowerCase(),
      );
    });

    it("Should revert if timestamps not in sequence", async () => {
      const lastOperatingCostsUpdate = await rollupFeeVault.lastOperatingCostsUpdate();
      const startTimestamp = lastOperatingCostsUpdate + 2n;
      const endTimestamp = startTimestamp + 86400n;

      await expectRevertWithCustomError(
        rollupFeeVault,
        rollupFeeVault.connect(invoiceSetter).sendOperatingCosts(startTimestamp, endTimestamp, 100n),
        "TimestampsNotInSequence",
      );
    });

    it("Should revert if endTimestamp is before startTimestamp", async () => {
      const lastOperatingCostsUpdate = await rollupFeeVault.lastOperatingCostsUpdate();
      const startTimestamp = lastOperatingCostsUpdate + 1n;
      const endTimestamp = startTimestamp - 1n;

      await expectRevertWithCustomError(
        rollupFeeVault,
        rollupFeeVault.connect(invoiceSetter).sendOperatingCosts(startTimestamp, endTimestamp, 100n),
        "EndTimestampMustBeGreaterThanStartTimestamp",
      );
    });

    it("Should revert if amount is zero", async () => {
      const lastOperatingCostsUpdate = await rollupFeeVault.lastOperatingCostsUpdate();
      const startTimestamp = lastOperatingCostsUpdate + 1n;
      const endTimestamp = startTimestamp + 86400n;

      await expectRevertWithCustomError(
        rollupFeeVault,
        rollupFeeVault.connect(invoiceSetter).sendOperatingCosts(startTimestamp, endTimestamp, 0n),
        "ZeroOperatingCosts",
      );
    });

    it("Should not send operating costs if the contract has no balance", async () => {
      const lastOperatingCostsUpdate = await rollupFeeVault.lastOperatingCostsUpdate();
      const startTimestamp = lastOperatingCostsUpdate + 1n;
      const endTimestamp = startTimestamp + 86400n;
      const operatingCostsAmount = ethers.parseEther("1");

      const operatingCostsReceiverBalanceBefore = await ethers.provider.getBalance(operatingCostsReceiver.address);

      await expectEvent(
        rollupFeeVault,
        rollupFeeVault.connect(invoiceSetter).sendOperatingCosts(startTimestamp, endTimestamp, operatingCostsAmount),
        "InvoiceProcessed",
        [operatingCostsReceiver, startTimestamp, endTimestamp, 0n, operatingCostsAmount],
      );

      const operatingCostsReceiverBalanceAfter = await ethers.provider.getBalance(operatingCostsReceiver.address);
      expect(operatingCostsReceiverBalanceAfter).to.equal(operatingCostsReceiverBalanceBefore);
      expect(await rollupFeeVault.lastOperatingCostsUpdate()).to.equal(endTimestamp);
    });

    it("should send the entire contract balance to the receiver when balanceAvailable < totalAmountCostsOwing", async () => {
      const lastOperatingCostsUpdate = await rollupFeeVault.lastOperatingCostsUpdate();
      const startTimestamp = lastOperatingCostsUpdate + 1n;
      const endTimestamp = startTimestamp + 86400n;
      const operatingCostsAmount = ethers.parseEther("1");
      const balanceAvailable = ethers.parseEther("0.6");

      const operatingCostsReceiverBalanceBefore = await ethers.provider.getBalance(operatingCostsReceiver.address);

      await admin.sendTransaction({ to: await rollupFeeVault.getAddress(), value: balanceAvailable });

      await expectEvent(
        rollupFeeVault,
        rollupFeeVault.connect(invoiceSetter).sendOperatingCosts(startTimestamp, endTimestamp, operatingCostsAmount),
        "InvoiceProcessed",
        [operatingCostsReceiver, startTimestamp, endTimestamp, balanceAvailable, operatingCostsAmount],
      );

      const operatingCostsReceiverBalanceAfter = await ethers.provider.getBalance(operatingCostsReceiver.address);
      expect(operatingCostsReceiverBalanceAfter).to.equal(operatingCostsReceiverBalanceBefore + balanceAvailable);
      expect(await rollupFeeVault.lastOperatingCostsUpdate()).to.equal(endTimestamp);
      expect(await rollupFeeVault.operatingCosts()).to.equal(operatingCostsAmount - balanceAvailable);
    });

    it("should send the entire totalAmountCostsOwing to the receiver when balanceAvailable >= totalAmountCostsOwing", async () => {
      const lastOperatingCostsUpdate = await rollupFeeVault.lastOperatingCostsUpdate();
      const startTimestamp = lastOperatingCostsUpdate + 1n;
      const endTimestamp = startTimestamp + 86400n;
      const operatingCostsAmount = ethers.parseEther("1");
      const balanceAvailable = ethers.parseEther("1.5");

      const operatingCostsReceiverBalanceBefore = await ethers.provider.getBalance(operatingCostsReceiver.address);

      await admin.sendTransaction({ to: await rollupFeeVault.getAddress(), value: balanceAvailable });

      await expectEvent(
        rollupFeeVault,
        rollupFeeVault.connect(invoiceSetter).sendOperatingCosts(startTimestamp, endTimestamp, operatingCostsAmount),
        "InvoiceProcessed",
        [operatingCostsReceiver, startTimestamp, endTimestamp, operatingCostsAmount, operatingCostsAmount],
      );

      const operatingCostsReceiverBalanceAfter = await ethers.provider.getBalance(operatingCostsReceiver.address);
      expect(operatingCostsReceiverBalanceAfter).to.equal(operatingCostsReceiverBalanceBefore + operatingCostsAmount);
      expect(await rollupFeeVault.lastOperatingCostsUpdate()).to.equal(endTimestamp);
      expect(await rollupFeeVault.operatingCosts()).to.equal(0n);
    });
  });

  describe("updateOperatingCosts", () => {
    it("Should revert if caller is not admin", async () => {
      await expectRevertWithReason(
        rollupFeeVault.connect(nonAuthorizedAccount).updateOperatingCosts(100n),
        "AccessControl: account " +
          nonAuthorizedAccount.address.toLowerCase() +
          " is missing role " +
          (await rollupFeeVault.DEFAULT_ADMIN_ROLE()).toLowerCase(),
      );
    });

    it("Should update operationCosts", async () => {
      const newOperatingCosts = 100n;

      await expectEvent(
        rollupFeeVault,
        rollupFeeVault.connect(admin).updateOperatingCosts(newOperatingCosts),
        "OperatingCostsUpdated",
        [newOperatingCosts],
      );

      expect(await rollupFeeVault.operatingCosts()).to.equal(newOperatingCosts);
    });
  });

  describe("updateL1BurnerContract", () => {
    it("Should revert if caller is not admin", async () => {
      const l1BurnerContractAddress = generateRandomBytes(20);
      await expectRevertWithReason(
        rollupFeeVault.connect(nonAuthorizedAccount).updateL1BurnerContract(l1BurnerContractAddress),
        "AccessControl: account " +
          nonAuthorizedAccount.address.toLowerCase() +
          " is missing role " +
          (await rollupFeeVault.DEFAULT_ADMIN_ROLE()).toLowerCase(),
      );
    });

    it("Should revert if l1BurnerContract address is zero address", async () => {
      await expectRevertWithCustomError(
        rollupFeeVault,
        rollupFeeVault.connect(admin).updateL1BurnerContract(ZeroAddress),
        "ZeroAddressNotAllowed",
      );
    });

    it("Should update l1BurnerContract address", async () => {
      const randomAddress = toChecksumAddress(generateRandomBytes(20));
      await expectEvent(
        rollupFeeVault,
        rollupFeeVault.connect(admin).updateL1BurnerContract(randomAddress),
        "L1BurnerContractUpdated",
        [randomAddress],
      );

      expect(await rollupFeeVault.l1BurnerContract()).to.equal(randomAddress);
    });
  });

  describe("updateDex", () => {
    it("Should revert if caller is not admin", async () => {
      const dexAddress = generateRandomBytes(20);
      await expectRevertWithReason(
        rollupFeeVault.connect(nonAuthorizedAccount).updateDex(dexAddress),
        "AccessControl: account " +
          nonAuthorizedAccount.address.toLowerCase() +
          " is missing role " +
          (await rollupFeeVault.DEFAULT_ADMIN_ROLE()).toLowerCase(),
      );
    });

    it("Should revert if Dex address is zero address", async () => {
      await expectRevertWithCustomError(
        rollupFeeVault,
        rollupFeeVault.connect(admin).updateDex(ZeroAddress),
        "ZeroAddressNotAllowed",
      );
    });

    it("Should update Dex address", async () => {
      const randomAddress = toChecksumAddress(generateRandomBytes(20));
      await expectEvent(rollupFeeVault, rollupFeeVault.connect(admin).updateDex(randomAddress), "DexUpdated", [
        randomAddress,
      ]);

      expect(await rollupFeeVault.v3Dex()).to.equal(randomAddress);
    });
  });

  describe("updateOperatingCostsReceiver", () => {
    it("Should revert if caller is not admin", async () => {
      const randomAddress = toChecksumAddress(generateRandomBytes(20));
      await expectRevertWithReason(
        rollupFeeVault.connect(nonAuthorizedAccount).updateOperatingCostsReceiver(randomAddress),
        "AccessControl: account " +
          nonAuthorizedAccount.address.toLowerCase() +
          " is missing role " +
          (await rollupFeeVault.DEFAULT_ADMIN_ROLE()).toLowerCase(),
      );
    });

    it("Should revert if operatingCostsReceiver address is zero address", async () => {
      await expectRevertWithCustomError(
        rollupFeeVault,
        rollupFeeVault.connect(admin).updateOperatingCostsReceiver(ZeroAddress),
        "ZeroAddressNotAllowed",
      );
    });

    it("Should update operatingCostsReceiver address", async () => {
      const randomAddress = toChecksumAddress(generateRandomBytes(20));
      await expectEvent(
        rollupFeeVault,
        rollupFeeVault.connect(admin).updateOperatingCostsReceiver(randomAddress),
        "OperatingCostsReceiverUpdated",
        [randomAddress],
      );

      expect(await rollupFeeVault.operatingCostsReceiver()).to.equal(randomAddress);
    });
  });

  describe("burnAndBridge", () => {
    const INITIAL_CONTRACT_BALANCE = ethers.parseEther("1");
    beforeEach(async () => {
      await admin.sendTransaction({ to: await rollupFeeVault.getAddress(), value: INITIAL_CONTRACT_BALANCE });
    });

    it("Should revert if caller is not burner", async () => {
      const minLineaOut = 200n;
      const deadline = (await time.latest()) + 86400;

      await expectRevertWithReason(
        rollupFeeVault.connect(nonAuthorizedAccount).burnAndBridge(minLineaOut, deadline, 0n),
        "AccessControl: account " +
          nonAuthorizedAccount.address.toLowerCase() +
          " is missing role " +
          (await rollupFeeVault.BURNER_ROLE()).toLowerCase(),
      );
    });

    it("Should revert if operating costs > 0", async () => {
      const lastOperatingCostsUpdate = await rollupFeeVault.lastOperatingCostsUpdate();
      const startTimestamp = lastOperatingCostsUpdate + 1n;
      const endTimestamp = startTimestamp + 86400n;

      await rollupFeeVault
        .connect(invoiceSetter)
        .sendOperatingCosts(startTimestamp, endTimestamp, ethers.parseEther("1.5"));

      const minLineaOut = 200n;
      const deadline = (await time.latest()) + 86400;

      await expectRevertWithCustomError(
        rollupFeeVault,
        rollupFeeVault.connect(burner).burnAndBridge(minLineaOut, deadline, 0n),
        "ZeroOperatingCosts",
      );
    });

    it("Should revert if contract balance is insufficient to cover minimum fee", async () => {
      const lastOperatingCostsUpdate = await rollupFeeVault.lastOperatingCostsUpdate();
      const startTimestamp = lastOperatingCostsUpdate + 1n;
      const endTimestamp = startTimestamp + 86400n;

      const minimumFee = await messageService.minimumFeeInWei();

      // Drain the contract balance
      await rollupFeeVault
        .connect(invoiceSetter)
        .sendOperatingCosts(startTimestamp, endTimestamp, INITIAL_CONTRACT_BALANCE - minimumFee / 2n);

      const minLineaOut = 200n;
      const deadline = (await time.latest()) + 86400;

      await expectRevertWithCustomError(
        rollupFeeVault,
        rollupFeeVault.connect(burner).burnAndBridge(minLineaOut, deadline, 0n),
        "InsufficientBalance",
      );
    });

    it("Should revert if minLineaOut is not met", async () => {
      const lastOperatingCostsUpdate = await rollupFeeVault.lastOperatingCostsUpdate();
      const startTimestamp = lastOperatingCostsUpdate + 1n;
      const endTimestamp = startTimestamp + 86400n;

      await rollupFeeVault
        .connect(invoiceSetter)
        .sendOperatingCosts(startTimestamp, endTimestamp, ethers.parseEther("0.5"));

      const minLineaOut = ethers.parseUnits("200", 18); // Big number to not be met
      const deadline = (await time.latest()) + 86400;

      await expectRevertWithCustomError(
        dex,
        rollupFeeVault.connect(burner).burnAndBridge(minLineaOut, deadline, 0n),
        "MinOutputAmountNotMet",
      );
    });

    it("Should burn ETH, swap to LINEA and bridge the tokens to L1 burner contract", async () => {
      const lastOperatingCostsUpdate = await rollupFeeVault.lastOperatingCostsUpdate();
      const startTimestamp = lastOperatingCostsUpdate + 1n;
      const endTimestamp = startTimestamp + 86400n;

      await rollupFeeVault
        .connect(invoiceSetter)
        .sendOperatingCosts(startTimestamp, endTimestamp, ethers.parseEther("0.5"));

      const minimumFee = await messageService.minimumFeeInWei();
      const balanceAvailable = (await ethers.provider.getBalance(rollupFeeVault.getAddress())) - minimumFee;

      const ethToBurn = (balanceAvailable * 20n) / 100n;

      const minLineaOut = 200n;
      const deadline = (await time.latest()) + 86400;

      await expectEvent(
        rollupFeeVault,
        rollupFeeVault.connect(burner).burnAndBridge(minLineaOut, deadline, 0n),
        "EthBurntSwappedAndBridged",
        [ethToBurn, (balanceAvailable - ethToBurn) * 2n], // We mock the swap to return amountIn * 2
      );
    });
  });
});
