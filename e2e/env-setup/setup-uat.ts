import { beforeAll, jest } from "@jest/globals";
import {
  DummyContract,
  DummyContract__factory,
  L2MessageService,
  L2MessageService__factory,
  L2TestContract__factory,
  L2TestContract,
  LineaRollup,
  LineaRollup__factory,
} from "../src/typechain";
import {
  getL1Provider,
  getL2Provider,
  CHAIN_ID,
  LINEA_ROLLUP_CONTRACT_ADDRESS,
  MESSAGE_SERVICE_ADDRESS,
  DUMMY_CONTRACT_ADDRESS,
  ACCOUNT_0_PRIVATE_KEY,
  TRANSACTION_CALLDATA_LIMIT,
  L1_DUMMY_CONTRACT_ADDRESS,
  SHOMEI_ENDPOINT,
  SHOMEI_FRONTEND_ENDPOINT,
} from "../src/utils/constants.uat";

jest.setTimeout(5 * 60 * 1000);

beforeAll(async () => {
  /*********** PROVIDERS SETUP ***********/
  const l1Provider = getL1Provider();
  const l2Provider = getL2Provider();

  const l1DummyContract: DummyContract = DummyContract__factory.connect(L1_DUMMY_CONTRACT_ADDRESS, l1Provider);
  const dummyContract: DummyContract = DummyContract__factory.connect(DUMMY_CONTRACT_ADDRESS, l2Provider);
  const l2TestContract: L2TestContract = L2TestContract__factory.connect(
    "0xaD711a736ae454Be345078C0bc849997b78A30B9",
    l2Provider,
  );
  // L2MessageService contract
  const l2MessageService: L2MessageService = L2MessageService__factory.connect(MESSAGE_SERVICE_ADDRESS, l2Provider);

  /*********** L1 Contracts ***********/

  // LineaRollup deployment
  const lineaRollup: LineaRollup = LineaRollup__factory.connect(LINEA_ROLLUP_CONTRACT_ADDRESS, l1Provider);

  global.l1Provider = l1Provider;
  global.l2Provider = l2Provider;
  global.l2TestContract = l2TestContract;
  global.l1DummyContract = l1DummyContract;
  global.dummyContract = dummyContract;
  global.l2MessageService = l2MessageService;
  global.lineaRollup = lineaRollup;
  global.chainId = CHAIN_ID;
  global.L1_ACCOUNT_0_PRIVATE_KEY = ACCOUNT_0_PRIVATE_KEY;
  global.TRANSACTION_CALLDATA_LIMIT = TRANSACTION_CALLDATA_LIMIT;
  global.SHOMEI_ENDPOINT = SHOMEI_ENDPOINT;
  global.SHOMEI_FRONTEND_ENDPOINT = SHOMEI_FRONTEND_ENDPOINT;
});
