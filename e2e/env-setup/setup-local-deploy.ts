import { beforeAll, jest } from "@jest/globals";
import { Wallet } from "ethers";
import {
  DummyContract,
  DummyContract__factory,
  L2MessageService,
  L2MessageService__factory,
  PlonkVerifier__factory,
  TransactionDecoder__factory,
  ZkEvmV2,
  ZkEvmV2__factory,
} from "../src/typechain";
import { ZkEvmV2LibraryAddresses } from "../src/typechain/factories/ZkEvmV2__factory";
import {
  ACCOUNT_0,
  ACCOUNT_0_PRIVATE_KEY,
  DEPLOYER_ACCOUNT_PRIVATE_KEY,
  INITIAL_WITHDRAW_LIMIT,
  ZKEVMV2_INITIAL_L2_BLOCK_NR,
  ZKEVMV2_INITIAL_STATE_ROOT_HASH,
  ZKEVMV2_OPERATORS,
  ZKEVMV2_RATE_LIMIT_AMOUNT,
  ZKEVMV2_RATE_LIMIT_PERIOD,
  ZKEVMV2_SECURITY_COUNCIL,
  getL1Provider,
  getL2Provider,
  TRANSACTION_CALLDATA_LIMIT,
  OPERATOR_0_PRIVATE_KEY
} from "../src/utils/constants.local";
import { deployContract, deployUpgradableContractWithProxyAdmin, encodeLibraryName } from "../src/utils/deployments";

jest.setTimeout(5 * 60 * 1000);

beforeAll(async () => {
  /*********** PROVIDERS SETUP ***********/
  const l1Provider = getL1Provider();
  const l2Provider = getL2Provider();

  const l2Deployer = new Wallet(DEPLOYER_ACCOUNT_PRIVATE_KEY, l2Provider);
  const dummyContract = (await deployContract(new DummyContract__factory(), l2Deployer)) as DummyContract;

  // L2MessageService deployment
  const l2MessageService = (await deployUpgradableContractWithProxyAdmin(new L2MessageService__factory(), l2Deployer, [
    ACCOUNT_0,
    86400,
    INITIAL_WITHDRAW_LIMIT,
  ])) as L2MessageService;

  /*********** L1 DEPLOYMENTS ***********/
  const l1Deployer = new Wallet(DEPLOYER_ACCOUNT_PRIVATE_KEY, l1Provider);

  // ZkEvmV2 deployment
  const transactionDecoder = await deployContract(new TransactionDecoder__factory(), l1Deployer);
  const plonkVerifier = await deployContract(new PlonkVerifier__factory(), l1Deployer);

  const transactionDecoderBytecodePlaceholder = encodeLibraryName(
    "contracts/messageService/lib/TransactionDecoder.sol:TransactionDecoder",
  );

  const zkEvmV2 = (await deployUpgradableContractWithProxyAdmin(
    new ZkEvmV2__factory({
      [`${transactionDecoderBytecodePlaceholder}`]: transactionDecoder.address,
    } as unknown as ZkEvmV2LibraryAddresses),
    l1Deployer,
    [
      ZKEVMV2_INITIAL_STATE_ROOT_HASH,
      ZKEVMV2_INITIAL_L2_BLOCK_NR,
      plonkVerifier.address,
      ZKEVMV2_SECURITY_COUNCIL,
      ZKEVMV2_OPERATORS,
      ZKEVMV2_RATE_LIMIT_PERIOD,
      ZKEVMV2_RATE_LIMIT_AMOUNT,
    ],
  )) as ZkEvmV2;

  global.l1Provider = l1Provider;
  global.l2Provider = l2Provider;
  global.dummyContract = dummyContract;
  global.l2MessageService = l2MessageService;
  global.zkEvmV2 = zkEvmV2;
  global.useLocalSetup = true;
  global.ACCOUNT_0_PRIVATE_KEY = ACCOUNT_0_PRIVATE_KEY;
  global.DEPLOYER_ACCOUNT_PRIVATE_KEY = DEPLOYER_ACCOUNT_PRIVATE_KEY;
  global.TRANSACTION_CALLDATA_LIMIT = TRANSACTION_CALLDATA_LIMIT;
  global.OPERATOR_0_PRIVATE_KEY = OPERATOR_0_PRIVATE_KEY;
});
