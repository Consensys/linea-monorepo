import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { expect } from "chai";
import {
  MockDashboard,
  MockLineaRollup,
  MockYieldProvider,
  SSZMerkleTree,
  TestCLProofVerifier,
  TestLidoStVaultYieldProvider,
  TestLineaRollup,
  TestYieldManager,
  YieldManager,
} from "contracts/typechain-types";
import { ethers } from "hardhat";
import {
  ADDRESS_ZERO,
  EMPTY_CALLDATA,
  MAX_0X2_VALIDATOR_EFFECTIVE_BALANCE_GWEI,
  ONE_ETHER,
  ONE_GWEI,
  VALIDATOR_WITNESS_TYPE,
  ZERO_VALUE,
} from "../../common/constants";
import { ClaimMessageWithProofParams } from "./types";
import { generateLidoUnstakePermissionlessWitness, randomBytes32 } from "./proof";
import { encodeSendMessage } from "../../common/helpers";
import { BaseContract } from "ethers";

// TODO - Existence of this setup function means that YieldManager has invariants that withdraw cannot underflow for userFunds and userFundsInYieldProvidersTotal
// Caution - assumes it will only be used once, will not work for consecutive uses in its current form
export const fundYieldProviderForWithdrawal = async (
  testYieldManager: TestYieldManager,
  mockYieldProvider: MockYieldProvider,
  signer: SignerWithAddress,
  withdrawAmount: bigint,
) => {
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

export const decrementBalance = async (address: string, decrement: bigint) => {
  const curBalance = await ethers.provider.getBalance(address);
  await ethers.provider.send("hardhat_setBalance", [address, ethers.toBeHex(curBalance - decrement)]);
};

export const setBalance = async (address: string, balance: bigint) => {
  await ethers.provider.send("hardhat_setBalance", [address, ethers.toBeHex(balance)]);
};

export const getBalance = async (contract: BaseContract) => {
  return await ethers.provider.getBalance(await contract.getAddress());
};

export const setWithdrawalReserveBalance = async (testYieldManager: TestYieldManager, balance: bigint) => {
  const l1MessageServiceAddress = await testYieldManager.getL1MessageService();
  await setBalance(l1MessageServiceAddress, balance);
};

export const setWithdrawalReserveToMinimum = async (testYieldManager: TestYieldManager) => {
  const minimumReserve = await testYieldManager.getEffectiveMinimumWithdrawalReserve();
  await setWithdrawalReserveBalance(testYieldManager, minimumReserve);
  return minimumReserve;
};

export const setWithdrawalReserveToTarget = async (testYieldManager: TestYieldManager) => {
  const targetReserve = await testYieldManager.getEffectiveTargetWithdrawalReserve();
  await setWithdrawalReserveBalance(testYieldManager, targetReserve);
  return targetReserve;
};

export const ossifyYieldProvider = async (
  yieldManager: TestYieldManager,
  yieldProviderAddress: string,
  securityCouncil: SignerWithAddress,
) => {
  await yieldManager.connect(securityCouncil).initiateOssification(yieldProviderAddress);
  await yieldManager.connect(securityCouncil).progressPendingOssification(yieldProviderAddress);
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

export const getWithdrawLSTCall = async (
  mockLineaRollup: MockLineaRollup,
  yieldManager: TestYieldManager,
  yieldProvider: TestLidoStVaultYieldProvider,
  signer: SignerWithAddress,
  withdrawAmount: bigint,
) => {
  const recipient = ethers.Wallet.createRandom().address;
  await fundLidoStVaultYieldProvider(yieldManager, yieldProvider, signer, withdrawAmount);

  // Add gas fees
  const l1MessageService = await yieldManager.L1_MESSAGE_SERVICE();
  await ethers.provider.send("hardhat_setBalance", [l1MessageService, ethers.toBeHex(ONE_ETHER)]);
  const l1Signer = await ethers.getImpersonatedSigner(l1MessageService);
  await mockLineaRollup.setWithdrawLSTAllowed(true);

  return yieldManager.connect(l1Signer).withdrawLST(await yieldProvider.getAddress(), withdrawAmount, recipient);
};

export const setupLineaRollupMessageMerkleTree = async (
  lineaRollup: TestLineaRollup,
  from: string,
  to: string,
  value: bigint,
  data: string,
): Promise<ClaimMessageWithProofParams> => {
  const messageNumber = await lineaRollup.nextMessageNumber();
  const expectedBytes = await encodeSendMessage(from, to, 0n, value, messageNumber, data);

  const messageHash = ethers.keccak256(expectedBytes);
  const proof = Array.from({ length: 32 }, () => randomBytes32());
  const leafIndex = 0n;
  const root = await lineaRollup.generateMerkleRoot(messageHash, proof, leafIndex);
  await lineaRollup.addL2MerkleRoots([root], proof.length);

  const claimParams: ClaimMessageWithProofParams = {
    proof,
    messageNumber: messageNumber,
    leafIndex,
    from: from,
    to: to,
    fee: 0n,
    value: value,
    feeRecipient: ADDRESS_ZERO,
    merkleRoot: root,
    data: data,
  };

  // Send empty message to increment the messageNumber
  await lineaRollup.sendMessage(ethers.Wallet.createRandom().address, ZERO_VALUE, EMPTY_CALLDATA);

  return claimParams;
};

export const incurPositiveYield = async (
  yieldManager: TestYieldManager,
  mockDashboard: MockDashboard,
  nativeYieldOperator: SignerWithAddress,
  mockStakingVaultAddress: string,
  yieldProviderAddress: string,
  l2YieldRecipient: SignerWithAddress,
  positiveYield: bigint,
  lstPrincipalPaid: bigint = 0n,
) => {
  await incrementBalance(mockStakingVaultAddress, positiveYield);
  const userFunds = await yieldManager.userFunds(yieldProviderAddress);
  const mockDashboardTotalValuePrev = await mockDashboard.totalValue();
  const prevNegativeYield = mockDashboardTotalValuePrev < userFunds ? userFunds - mockDashboardTotalValuePrev : 0n;
  await incrementMockDashboardTotalValue(mockDashboard, positiveYield);
  const yieldProviderYieldReportedCumulativePrev =
    await yieldManager.getYieldProviderYieldReportedCumulative(yieldProviderAddress);

  // Act
  const [newReportedYield, outstandingNegativeYield] = await yieldManager
    .connect(nativeYieldOperator)
    .reportYield.staticCall(yieldProviderAddress, l2YieldRecipient);
  expect(newReportedYield).eq(positiveYield - lstPrincipalPaid - prevNegativeYield);
  expect(outstandingNegativeYield).eq(0);
  await yieldManager.connect(nativeYieldOperator).reportYield(yieldProviderAddress, l2YieldRecipient);
  if (lstPrincipalPaid > 0n) {
    await decrementMockDashboardTotalValue(mockDashboard, lstPrincipalPaid);
  }
  // Obligations paid
  expect(await yieldManager.userFunds(yieldProviderAddress)).eq(userFunds + newReportedYield - lstPrincipalPaid);
  expect(await yieldManager.getYieldProviderYieldReportedCumulative(yieldProviderAddress)).eq(
    yieldProviderYieldReportedCumulativePrev + positiveYield - lstPrincipalPaid - prevNegativeYield,
  );
  expect(await yieldManager.userFundsInYieldProvidersTotal()).eq(await yieldManager.userFunds(yieldProviderAddress));
};

export const incurNegativeYield = async (
  yieldManager: TestYieldManager,
  mockDashboard: MockDashboard,
  nativeYieldOperator: SignerWithAddress,
  mockStakingVaultAddress: string,
  yieldProviderAddress: string,
  l2YieldRecipient: SignerWithAddress,
  negativeYield: bigint,
) => {
  await decrementBalance(mockStakingVaultAddress, negativeYield);
  await mockDashboard.setTotalValueReturn((await mockDashboard.totalValue()) - negativeYield);
  const userFunds = await yieldManager.userFunds(yieldProviderAddress);
  const yieldProviderYieldReportedCumulativePrev =
    await yieldManager.getYieldProviderYieldReportedCumulative(yieldProviderAddress);
  {
    const [newReportedYield, outstandingNegativeYield] = await yieldManager
      .connect(nativeYieldOperator)
      .reportYield.staticCall(yieldProviderAddress, l2YieldRecipient);
    expect(newReportedYield).eq(0);
    expect(outstandingNegativeYield).eq(negativeYield);
  }
  await yieldManager.connect(nativeYieldOperator).reportYield(yieldProviderAddress, l2YieldRecipient);
  expect(await yieldManager.userFunds(yieldProviderAddress)).eq(userFunds);
  expect(await yieldManager.userFundsInYieldProvidersTotal()).eq(await yieldManager.userFunds(yieldProviderAddress));
  expect(await yieldManager.getYieldProviderYieldReportedCumulative(yieldProviderAddress)).eq(
    yieldProviderYieldReportedCumulativePrev,
  );
};

export const withdrawLST = async (
  lineaRollup: TestLineaRollup,
  nonAuthorizedAccount: SignerWithAddress,
  yieldProviderAddress: string,
  amount: bigint,
) => {
  const recipientAddress = await nonAuthorizedAccount.getAddress();
  const claimParams = await setupLineaRollupMessageMerkleTree(
    lineaRollup,
    recipientAddress,
    recipientAddress,
    amount,
    EMPTY_CALLDATA,
  );
  await lineaRollup
    .connect(nonAuthorizedAccount)
    .claimMessageWithProofAndWithdrawLST(claimParams, yieldProviderAddress);
};

export const executeUnstakePermissionless = async (
  sszMerkleTree: SSZMerkleTree,
  verifier: TestCLProofVerifier,
  yieldManager: YieldManager,
  yieldProviderAddress: string,
  mockStakingVaultAddress: string,
  refundAddress: string,
  unstakeAmount: bigint,
) => {
  const { validatorWitness } = await generateLidoUnstakePermissionlessWitness(
    sszMerkleTree,
    verifier,
    mockStakingVaultAddress,
    MAX_0X2_VALIDATOR_EFFECTIVE_BALANCE_GWEI,
  );
  const withdrawalParams = ethers.AbiCoder.defaultAbiCoder().encode(
    ["bytes", "uint64[]", "address"],
    [validatorWitness.pubkey, [unstakeAmount / ONE_GWEI], refundAddress],
  );
  const withdrawalParamsProof = ethers.AbiCoder.defaultAbiCoder().encode([VALIDATOR_WITNESS_TYPE], [validatorWitness]);
  // Arrange - first unstake
  await yieldManager.unstakePermissionless(yieldProviderAddress, withdrawalParams, withdrawalParamsProof);
};

export const incrementMockDashboardTotalValue = async (mockDashboard: MockDashboard, amount: bigint) => {
  await mockDashboard.setTotalValueReturn((await mockDashboard.totalValue()) + amount);
};

export const decrementMockDashboardTotalValue = async (mockDashboard: MockDashboard, amount: bigint) => {
  await mockDashboard.setTotalValueReturn((await mockDashboard.totalValue()) - amount);
};
