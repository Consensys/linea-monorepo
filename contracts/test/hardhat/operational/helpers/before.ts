import { ethers } from "hardhat";

export async function getRollupFeeVaultAccountsFixture() {
  const [
    admin,
    invoiceSetter,
    burner,
    operatingCostsReceiver,
    l1l2MessageSetter,
    l1BurnerContract,
    nonAuthorizedAccount,
  ] = await ethers.getSigners();
  return {
    admin,
    invoiceSetter,
    burner,
    operatingCostsReceiver,
    l1l2MessageSetter,
    l1BurnerContract,
    nonAuthorizedAccount,
  };
}
