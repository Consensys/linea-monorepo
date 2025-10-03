import { ethers } from "hardhat";

export async function getAccountsFixture() {
  const [admin, securityCouncil, operator, nonAuthorizedAccount, nativeYieldOperator, l2YieldRecipient] =
    await ethers.getSigners();
  return { admin, securityCouncil, operator, nonAuthorizedAccount, nativeYieldOperator, l2YieldRecipient };
}
