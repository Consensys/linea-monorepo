import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { TimeLock } from "../typechain-types";
import { CANCELLER_ROLE, EXECUTOR_ROLE, PROPOSER_ROLE, TIMELOCK_ADMIN_ROLE } from "./utils/constants";
import { deployFromFactory } from "./utils/deployment";

describe("Timelock", () => {
  let contract: TimeLock;
  let proposer: SignerWithAddress;
  let executor: SignerWithAddress;

  async function deployTimeLockFixture() {
    return deployFromFactory(
      "TimeLock",
      10,
      [proposer.address],
      [executor.address],
      ethers.constants.AddressZero,
    ) as Promise<TimeLock>;
  }

  before(async () => {
    [, proposer, executor] = await ethers.getSigners();
  });

  beforeEach(async () => {
    contract = await loadFixture(deployTimeLockFixture);
  });

  describe("Initialization", () => {
    it("Timelock contract should have the 'TIMELOCK_ADMIN_ROLE' role", async () => {
      expect(await contract.hasRole(TIMELOCK_ADMIN_ROLE, contract.address)).to.be.true;
    });

    it("Proposer address should have the 'PROPOSER_ROLE' role", async () => {
      expect(await contract.hasRole(PROPOSER_ROLE, proposer.address)).to.be.true;
    });

    it("Proposer address should have the 'CANCELLER_ROLE' role", async () => {
      expect(await contract.hasRole(CANCELLER_ROLE, proposer.address)).to.be.true;
    });

    it("Executor address should have the 'EXECUTOR_ROLE' role", async () => {
      expect(await contract.hasRole(EXECUTOR_ROLE, executor.address)).to.be.true;
    });

    it("Should set the minDelay state variable with the value passed in the contructor params", async () => {
      expect(await contract.getMinDelay()).to.equal(10);
    });
  });
});
