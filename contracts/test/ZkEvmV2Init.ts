import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { TestZkEvmV2, ZkEvmV2Init__factory } from "../typechain-types";
import { INITIAL_WITHDRAW_LIMIT, ONE_DAY_IN_SECONDS } from "./utils/constants";
import { deployUpgradableFromFactory } from "./utils/deployment";
import { getProverTestData } from "./utils/helpers";

describe("ZK EVM V2 contract", () => {
  let zkEvm: TestZkEvmV2;

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  let admin: SignerWithAddress;
  let verifier: string;
  let securityCouncil: SignerWithAddress;
  let operator: SignerWithAddress;

  const { parentStateRootHash, firstBlockNumber } = getProverTestData("Light", "output-file.json");

  async function deployZkEvmFixture() {
    const PlonkVerifierFactory = await ethers.getContractFactory("PlonkVerifier");
    const plonkVerifier = await PlonkVerifierFactory.deploy();
    await plonkVerifier.deployed();

    verifier = plonkVerifier.address;

    const zkEvm = (await deployUpgradableFromFactory(
      "TestZkEvmV2",
      [
        parentStateRootHash,
        firstBlockNumber - 1,
        verifier,
        securityCouncil.address,
        [operator.address],
        ONE_DAY_IN_SECONDS,
        INITIAL_WITHDRAW_LIMIT,
      ],
      {
        initializer: "initialize(bytes32,uint256,address,address,address[],uint256,uint256)",
        unsafeAllow: ["constructor"],
      },
    )) as TestZkEvmV2;

    return { zkEvm };
  }

  before(async () => {
    [admin, securityCouncil, operator] = await ethers.getSigners();
  });

  beforeEach(async () => {
    const contracts = await loadFixture(deployZkEvmFixture);
    zkEvm = contracts.zkEvm;
  });

  describe("Re-initialisation", () => {
    ZkEvmV2Init__factory.createInterface();

    it("Should set the initial block number ", async () => {
      const l2block = ethers.BigNumber.from(12121);
      const l2BlockNumber = await zkEvm.currentL2BlockNumber();
      const zkEvmContract = await deployUpgradableFromFactory("ZkEvmV2Init", [l2block, parentStateRootHash], {
        initializer: "initializeV2(uint256,bytes32)",
        unsafeAllow: ["constructor"],
      });
      const currentL2BlockNumber = await zkEvmContract.currentL2BlockNumber();

      expect(currentL2BlockNumber).to.be.equal(l2block);
      expect(currentL2BlockNumber).to.not.be.equal(l2BlockNumber);
      expect(await zkEvm.periodInSeconds()).to.be.equal(ONE_DAY_IN_SECONDS);
      expect(zkEvmContract.stateRootHashes(l2block)).to.not.be.equal(zkEvm.stateRootHashes(parentStateRootHash));
    });
  });
});
