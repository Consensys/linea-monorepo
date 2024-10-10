import { beforeAll, jest } from "@jest/globals";
import { Wallet } from "ethers";
import { TestContract, TestContract__factory } from "../src/typechain";
import {
  getL2Provider,
  getProvider,
  L2_DEPLOYER_ACCOUNT_PRIVATE_KEY,
  L2_ACCOUNT_0_PRIVATE_KEY,
  L2_ACCOUNT_1_PRIVATE_KEY,
  TRANSACTION_EXCLUSION_ENDPOINT,
  SEQUENCER_RPC_URL,
  L2_BESU_NODE_RPC_URL,
} from "../src/utils/constants.local";
import { deployContract } from "../src/utils/deployments";

jest.setTimeout(3 * 60 * 1000);

beforeAll(async () => {
  /*********** PROVIDERS SETUP ***********/
  const l2JsonRpcProvider = getL2Provider();
  const sequencerJsonRpcProvider = getProvider(SEQUENCER_RPC_URL);
  const l2BesuNodeJsonRpcProvider = getProvider(L2_BESU_NODE_RPC_URL);

  const l2Deployer = new Wallet(L2_DEPLOYER_ACCOUNT_PRIVATE_KEY, l2JsonRpcProvider);
  const testContract = (await deployContract(new TestContract__factory(), l2Deployer)) as unknown as TestContract;

  global.l2Provider = l2JsonRpcProvider;
  global.l2BesuNodeProvider = l2BesuNodeJsonRpcProvider;
  global.sequencerProvider = sequencerJsonRpcProvider;
  global.testContract = testContract;
  global.L2_ACCOUNT_0_PRIVATE_KEY = L2_ACCOUNT_0_PRIVATE_KEY;
  global.L2_ACCOUNT_1_PRIVATE_KEY = L2_ACCOUNT_1_PRIVATE_KEY;
  global.L2_DEPLOYER_ACCOUNT_PRIVATE_KEY = L2_DEPLOYER_ACCOUNT_PRIVATE_KEY;
  global.TRANSACTION_EXCLUSION_ENDPOINT = TRANSACTION_EXCLUSION_ENDPOINT;
});
