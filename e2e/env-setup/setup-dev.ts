import { beforeAll, jest } from "@jest/globals";
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
  DEPLOYER_ACCOUNT_PRIVATE_KEY,
  MESSAGE_SERVICE_ADDRESS,
  DUMMY_CONTRACT_ADDRESS,
  ACCOUNT_0_PRIVATE_KEY,
  TRANSACTION_CALLDATA_LIMIT,
  OPERATOR_0_PRIVATE_KEY,
  SHOMEI_ENDPOINT,
  SHOMEI_FRONTEND_ENDPOINT,
  TRANSACTION_EXCLUSION_ENDPOINT,
  LINEA_ROLLUP_CONTRACT_ADDRESS,
} from "../src/utils/constants.dev";

jest.setTimeout(5 * 60 * 1000);

beforeAll(async () => {
  /*********** PROVIDERS SETUP ***********/
  const l1Provider = getL1Provider();
  const l2Provider = getL2Provider();

  const dummyContract: DummyContract = DummyContract__factory.connect(DUMMY_CONTRACT_ADDRESS, l2Provider);
  // L2MessageService contract
  const l2MessageService: L2MessageService = L2MessageService__factory.connect(MESSAGE_SERVICE_ADDRESS, l2Provider);

  /*********** L1 Contracts ***********/
  // LineaRollup deployment
  const lineaRollup: LineaRollup = LineaRollup__factory.connect(LINEA_ROLLUP_CONTRACT_ADDRESS, l1Provider);

  global.l1Provider = l1Provider;
  global.l2Provider = l2Provider;
  global.dummyContract = dummyContract;
  global.l2MessageService = l2MessageService;
  global.lineaRollup = lineaRollup;
  global.chainId = CHAIN_ID;
  global.L1_ACCOUNT_0_PRIVATE_KEY = ACCOUNT_0_PRIVATE_KEY;
  global.L1_DEPLOYER_ACCOUNT_PRIVATE_KEY = DEPLOYER_ACCOUNT_PRIVATE_KEY;
  global.L2_DEPLOYER_ACCOUNT_PRIVATE_KEY = DEPLOYER_ACCOUNT_PRIVATE_KEY;
  global.TRANSACTION_CALLDATA_LIMIT = TRANSACTION_CALLDATA_LIMIT;
  global.OPERATOR_0_PRIVATE_KEY = OPERATOR_0_PRIVATE_KEY;
  global.SHOMEI_ENDPOINT = SHOMEI_ENDPOINT;
  global.SHOMEI_FRONTEND_ENDPOINT = SHOMEI_FRONTEND_ENDPOINT;
  global.TRANSACTION_EXCLUSION_ENDPOINT = TRANSACTION_EXCLUSION_ENDPOINT;
});
