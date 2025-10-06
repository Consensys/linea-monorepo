import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { MockYieldProvider, TestYieldManager } from "contracts/typechain-types";
import { ethers } from "hardhat";

export const setupReceiveCallerForSuccessfulYieldProviderWithdrawal = async (
  testYieldManager: TestYieldManager,
  mockYieldProvider: MockYieldProvider,
  signer: SignerWithAddress,
) => {
  const testYieldManagerAddress = await testYieldManager.getAddress();
  const mockYieldProviderAddress = await mockYieldProvider.getAddress();
  await testYieldManager
    .connect(signer)
    .setYieldProviderReceiveCaller(mockYieldProviderAddress, testYieldManagerAddress);
};

// TODO - Existence of this setup function means that YieldManager has invariants that withdraw cannot underflow for userFunds and userFundsInYieldProvidersTotal
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
  await testYieldManager.connect(signer).fundYieldProvider(mockYieldProviderAddress, withdrawAmount);
};

export const incrementBalance = async (address: string, increment: bigint) => {
  const curBalance = await ethers.provider.getBalance(address);
  await ethers.provider.send("hardhat_setBalance", [address, ethers.toBeHex(curBalance + increment)]);
};
