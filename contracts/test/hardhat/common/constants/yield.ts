import { ethers } from "hardhat";

export const MINIMUM_WITHDRAWAL_RESERVE_PERCENTAGE_BPS = 2000;
export const TARGET_WITHDRAWAL_RESERVE_PERCENTAGE_BPS = 2500;
export const MINIMUM_WITHDRAWAL_RESERVE_AMOUNT = ethers.parseEther("1000");
export const TARGET_WITHDRAWAL_RESERVE_AMOUNT = ethers.parseEther("1250");
