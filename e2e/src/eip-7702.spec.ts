import { Authorization, ethers } from "ethers";
import { describe, expect, it } from "@jest/globals";
import { config } from "./config/tests-config";
import { LineaEstimateGasClient, waitForEvents } from "./common/utils";

const l2AccountManager = config.getL2AccountManager();

// Constants
const DEFAULT_GAS_LIMIT = 500000n;

describe("EIP-7702 test suite", () => {
  const lineaEstimateGasClient = new LineaEstimateGasClient(config.getL2BesuNodeEndpoint()!);

  // Helper function to deploy a simple test contract for delegation
  // This matches the TestEIP7702Delegation contract from contracts/src/_testing/mocks/base/TestEIP7702Delegation.sol
  const deployTestContract = async (deployer: ethers.Wallet) => {
    const TestEIP7702DelegationABI = ["event Log(string message)", "function initialize() external"];

    // Bytecode for TestEIP7702Delegation contract
    // This is the compiled output of the contract which emits a Log event with message "Hello, world computer!"
    // Note: In production, this should ideally be imported from compiled artifacts
    const bytecode =
      "0x608060405234801561001057600080fd5b5060c78061001f6000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c8063e1c7392a14602d575b600080fd5b60336035565b005b7f0b2e13ff20ac7b474198655583edf70dedd2c1dc980e329c4fbb2fc0748b796b6040518060400160405280601981526020017f48656c6c6f2c20776f726c6420636f6d7075746572210000000000000000000081525060405180806020018281038252602081526020018280519060200190809190601f01602080910402602001604051908101604052809392919081815260200183838082843760405192018290039091209050607160f01b90529091019060200390a1565b00";

    const factory = new ethers.ContractFactory(TestEIP7702DelegationABI, bytecode, deployer);
    const contract = await factory.deploy();
    await contract.waitForDeployment();
    return contract;
  };

  it.concurrent(
    "Should successfully send a Type 4 EIP-7702 transaction and verify finalization",
    async () => {
      // Generate account for testing
      const account = await l2AccountManager.generateAccount();
      const provider = config.getL2Provider();
      const lineaRollup = config.getLineaRollupContract();

      logger.debug(`Testing EIP-7702 transaction with account: ${account.address}`);

      // Deploy TestEIP7702Delegation contract
      const delegationContract = await deployTestContract(account);
      const targetAddress = await delegationContract.getAddress();

      logger.debug(`TestEIP7702Delegation deployed at: ${targetAddress}`);

      // Get current network details
      const network = await provider.getNetwork();
      const currentChainId = network.chainId;
      const currentNonce = await provider.getTransactionCount(account.address);

      // Create authorization - nonce should be currentNonce + 1 when the sender is also the authorization signer
      const authNonce = currentNonce + 1;
      const authorization: Authorization = await account.authorize({
        address: targetAddress,
        nonce: authNonce,
        chainId: currentChainId,
      });

      logger.debug(`Authorization created with nonce: ${authorization.nonce}`);

      // Get current L2 block number for finalization tracking
      const currentL2BlockNumber = await lineaRollup.currentL2BlockNumber();
      logger.debug(`Current L2 block number before transaction: ${currentL2BlockNumber}`);

      // Prepare transaction data
      const calldata = delegationContract.interface.encodeFunctionData("initialize");

      // Get gas estimates from Linea
      const { maxPriorityFeePerGas, maxFeePerGas, gasLimit } = await lineaEstimateGasClient.lineaEstimateGas(
        account.address,
        account.address, // Calling own address after delegation
        calldata,
      );

      logger.debug(
        `Gas estimates: maxPriorityFeePerGas=${maxPriorityFeePerGas}, maxFeePerGas=${maxFeePerGas}, gasLimit=${gasLimit}`,
      );

      // Send Type 4 (EIP-7702) transaction
      const tx = await account.sendTransaction({
        type: 4,
        to: account.address, // Send to self (EOA will be delegated to contract)
        authorizationList: [authorization],
        data: calldata,
        gasLimit: gasLimit > DEFAULT_GAS_LIMIT ? gasLimit : DEFAULT_GAS_LIMIT,
        value: 0n,
        maxPriorityFeePerGas: maxPriorityFeePerGas,
        maxFeePerGas: maxFeePerGas,
      });

      logger.debug(`Type 4 (EIP-7702) transaction sent: ${tx.hash}`);

      // Wait for transaction receipt
      const receipt = await tx.wait();
      logger.debug(`Transaction mined in block: ${receipt?.blockNumber}, status: ${receipt?.status}`);

      // Verify transaction was included successfully
      expect(receipt).not.toBeNull();
      expect(receipt?.status).toEqual(1);
      expect(receipt?.type).toEqual(4);

      // Verify delegation was set up correctly
      const code = await provider.getCode(account.address);
      logger.debug(`Code at EOA after delegation: ${code}`);

      // EIP-7702 delegation code starts with 0xef0100
      expect(code.startsWith("0xef0100")).toBe(true);

      logger.debug("Waiting for transaction to be finalized on L1...");

      // Wait for the block containing our transaction to be finalized
      // This ensures the full lifecycle: submission → inclusion → finalization
      const transactionBlockNumber = receipt!.blockNumber;

      // Wait for DataFinalizedV3 event that includes our transaction's block
      const [dataFinalizedEvent] = await waitForEvents(
        lineaRollup,
        lineaRollup.filters.DataFinalizedV3(),
        1_000,
        0,
        "latest",
        async (events) => events.filter((event) => event.args.endBlockNumber >= transactionBlockNumber),
      );

      logger.debug(
        `Finalization event received. Start block: ${dataFinalizedEvent.args.startBlockNumber}, End block: ${dataFinalizedEvent.args.endBlockNumber}`,
      );

      // Verify finalization details
      const [lastBlockFinalized, stateRootHash] = await Promise.all([
        lineaRollup.currentL2BlockNumber(),
        lineaRollup.stateRootHashes(dataFinalizedEvent.args.endBlockNumber),
      ]);

      expect(lastBlockFinalized).toBeGreaterThanOrEqual(transactionBlockNumber);
      expect(lastBlockFinalized).toBeGreaterThanOrEqual(dataFinalizedEvent.args.endBlockNumber);
      expect(stateRootHash).toEqual(dataFinalizedEvent.args.finalStateRootHash);

      logger.debug(
        `EIP-7702 transaction successfully finalized. Last finalized block: ${lastBlockFinalized}, Transaction block: ${transactionBlockNumber}`,
      );
    },
    150_000, // 150 second timeout
  );
});
