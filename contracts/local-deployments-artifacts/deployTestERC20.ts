import { ethers } from "ethers";
import { abi as TestERC20Abi, bytecode as TestERC20Bytecode } from "./static-artifacts/TestERC20.json";
import { deployContractFromArtifacts } from "../common/helpers/deployments";
import { get1559Fees } from "../scripts/utils";
import { getRequiredEnvVar } from "../common/helpers/environment";

async function main() {
  const provider = new ethers.JsonRpcProvider(process.env.RPC_URL);
  const wallet = new ethers.Wallet(process.env.PRIVATE_KEY!, provider);

  const erc20Name = getRequiredEnvVar("TEST_ERC20_NAME");
  const erc20Symbol = getRequiredEnvVar("TEST_ERC20_SYMBOL");
  const erc20Supply = getRequiredEnvVar("TEST_ERC20_INITIAL_SUPPLY");

  const [walletNonce, { gasPrice }] = await Promise.all([wallet.getNonce(), get1559Fees(provider)]);

  const testERC20 = await deployContractFromArtifacts(
    TestERC20Abi,
    TestERC20Bytecode,
    wallet,
    erc20Name,
    erc20Symbol,
    erc20Supply,
    {
      nonce: walletNonce,
      gasPrice,
    },
  );

  const testERC20Address = await testERC20.getAddress();

  const chainId = (await provider.getNetwork()).chainId;

  console.log(`testERC20 deployed: address=${testERC20Address} chainId=${chainId}`);
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
