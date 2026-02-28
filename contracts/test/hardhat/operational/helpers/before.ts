import hre from "hardhat";
const { ethers } = await hre.network.connect();

export async function getRollupRevenueVaultAccountsFixture() {
  const [
    admin,
    invoiceSubmitter,
    burner,
    invoicePaymentReceiver,
    l1l2MessageSetter,
    l1LineaTokenBurner,
    nonAuthorizedAccount,
  ] = await ethers.getSigners();
  return {
    admin,
    invoiceSubmitter,
    burner,
    invoicePaymentReceiver,
    l1l2MessageSetter,
    l1LineaTokenBurner,
    nonAuthorizedAccount,
  };
}
