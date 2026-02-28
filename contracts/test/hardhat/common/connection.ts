import hre from "hardhat";

const connection = await hre.network.connect();

export const ethers = connection.ethers;
export const networkHelpers = connection.networkHelpers;
