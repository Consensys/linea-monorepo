/* eslint-disable no-var */
import { JsonRpcProvider } from "@ethersproject/providers";
import { TestEnvironment } from "./test-env";
import { DummyContract, L2MessageService, L2TestContract, LineaRollup } from "src/typechain";

declare global {
  var testingEnv: TestEnvironment;
  var l1Provider: JsonRpcProvider;
  var l2Provider: JsonRpcProvider;
  var l2MessageService: L2MessageService;
  var dummyContract: DummyContract;
  var l1DummyContract: DummyContract;
  var l2TestContract: L2TestContract;
  var lineaRollup: LineaRollup;
  var useLocalSetup: boolean;
  var chainId: number;
  var CHAIN_ID: number;
  var L2_MESSAGE_SERVICE_ADDRESS: string;
  var LINEA_ROLLUP_CONTRACT_ADDRESS: string;
  var L1_ACCOUNT_0_PRIVATE_KEY: string;
  var L2_ACCOUNT_0_PRIVATE_KEY: string;
  var L2_ACCOUNT_1_PRIVATE_KEY: string;
  var L1_DEPLOYER_ACCOUNT_PRIVATE_KEY: string;
  var L2_DEPLOYER_ACCOUNT_PRIVATE_KEY: string;
  var TRANSACTION_CALLDATA_LIMIT: number;
  var OPERATOR_0_PRIVATE_KEY: string;
  var SHOMEI_ENDPOINT: URL | null;
  var SHOMEI_FRONTEND_ENDPOINT: URL | null;
  var SEQUENCER_ENDPOINT: URL | null;
  var TRANSACTION_EXCLUSION_ENDPOINT: URL | null;
  var OPERATOR_1_ADDRESS: string;
  var SECURITY_COUNCIL_PRIVATE_KEY: string;
  var CONTRACT_GAS_OPTIMIZATION_SWITCH_BLOCK: number;
}
