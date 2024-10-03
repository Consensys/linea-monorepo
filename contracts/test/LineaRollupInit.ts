import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { TestLineaRollup, LineaRollupInit__factory } from "../typechain-types";
import {
  DEFAULT_ADMIN_ROLE,
  GENESIS_L2_TIMESTAMP,
  INITIALIZED_ALREADY_MESSAGE,
  INITIAL_WITHDRAW_LIMIT,
  LINEA_ROLLUP_INITIALIZE_SIGNATURE,
  ONE_DAY_IN_SECONDS,
  OPERATOR_ROLE,
  VERIFIER_SETTER_ROLE,
  pauseTypeRoles,
  unpauseTypeRoles,
} from "./common/constants";
import { deployUpgradableFromFactory } from "./common/deployment";
import { expectRevertWithReason, generateRandomBytes } from "./common/helpers";

describe("LineaRollup Init contract", () => {
  let LineaRollup: TestLineaRollup;

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  let admin: SignerWithAddress;
  let verifier: string;
  let securityCouncil: SignerWithAddress;
  let operator: SignerWithAddress;

  const multiCallAddress = "0xcA11bde05977b3631167028862bE2a173976CA11";

  const parentStateRootHash = generateRandomBytes(32);

  const firstBlockNumber = 199;

  async function deployLineaRollupFixture() {
    const PlonkVerifierFactory = await ethers.getContractFactory("TestPlonkVerifierForDataAggregation");
    const plonkVerifier = await PlonkVerifierFactory.deploy();
    await plonkVerifier.waitForDeployment();

    verifier = await plonkVerifier.getAddress();

    const genesisData = {
      initialStateRootHash: parentStateRootHash,
      initialL2BlockNumber: firstBlockNumber - 1,
      genesisTimestamp: GENESIS_L2_TIMESTAMP,
      defaultVerifier: verifier,
      rateLimitPeriodInSeconds: ONE_DAY_IN_SECONDS,
      rateLimitAmountInWei: INITIAL_WITHDRAW_LIMIT,
      roleAddresses: [
        { addressWithRole: securityCouncil.address, role: DEFAULT_ADMIN_ROLE },
        { addressWithRole: securityCouncil.address, role: VERIFIER_SETTER_ROLE },
        { addressWithRole: operator.address, role: OPERATOR_ROLE },
      ],
      pauseTypeRoles: pauseTypeRoles,
      unpauseTypeRoles: unpauseTypeRoles,
      fallbackOperator: multiCallAddress,
    };

    const LineaRollup = (await deployUpgradableFromFactory("TestLineaRollup", [genesisData], {
      initializer: LINEA_ROLLUP_INITIALIZE_SIGNATURE,
    })) as unknown as TestLineaRollup;

    return { LineaRollup };
  }

  before(async () => {
    [admin, securityCouncil, operator] = await ethers.getSigners();
  });

  beforeEach(async () => {
    const contracts = await loadFixture(deployLineaRollupFixture);
    LineaRollup = contracts.LineaRollup;
  });

  describe("Re-initialisation", () => {
    LineaRollupInit__factory.createInterface();

    it("Should set the initial block number", async () => {
      const l2block = 12121n;
      const l2BlockNumber = await LineaRollup.currentL2BlockNumber();
      const lineaRollupContract = await deployUpgradableFromFactory("LineaRollupInit", [l2block, parentStateRootHash], {
        initializer: "initializeV2(uint256,bytes32)",
        unsafeAllow: ["constructor"],
      });
      const currentL2BlockNumber = await lineaRollupContract.currentL2BlockNumber();

      expect(currentL2BlockNumber).to.be.equal(l2block);
      expect(currentL2BlockNumber).to.not.be.equal(l2BlockNumber);
      expect(await LineaRollup.periodInSeconds()).to.be.equal(ONE_DAY_IN_SECONDS);
      expect(lineaRollupContract.stateRootHashes(l2block)).to.not.be.equal(
        LineaRollup.stateRootHashes(parentStateRootHash),
      );
    });

    it("Cannot initialize twice", async () => {
      const l2block = 12121n;
      const l2BlockNumber = await LineaRollup.currentL2BlockNumber();
      const lineaRollupContract = await deployUpgradableFromFactory("LineaRollupInit", [l2block, parentStateRootHash], {
        initializer: "initializeV2(uint256,bytes32)",
        unsafeAllow: ["constructor"],
      });

      await expectRevertWithReason(
        lineaRollupContract.initializeV2(l2BlockNumber, parentStateRootHash),
        INITIALIZED_ALREADY_MESSAGE,
      );
    });
  });
});
