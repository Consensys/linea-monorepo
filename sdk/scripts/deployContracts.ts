import { BaseContract, ContractFactory, JsonRpcProvider, Wallet, ethers } from "ethers";
import { config } from "dotenv";
import fs from "fs";
import yargs from "yargs";
import { hideBin } from "yargs/helpers";
import {
  ZkEvmV2__factory,
  TransparentUpgradeableProxy__factory,
  ProxyAdmin__factory,
  PlonkVerifier__factory,
  L2MessageService__factory,
} from "../src/clients/blockchain/typechain/factories";
import { ProxyAdmin } from "../src/clients/blockchain/typechain/ProxyAdmin";
import { sanitizePrivKey } from "./cli";

config();

const argv = yargs(hideBin(process.argv))
  .option("l1-rpc-url", {
    describe: "L1 rpc url",
    type: "string",
    demandOption: true,
  })
  .option("l2-rpc-url", {
    describe: "L2 rpc url",
    type: "string",
    demandOption: true,
  })
  .option("l1-deployer-priv-key", {
    describe: "L1 deployer private key",
    type: "string",
    demandOption: true,
    coerce: sanitizePrivKey("priv-key"),
  })
  .option("l2-deployer-priv-key", {
    describe: "L2 deployer private key",
    type: "string",
    demandOption: true,
    coerce: sanitizePrivKey("priv-key"),
  })
  .parseSync();

const getInitializerData = (contractInterface: ethers.Interface, args: unknown[]) => {
  const initializer = "initialize";
  const fragment = contractInterface.getFunction(initializer);
  return contractInterface.encodeFunctionData(fragment!, args);
};

const deployUpgradableContract = async <T extends ContractFactory>(
  contractFactory: T,
  deployer: Wallet,
  admin: ProxyAdmin,
  initializerData = "0x",
): Promise<BaseContract> => {
  const instance = await contractFactory.connect(deployer).deploy();
  await instance.waitForDeployment();

  const proxy = await new TransparentUpgradeableProxy__factory()
    .connect(deployer)
    .deploy(await instance.getAddress(), await admin.getAddress(), initializerData);
  await proxy.waitForDeployment();

  return instance.attach(await proxy.getAddress());
};

const deployLineaRollup = async (deployer: Wallet): Promise<{ zkevmV2ContractAddress: string }> => {
  const proxyFactory = new ProxyAdmin__factory(deployer);
  const proxyAdmin = await proxyFactory.connect(deployer).deploy();
  await proxyAdmin.waitForDeployment();
  console.log(`ProxyAdmin contract deployed at address: ${await proxyAdmin.getAddress()}`);

  const plonkVerifierFactory = new PlonkVerifier__factory(deployer);
  const plonkVerifier = await plonkVerifierFactory.deploy();
  await plonkVerifier.waitForDeployment();
  console.log(`PlonkVerifier contract deployed at address: ${await plonkVerifier.getAddress()}`);

  const zkevmV2Contract = await deployUpgradableContract(
    new ZkEvmV2__factory(deployer),
    deployer,
    proxyAdmin,
    getInitializerData(ZkEvmV2__factory.createInterface(), [
      ethers.ZeroHash,
      0,
      await plonkVerifier.getAddress(),
      deployer.address,
      [deployer.address],
      86400,
      ethers.parseEther("5"),
    ]),
  );
  const zkEvmV2Address = await zkevmV2Contract.getAddress();
  console.log(`ZkEvmV2 contract deployed at address: ${zkEvmV2Address}`);
  return { zkevmV2ContractAddress: zkEvmV2Address };
};

const deployL2MessageService = async (deployer: Wallet): Promise<string> => {
  const proxyFactory = new ProxyAdmin__factory(deployer);
  const proxyAdmin = await proxyFactory.connect(deployer).deploy();
  await proxyAdmin.waitForDeployment();
  console.log(`L2 ProxyAdmin contract deployed at address: ${await proxyAdmin.getAddress()}`);

  const l2MessageService = await deployUpgradableContract(
    new L2MessageService__factory(deployer),
    deployer,
    proxyAdmin,
    getInitializerData(L2MessageService__factory.createInterface(), [
      deployer.address,
      deployer.address,
      86400,
      ethers.parseEther("5"),
    ]),
  );
  const l2MessageServiceAddress = await l2MessageService.getAddress();
  console.log(`L2MessageService contract deployed at address: ${l2MessageServiceAddress}`);
  return l2MessageServiceAddress;
};

const main = async (args: typeof argv) => {
  const l1Provider = new JsonRpcProvider(args.l1RpcUrl);
  const l2Provider = new JsonRpcProvider(args.l2RpcUrl);
  // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
  const l1Deployer = new Wallet(args.l1DeployerPrivKey, l1Provider);
  // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
  const l2Deployer = new Wallet(args.l2DeployerPrivKey, l2Provider);

  const { zkevmV2ContractAddress } = await deployLineaRollup(l1Deployer);
  const l2MessageServiceAddress = await deployL2MessageService(l2Deployer);

  const tx = await l2Deployer.sendTransaction({
    to: l2MessageServiceAddress,
    value: ethers.parseEther("1000"),
    data: "0x",
  });

  await tx.wait();

  const tx2 = await l1Deployer.sendTransaction({
    to: zkevmV2ContractAddress,
    value: ethers.parseEther("1000"),
    data: "0x",
  });

  await tx2.wait();

  fs.writeFileSync(
    "./scripts/contractAddresses.json",
    JSON.stringify({ zkevmV2ContractAddress, l2MessageServiceAddress }, null, 2),
  );
};

main(argv)
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
