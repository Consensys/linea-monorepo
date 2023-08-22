/* eslint-disable no-var */
import { JsonRpcProvider } from "@ethersproject/providers";
import { TestEnvironment } from "./test-env";
import { DummyContract, L2MessageService, L2TestContract, ZkEvmV2 } from "src/typechain";

declare global {
  var testingEnv: TestEnvironment;
  var l1Provider: JsonRpcProvider;
  var l2Provider: JsonRpcProvider;
  var l2MessageService: L2MessageService;
  var dummyContract: DummyContract;
  var l1DummyContract: DummyContract;
  var l2TestContract: L2TestContract;
  var zkEvmV2: ZkEvmV2;
  var useLocalSetup: Boolean;
  var chainId: number
  var CHAIN_ID: number;
  var ACCOUNT_0_PRIVATE_KEY: string;
  var DEPLOYER_ACCOUNT_PRIVATE_KEY: string;
  var TRANSACTION_CALLDATA_LIMIT: number;
  var OPERATOR_0_PRIVATE_KEY: string;
}
