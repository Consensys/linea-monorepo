// // Test scenarios with LineaRollup + YieldManager + LidoStVaultYieldProvider
// import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
// import { expectRevertWithCustomError, getAccountsFixture } from "../../common/helpers";
// import {
//   deployAndAddSingleLidoStVaultYieldProvider,
//   deployMockStakingVault,
//   fundLidoStVaultYieldProvider,
//   getWithdrawLSTCall,
//   incrementBalance,
//   ossifyYieldProvider,
//   setWithdrawalReserveToMinimum,
//   YieldProviderRegistration,
// } from "../helpers";
// import {
//   MockVaultHub,
//   MockSTETH,
//   MockLineaRollup,
//   TestYieldManager,
//   MockDashboard,
//   MockStakingVault,
//   TestLidoStVaultYieldProvider,
//   TestCLProofVerifier,
//   SSZMerkleTree,
// } from "contracts/typechain-types";
// import { expect } from "chai";
// import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
// import { ethers } from "hardhat";
// import { parseUnits, ZeroAddress } from "ethers";
// import {
//   GI_FIRST_VALIDATOR,
//   GI_FIRST_VALIDATOR_AFTER_CHANGE,
//   CHANGE_SLOT,
//   ONE_ETHER,
//   ZERO_VALUE,
//   LIDO_ST_VAULT_YIELD_PROVIDER_VENDOR,
//   UNUSED_YIELD_PROVIDER_VENDOR,
//   LIDO_DASHBOARD_NOT_LINKED_TO_VAULT,
//   LIDO_VAULT_IS_EXPECTED_RECEIVE_CALLER_AND_OSSIFIED_ENTRYPOINT,
//   EMPTY_CALLDATA,
//   VALIDATOR_WITNESS_TYPE,
//   THIRTY_TWO_ETH_IN_GWEI,
//   ONE_GWEI,
// } from "../../common/constants";
// import { generateLidoUnstakePermissionlessWitness } from "../helpers/proof";

// describe("Integration tests with LineaRollup, YieldManager and LidoStVaultYieldProvider", () => {
//   let yieldProvider: TestLidoStVaultYieldProvider;
//   let nativeYieldOperator: SignerWithAddress;
//   let securityCouncil: SignerWithAddress;
//   let mockVaultHub: MockVaultHub;
//   let mockSTETH: MockSTETH;
//   let mockLineaRollup: MockLineaRollup;
//   let yieldManager: TestYieldManager;
//   let mockDashboard: MockDashboard;
//   let mockStakingVault: MockStakingVault;
//   let sszMerkleTree: SSZMerkleTree;
//   let verifier: TestCLProofVerifier;

//   let l1MessageServiceAddress: string;
//   let yieldManagerAddress: string;
//   let vaultHubAddress: string;
//   let stethAddress: string;
//   let mockDashboardAddress: string;
//   let mockStakingVaultAddress: string;
//   let yieldProviderAddress: string;
//   before(async () => {
//     ({ nativeYieldOperator, securityCouncil } = await loadFixture(getAccountsFixture));
//   });

//   beforeEach(async () => {
//     ({
//       yieldProvider,
//       yieldProviderAddress,
//       mockDashboard,
//       mockStakingVault,
//       yieldManager,
//       mockVaultHub,
//       mockSTETH,
//       mockLineaRollup,
//       sszMerkleTree,
//       verifier,
//     } = await loadFixture(deployAndAddSingleLidoStVaultYieldProvider));

//     l1MessageServiceAddress = await mockLineaRollup.getAddress();
//     yieldManagerAddress = await yieldManager.getAddress();
//     vaultHubAddress = await mockVaultHub.getAddress();
//     stethAddress = await mockSTETH.getAddress();
//     mockDashboardAddress = await mockDashboard.getAddress();
//     mockStakingVaultAddress = await mockStakingVault.getAddress();
//   });

//   describe("Constructor", () => {
//     // it("Should revert if 0 address provided for _l1MessageService", async () => {
//     //   const contractFactory = await ethers.getContractFactory("TestLidoStVaultYieldProvider");
//     //   const call = contractFactory.deploy(
//     //     ZeroAddress,
//     //     yieldManagerAddress,
//     //     vaultHubAddress,
//     //     stethAddress,
//     //     GI_FIRST_VALIDATOR,
//     //     GI_FIRST_VALIDATOR_AFTER_CHANGE,
//     //     CHANGE_SLOT,
//     //   );
//     //   await expectRevertWithCustomError(contractFactory, call, "ZeroAddressNotAllowed");
//     // });
//   });
// });
