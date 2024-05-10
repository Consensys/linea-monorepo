import { ethers } from "ethers";
import fs from "fs";
import path from "path";
import { requireEnv } from "../hardhat/utils";

const proxyContract = "0x41B186Dc7C46f08ADFdCe21Da1b07f605819E9Ab";
// eslint-disable-next-line @typescript-eslint/no-unused-vars
const timelockAddress = "0x0169659ab31d3857C11b36E8418a97D422D5C363";

const newVerifierAddress = "0x0169659ab31d3857C11b36E8418a97D422D5C363";
const proofType = 0;

const RPC_URL = requireEnv("RPC_URL");
const safeAddress = requireEnv("SAFE_ADDRESS");
const ownerPrivateKey = requireEnv("OWNER_PRIVATE_KEY");

async function main() {
  // Connect to the network
  const provider = new ethers.JsonRpcProvider(RPC_URL);
  const wallet = new ethers.Wallet(ownerPrivateKey, provider);

  // Load Gnosis Safe contract ABI
  const safeAbiPath = path.join(__dirname, "./SafeABI.json");
  const safeAbi = JSON.parse(fs.readFileSync(safeAbiPath, "utf8"));

  // Create a contract instance
  const safeContract = new ethers.Contract(safeAddress, safeAbi, wallet);

  // Encode Tx
  const setVerifierRole = ethers.concat([
    "0xc2116974",
    ethers.AbiCoder.defaultAbiCoder().encode(["address", "uint256"], [newVerifierAddress, proofType]),
  ]);
  const setVerifierScheduleTransaction = ethers.concat([
    "0x01d5062a",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["address", "uint256", "bytes", "bytes32", "bytes32", "uint256"],
      [
        proxyContract,
        0, //value
        setVerifierRole,
        ethers.ZeroHash,
        ethers.ZeroHash,
        0, //timelock delay
      ],
    ),
  ]);

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const setVerifierExecuteTransaction = ethers.concat([
    "0x134008d3",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["address", "uint256", "bytes", "bytes32", "bytes32"],
      [
        proxyContract,
        0, //value
        setVerifierRole,
        ethers.ZeroHash,
        ethers.ZeroHash,
      ],
    ),
  ]);

  const to = timelockAddress; // Timelock
  const value = 0;
  const data = setVerifierScheduleTransaction; // Encoded Schedule TX data
  // const data = setVerifierExecuteTransaction; // Encoded Execute TX data

  // get signatures
  const encodeSignature = ethers.concat([
    `0x000000000000000000000000${wallet.address.slice(2)}000000000000000000000000000000000000000000000000000000000000000001`,
  ]);

  // Estimate Gas
  const gasEstimate = await safeContract.execTransaction.estimateGas(
    to,
    value,
    data,
    0, // operation
    0, // safeTxGas
    0, // baseGas
    0, // gasPrice
    ethers.ZeroAddress, // gasToken
    ethers.ZeroAddress, // refundReceiver
    encodeSignature, // signatures
  );

  // Create the transaction
  const tx = await safeContract.execTransaction(
    to,
    value,
    data,
    0, // operation
    0, // safeTxGas , previously gasEstimate
    0, // baseGas
    0, // gasPrice
    ethers.ZeroAddress, // gasToken
    ethers.ZeroAddress, // refundReceiver
    encodeSignature, // signatures
    { gasLimit: gasEstimate + 50_000n }, // add some extra gas
  );
  await tx.wait();

  console.log(`Transaction hash: ${tx.hash}`);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
