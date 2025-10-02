import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { ethers } from "hardhat";
import { deployFromFactory, deployUpgradableFromFactory } from "../../common/deployment";
import { ROLLUP_FEE_VAULT_INITIALIZE_SIGNATURE } from "../constants";
import { L2MessageService, RollupFeeVault, TestDexSwap, TestERC20 } from "contracts/typechain-types";
import { getRollupFeeVaultAccountsFixture } from "./before";
import { deployTokenBridge } from "contracts/scripts/tokenBridge/test/deployTokenBridges";
import { INITIAL_WITHDRAW_LIMIT, L1_L2_MESSAGE_SETTER_ROLE, ONE_DAY_IN_SECONDS } from "../../common/constants";
import { generateRoleAssignments } from "contracts/common/helpers";
import {
  L2_MESSAGE_SERVICE_PAUSE_TYPES_ROLES,
  L2_MESSAGE_SERVICE_ROLES,
  L2_MESSAGE_SERVICE_UNPAUSE_TYPES_ROLES,
} from "contracts/common/constants";
import { generateRandomBytes } from "../../common/helpers";

export async function deployWETH9Fixture(): Promise<string> {
  const weth9Factory = await ethers.getContractFactory("TestWETH9");
  const weth9 = await weth9Factory.deploy();
  await weth9.waitForDeployment();
  return weth9.getAddress();
}

async function deployL2MessageService(adminAddress: string, l1l2MessageSetterAddress: string) {
  const roleAddresses = generateRoleAssignments(L2_MESSAGE_SERVICE_ROLES, adminAddress, [
    { role: L1_L2_MESSAGE_SETTER_ROLE, addresses: [l1l2MessageSetterAddress] },
  ]);

  const messageService = await deployUpgradableFromFactory("L2MessageService", [
    ONE_DAY_IN_SECONDS,
    INITIAL_WITHDRAW_LIMIT,
    adminAddress,
    roleAddresses,
    L2_MESSAGE_SERVICE_PAUSE_TYPES_ROLES,
    L2_MESSAGE_SERVICE_UNPAUSE_TYPES_ROLES,
  ]);

  return messageService as unknown as L2MessageService;
}

async function deployDexSwapFixture(rollupFeeVaultAddress: string, lineaTokenAddress: string) {
  const testWETH9 = await loadFixture(deployWETH9Fixture);
  const router = await deployFromFactory("TestDexRouter");
  const dexSwap = await deployFromFactory(
    "TestDexSwap",
    rollupFeeVaultAddress,
    lineaTokenAddress,
    testWETH9,
    await router.getAddress(),
  );
  return dexSwap as TestDexSwap;
}

export async function deployRollupFeeVaultFixture() {
  const { admin, invoiceSetter, burner, operatingCostsReceiver, l1BurnerContract, l1l2MessageSetter } =
    await loadFixture(getRollupFeeVaultAccountsFixture);

  const messageService = await deployL2MessageService(admin.address, l1l2MessageSetter.address);
  const { l2TokenBridge: tokenBridge } = await deployTokenBridge(await messageService.getAddress(), false);

  const l2LineaToken = (await deployFromFactory(
    "TestERC20",
    "TestERC20",
    "TEST",
    ethers.parseUnits("1000000000", 18),
  )) as TestERC20;

  const rollupFeeVault = (await deployUpgradableFromFactory(
    "RollupFeeVault",
    [
      admin.address,
      invoiceSetter.address,
      burner.address,
      operatingCostsReceiver.address,
      await tokenBridge.getAddress(),
      await messageService.getAddress(),
      l1BurnerContract.address,
      await l2LineaToken.getAddress(),
      generateRandomBytes(20), // Will be set after RollupFeeVault deployment
    ],
    {
      initializer: ROLLUP_FEE_VAULT_INITIALIZE_SIGNATURE,
      unsafeAllow: ["constructor"],
    },
  )) as unknown as RollupFeeVault;

  const dex = await deployDexSwapFixture(await rollupFeeVault.getAddress(), await l2LineaToken.getAddress());
  await rollupFeeVault.updateDex(dex.getAddress());

  return { rollupFeeVault, l2LineaToken, tokenBridge, l1BurnerContract, messageService, dex };
}
