import { BigNumber, ethers } from "ethers";

export async function encodeSendMessage(
  sender: string,
  receiver: string,
  fee: BigNumber,
  amount: BigNumber,
  messageNonce: BigNumber,
  calldata: string,
) {
  const abiCoder = ethers.utils.defaultAbiCoder;
  const data = abiCoder.encode(
    ["address", "address", "uint256", "uint256", "uint256", "bytes"],
    [sender, receiver, fee, amount, messageNonce, calldata],
  );

  return ethers.utils.keccak256(data);
}
