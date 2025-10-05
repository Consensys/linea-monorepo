import { ethers } from "hardhat";
import { YieldProviderRegistration } from "./types";

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
