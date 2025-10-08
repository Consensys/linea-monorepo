import { ethers } from "hardhat";
import { YieldProviderRegistration } from "./types";
import { deployMockWithdrawTarget, deployMockYieldProvider } from "./deploy";
import { TestYieldManager } from "contracts/typechain-types";
import { getAccountsFixture } from "../../common/helpers";

export const buildMockYieldProviderRegistration = (
  overrides: Partial<{
    yieldProviderVendor: number;
    primaryEntrypoint: string;
    ossifiedEntrypoint: string;
    receiveCaller: string;
  }> = {},
): YieldProviderRegistration => ({
  yieldProviderVendor: overrides.yieldProviderVendor ?? 0,
  primaryEntrypoint: overrides.primaryEntrypoint ?? ethers.Wallet.createRandom().address,
  ossifiedEntrypoint: overrides.ossifiedEntrypoint ?? ethers.Wallet.createRandom().address,
  receiveCaller: overrides.receiveCaller ?? ethers.Wallet.createRandom().address,
});

export const addMockYieldProvider = async (yieldManager: TestYieldManager) => {
  const { securityCouncil } = await getAccountsFixture();
  const mockYieldProvider = await deployMockYieldProvider();
  const mockYieldProviderAddress = await mockYieldProvider.getAddress();
  const mockRegistration = buildMockYieldProviderRegistration();
  await yieldManager.connect(securityCouncil).addYieldProvider(mockYieldProviderAddress, mockRegistration);
  const mockWithdrawTarget = await deployMockWithdrawTarget();
  const mockWithdrawTargetAddress = await mockWithdrawTarget.getAddress();
  await yieldManager
    .connect(securityCouncil)
    .setMockWithdrawTarget(mockYieldProviderAddress, mockWithdrawTargetAddress);
  return { mockWithdrawTarget, mockYieldProvider, mockYieldProviderAddress, mockRegistration };
};
