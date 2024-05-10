import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { TestPauseManager } from "../typechain-types";
import {
  DEFAULT_ADMIN_ROLE,
  GENERAL_PAUSE_TYPE,
  L1_L2_PAUSE_TYPE,
  L2_L1_PAUSE_TYPE,
  PAUSE_MANAGER_ROLE,
  PROVING_SYSTEM_PAUSE_TYPE,
} from "./utils/constants";
import { deployUpgradableFromFactory } from "./utils/deployment";

async function deployTestPauseManagerFixture(): Promise<TestPauseManager> {
  return deployUpgradableFromFactory("TestPauseManager") as unknown as Promise<TestPauseManager>;
}

describe("PauseManager", () => {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  let defaultAdmin: SignerWithAddress;
  let pauseManagerAccount: SignerWithAddress;
  let nonManager: SignerWithAddress;
  let pauseManager: TestPauseManager;

  beforeEach(async () => {
    [defaultAdmin, pauseManagerAccount, nonManager] = await ethers.getSigners();
    pauseManager = await loadFixture(deployTestPauseManagerFixture);

    await pauseManager.grantRole(PAUSE_MANAGER_ROLE, pauseManagerAccount.address);
  });

  describe("Initialization checks", () => {
    it("Deployer has default admin role", async () => {
      expect(await pauseManager.hasRole(DEFAULT_ADMIN_ROLE, defaultAdmin.address)).to.be.true;
    });

    it("Second initialisation while initializing fails", async () => {
      await expect(pauseManager.initialize()).to.be.revertedWith("Initializable: contract is already initialized");
    });
  });

  describe("General pausing", () => {
    // can pause as PAUSE_MANAGER_ROLE
    it("should pause the contract if PAUSE_MANAGER_ROLE", async () => {
      await pauseManager.connect(pauseManagerAccount).pauseByType(GENERAL_PAUSE_TYPE);
      expect(await pauseManager.isPaused(GENERAL_PAUSE_TYPE)).to.be.true;
    });

    // cannot pause as non-PAUSE_MANAGER_ROLE
    it("should revert pause attempt if not PAUSE_MANAGER_ROLE", async () => {
      await expect(pauseManager.connect(nonManager).pauseByType(GENERAL_PAUSE_TYPE)).to.be.revertedWith(
        "AccessControl: account " + nonManager.address.toLowerCase() + " is missing role " + PAUSE_MANAGER_ROLE,
      );
    });

    // can unpause as PAUSE_MANAGER_ROLE
    it("should unpause the contract if PAUSE_MANAGER_ROLE", async () => {
      await pauseManager.connect(pauseManagerAccount).pauseByType(GENERAL_PAUSE_TYPE);
      await pauseManager.connect(pauseManagerAccount).unPauseByType(GENERAL_PAUSE_TYPE);
      expect(await pauseManager.isPaused(GENERAL_PAUSE_TYPE)).to.be.false;
    });

    // cannot unpause as non-PAUSE_MANAGER_ROLE
    it("should revert unpause attempt if not PAUSE_MANAGER_ROLE", async () => {
      await pauseManager.connect(pauseManagerAccount).pauseByType(GENERAL_PAUSE_TYPE);
      await expect(pauseManager.connect(nonManager).unPauseByType(GENERAL_PAUSE_TYPE)).to.be.revertedWith(
        "AccessControl: account " + nonManager.address.toLowerCase() + " is missing role " + PAUSE_MANAGER_ROLE,
      );
    });
  });

  describe("Pause and unpause event emitting", () => {
    it("should pause the L1_L2_PAUSE_TYPE", async () => {
      await expect(pauseManager.connect(pauseManagerAccount).pauseByType(L1_L2_PAUSE_TYPE))
        .to.emit(pauseManager, "Paused")
        .withArgs(pauseManagerAccount.address, L1_L2_PAUSE_TYPE);
    });

    it("should unpause the L1_L2_PAUSE_TYPE", async () => {
      await pauseManager.connect(pauseManagerAccount).pauseByType(L1_L2_PAUSE_TYPE);
      await expect(pauseManager.connect(pauseManagerAccount).unPauseByType(L1_L2_PAUSE_TYPE))
        .to.emit(pauseManager, "UnPaused")
        .withArgs(pauseManagerAccount.address, L1_L2_PAUSE_TYPE);
    });
  });

  describe("Specific type pausing", () => {
    describe("With permissions as PAUSE_MANAGER_ROLE", () => {
      it("should pause the L1_L2_PAUSE_TYPE", async () => {
        await pauseManager.connect(pauseManagerAccount).pauseByType(L1_L2_PAUSE_TYPE);
        expect(await pauseManager.isPaused(L1_L2_PAUSE_TYPE)).to.be.true;
      });

      it("should unpause the L1_L2_PAUSE_TYPE", async () => {
        await pauseManager.connect(pauseManagerAccount).pauseByType(L1_L2_PAUSE_TYPE);
        await pauseManager.connect(pauseManagerAccount).unPauseByType(L1_L2_PAUSE_TYPE);
        expect(await pauseManager.isPaused(L1_L2_PAUSE_TYPE)).to.be.false;
      });

      it("should pause the L2_L1_PAUSE_TYPE", async () => {
        await pauseManager.connect(pauseManagerAccount).pauseByType(L2_L1_PAUSE_TYPE);
        expect(await pauseManager.isPaused(L2_L1_PAUSE_TYPE)).to.be.true;
      });

      it("should unpause the L2_L1_PAUSE_TYPE", async () => {
        await pauseManager.connect(pauseManagerAccount).pauseByType(L2_L1_PAUSE_TYPE);
        await pauseManager.connect(pauseManagerAccount).unPauseByType(L2_L1_PAUSE_TYPE);
        expect(await pauseManager.isPaused(L2_L1_PAUSE_TYPE)).to.be.false;
      });

      it("should pause the PROVING_SYSTEM_PAUSE_TYPE", async () => {
        await pauseManager.connect(pauseManagerAccount).pauseByType(PROVING_SYSTEM_PAUSE_TYPE);
        expect(await pauseManager.isPaused(PROVING_SYSTEM_PAUSE_TYPE)).to.be.true;
      });

      it("should unpause the PROVING_SYSTEM_PAUSE_TYPE", async () => {
        await pauseManager.connect(pauseManagerAccount).pauseByType(PROVING_SYSTEM_PAUSE_TYPE);
        await pauseManager.connect(pauseManagerAccount).unPauseByType(PROVING_SYSTEM_PAUSE_TYPE);
        expect(await pauseManager.isPaused(PROVING_SYSTEM_PAUSE_TYPE)).to.be.false;
      });
    });

    describe("Without permissions - non-PAUSE_MANAGER_ROLE", () => {
      it("cannot pause the L1_L2_PAUSE_TYPE as non-manager", async () => {
        await expect(pauseManager.connect(nonManager).pauseByType(L1_L2_PAUSE_TYPE)).to.be.revertedWith(
          "AccessControl: account " + nonManager.address.toLowerCase() + " is missing role " + PAUSE_MANAGER_ROLE,
        );
      });

      it("cannot unpause the L2_L1_PAUSE_TYPE", async () => {
        await pauseManager.connect(pauseManagerAccount).pauseByType(L1_L2_PAUSE_TYPE);

        await expect(pauseManager.connect(nonManager).unPauseByType(L1_L2_PAUSE_TYPE)).to.be.revertedWith(
          "AccessControl: account " + nonManager.address.toLowerCase() + " is missing role " + PAUSE_MANAGER_ROLE,
        );
      });

      it("cannot pause the L2_L1_PAUSE_TYPE as non-manager", async () => {
        await expect(pauseManager.connect(nonManager).pauseByType(L2_L1_PAUSE_TYPE)).to.be.revertedWith(
          "AccessControl: account " + nonManager.address.toLowerCase() + " is missing role " + PAUSE_MANAGER_ROLE,
        );
      });

      it("cannot unpause the L2_L1_PAUSE_TYPE", async () => {
        await pauseManager.connect(pauseManagerAccount).pauseByType(L2_L1_PAUSE_TYPE);

        await expect(pauseManager.connect(nonManager).unPauseByType(L2_L1_PAUSE_TYPE)).to.be.revertedWith(
          "AccessControl: account " + nonManager.address.toLowerCase() + " is missing role " + PAUSE_MANAGER_ROLE,
        );
      });

      it("cannot pause the PROVING_SYSTEM_PAUSE_TYPE as non-manager", async () => {
        await expect(pauseManager.connect(nonManager).pauseByType(PROVING_SYSTEM_PAUSE_TYPE)).to.be.revertedWith(
          "AccessControl: account " + nonManager.address.toLowerCase() + " is missing role " + PAUSE_MANAGER_ROLE,
        );
      });

      it("cannot unpause the PROVING_SYSTEM_PAUSE_TYPE", async () => {
        await pauseManager.connect(pauseManagerAccount).pauseByType(PROVING_SYSTEM_PAUSE_TYPE);

        await expect(pauseManager.connect(nonManager).unPauseByType(PROVING_SYSTEM_PAUSE_TYPE)).to.be.revertedWith(
          "AccessControl: account " + nonManager.address.toLowerCase() + " is missing role " + PAUSE_MANAGER_ROLE,
        );
      });
    });

    describe("Incorrect states for pausing and unpausing", () => {
      it("Should pause and fail to pause when paused", async () => {
        await pauseManager.connect(pauseManagerAccount).pauseByType(L1_L2_PAUSE_TYPE);

        await expect(
          pauseManager.connect(pauseManagerAccount).pauseByType(L1_L2_PAUSE_TYPE),
        ).to.be.revertedWithCustomError(pauseManager, "IsPaused");
      });

      it("Should allow other types to pause if one is paused", async () => {
        await pauseManager.connect(pauseManagerAccount).pauseByType(L1_L2_PAUSE_TYPE);

        await expect(
          pauseManager.connect(pauseManagerAccount).pauseByType(L1_L2_PAUSE_TYPE),
        ).to.be.revertedWithCustomError(pauseManager, "IsPaused");

        await expect(pauseManager.connect(pauseManagerAccount).pauseByType(PROVING_SYSTEM_PAUSE_TYPE))
          .to.emit(pauseManager, "Paused")
          .withArgs(pauseManagerAccount.address, PROVING_SYSTEM_PAUSE_TYPE);
      });

      it("Should fail to unpause if not paused", async () => {
        await expect(
          pauseManager.connect(pauseManagerAccount).unPauseByType(L1_L2_PAUSE_TYPE),
        ).to.be.revertedWithCustomError(pauseManager, "IsNotPaused");
      });
    });
  });
});
