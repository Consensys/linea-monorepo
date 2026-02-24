import { describe, expect, it } from "@jest/globals";
import { encodeFunctionData, getAddress, parseEventLogs } from "viem";

import { estimateLineaGas, sendTransactionWithRetry } from "./common/utils";
import { L2RpcEndpoint } from "./config/clients/l2-client";
import { createTestContext } from "./config/setup";
import { TestEIP7702DelegationAbi, TestEIP7702DelegationAbiBytecode } from "./generated";

const EIP7702_DELEGATION_PREFIX = "0xef0100";

const context = createTestContext();
const l2AccountManager = context.getL2AccountManager();

describe("EIP-7702 test suite", () => {
  const lineaEstimateGasClient = context.l2PublicClient({ type: L2RpcEndpoint.BesuNode });

  it.concurrent("Should successfully send a non-sponsored EIP-7702 transaction", async () => {
    const [deployer, eoa] = await l2AccountManager.generateAccounts(2);
    const l2PublicClient = context.l2PublicClient();

    // Deploy the TestEIP7702Delegation target contract
    const deployerWalletClient = context.l2WalletClient({ account: deployer });
    const deployNonce = await l2PublicClient.getTransactionCount({ address: deployer.address });

    const deployEstimate = await estimateLineaGas(lineaEstimateGasClient, {
      account: deployer,
      data: TestEIP7702DelegationAbiBytecode,
    });

    const { receipt: deployReceipt } = await sendTransactionWithRetry(l2PublicClient, (fees) =>
      deployerWalletClient.sendTransaction({
        data: TestEIP7702DelegationAbiBytecode,
        nonce: deployNonce,
        ...deployEstimate,
        ...fees,
      }),
    );

    expect(deployReceipt.status).toEqual("success");
    expect(deployReceipt.contractAddress).toBeTruthy();
    const targetContractAddress = getAddress(deployReceipt.contractAddress!);

    logger.debug(`TestEIP7702Delegation deployed. address=${targetContractAddress}`);

    // Send through the sequencer endpoint: the RPC node (default endpoint) has
    // tx-pool simulation checks enabled which cause "Internal error" for type 4
    // transactions because the simulator cannot process authorization lists.
    const eoaWalletClient = context.l2WalletClient({ account: eoa, type: L2RpcEndpoint.Sequencer });

    // Sign EIP-7702 authorization: EOA delegates to the target contract.
    // `executor: 'self'` tells viem the EOA is both the signer and sender,
    // so it auto-increments the authorization nonce by 1 per EIP-7702 spec.
    const authorization = await eoaWalletClient.signAuthorization({
      contractAddress: targetContractAddress,
      executor: "self",
    });

    logger.debug(`EIP-7702 authorization signed. eoaAddress=${eoa.address} target=${targetContractAddress}`);

    // Send type 4 transaction calling initialize() through the delegated EOA
    const initializeData = encodeFunctionData({
      abi: TestEIP7702DelegationAbi,
      functionName: "initialize",
    });

    const eoaNonce = await l2PublicClient.getTransactionCount({ address: eoa.address });

    // Estimate fees against the target contract for accurate maxFeePerGas / maxPriorityFeePerGas.
    // The gas limit from this estimate is too low for EIP-7702 because it doesn't account for
    // per-authorization intrinsic costs (PER_AUTH_BASE_COST + PER_EMPTY_ACCOUNT_COST ≈ 27,500).
    const { maxFeePerGas, maxPriorityFeePerGas } = await estimateLineaGas(lineaEstimateGasClient, {
      account: deployer,
      to: targetContractAddress,
      data: initializeData,
    });

    // Send type 4 tx via sequencer and wait for receipt on the default RPC node.
    // sendTransactionWithRetry is not used here because its retry loop resends with
    // bumped fees and a fixed nonce, but the signed authorization embeds a one-time
    // nonce that becomes invalid after the first inclusion attempt.
    const hash = await eoaWalletClient.sendTransaction({
      authorizationList: [authorization],
      to: eoa.address,
      data: initializeData,
      nonce: eoaNonce,
      gas: 100_000n,
      maxFeePerGas,
      maxPriorityFeePerGas,
    });

    logger.debug(`EIP-7702 transaction sent. transactionHash=${hash}`);

    const receipt = await l2PublicClient.waitForTransactionReceipt({ hash, timeout: 60_000 });

    logger.debug(`EIP-7702 transaction receipt received. transactionHash=${hash} status=${receipt.status}`);
    expect(receipt.status).toEqual("success");

    // Verify delegation: EOA code should start with 0xef0100 followed by the target contract address
    const eoaCode = await l2PublicClient.getCode({ address: eoa.address });
    expect(eoaCode).toBeDefined();
    expect(eoaCode!.toLowerCase().startsWith(EIP7702_DELEGATION_PREFIX)).toBe(true);

    const delegatedAddress = getAddress(`0x${eoaCode!.slice(EIP7702_DELEGATION_PREFIX.length)}`);
    expect(delegatedAddress).toEqual(targetContractAddress);

    logger.debug(`Delegation verified. eoaAddress=${eoa.address} delegatedTo=${delegatedAddress}`);

    // Verify the Log event was emitted by calling initialize() through the delegated EOA
    const logs = parseEventLogs({
      abi: TestEIP7702DelegationAbi,
      logs: receipt.logs,
      eventName: "Log",
    });

    expect(logs).toHaveLength(1);
    expect(logs[0].args.message).toEqual("Hello, world computer!");

    logger.debug("EIP-7702 Log event verified.");
  });
});
