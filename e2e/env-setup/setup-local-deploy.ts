import { beforeAll, jest } from "@jest/globals";
import { Wallet } from "ethers";
import {
  DummyContract,
  DummyContract__factory,
  L2MessageService,
  L2MessageService__factory,
  LineaRollup,
  LineaRollup__factory,
  TestPlonkVerifierForDataAggregation__factory,
} from "../src/typechain";
import {
  L2_ACCOUNT_0,
  L2_ACCOUNT_0_PRIVATE_KEY,
  L2_ACCOUNT_1_PRIVATE_KEY,
  L1_DEPLOYER_ACCOUNT_PRIVATE_KEY,
  L2_DEPLOYER_ACCOUNT_PRIVATE_KEY,
  INITIAL_WITHDRAW_LIMIT,
  LINEA_ROLLUP_INITIAL_L2_BLOCK_NR,
  LINEA_ROLLUP_INITIAL_STATE_ROOT_HASH,
  LINEA_ROLLUP_OPERATORS,
  LINEA_ROLLUP_RATE_LIMIT_AMOUNT,
  LINEA_ROLLUP_RATE_LIMIT_PERIOD,
  LINEA_ROLLUP_SECURITY_COUNCIL,
  getL1Provider,
  getL2Provider,
  TRANSACTION_CALLDATA_LIMIT,
  OPERATOR_0_PRIVATE_KEY,
  SHOMEI_ENDPOINT,
  SHOMEI_FRONTEND_ENDPOINT,
} from "../src/utils/constants.local";
import { deployContract, deployUpgradableContractWithProxyAdmin } from "../src/utils/deployments";

jest.setTimeout(5 * 60 * 1000);

beforeAll(async () => {
  /*********** PROVIDERS SETUP ***********/
  const l1Provider = getL1Provider();
  const l2Provider = getL2Provider();

  const l2Deployer = new Wallet(L2_DEPLOYER_ACCOUNT_PRIVATE_KEY, l2Provider);
  const dummyContract = (await deployContract(new DummyContract__factory(), l2Deployer)) as DummyContract;

  // L2MessageService deployment
  const l2MessageService = (await deployUpgradableContractWithProxyAdmin(new L2MessageService__factory(), l2Deployer, [
    L2_ACCOUNT_0,
    86400,
    INITIAL_WITHDRAW_LIMIT,
  ])) as L2MessageService;

  /*********** L1 DEPLOYMENTS ***********/
  const l1Deployer = new Wallet(L1_DEPLOYER_ACCOUNT_PRIVATE_KEY, l1Provider);

  // PlonkVerifier and LineaRollup deployment
  const plonkVerifier = await deployContract(new TestPlonkVerifierForDataAggregation__factory(), l1Deployer);

  const lineaRollup = (await deployUpgradableContractWithProxyAdmin(new LineaRollup__factory(), l1Deployer, [
    LINEA_ROLLUP_INITIAL_STATE_ROOT_HASH,
    LINEA_ROLLUP_INITIAL_L2_BLOCK_NR,
    plonkVerifier.address,
    LINEA_ROLLUP_SECURITY_COUNCIL,
    LINEA_ROLLUP_OPERATORS,
    LINEA_ROLLUP_RATE_LIMIT_PERIOD,
    LINEA_ROLLUP_RATE_LIMIT_AMOUNT,
  ])) as LineaRollup;

  global.l1Provider = l1Provider;
  global.l2Provider = l2Provider;
  global.dummyContract = dummyContract;
  global.l2MessageService = l2MessageService;
  global.lineaRollup = lineaRollup;
  global.useLocalSetup = true;
  global.L2_ACCOUNT_0_PRIVATE_KEY = L2_ACCOUNT_0_PRIVATE_KEY;
  global.L2_ACCOUNT_1_PRIVATE_KEY = L2_ACCOUNT_1_PRIVATE_KEY;
  global.L1_DEPLOYER_ACCOUNT_PRIVATE_KEY = L1_DEPLOYER_ACCOUNT_PRIVATE_KEY;
  global.L2_DEPLOYER_ACCOUNT_PRIVATE_KEY = L2_DEPLOYER_ACCOUNT_PRIVATE_KEY;
  global.TRANSACTION_CALLDATA_LIMIT = TRANSACTION_CALLDATA_LIMIT;
  global.OPERATOR_0_PRIVATE_KEY = OPERATOR_0_PRIVATE_KEY;
  global.SHOMEI_ENDPOINT = SHOMEI_ENDPOINT;
  global.SHOMEI_FRONTEND_ENDPOINT = SHOMEI_FRONTEND_ENDPOINT;
});
