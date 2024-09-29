import { ethers } from "ethers";

export function calculateRewards(balance: bigint): bigint {
  const oneEth = ethers.parseEther("1");

  if (balance < oneEth) {
    return 0n;
  }

  const quotient = (balance - oneEth) / oneEth;
  const flooredBalance = quotient * oneEth;
  return flooredBalance;
}
