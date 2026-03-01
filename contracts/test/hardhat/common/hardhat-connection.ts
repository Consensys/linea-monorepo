import hre from "hardhat";

export const connection = await hre.network.connect();
export const ethers = connection.ethers;
export const networkHelpers = connection.networkHelpers;
export const time = networkHelpers.time;
export const loadFixture = networkHelpers.loadFixture;
export const mine = networkHelpers.mine;
export const reset = networkHelpers.reset;
export const setBalance = networkHelpers.setBalance;
export const setCode = networkHelpers.setCode;
export const setNonce = networkHelpers.setNonce;
export const setStorageAt = networkHelpers.setStorageAt;
export const stopImpersonatingAccount = networkHelpers.stopImpersonatingAccount;
export const impersonateAccount = networkHelpers.impersonateAccount;
export const takeSnapshot = networkHelpers.takeSnapshot;
export { hre };
