import { beforeAll, jest } from "@jest/globals";
import { Wallet } from "ethers";
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
} from "../src/utils/constants.local";
import { deployContract } from "../src/utils/deployments";

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

  // Old LineaRollup and old L2MessageService contracts instanciation
  // New contracts ABI are used because there are backward compatible
  const lineaRollup: LineaRollup = LineaRollup__factory.connect(LINEA_ROLLUP_CONTRACT_ADDRESS, l1JsonRpcProvider);
  const l2MessageService: L2MessageService = L2MessageService__factory.connect(
    MESSAGE_SERVICE_ADDRESS,
    l2JsonRpcProvider,
  );

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
});
