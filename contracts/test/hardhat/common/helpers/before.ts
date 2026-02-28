import { ethers } from "../../common/hardhat-ethers.js";

export async function getAccountsFixture() {
  const [
    admin,
    securityCouncil,
    operator,
    nonAuthorizedAccount,
    nativeYieldOperator,
    l2YieldRecipient,
    operationalSafe,
  ] = await ethers.getSigners();
  return {
    admin,
    securityCouncil,
    operator,
    nonAuthorizedAccount,
    nativeYieldOperator,
    l2YieldRecipient,
    operationalSafe,
  };
}
