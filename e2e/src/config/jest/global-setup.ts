import { deployContract, linkBytecode } from "../../common/deployments";
import { estimateLineaGas, etherToWei, sendTransactionsToGenerateTrafficWithInterval } from "../../common/utils";
import { createTestLogger } from "../logger";
import { encodeDeployData, formatEther, parseGwei } from "viem";
import {
  DummyContractAbi,
  DummyContractAbiBytecode,
  LineaSequencerUptimeFeedAbi,
  LineaSequencerUptimeFeedAbiBytecode,
  MimcAbi,
  MimcAbiBytecode,
  SparseMerkleProofAbi,
  SparseMerkleProofAbiBytecode,
  SparseMerkleProofAbiLinkReferences,
  TestContractAbi,
  TestContractAbiBytecode,
} from "../../generated";
import { config } from "../tests-config/setup";
import { L2RpcEndpoint } from "../tests-config/setup/clients/l2-client";
import { DEPLOYER_ACCOUNT_INDEX, LIVENESS_ACCOUNT_INDEX } from "../../common/constants";

const logger = createTestLogger();

export default async (): Promise<void> => {
  const dummyContractCode = await config
    .l1PublicClient()
    .getCode({ address: config.l1PublicClient().getDummyContract().address });

  // If this is empty, we have not deployed and prerequisites or configured token bridges.
  if (!dummyContractCode) {
    logger.info("Configuring once-off prerequisite contracts");
    await configureOnceOffPrerequisities();
  }

  logger.info("Generating L2 traffic...");
  const pollingAccount = await config.getL2AccountManager().generateAccount(etherToWei("200"));
  const walletClient = config.l2WalletClient({ account: pollingAccount });
  const publicClient = config.l2PublicClient();
  const stopPolling = await sendTransactionsToGenerateTrafficWithInterval(walletClient, publicClient, {
    pollingInterval: 2_000,
  });

  global.stopL2TrafficGeneration = stopPolling;
};

async function configureOnceOffPrerequisities() {
  const account = config.getL1AccountManager().whaleAccount(DEPLOYER_ACCOUNT_INDEX);
  const l2Account = config.getL2AccountManager().whaleAccount(DEPLOYER_ACCOUNT_INDEX);

  const l1PublicClient = config.l1PublicClient();
  const l1WalletClient = config.l1WalletClient({ account });
  const l2SequencerPublicClient = config.l2PublicClient({ type: L2RpcEndpoint.Sequencer });
  const l2SequencerWalletClient = config.l2WalletClient({ type: L2RpcEndpoint.Sequencer, account: l2Account });
  /**
   * Account index {@link LIVENESS_ACCOUNT_INDEX} is reserved for liveness testing to avoid nonce conflicts with other concurrent e2e tests.
   */
  const livenessSignerAccount = config.getL2AccountManager().whaleAccount(LIVENESS_ACCOUNT_INDEX);

  const lineaRollup = l1WalletClient.getLineaRollup();

  const [l1AccountNonce, l2AccountNonce] = await Promise.all([
    l1PublicClient.getTransactionCount({ address: account.address }),
    l2SequencerPublicClient?.getTransactionCount({ address: l2Account.address }),
  ]);

  const fee = etherToWei("3");
  const to = "0x8D97689C9818892B700e27F316cc3E41e17fBeb9";
  const calldata = "0x";

  const l2BesuNodePublicClient = config.l2PublicClient({ type: L2RpcEndpoint.BesuNode })!;
  const [
    { maxPriorityFeePerGas, maxFeePerGas },
    { maxPriorityFeePerGas: maxPriorityFeePerGasTestContract, maxFeePerGas: maxFeePerGasTestContract },
    { maxPriorityFeePerGas: maxPriorityFeePerGasMimc, maxFeePerGas: maxFeePerGasMimc },
    {
      maxPriorityFeePerGas: maxPriorityFeePerGasLineaSequencerUptimeFeed,
      maxFeePerGas: maxFeePerGasLineaSequencerUptimeFeed,
    },
  ] = await Promise.all([
    estimateLineaGas(l2BesuNodePublicClient, {
      account: l2Account.address,
      data: encodeDeployData({
        abi: DummyContractAbi,
        bytecode: DummyContractAbiBytecode,
      }),
      value: 0n,
    }),
    estimateLineaGas(l2BesuNodePublicClient, {
      account: l2Account.address,
      data: encodeDeployData({
        abi: TestContractAbi,
        bytecode: TestContractAbiBytecode,
      }),
      value: 0n,
    }),
    estimateLineaGas(l2BesuNodePublicClient, {
      account: l2Account.address,
      data: encodeDeployData({
        abi: MimcAbi,
        bytecode: MimcAbiBytecode,
      }),
      value: 0n,
    }),
    estimateLineaGas(l2BesuNodePublicClient, {
      account: l2Account.address,
      data: encodeDeployData({
        abi: LineaSequencerUptimeFeedAbi,
        bytecode: LineaSequencerUptimeFeedAbiBytecode,
        args: [false, livenessSignerAccount.address, livenessSignerAccount.address],
      }),
      value: 0n,
    }),
  ]);

  const [
    dummyContractAddress,
    l2DummyContractAddress,
    l2TestContractAddress,
    l2MimcContractAddress,
    l2LineaSequencerUptimeFeedContractAddress,
  ] = await Promise.all([
    deployContract(l1WalletClient, {
      account,
      chain: l1WalletClient.chain,
      abi: DummyContractAbi,
      bytecode: DummyContractAbiBytecode,
      nonce: l1AccountNonce,
    }),
    deployContract(l2SequencerWalletClient!, {
      account: l2Account,
      chain: l2SequencerWalletClient!.chain,
      abi: DummyContractAbi,
      bytecode: DummyContractAbiBytecode,
      nonce: l2AccountNonce,
      maxPriorityFeePerGas,
      maxFeePerGas,
    }),
    deployContract(l2SequencerWalletClient!, {
      account: l2Account,
      chain: l2SequencerWalletClient!.chain,
      abi: TestContractAbi,
      bytecode: TestContractAbiBytecode,
      nonce: l2AccountNonce! + 1,
      maxPriorityFeePerGas: maxPriorityFeePerGasTestContract,
      maxFeePerGas: maxFeePerGasTestContract,
    }),
    deployContract(l2SequencerWalletClient!, {
      account: l2Account,
      chain: l2SequencerWalletClient!.chain,
      abi: MimcAbi,
      bytecode: MimcAbiBytecode,
      nonce: l2AccountNonce! + 2,
      maxPriorityFeePerGas: maxPriorityFeePerGasMimc,
      maxFeePerGas: maxFeePerGasMimc,
    }),
    deployContract(l2SequencerWalletClient!, {
      account: l2Account,
      chain: l2SequencerWalletClient!.chain,
      abi: LineaSequencerUptimeFeedAbi,
      bytecode: LineaSequencerUptimeFeedAbiBytecode,
      args: [false, livenessSignerAccount.address, livenessSignerAccount.address],
      nonce: l2AccountNonce! + 3,
      maxPriorityFeePerGas: maxPriorityFeePerGasLineaSequencerUptimeFeed,
      maxFeePerGas: maxFeePerGasLineaSequencerUptimeFeed,
    }),
    // Send ETH to the LineaRollup contract
    l1PublicClient.waitForTransactionReceipt({
      hash: await lineaRollup.write.sendMessage([to, fee, calldata], {
        value: etherToWei("500"),
        gasPrice: parseGwei("300"),
        nonce: l1AccountNonce + 1,
      }),
    }),
  ]);

  const lineaRollupBalance = await l1PublicClient.getBalance({ address: lineaRollup.address });
  if (lineaRollupBalance < etherToWei("500")) {
    throw new Error("LineaRollup funding failed");
  }

  const { maxPriorityFeePerGas: maxPriorityFeePerGasSparseMerkleProof, maxFeePerGas: maxFeePerGasSparseMerkleProof } =
    await estimateLineaGas(l2BesuNodePublicClient, {
      account: l2Account.address,
      data: encodeDeployData({
        abi: SparseMerkleProofAbi,
        bytecode: linkBytecode(SparseMerkleProofAbiBytecode, SparseMerkleProofAbiLinkReferences, {
          Mimc: l2MimcContractAddress!,
        }),
      }),
      value: 0n,
    });

  const l2SparseMerkleProofContractAddress = await deployContract(l2SequencerWalletClient!, {
    account: l2Account,
    chain: l2SequencerWalletClient!.chain,
    abi: SparseMerkleProofAbi,
    bytecode: linkBytecode(SparseMerkleProofAbiBytecode, SparseMerkleProofAbiLinkReferences, {
      Mimc: l2MimcContractAddress!,
    }),
    nonce: l2AccountNonce! + 4,
    maxPriorityFeePerGas: maxPriorityFeePerGasSparseMerkleProof,
    maxFeePerGas: maxFeePerGasSparseMerkleProof,
  });

  logger.info(`L1 Dummy contract deployed. address=${dummyContractAddress}`);
  logger.info(`L2 Dummy contract deployed. address=${l2DummyContractAddress}`);
  logger.info(`L2 Test contract deployed. address=${l2TestContractAddress}`);
  logger.info(`L2 Mimc contract deployed. address=${l2MimcContractAddress}`);
  logger.info(`L2 LineaSequencerUptimeFeed contract deployed. address=${l2LineaSequencerUptimeFeedContractAddress}`);
  logger.info(`L2 SparseMerkleProof contract deployed. address=${l2SparseMerkleProofContractAddress}`);
  logger.info(`LineaRollup funded with ${formatEther(lineaRollupBalance)} ETH on L1`);
}
