// TODO rename to LineaRollupYieldExtension
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
// import { ethers } from "hardhat";

import { TestYieldManager } from "contracts/typechain-types";
import { deployYieldManagerForUnitTest } from "./helpers/deploy";
import { MINIMUM_FEE, EMPTY_CALLDATA } from "../common/constants";
import {
  // expectEvent,
  // buildAccessErrorMessage,
  expectRevertWithCustomError,
  // expectRevertWithReason,
  getAccountsFixture,
} from "../common/helpers";

describe("Linea Rollup contract", () => {
  let yieldManager: TestYieldManager;

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  // let securityCouncil: SignerWithAddress;
  // let nonAuthorizedAccount: SignerWithAddress;
  let nativeYieldOperator: SignerWithAddress;

  before(async () => {
    ({ nativeYieldOperator } = await loadFixture(getAccountsFixture));
  });

  beforeEach(async () => {
    ({ yieldManager } = await loadFixture(deployYieldManagerForUnitTest));
  });

  describe("Fallback/Receive tests", () => {
    const sendEthToContract = async (data: string) => {
      return nativeYieldOperator.sendTransaction({
        to: await yieldManager.getAddress(),
        value: MINIMUM_FEE,
        data,
      });
    };

    it("Should fail to send eth to the yieldManager contract through the fallback", async () => {
      await expect(sendEthToContract(EMPTY_CALLDATA)).to.be.reverted;
    });

    it("Should fail to send eth to the yieldManager contract through the receive function", async () => {
      await expectRevertWithCustomError(yieldManager, sendEthToContract("0x1234"), "UnexpectedReceiveCaller");
    });
  });

  // describe("Initialisation", () => {
  //   it("Should revert if verifier address is zero address", async () => {
  //     const initializationData = {
  //       initialStateRootHash: parentStateRootHash,
  //       initialL2BlockNumber: INITIAL_MIGRATION_BLOCK,
  //       genesisTimestamp: GENESIS_L2_TIMESTAMP,
  //       defaultVerifier: ADDRESS_ZERO,
  //       rateLimitPeriodInSeconds: ONE_DAY_IN_SECONDS,
  //       rateLimitAmountInWei: INITIAL_WITHDRAW_LIMIT,
  //       roleAddresses,
  //       pauseTypeRoles: LINEA_ROLLUP_PAUSE_TYPES_ROLES,
  //       unpauseTypeRoles: LINEA_ROLLUP_UNPAUSE_TYPES_ROLES,
  //       initialYieldManager: mockYieldManager,
  //       fallbackOperator: FALLBACK_OPERATOR_ADDRESS,
  //       defaultAdmin: securityCouncil.address,
  //     };

  //     const deployCall = deployUpgradableFromFactory("src/rollup/LineaRollup.sol:LineaRollup", [initializationData], {
  //       initializer: LINEA_ROLLUP_INITIALIZE_SIGNATURE,
  //       unsafeAllow: ["constructor", "incorrect-initializer-order"],
  //     });

  //     await expectRevertWithCustomError(lineaRollup, deployCall, "ZeroAddressNotAllowed");
  //   });

  //   it("Should revert if the fallback operator address is zero address", async () => {
  //     const initializationData = {
  //       initialStateRootHash: parentStateRootHash,
  //       initialL2BlockNumber: INITIAL_MIGRATION_BLOCK,
  //       genesisTimestamp: GENESIS_L2_TIMESTAMP,
  //       defaultVerifier: verifier,
  //       rateLimitPeriodInSeconds: ONE_DAY_IN_SECONDS,
  //       rateLimitAmountInWei: INITIAL_WITHDRAW_LIMIT,
  //       roleAddresses: [...roleAddresses.slice(1)],
  //       pauseTypeRoles: LINEA_ROLLUP_PAUSE_TYPES_ROLES,
  //       unpauseTypeRoles: LINEA_ROLLUP_UNPAUSE_TYPES_ROLES,
  //       initialYieldManager: mockYieldManager,
  //       fallbackOperator: ADDRESS_ZERO,
  //       defaultAdmin: securityCouncil.address,
  //     };

  //     const deployCall = deployUpgradableFromFactory("TestLineaRollup", [initializationData], {
  //       initializer: LINEA_ROLLUP_INITIALIZE_SIGNATURE,
  //       unsafeAllow: ["constructor", "incorrect-initializer-order"],
  //     });

  //     await expectRevertWithCustomError(lineaRollup, deployCall, "ZeroAddressNotAllowed");
  //   });

  //   it("Should revert if the default admin address is zero address", async () => {
  //     const initializationData = {
  //       initialStateRootHash: parentStateRootHash,
  //       initialL2BlockNumber: INITIAL_MIGRATION_BLOCK,
  //       genesisTimestamp: GENESIS_L2_TIMESTAMP,
  //       defaultVerifier: verifier,
  //       rateLimitPeriodInSeconds: ONE_DAY_IN_SECONDS,
  //       rateLimitAmountInWei: INITIAL_WITHDRAW_LIMIT,
  //       roleAddresses: [...roleAddresses.slice(1)],
  //       pauseTypeRoles: LINEA_ROLLUP_PAUSE_TYPES_ROLES,
  //       unpauseTypeRoles: LINEA_ROLLUP_UNPAUSE_TYPES_ROLES,
  //       initialYieldManager: mockYieldManager,
  //       fallbackOperator: FALLBACK_OPERATOR_ADDRESS,
  //       defaultAdmin: ADDRESS_ZERO,
  //     };

  //     const deployCall = deployUpgradableFromFactory("TestLineaRollup", [initializationData], {
  //       initializer: LINEA_ROLLUP_INITIALIZE_SIGNATURE,
  //       unsafeAllow: ["constructor", "incorrect-initializer-order"],
  //     });

  //     await expectRevertWithCustomError(lineaRollup, deployCall, "ZeroAddressNotAllowed");
  //   });

  //   it("Should revert if an operator address is zero address", async () => {
  //     const initializationData = {
  //       initialStateRootHash: parentStateRootHash,
  //       initialL2BlockNumber: INITIAL_MIGRATION_BLOCK,
  //       genesisTimestamp: GENESIS_L2_TIMESTAMP,
  //       defaultVerifier: verifier,
  //       rateLimitPeriodInSeconds: ONE_DAY_IN_SECONDS,
  //       rateLimitAmountInWei: INITIAL_WITHDRAW_LIMIT,
  //       roleAddresses: [{ addressWithRole: ADDRESS_ZERO, role: DEFAULT_ADMIN_ROLE }, ...roleAddresses.slice(1)],
  //       pauseTypeRoles: LINEA_ROLLUP_PAUSE_TYPES_ROLES,
  //       unpauseTypeRoles: LINEA_ROLLUP_UNPAUSE_TYPES_ROLES,
  //       initialYieldManager: mockYieldManager,
  //       fallbackOperator: FALLBACK_OPERATOR_ADDRESS,
  //       defaultAdmin: securityCouncil.address,
  //     };

  //     const deployCall = deployUpgradableFromFactory("TestLineaRollup", [initializationData], {
  //       initializer: LINEA_ROLLUP_INITIALIZE_SIGNATURE,
  //       unsafeAllow: ["constructor", "incorrect-initializer-order"],
  //     });

  //     await expectRevertWithCustomError(lineaRollup, deployCall, "ZeroAddressNotAllowed");
  //   });

  //   it("Should revert if an initialYieldManager is zero address", async () => {
  //     const initializationData = {
  //       initialStateRootHash: parentStateRootHash,
  //       initialL2BlockNumber: INITIAL_MIGRATION_BLOCK,
  //       genesisTimestamp: GENESIS_L2_TIMESTAMP,
  //       defaultVerifier: verifier,
  //       rateLimitPeriodInSeconds: ONE_DAY_IN_SECONDS,
  //       rateLimitAmountInWei: INITIAL_WITHDRAW_LIMIT,
  //       roleAddresses: [{ addressWithRole: ADDRESS_ZERO, role: DEFAULT_ADMIN_ROLE }, ...roleAddresses.slice(1)],
  //       pauseTypeRoles: LINEA_ROLLUP_PAUSE_TYPES_ROLES,
  //       unpauseTypeRoles: LINEA_ROLLUP_UNPAUSE_TYPES_ROLES,
  //       initialYieldManager: ADDRESS_ZERO,
  //       fallbackOperator: FALLBACK_OPERATOR_ADDRESS,
  //       defaultAdmin: securityCouncil.address,
  //     };

  //     const deployCall = deployUpgradableFromFactory("TestLineaRollup", [initializationData], {
  //       initializer: LINEA_ROLLUP_INITIALIZE_SIGNATURE,
  //       unsafeAllow: ["constructor", "incorrect-initializer-order"],
  //     });

  //     await expectRevertWithCustomError(lineaRollup, deployCall, "ZeroAddressNotAllowed");
  //   });

  //   it("Should store verifier address in storage", async () => {
  //     ({ verifier, lineaRollup } = await loadFixture(deployLineaRollupFixture));
  //     expect(await lineaRollup.verifiers(0)).to.be.equal(verifier);
  //   });

  //   it("Should store yield manager address in storage", async () => {
  //     ({ verifier, lineaRollup } = await loadFixture(deployLineaRollupFixture));
  //     expect(await lineaRollup.yieldManager()).to.be.equal(mockYieldManager);
  //   });

  //   it("Should assign the OPERATOR_ROLE to operator addresses", async () => {
  //     ({ verifier, lineaRollup } = await loadFixture(deployLineaRollupFixture));
  //     expect(await lineaRollup.hasRole(OPERATOR_ROLE, operator.address)).to.be.true;
  //   });

  //   it("Should assign the VERIFIER_SETTER_ROLE to securityCouncil addresses", async () => {
  //     ({ verifier, lineaRollup } = await loadFixture(deployLineaRollupFixture));
  //     expect(await lineaRollup.hasRole(VERIFIER_SETTER_ROLE, securityCouncil.address)).to.be.true;
  //   });

  //   it("Should assign the VERIFIER_UNSETTER_ROLE to securityCouncil addresses", async () => {
  //     ({ verifier, lineaRollup } = await loadFixture(deployLineaRollupFixture));
  //     expect(await lineaRollup.hasRole(VERIFIER_UNSETTER_ROLE, securityCouncil.address)).to.be.true;
  //   });

  //   it("Should assign the SET_YIELD_MANAGER_ROLE to securityCouncil addresses", async () => {
  //     ({ verifier, lineaRollup } = await loadFixture(deployLineaRollupFixture));
  //     expect(await lineaRollup.hasRole(SET_YIELD_MANAGER_ROLE, securityCouncil.address)).to.be.true;
  //   });

  //   it("Should assign the RESERVE_OPERATOR_ROLE to securityCouncil addresses", async () => {
  //     ({ verifier, lineaRollup } = await loadFixture(deployLineaRollupFixture));
  //     expect(await lineaRollup.hasRole(RESERVE_OPERATOR_ROLE, securityCouncil.address)).to.be.true;
  //   });

  //   it("Should assign the FUNDER_ROLE to securityCouncil addresses", async () => {
  //     ({ verifier, lineaRollup } = await loadFixture(deployLineaRollupFixture));
  //     expect(await lineaRollup.hasRole(FUNDER_ROLE, securityCouncil.address)).to.be.true;
  //   });
  //   it("Should store the startingRootHash in storage for the first block number", async () => {
  //     const initializationData = {
  //       initialStateRootHash: parentStateRootHash,
  //       initialL2BlockNumber: INITIAL_MIGRATION_BLOCK,
  //       genesisTimestamp: GENESIS_L2_TIMESTAMP,
  //       defaultVerifier: verifier,
  //       rateLimitPeriodInSeconds: ONE_DAY_IN_SECONDS,
  //       rateLimitAmountInWei: INITIAL_WITHDRAW_LIMIT,
  //       roleAddresses,
  //       pauseTypeRoles: LINEA_ROLLUP_PAUSE_TYPES_ROLES,
  //       unpauseTypeRoles: LINEA_ROLLUP_UNPAUSE_TYPES_ROLES,
  //       initialYieldManager: mockYieldManager,
  //       fallbackOperator: FALLBACK_OPERATOR_ADDRESS,
  //       defaultAdmin: securityCouncil.address,
  //     };

  //     const lineaRollup = await deployUpgradableFromFactory(
  //       "src/rollup/LineaRollup.sol:LineaRollup",
  //       [initializationData],
  //       {
  //         initializer: LINEA_ROLLUP_INITIALIZE_SIGNATURE,
  //         unsafeAllow: ["constructor", "incorrect-initializer-order"],
  //       },
  //     );

  //     expect(await lineaRollup.stateRootHashes(INITIAL_MIGRATION_BLOCK)).to.be.equal(parentStateRootHash);
  //   });

  //   it("Should assign the VERIFIER_SETTER_ROLE to both SecurityCouncil and Operator", async () => {
  //     const initializationData = {
  //       initialStateRootHash: parentStateRootHash,
  //       initialL2BlockNumber: INITIAL_MIGRATION_BLOCK,
  //       genesisTimestamp: GENESIS_L2_TIMESTAMP,
  //       defaultVerifier: verifier,
  //       rateLimitPeriodInSeconds: ONE_DAY_IN_SECONDS,
  //       rateLimitAmountInWei: INITIAL_WITHDRAW_LIMIT,
  //       roleAddresses: [...roleAddresses, { addressWithRole: operator.address, role: VERIFIER_SETTER_ROLE }],
  //       pauseTypeRoles: LINEA_ROLLUP_PAUSE_TYPES_ROLES,
  //       unpauseTypeRoles: LINEA_ROLLUP_UNPAUSE_TYPES_ROLES,
  //       initialYieldManager: mockYieldManager,
  //       fallbackOperator: FALLBACK_OPERATOR_ADDRESS,
  //       defaultAdmin: securityCouncil.address,
  //     };

  //     const lineaRollup = await deployUpgradableFromFactory(
  //       "src/rollup/LineaRollup.sol:LineaRollup",
  //       [initializationData],
  //       {
  //         initializer: LINEA_ROLLUP_INITIALIZE_SIGNATURE,
  //         unsafeAllow: ["constructor", "incorrect-initializer-order"],
  //       },
  //     );

  //     expect(await lineaRollup.hasRole(VERIFIER_SETTER_ROLE, securityCouncil.address)).to.be.true;
  //     expect(await lineaRollup.hasRole(VERIFIER_SETTER_ROLE, operator.address)).to.be.true;
  //   });

  //   it("Should have the correct contract version", async () => {
  //     ({ verifier, lineaRollup } = await loadFixture(deployLineaRollupFixture));
  //     expect(await lineaRollup.CONTRACT_VERSION()).to.equal("7.0");
  //   });

  //   it("Should revert if the initialize function is called a second time", async () => {
  //     ({ verifier, lineaRollup } = await loadFixture(deployLineaRollupFixture));
  //     const initializeCall = lineaRollup.initialize({
  //       initialStateRootHash: parentStateRootHash,
  //       initialL2BlockNumber: INITIAL_MIGRATION_BLOCK,
  //       genesisTimestamp: GENESIS_L2_TIMESTAMP,
  //       defaultVerifier: verifier,
  //       rateLimitPeriodInSeconds: ONE_DAY_IN_SECONDS,
  //       rateLimitAmountInWei: INITIAL_WITHDRAW_LIMIT,
  //       roleAddresses,
  //       pauseTypeRoles: LINEA_ROLLUP_PAUSE_TYPES_ROLES,
  //       unpauseTypeRoles: LINEA_ROLLUP_UNPAUSE_TYPES_ROLES,
  //       initialYieldManager: mockYieldManager,
  //       fallbackOperator: FALLBACK_OPERATOR_ADDRESS,
  //       defaultAdmin: securityCouncil.address,
  //     });

  //     await expectRevertWithReason(initializeCall, INITIALIZED_ALREADY_MESSAGE);
  //   });
  // });
});
