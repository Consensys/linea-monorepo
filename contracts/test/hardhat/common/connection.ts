import hre from "hardhat";

const connection = await hre.network.connect("hardhat");

export const ethers = connection.ethers;
export const networkHelpers = connection.networkHelpers;
