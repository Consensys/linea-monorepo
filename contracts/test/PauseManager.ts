import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { TestPauseManager } from "../typechain-types";
import {
  DEFAULT_ADMIN_ROLE,
  GENERAL_PAUSE_TYPE,
  INITIALIZED_ALREADY_MESSAGE,
  L1_L2_PAUSE_TYPE,
  L2_L1_PAUSE_TYPE,
  PAUSE_MANAGER_ROLE,
  PROVING_SYSTEM_PAUSE_TYPE,
} from "./utils/constants";
import { deployUpgradableFromFactory } from "./utils/deployment";
import { buildAccessErrorMessage, expectEvent } from "./utils/helpers";

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

  async function pauseByType(pauseType: number, account: SignerWithAddress = pauseManagerAccount) {
    return pauseManager.connect(account).pauseByType(pauseType);
  }

  async function unPauseByType(pauseType: number, account: SignerWithAddress = pauseManagerAccount) {
    return pauseManager.connect(account).unPauseByType(pauseType);
  }

  describe("Initialization checks", () => {
    it("Deployer has default admin role", async () => {
      expect(await pauseManager.hasRole(DEFAULT_ADMIN_ROLE, defaultAdmin.address)).to.be.true;
    });

    it("Second initialisation while initializing fails", async () => {
      await expect(pauseManager.initialize()).to.be.revertedWith(INITIALIZED_ALREADY_MESSAGE);
    });
  });

  describe("General pausing", () => {
    // can pause as PAUSE_MANAGER_ROLE
    it("should pause the contract if PAUSE_MANAGER_ROLE", async () => {
      await pauseByType(GENERAL_PAUSE_TYPE);
      expect(await pauseManager.isPaused(GENERAL_PAUSE_TYPE)).to.be.true;
    });

    // cannot pause as non-PAUSE_MANAGER_ROLE
    it("should revert pause attempt if not PAUSE_MANAGER_ROLE", async () => {
      await expect(pauseByType(GENERAL_PAUSE_TYPE, nonManager)).to.be.revertedWith(
        buildAccessErrorMessage(nonManager, PAUSE_MANAGER_ROLE),
      );
    });

    // can unpause as PAUSE_MANAGER_ROLE
    it("should unpause the contract if PAUSE_MANAGER_ROLE", async () => {
      await pauseByType(GENERAL_PAUSE_TYPE);
      await unPauseByType(GENERAL_PAUSE_TYPE);

      expect(await pauseManager.isPaused(GENERAL_PAUSE_TYPE)).to.be.false;
    });

    // cannot unpause as non-PAUSE_MANAGER_ROLE
    it("should revert unpause attempt if not PAUSE_MANAGER_ROLE", async () => {
      await pauseByType(GENERAL_PAUSE_TYPE);

      await expect(unPauseByType(GENERAL_PAUSE_TYPE, nonManager)).to.be.revertedWith(
        buildAccessErrorMessage(nonManager, PAUSE_MANAGER_ROLE),
      );
    });
  });

  describe("Pause and unpause event emitting", () => {
    it("should pause the L1_L2_PAUSE_TYPE", async () => {
      await expectEvent(pauseManager, pauseByType(L1_L2_PAUSE_TYPE), "Paused", [
        pauseManagerAccount.address,
        L1_L2_PAUSE_TYPE,
      ]);
    });

    it("should unpause the L1_L2_PAUSE_TYPE", async () => {
      await pauseByType(L1_L2_PAUSE_TYPE);

      await expectEvent(pauseManager, unPauseByType(L1_L2_PAUSE_TYPE), "UnPaused", [
        pauseManagerAccount.address,
        L1_L2_PAUSE_TYPE,
      ]);
    });
  });

  describe("Specific type pausing", () => {
    describe("With permissions as PAUSE_MANAGER_ROLE", () => {
      it("should pause the L1_L2_PAUSE_TYPE", async () => {
        await pauseByType(L1_L2_PAUSE_TYPE);
        expect(await pauseManager.isPaused(L1_L2_PAUSE_TYPE)).to.be.true;
      });

      it("should unpause the L1_L2_PAUSE_TYPE", async () => {
        await pauseByType(L1_L2_PAUSE_TYPE);

        await unPauseByType(L1_L2_PAUSE_TYPE);
        expect(await pauseManager.isPaused(L1_L2_PAUSE_TYPE)).to.be.false;
      });

      it("should pause the L2_L1_PAUSE_TYPE", async () => {
        await pauseByType(L2_L1_PAUSE_TYPE);
        expect(await pauseManager.isPaused(L2_L1_PAUSE_TYPE)).to.be.true;
      });

      it("should unpause the L2_L1_PAUSE_TYPE", async () => {
        await pauseByType(L2_L1_PAUSE_TYPE);

        await unPauseByType(L2_L1_PAUSE_TYPE);
        expect(await pauseManager.isPaused(L2_L1_PAUSE_TYPE)).to.be.false;
      });

      it("should pause the PROVING_SYSTEM_PAUSE_TYPE", async () => {
        await pauseByType(PROVING_SYSTEM_PAUSE_TYPE);
        expect(await pauseManager.isPaused(PROVING_SYSTEM_PAUSE_TYPE)).to.be.true;
      });

      it("should unpause the PROVING_SYSTEM_PAUSE_TYPE", async () => {
        await pauseByType(PROVING_SYSTEM_PAUSE_TYPE);

        await unPauseByType(PROVING_SYSTEM_PAUSE_TYPE);
        expect(await pauseManager.isPaused(PROVING_SYSTEM_PAUSE_TYPE)).to.be.false;
      });
    });

    describe("Without permissions - non-PAUSE_MANAGER_ROLE", () => {
      it("cannot pause the L1_L2_PAUSE_TYPE as non-manager", async () => {
        await expect(pauseByType(L1_L2_PAUSE_TYPE, nonManager)).to.be.revertedWith(
          buildAccessErrorMessage(nonManager, PAUSE_MANAGER_ROLE),
        );
      });

      it("cannot unpause the L2_L1_PAUSE_TYPE", async () => {
        await pauseByType(L1_L2_PAUSE_TYPE);

        await expect(unPauseByType(L1_L2_PAUSE_TYPE, nonManager)).to.be.revertedWith(
          buildAccessErrorMessage(nonManager, PAUSE_MANAGER_ROLE),
        );
      });

      it("cannot pause the L2_L1_PAUSE_TYPE as non-manager", async () => {
        await expect(pauseByType(L2_L1_PAUSE_TYPE, nonManager)).to.be.revertedWith(
          buildAccessErrorMessage(nonManager, PAUSE_MANAGER_ROLE),
        );
      });

      it("cannot unpause the L2_L1_PAUSE_TYPE", async () => {
        await pauseByType(L2_L1_PAUSE_TYPE);

        await expect(unPauseByType(L2_L1_PAUSE_TYPE, nonManager)).to.be.revertedWith(
          buildAccessErrorMessage(nonManager, PAUSE_MANAGER_ROLE),
        );
      });

      it("cannot pause the PROVING_SYSTEM_PAUSE_TYPE as non-manager", async () => {
        await expect(pauseByType(PROVING_SYSTEM_PAUSE_TYPE, nonManager)).to.be.revertedWith(
          buildAccessErrorMessage(nonManager, PAUSE_MANAGER_ROLE),
        );
      });

      it("cannot unpause the PROVING_SYSTEM_PAUSE_TYPE", async () => {
        await pauseByType(PROVING_SYSTEM_PAUSE_TYPE);

        await expect(unPauseByType(PROVING_SYSTEM_PAUSE_TYPE, nonManager)).to.be.revertedWith(
          buildAccessErrorMessage(nonManager, PAUSE_MANAGER_ROLE),
        );
      });
    });

    describe("Incorrect states for pausing and unpausing", () => {
      it("Should pause and fail to pause when paused", async () => {
        await pauseByType(L1_L2_PAUSE_TYPE);

        await expect(pauseByType(L1_L2_PAUSE_TYPE)).to.be.revertedWithCustomError(pauseManager, "IsPaused");
      });

      it("Should allow other types to pause if one is paused", async () => {
        await pauseByType(L1_L2_PAUSE_TYPE);

        await expect(pauseByType(L1_L2_PAUSE_TYPE)).to.be.revertedWithCustomError(pauseManager, "IsPaused");

        await expectEvent(pauseManager, pauseByType(PROVING_SYSTEM_PAUSE_TYPE), "Paused", [
          pauseManagerAccount.address,
          PROVING_SYSTEM_PAUSE_TYPE,
        ]);
      });

      it("Should fail to unpause if not paused", async () => {
        await expect(unPauseByType(L1_L2_PAUSE_TYPE)).to.be.revertedWithCustomError(pauseManager, "IsNotPaused");
      });
    });
  });
});
