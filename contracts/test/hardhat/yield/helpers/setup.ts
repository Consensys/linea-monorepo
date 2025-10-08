import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { expect } from "chai";
import { MockYieldProvider, TestLidoStVaultYieldProvider, TestYieldManager } from "contracts/typechain-types";
import { ethers } from "hardhat";

export const setupReceiveCallerForSuccessfulYieldProviderWithdrawal = async (
  testYieldManager: TestYieldManager,
  mockYieldProvider: MockYieldProvider,
  signer: SignerWithAddress,
) => {
  // const testYieldManagerAddress = await testYieldManager.getAddress();
  const mockYieldProviderAddress = await mockYieldProvider.getAddress();
  const mockWithdrawTarget = await testYieldManager.getMockWithdrawTarget(mockYieldProviderAddress);
  await testYieldManager.connect(signer).setYieldProviderReceiveCaller(mockYieldProviderAddress, mockWithdrawTarget);
};

// TODO - Existence of this setup function means that YieldManager has invariants that withdraw cannot underflow for userFunds and userFundsInYieldProvidersTotal
// Caution - assumes it will only be used once, will not work for consecutive uses in its current form
export const fundYieldProviderForWithdrawal = async (
  testYieldManager: TestYieldManager,
  mockYieldProvider: MockYieldProvider,
  signer: SignerWithAddress,
  withdrawAmount: bigint,
) => {
  await setupReceiveCallerForSuccessfulYieldProviderWithdrawal(testYieldManager, mockYieldProvider, signer);
  const mockYieldProviderAddress = await mockYieldProvider.getAddress();
  const yieldManagerAddress = await testYieldManager.getAddress();
  // Funding cannot happen if withdrawal reserve in deficit
  const minimumReserveAmount = await testYieldManager.minimumWithdrawalReserveAmount();
  const l1MessageServiceAddress = await testYieldManager.getL1MessageService();
  await ethers.provider.send("hardhat_setBalance", [l1MessageServiceAddress, ethers.toBeHex(minimumReserveAmount)]);
  await ethers.provider.send("hardhat_setBalance", [yieldManagerAddress, ethers.toBeHex(withdrawAmount)]);
  await testYieldManager.connect(signer).setWithdrawableValueReturnVal(mockYieldProviderAddress, withdrawAmount);
  await testYieldManager.connect(signer).fundYieldProvider(mockYieldProviderAddress, withdrawAmount);
};

export const incrementBalance = async (address: string, increment: bigint) => {
  const curBalance = await ethers.provider.getBalance(address);
  await ethers.provider.send("hardhat_setBalance", [address, ethers.toBeHex(curBalance + increment)]);
};

export const setWithdrawalReserveBalance = async (testYieldManager: TestYieldManager, balance: bigint) => {
  const l1MessageServiceAddress = await testYieldManager.getL1MessageService();
  await ethers.provider.send("hardhat_setBalance", [l1MessageServiceAddress, ethers.toBeHex(balance)]);
};

export const setWithdrawalReserveToMinimum = async (testYieldManager: TestYieldManager) => {
  const minimumReserve = await testYieldManager.getEffectiveMinimumWithdrawalReserve();
  await setWithdrawalReserveBalance(testYieldManager, minimumReserve);
  return minimumReserve;
};

export const ossifyYieldProvider = async (
  yieldManager: TestYieldManager,
  yieldProviderAddress: string,
  securityCouncil: SignerWithAddress,
) => {
  await yieldManager.connect(securityCouncil).initiateOssification(yieldProviderAddress);
  await yieldManager.connect(securityCouncil).processPendingOssification(yieldProviderAddress);
  expect(await yieldManager.isOssified(yieldProviderAddress)).to.be.true;
};

export const fundLidoStVaultYieldProvider = async (
  testYieldManager: TestYieldManager,
  yieldProvider: TestLidoStVaultYieldProvider,
  signer: SignerWithAddress,
  withdrawAmount: bigint,
) => {
  const yieldProviderAddress = await yieldProvider.getAddress();
  const yieldManagerAddress = await testYieldManager.getAddress();
  // Funding cannot happen if withdrawal reserve in deficit
  const minimumReserveAmount = await testYieldManager.minimumWithdrawalReserveAmount();
  const l1MessageServiceAddress = await testYieldManager.getL1MessageService();
  await ethers.provider.send("hardhat_setBalance", [l1MessageServiceAddress, ethers.toBeHex(minimumReserveAmount)]);
  await ethers.provider.send("hardhat_setBalance", [yieldManagerAddress, ethers.toBeHex(withdrawAmount)]);
  await testYieldManager.connect(signer).fundYieldProvider(yieldProviderAddress, withdrawAmount);
};
