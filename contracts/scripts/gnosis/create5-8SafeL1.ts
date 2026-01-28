import { EthersAdapter, SafeFactory } from "@safe-global/protocol-kit";
import { ethers } from "ethers";
import { requireEnv } from "../hardhat/utils";

const main = async () => {
  const RPC_URL = requireEnv("BLOCKCHAIN_NODE");
  const SAFE_OWNER1_PRIVATE_KEY = requireEnv("SAFE_OWNER1_PRIVATE_KEY");
  const SAFE_OWNERS = requireEnv("SAFE_OWNERS");
  const provider = new ethers.JsonRpcProvider(RPC_URL);
  const signer = new ethers.Wallet(SAFE_OWNER1_PRIVATE_KEY, provider);

  const ethAdapter = new EthersAdapter({ ethers, signerOrProvider: signer });

  const chainId = await ethAdapter.getChainId();
  console.log(`ChainId: ${chainId}`);

  // const safeVersion = '1.3.0'
  // const isL1SafeMasterCopy = false
  const safeFactory = await SafeFactory.create({ ethAdapter: ethAdapter });

  const safeAccountConfig = {
    threshold: 5, // Setting the Threshold to 5
    owners: SAFE_OWNERS?.split(","),
  };
  console.log("Deploying 5/8 safe..");

  const safeSdkOwner1 = await safeFactory.deploySafe({ safeAccountConfig });
  const safeAddress = safeSdkOwner1.getAddress();

  console.log(`5/8 Safe deployed at: ${safeAddress}`);
};

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
