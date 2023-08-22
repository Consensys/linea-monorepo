import { beforeAll, jest } from "@jest/globals";
import { Wallet } from "ethers";
import {
  DummyContract,
  DummyContract__factory,
  L2MessageService,
  L2MessageService__factory,
  ZkEvmV2,
  ZkEvmV2__factory,
} from "../src/typechain";
import {
  getL1Provider,
  getL2Provider,
  CHAIN_ID,
  DEPLOYER_ACCOUNT_PRIVATE_KEY,
  ZKEVMV2_CONTRACT_ADDRESS,
  MESSAGE_SERVICE_ADDRESS,
  ACCOUNT_0,
  ACCOUNT_0_PRIVATE_KEY,
  TRANSACTION_CALLDATA_LIMIT,
  OPERATOR_0_PRIVATE_KEY
} from "../src/utils/constants.local";
import { deployContract } from "../src/utils/deployments";

jest.setTimeout(3 * 60 * 1000);

beforeAll(async () => {
  /*********** PROVIDERS SETUP ***********/
  const l1Provider = getL1Provider();
  const l2Provider = getL2Provider();

  const l1Deployer = new Wallet(DEPLOYER_ACCOUNT_PRIVATE_KEY, l1Provider);
  const l2Deployer = new Wallet(DEPLOYER_ACCOUNT_PRIVATE_KEY, l2Provider);
  const dummyContract = (await deployContract(new DummyContract__factory(), l2Deployer)) as DummyContract;
  const l1DummyContract = (await deployContract(new DummyContract__factory(), l1Deployer)) as DummyContract;

  const l2MessageService: L2MessageService = L2MessageService__factory.connect(MESSAGE_SERVICE_ADDRESS, l2Provider);
  // ZkEvmV2 deployment
  const zkEvmV2: ZkEvmV2 = ZkEvmV2__factory.connect(ZKEVMV2_CONTRACT_ADDRESS, l1Provider);

  global.l1Provider = l1Provider;
  global.l2Provider = l2Provider;
  global.dummyContract = dummyContract;
  global.l1DummyContract = l1DummyContract;
  global.l2MessageService = l2MessageService;
  global.zkEvmV2 = zkEvmV2;
  global.useLocalSetup = true;
  global.chainId = CHAIN_ID;
  global.ACCOUNT_0_PRIVATE_KEY = ACCOUNT_0_PRIVATE_KEY;
  global.DEPLOYER_ACCOUNT_PRIVATE_KEY = DEPLOYER_ACCOUNT_PRIVATE_KEY;
  global.TRANSACTION_CALLDATA_LIMIT = TRANSACTION_CALLDATA_LIMIT;
  global.OPERATOR_0_PRIVATE_KEY = OPERATOR_0_PRIVATE_KEY;
});
