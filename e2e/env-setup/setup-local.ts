import { beforeAll, jest } from "@jest/globals";
import { Wallet, ethers } from "ethers";
import {
  DummyContract,
  DummyContract__factory,
  L2MessageService,
  L2MessageService__factory,
  LineaRollup,
  LineaRollup__factory,
} from "../src/typechain";
import {
  getL1Provider,
  getL2Provider,
  CHAIN_ID,
  L1_DEPLOYER_ACCOUNT_PRIVATE_KEY,
  L2_DEPLOYER_ACCOUNT_PRIVATE_KEY,
  LINEA_ROLLUP_CONTRACT_ADDRESS,
  MESSAGE_SERVICE_ADDRESS,
  L1_ACCOUNT_0_PRIVATE_KEY,
  L2_ACCOUNT_0_PRIVATE_KEY,
  TRANSACTION_CALLDATA_LIMIT,
  OPERATOR_0_PRIVATE_KEY,
  SHOMEI_ENDPOINT,
  SHOMEI_FRONTEND_ENDPOINT,
  OPERATOR_1,
  SECURITY_COUNCIL_PRIVATE_KEY,
  CONTRACT_GAS_OPTIMIZATION_SWITCH_BLOCK,
} from "../src/utils/constants.local";
import { deployContract } from "../src/utils/deployments";
import { getAndIncreaseFeeData } from "../src/utils/helpers";

jest.setTimeout(3 * 60 * 1000);

beforeAll(async () => {
  /*********** PROVIDERS SETUP ***********/
  const l1JsonRpcProvider = getL1Provider();
  const l2JsonRpcProvider = getL2Provider();

  const l1Deployer = new Wallet(L1_DEPLOYER_ACCOUNT_PRIVATE_KEY, l1JsonRpcProvider);
  const l2Deployer = new Wallet(L2_DEPLOYER_ACCOUNT_PRIVATE_KEY, l2JsonRpcProvider);

  const [dummyContract, l1DummyContract] = await Promise.all([
    deployContract(new DummyContract__factory(), l2Deployer) as unknown as DummyContract,
    deployContract(new DummyContract__factory(), l1Deployer) as unknown as DummyContract,
  ]);

  // Old LineaRollup contract instanciation
  // LineaRollup ABI contains both old and new functions
  const lineaRollup: LineaRollup = LineaRollup__factory.connect(LINEA_ROLLUP_CONTRACT_ADDRESS, l1JsonRpcProvider);
  // L2 MessageService contract instanciation
  const l2MessageService: L2MessageService = L2MessageService__factory.connect(
    MESSAGE_SERVICE_ADDRESS,
    l2JsonRpcProvider,
  );

  // Send ETH to the LineaRollup contract
  const value = ethers.utils.parseEther("500");
  const fee = ethers.utils.parseEther("3");
  const to = "0x8D97689C9818892B700e27F316cc3E41e17fBeb9";
  const calldata = "0x";
  const [maxPriorityFeePerGas, maxFeePerGas] = getAndIncreaseFeeData(await l1JsonRpcProvider.getFeeData());
  const tx = await lineaRollup
    .connect(l1Deployer)
    .sendMessage(to, fee, calldata, { value, maxPriorityFeePerGas, maxFeePerGas });
  await tx.wait();

  global.l1Provider = l1JsonRpcProvider;
  global.l2Provider = l2JsonRpcProvider;
  global.dummyContract = dummyContract;
  global.l1DummyContract = l1DummyContract;
  global.l2MessageService = l2MessageService;
  global.L2_MESSAGE_SERVICE_ADDRESS = MESSAGE_SERVICE_ADDRESS;
  global.LINEA_ROLLUP_CONTRACT_ADDRESS = LINEA_ROLLUP_CONTRACT_ADDRESS;
  global.lineaRollup = lineaRollup;
  global.useLocalSetup = true;
  global.chainId = CHAIN_ID;
  global.L1_ACCOUNT_0_PRIVATE_KEY = L1_ACCOUNT_0_PRIVATE_KEY;
  global.L2_ACCOUNT_0_PRIVATE_KEY = L2_ACCOUNT_0_PRIVATE_KEY;
  global.L1_DEPLOYER_ACCOUNT_PRIVATE_KEY = L1_DEPLOYER_ACCOUNT_PRIVATE_KEY;
  global.L2_DEPLOYER_ACCOUNT_PRIVATE_KEY = L2_DEPLOYER_ACCOUNT_PRIVATE_KEY;
  global.TRANSACTION_CALLDATA_LIMIT = TRANSACTION_CALLDATA_LIMIT;
  global.OPERATOR_0_PRIVATE_KEY = OPERATOR_0_PRIVATE_KEY;
  global.SHOMEI_ENDPOINT = SHOMEI_ENDPOINT;
  global.SHOMEI_FRONTEND_ENDPOINT = SHOMEI_FRONTEND_ENDPOINT;
  global.OPERATOR_1_ADDRESS = OPERATOR_1;
  global.SECURITY_COUNCIL_PRIVATE_KEY = SECURITY_COUNCIL_PRIVATE_KEY;
  global.CONTRACT_GAS_OPTIMIZATION_SWITCH_BLOCK = CONTRACT_GAS_OPTIMIZATION_SWITCH_BLOCK;
});
