import { loadFixture, time } from "@nomicfoundation/hardhat-network-helpers";
import { ethers } from "hardhat";
import { deployFromFactory, deployUpgradableFromFactory } from "../../common/deployment";
import { ROLLUP_REVENUE_VAULT_INITIALIZE_SIGNATURE } from "../constants";
import { DexSwap, L2MessageService, RollupRevenueVault, TestERC20 } from "../../../../typechain-types";
import { getRollupRevenueVaultAccountsFixture } from "./before";
import { deployTokenBridge } from "../../../../scripts/tokenBridge/test/deployTokenBridges";
import { INITIAL_WITHDRAW_LIMIT, L1_L2_MESSAGE_SETTER_ROLE, ONE_DAY_IN_SECONDS } from "../../common/constants";
import { generateRoleAssignments } from "../../../../common/helpers";
import {
  L2_MESSAGE_SERVICE_PAUSE_TYPES_ROLES,
  L2_MESSAGE_SERVICE_ROLES,
  L2_MESSAGE_SERVICE_UNPAUSE_TYPES_ROLES,
} from "../../../../common/constants";
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

  const messageServiceFn = () =>
    deployUpgradableFromFactory("L2MessageService", [
      ONE_DAY_IN_SECONDS,
      INITIAL_WITHDRAW_LIMIT,
      adminAddress,
      roleAddresses,
      L2_MESSAGE_SERVICE_PAUSE_TYPES_ROLES,
      L2_MESSAGE_SERVICE_UNPAUSE_TYPES_ROLES,
    ]);

  const messageService = await loadFixture(messageServiceFn);

  return messageService as unknown as L2MessageService;
}

export async function deployDexSwapFixture(rollupRevenueVaultAddress: string, lineaTokenAddress: string) {
  const testWETH9 = await loadFixture(deployWETH9Fixture);
  const router = await deployFromFactory("TestDexRouter");
  const dexSwap = await deployFromFactory(
    "DexSwap",
    await router.getAddress(),
    testWETH9,
    lineaTokenAddress,
    rollupRevenueVaultAddress,
  );
  return dexSwap as DexSwap;
}

export async function deployRollupRevenueVaultFixture() {
  const { admin, invoiceSubmitter, burner, invoicePaymentReceiver, l1LineaTokenBurner, l1l2MessageSetter } =
    await loadFixture(getRollupRevenueVaultAccountsFixture);

  const messageServiceFn = () => deployL2MessageService(admin.address, l1l2MessageSetter.address);
  const messageService = await loadFixture(messageServiceFn);
  const tokenBridgeFn = async () => deployTokenBridge(await messageService.getAddress(), false);
  const { l2TokenBridge: tokenBridge } = await loadFixture(tokenBridgeFn);

  const l2LineaTokenFn = async () =>
    deployFromFactory("TestERC20", "TestERC20", "TEST", ethers.parseUnits("1000000000", 18)) as unknown as TestERC20;

  const l2LineaToken = await loadFixture(l2LineaTokenFn);

  const rollupRevenueVaultFn = async () =>
    (await deployUpgradableFromFactory(
      "RollupRevenueVault",
      [
        await time.latest(),
        admin.address,
        invoiceSubmitter.address,
        burner.address,
        invoicePaymentReceiver.address,
        await tokenBridge.getAddress(),
        await messageService.getAddress(),
        l1LineaTokenBurner.address,
        await l2LineaToken.getAddress(),
        generateRandomBytes(20), // Will be set after RollupRevenueVault deployment
      ],
      {
        initializer: ROLLUP_REVENUE_VAULT_INITIALIZE_SIGNATURE,
        unsafeAllow: ["constructor", "incorrect-initializer-order"],
      },
    )) as unknown as RollupRevenueVault;

  const rollupRevenueVault = await loadFixture(rollupRevenueVaultFn);

  const dexFn = async () =>
    await deployDexSwapFixture(await rollupRevenueVault.getAddress(), await l2LineaToken.getAddress());

  const dex = await loadFixture(dexFn);
  await rollupRevenueVault.updateDex(dex.getAddress());

  return { rollupRevenueVault, l2LineaToken, tokenBridge, l1LineaTokenBurner, messageService, dex };
}
