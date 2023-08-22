import { beforeAll, jest } from "@jest/globals";
import { Wallet } from "ethers";
import {
  DummyContract,
  DummyContract__factory,
  L2MessageService,
  L2MessageService__factory,
  ZkEvmV2,
  ZkEvmV2__factory,
  L2TestContract__factory,
  L2TestContract
} from "../src/typechain";
import {
  getL1Provider,
  getL2Provider,
  CHAIN_ID,
  SHADOW_ZKEVMV2_CONTRACT_ADDRESS,
  SHADOW_MESSAGE_SERVICE_ADDRESS,
  DUMMY_CONTRACT_ADDRESS,
  ACCOUNT_0_PRIVATE_KEY,
  TRANSACTION_CALLDATA_LIMIT,
  L1_DUMMY_CONTRACT_ADDRESS
} from "../src/utils/constants.uat";
import { deployContract, deployUpgradableContractWithProxyAdmin, encodeLibraryName } from "../src/utils/deployments";

jest.setTimeout(5 * 60 * 1000);

beforeAll(async () => {
  /*********** PROVIDERS SETUP ***********/
  const l1Provider = getL1Provider();
  const l2Provider = getL2Provider();

  const l1DummyContract: DummyContract = DummyContract__factory.connect(L1_DUMMY_CONTRACT_ADDRESS, l1Provider);
  const dummyContract: DummyContract = DummyContract__factory.connect(DUMMY_CONTRACT_ADDRESS, l2Provider);
  const l2TestContract: L2TestContract = L2TestContract__factory.connect("0xaD711a736ae454Be345078C0bc849997b78A30B9", l2Provider);
  // L2MessageService contract
  const l2MessageService: L2MessageService = L2MessageService__factory.connect(SHADOW_MESSAGE_SERVICE_ADDRESS, l2Provider);

  /*********** L1 Contracts ***********/

  // ZkEvmV2 deployment
  const zkEvmV2: ZkEvmV2 = ZkEvmV2__factory.connect(SHADOW_ZKEVMV2_CONTRACT_ADDRESS, l1Provider);

  global.l1Provider = l1Provider;
  global.l2Provider = l2Provider;
  global.l2TestContract = l2TestContract;
  global.l1DummyContract = l1DummyContract;
  global.dummyContract = dummyContract;
  global.l2MessageService = l2MessageService;
  global.zkEvmV2 = zkEvmV2;
  global.chainId = CHAIN_ID;
  global.ACCOUNT_0_PRIVATE_KEY = ACCOUNT_0_PRIVATE_KEY;
  global.TRANSACTION_CALLDATA_LIMIT = TRANSACTION_CALLDATA_LIMIT;
});
