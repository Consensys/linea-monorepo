import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { MockYieldProvider, TestYieldManager } from "contracts/typechain-types";

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
export const setupSuccessfulYieldProviderWithdrawal = async (
  testYieldManager: TestYieldManager,
  mockYieldProvider: MockYieldProvider,
  signer: SignerWithAddress,
  withdrawAmount: bigint,
) => {
  await setupReceiveCallerForSuccessfulYieldProviderWithdrawal(testYieldManager, mockYieldProvider, signer);
  const mockYieldProviderAddress = await mockYieldProvider.getAddress();
  await testYieldManager.connect(signer).setYieldProviderUserFunds(mockYieldProviderAddress, withdrawAmount);
  await testYieldManager.connect(signer).setUserFundsInYieldProvidersTotal(withdrawAmount);
};
