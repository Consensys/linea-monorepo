import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { expect } from "chai";
import {
  MockLineaRollup,
  MockYieldProvider,
  TestLidoStVaultYieldProvider,
  TestLineaRollup,
  TestYieldManager,
} from "contracts/typechain-types";
import { ethers } from "hardhat";
import { ADDRESS_ZERO, ONE_ETHER } from "../../common/constants";
import { ClaimMessageWithProofParams } from "./types";
import { randomBytes32 } from "./proof";
import { encodeSendMessage } from "../../common/helpers";

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

export const setBalance = async (address: string, balance: bigint) => {
  await ethers.provider.send("hardhat_setBalance", [address, ethers.toBeHex(balance)]);
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

let setupLineaRollupMessageMerkleTreeMessageNumber = 1n;

export const setupLineaRollupMessageMerkleTree = async (
  lineaRollup: TestLineaRollup,
  from: string,
  to: string,
  value: bigint,
  data: string,
): Promise<ClaimMessageWithProofParams> => {
  const expectedBytes = await encodeSendMessage(
    from,
    to,
    0n,
    value,
    setupLineaRollupMessageMerkleTreeMessageNumber,
    data,
  );
  setupLineaRollupMessageMerkleTreeMessageNumber = setupLineaRollupMessageMerkleTreeMessageNumber + 1n;

  const messageHash = ethers.keccak256(expectedBytes);
  const proof = Array.from({ length: 32 }, () => randomBytes32());
  const leafIndex = 0n;
  const root = await lineaRollup.generateMerkleRoot(messageHash, proof, leafIndex);
  await lineaRollup.addL2MerkleRoots([root], proof.length);

  const claimParams: ClaimMessageWithProofParams = {
    proof,
    messageNumber: 1n,
    leafIndex,
    from: from,
    to: to,
    fee: 0n,
    value: value,
    feeRecipient: ADDRESS_ZERO,
    merkleRoot: root,
    data: data,
  };

  return claimParams;
};
