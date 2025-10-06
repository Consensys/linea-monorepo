import { ethers } from "hardhat";
import { YieldProviderRegistration } from "./types";
import { deployMockYieldProvider } from "./deploy";
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
  const { operationalSafe } = await getAccountsFixture();
  const mockYieldProvider = await deployMockYieldProvider();
  const mockYieldProviderAddress = await mockYieldProvider.getAddress();
  const mockRegistration = buildMockYieldProviderRegistration();
  // Enable delegatecall into MockYieldProvider to call() back into YieldManager
  mockRegistration.receiveCaller = await yieldManager.getAddress();
  await yieldManager.connect(operationalSafe).addYieldProvider(mockYieldProviderAddress, mockRegistration);
  return { mockYieldProvider, mockYieldProviderAddress, mockRegistration };
};
