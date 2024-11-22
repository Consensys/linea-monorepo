import { ethers } from "ethers";

function computeCreateAddress(deployerAddress: string, nonce: number): string {
  // Encode deployerAddress and nonce in RLP format
  const rlpEncoded = ethers.encodeRlp([deployerAddress, ethers.toBeHex(nonce)]);

  // Compute the Keccak256 hash of the encoded value
  const keccakHash = ethers.keccak256(rlpEncoded);

  // Extract the last 20 bytes (address) from the hash
  const computedAddress = `0x${keccakHash.slice(-40)}`;
  return ethers.getAddress(computedAddress); // Convert to checksum address
}

// Example usage:
const deployer = "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"; // Replace with your deployer address
const nonce = 3; // Replace with the actual nonce
const computedAddress = computeCreateAddress(deployer, nonce);

console.log("Computed Address:", computedAddress);
