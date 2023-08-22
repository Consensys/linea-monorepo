import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber, Contract } from "ethers";
import { ethers } from "hardhat";

const Permit = [
  { name: "owner", type: "address" },
  { name: "spender", type: "address" },
  { name: "value", type: "uint256" },
  { name: "nonce", type: "uint256" },
  { name: "deadline", type: "uint256" },
];

function buildData(
  owner: string,
  name: string,
  version: string,
  chainId: number,
  verifyingContract: string,
  spender: string,
  value: number,
  nonce: BigNumber,
  deadline: BigNumber,
) {
  return {
    domain: { name, version, chainId, verifyingContract },
    types: { Permit },
    value: { owner, spender, value, nonce, deadline },
  };
}

export async function getPermitData(
  wallet: SignerWithAddress,
  token: Contract,
  nonce: BigNumber,
  chainId: number,
  spender: string,
  value: number,
  deadline: BigNumber,
) {
  const name = await token.name();
  const data = buildData(wallet.address, name, "1", chainId, token.address, spender, value, nonce, deadline);
  const signature = await wallet._signTypedData(data.domain, data.types, data.value);
  const { v, r, s } = ethers.utils.splitSignature(signature);
  const permitCall = token.interface.encodeFunctionData("permit", [wallet.address, spender, value, deadline, v, r, s]);
  return permitCall;
}
