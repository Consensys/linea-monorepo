import { mock, MockProxy } from "jest-mock-extended";
import type { ILogger, IBlockchainClient } from "@consensys/linea-shared-utils";
import { absDiff, ONE_ETHER, weiToGweiNumber } from "@consensys/linea-shared-utils";
import type { Address, Hex, PublicClient, TransactionReceipt } from "viem";
import { YieldManagerABI } from "../../../core/abis/YieldManager.js";
import { StakingVaultABI } from "../../../core/abis/StakingVault.js";
import { DashboardErrorsABI } from "../../../core/abis/errors/DashboardErrors.js";
import { LidoStVaultYieldProviderErrorsABI } from "../../../core/abis/errors/LidoStVaultYieldProviderErrors.js";
import { StakingVaultErrorsABI } from "../../../core/abis/errors/StakingVaultErrors.js";
import { VaultHubErrorsABI } from "../../../core/abis/errors/VaultHubErrors.js";
import { RebalanceDirection } from "../../../core/entities/RebalanceRequirement.js";
import type { WithdrawalRequests } from "../../../core/entities/LidoStakingVaultWithdrawalParams.js";
import type { INativeYieldAutomationMetricsUpdater } from "../../../core/metrics/INativeYieldAutomationMetricsUpdater.js";
import type { IRebalanceQuotaService } from "../../../core/services/IRebalanceQuotaService.js";

const YieldManagerCombinedABI = [
  ...YieldManagerABI,
  ...DashboardErrorsABI,
  ...LidoStVaultYieldProviderErrorsABI,
  ...StakingVaultErrorsABI,
  ...VaultHubErrorsABI,
] as const;

jest.mock("viem", () => {
  const actual = jest.requireActual("viem");
  return {
    ...actual,
    getContract: jest.fn(),
    encodeFunctionData: jest.fn(),
    parseEventLogs: jest.fn(),
    encodeAbiParameters: jest.fn(),
  };
});

import { concat, getContract, encodeFunctionData, parseEventLogs, encodeAbiParameters } from "viem";

jest.mock("../DashboardContractClient.js", () => ({
  DashboardContractClient: {
    getOrCreate: jest.fn(),
    initialize: jest.fn(),
  },
}));

import { DashboardContractClient } from "../DashboardContractClient.js";

const mockedGetContract = getContract as jest.MockedFunction<typeof getContract>;
const mockedEncodeFunctionData = encodeFunctionData as jest.MockedFunction<typeof encodeFunctionData>;
const mockedParseEventLogs = parseEventLogs as jest.MockedFunction<typeof parseEventLogs>;
const mockedEncodeAbiParameters = encodeAbiParameters as jest.MockedFunction<typeof encodeAbiParameters>;
const mockedDashboardContractClientGetOrCreate = DashboardContractClient.getOrCreate as jest.MockedFunction<
  typeof DashboardContractClient.getOrCreate
>;

let YieldManagerContractClient: typeof import("../YieldManagerContractClient.js").YieldManagerContractClient;

beforeAll(async () => {
  ({ YieldManagerContractClient } = await import("../YieldManagerContractClient.js"));
});

describe("YieldManagerContractClient", () => {
  const contractAddress = "0xcccccccccccccccccccccccccccccccccccccccc" as Address;
  const yieldProvider = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" as Address;
  const l2Recipient = "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" as Address;
  const l1MessageServiceAddress = "0x9999999999999999999999999999999999999999" as Address;
  const stakingVaultAddress = "0x8888888888888888888888888888888888888888" as Address;
  const signerAddress = "0xdddddddddddddddddddddddddddddddddddddddd" as Address;

  let logger: MockProxy<ILogger>;
  let blockchainClient: MockProxy<IBlockchainClient<PublicClient, TransactionReceipt>>;
  let publicClient: { readContract: jest.Mock } & Record<string, unknown>;
  let contractStub: {
    abi: typeof YieldManagerABI;
    read: Record<string, jest.Mock>;
    simulate: Record<string, jest.Mock>;
  };

  const defaultYieldProviderData = {
    yieldProviderVendor: 0,
    isStakingPaused: false,
    isOssificationInitiated: false,
    isOssified: false,
    primaryEntrypoint: l2Recipient,
    ossifiedEntrypoint: stakingVaultAddress,
    yieldProviderIndex: 0n,
    userFunds: 0n,
    yieldReportedCumulative: 0n,
    lstLiabilityPrincipal: 0n,
    lastReportedNegativeYield: 0n,
  };

  const buildReceipt = (logs: Array<{ address: string; data: string; topics: string[] }>): TransactionReceipt =>
    ({
      logs,
    }) as unknown as TransactionReceipt;

  const createDefaultQuotaService = (): IRebalanceQuotaService => ({
    getRebalanceAmountAfterQuota: jest.fn((_vaultAddress, _totalSystemBalance, reBalanceAmountWei) => reBalanceAmountWei),
    getStakingDirection: jest.fn(() => RebalanceDirection.STAKE),
  });

  const createClient = ({
    rebalanceToleranceAmountWei = 1000000000000000000n, // 1 ETH default
    minWithdrawalThresholdEth = 0n,
    rebalanceQuotaService = createDefaultQuotaService(),
    metricsUpdater,
  }: {
    rebalanceToleranceAmountWei?: bigint;
    minWithdrawalThresholdEth?: bigint;
    rebalanceQuotaService?: IRebalanceQuotaService;
    metricsUpdater?: Partial<INativeYieldAutomationMetricsUpdater>;
  } = {}) =>
    new YieldManagerContractClient(
      logger,
      blockchainClient,
      contractAddress,
      rebalanceToleranceAmountWei,
      minWithdrawalThresholdEth,
      rebalanceQuotaService,
      metricsUpdater as INativeYieldAutomationMetricsUpdater | undefined,
    );

  beforeEach(() => {
    jest.clearAllMocks();
    logger = mock<ILogger>();

    blockchainClient = mock<IBlockchainClient<PublicClient, TransactionReceipt>>();
    publicClient = {
      readContract: jest.fn(),
    };

    blockchainClient.getBlockchainClient.mockReturnValue(publicClient as unknown as PublicClient);

    contractStub = {
      abi: YieldManagerABI,
      read: {
        L1_MESSAGE_SERVICE: jest.fn(),
        getTotalSystemBalance: jest.fn(),
        getEffectiveTargetWithdrawalReserve: jest.fn(),
        getTargetReserveDeficit: jest.fn(),
        isStakingPaused: jest.fn(),
        isOssificationInitiated: jest.fn(),
        isOssified: jest.fn(),
        getYieldProviderData: jest.fn().mockResolvedValue(defaultYieldProviderData),
        userFunds: jest.fn(),
      },
      simulate: {
        withdrawableValue: jest.fn(),
        reportYield: jest.fn(),
      },
    };

    mockedGetContract.mockReturnValue(contractStub as any);
    blockchainClient.getBalance.mockResolvedValue(0n);
    blockchainClient.getSignerAddress.mockReturnValue(signerAddress);
    blockchainClient.sendSignedTransaction.mockResolvedValue({
      transactionHash: "0xreceipt",
    } as unknown as TransactionReceipt);
    contractStub.read.L1_MESSAGE_SERVICE.mockResolvedValue(l1MessageServiceAddress);
    contractStub.read.getTotalSystemBalance.mockResolvedValue(0n);
    contractStub.read.getEffectiveTargetWithdrawalReserve.mockResolvedValue(0n);
    contractStub.read.getTargetReserveDeficit.mockResolvedValue(0n);
    contractStub.read.isStakingPaused.mockResolvedValue(false);
    contractStub.read.isOssificationInitiated.mockResolvedValue(false);
    contractStub.read.isOssified.mockResolvedValue(false);
    contractStub.simulate.withdrawableValue.mockResolvedValue({ result: 0n });
  });

  it("initializes the viem contract and exposes address & contract accessors", () => {
    const client = createClient();

    expect(mockedGetContract).toHaveBeenCalledWith({
      abi: YieldManagerCombinedABI,
      address: contractAddress,
      client: publicClient,
    });
    expect(client.getAddress()).toBe(contractAddress);
    expect(client.getContract()).toBe(contractStub);
  });

  it("gets the contract balance", async () => {
    const balance = 1_000_000_000_000_000_000n; // 1 ETH
    blockchainClient.getBalance.mockResolvedValueOnce(balance);

    const client = createClient();
    await expect(client.getBalance()).resolves.toBe(balance);

    expect(blockchainClient.getBalance).toHaveBeenCalledWith(contractAddress);
  });

  it("delegates simple reads to the viem contract", async () => {
    const stakingPaused = true;
    const ossificationInitiated = true;
    const ossified = true;
    contractStub.read.getTotalSystemBalance.mockResolvedValueOnce(123n);
    contractStub.read.getEffectiveTargetWithdrawalReserve.mockResolvedValueOnce(456n);
    contractStub.read.getTargetReserveDeficit.mockResolvedValueOnce(789n);
    contractStub.read.isStakingPaused.mockResolvedValueOnce(stakingPaused);
    contractStub.read.isOssificationInitiated.mockResolvedValueOnce(ossificationInitiated);
    contractStub.read.isOssified.mockResolvedValueOnce(ossified);

    const client = createClient();

    await expect(client.getTotalSystemBalance()).resolves.toBe(123n);
    await expect(client.getEffectiveTargetWithdrawalReserve()).resolves.toBe(456n);
    await expect(client.getTargetReserveDeficit()).resolves.toBe(789n);
    await expect(client.L1_MESSAGE_SERVICE()).resolves.toBe(l1MessageServiceAddress);
    await expect(client.isStakingPaused(yieldProvider)).resolves.toBe(stakingPaused);
    await expect(client.isOssificationInitiated(yieldProvider)).resolves.toBe(ossificationInitiated);
    await expect(client.isOssified(yieldProvider)).resolves.toBe(ossified);

    expect(contractStub.read.getTotalSystemBalance).toHaveBeenCalledTimes(1);
    expect(contractStub.read.getEffectiveTargetWithdrawalReserve).toHaveBeenCalledTimes(1);
    expect(contractStub.read.getTargetReserveDeficit).toHaveBeenCalledTimes(1);
    expect(contractStub.read.L1_MESSAGE_SERVICE).toHaveBeenCalledTimes(1);
    expect(contractStub.read.isStakingPaused).toHaveBeenCalledWith([yieldProvider]);
    expect(contractStub.read.isOssificationInitiated).toHaveBeenCalledWith([yieldProvider]);
    expect(contractStub.read.isOssified).toHaveBeenCalledWith([yieldProvider]);
  });

  it("reads withdrawableValue via simulate and returns the result", async () => {
    const withdrawable = 42n;
    contractStub.simulate.withdrawableValue.mockResolvedValueOnce({ result: withdrawable });

    const client = createClient();
    const result = await client.withdrawableValue(yieldProvider);

    expect(contractStub.simulate.withdrawableValue).toHaveBeenCalledWith([yieldProvider]);
    expect(result).toBe(withdrawable);
  });

  it("peeks yield report via simulate and returns mapped YieldReport", async () => {
    const newReportedYield = 100n;
    const outstandingNegativeYield = 25n;
    contractStub.simulate.reportYield.mockResolvedValueOnce({
      result: [newReportedYield, outstandingNegativeYield],
    });

    const client = createClient();
    const result = await client.peekYieldReport(yieldProvider, l2Recipient);

    expect(contractStub.simulate.reportYield).toHaveBeenCalledWith([yieldProvider, l2Recipient], {
      account: signerAddress,
    });
    expect(result).toEqual({
      yieldAmount: newReportedYield,
      outstandingNegativeYield,
      yieldProvider,
    });
  });

  it("returns undefined when peekYieldReport simulation fails", async () => {
    contractStub.simulate.reportYield.mockRejectedValueOnce(new Error("Simulation failed"));

    const client = createClient();
    const result = await client.peekYieldReport(yieldProvider, l2Recipient);

    expect(contractStub.simulate.reportYield).toHaveBeenCalledWith([yieldProvider, l2Recipient], {
      account: signerAddress,
    });
    expect(result).toBeUndefined();
    expect(logger.error).toHaveBeenCalledWith(
      `peekYieldReport failed, yieldProvider=${yieldProvider}, l2YieldRecipient=${l2Recipient}`,
      { error: expect.any(Error) },
    );
  });

  it("encodes calldata and sends fundYieldProvider transactions", async () => {
    const amount = 100n;
    const calldata = "0xcalldata" as Hex;
    const txReceipt = { transactionHash: "0xhash" } as unknown as TransactionReceipt;
    mockedEncodeFunctionData.mockReturnValueOnce(calldata);
    blockchainClient.sendSignedTransaction.mockResolvedValueOnce(txReceipt);

    const client = createClient();
    const receipt = await client.fundYieldProvider(yieldProvider, amount);

    expect(logger.info).toHaveBeenCalledWith(
      `fundYieldProvider started, yieldProvider=${yieldProvider}, amount=${amount.toString()}`,
    );
    expect(mockedEncodeFunctionData).toHaveBeenCalledWith({
      abi: contractStub.abi,
      functionName: "fundYieldProvider",
      args: [yieldProvider, amount],
    });
    expect(blockchainClient.sendSignedTransaction).toHaveBeenCalledWith(
      contractAddress,
      calldata,
      undefined,
      YieldManagerCombinedABI,
    );
    expect(logger.info).toHaveBeenCalledWith(
      `fundYieldProvider succeeded, yieldProvider=${yieldProvider}, amount=${amount.toString()}, txHash=${txReceipt.transactionHash}`,
    );
    expect(receipt).toBe(txReceipt);
  });

  it("reports yield with encoded calldata", async () => {
    const calldata = "0xreport" as Hex;
    const txReceipt = { transactionHash: "0xreporthash" } as unknown as TransactionReceipt;
    mockedEncodeFunctionData.mockReturnValueOnce(calldata);
    blockchainClient.sendSignedTransaction.mockResolvedValueOnce(txReceipt);

    const client = createClient();
    const receipt = await client.reportYield(yieldProvider, l2Recipient);

    expect(logger.info).toHaveBeenCalledWith(
      `reportYield started, yieldProvider=${yieldProvider}, l2YieldRecipient=${l2Recipient}`,
    );
    expect(mockedEncodeFunctionData).toHaveBeenCalledWith({
      abi: contractStub.abi,
      functionName: "reportYield",
      args: [yieldProvider, l2Recipient],
    });
    expect(blockchainClient.sendSignedTransaction).toHaveBeenCalledWith(
      contractAddress,
      calldata,
      undefined,
      YieldManagerCombinedABI,
    );
    expect(logger.info).toHaveBeenCalledWith(
      `reportYield succeeded, yieldProvider=${yieldProvider}, l2YieldRecipient=${l2Recipient}, txHash=${txReceipt.transactionHash}`,
    );
    expect(receipt).toBe(txReceipt);
  });

  it("unstakes with encoded withdrawal params and pays validator fee", async () => {
    const withdrawalParams: WithdrawalRequests = {
      pubkeys: ["0x01"] as Hex[],
      amountsGwei: [32n],
    };
    const encodedWithdrawalParams = "0xencoded" as Hex;
    const calldata = "0xunstake" as Hex;
    const fee = 123n;
    const txReceipt = { transactionHash: "0xunstakehash" } as unknown as TransactionReceipt;

    mockedEncodeAbiParameters.mockReturnValueOnce(encodedWithdrawalParams);
    mockedEncodeFunctionData.mockReturnValueOnce(calldata);
    blockchainClient.sendSignedTransaction.mockResolvedValueOnce(txReceipt);
    (publicClient.readContract as jest.Mock).mockResolvedValueOnce(fee);

    const client = createClient();
    const receipt = await client.unstake(yieldProvider, withdrawalParams);

    expect(logger.info).toHaveBeenCalledWith(
      `unstake started, yieldProvider=${yieldProvider}, validatorCount=${withdrawalParams.pubkeys.length}`,
    );
    expect(logger.debug).toHaveBeenCalledWith(`unstake started withdrawalParams`, {
      withdrawalParams,
    });
    expect(mockedEncodeAbiParameters).toHaveBeenCalledWith(
      [
        { name: "pubkeys", type: "bytes" },
        { name: "amounts", type: "uint64[]" },
        { name: "refundRecipient", type: "address" },
      ],
      [concat(withdrawalParams.pubkeys), withdrawalParams.amountsGwei, contractAddress],
    );
    expect(publicClient.readContract).toHaveBeenCalledWith({
      address: stakingVaultAddress,
      abi: StakingVaultABI,
      functionName: "calculateValidatorWithdrawalFee",
      args: [BigInt(withdrawalParams.pubkeys.length)],
    });
    expect(mockedEncodeFunctionData).toHaveBeenCalledWith({
      abi: contractStub.abi,
      functionName: "unstake",
      args: [yieldProvider, encodedWithdrawalParams],
    });
    expect(blockchainClient.sendSignedTransaction).toHaveBeenCalledWith(
      contractAddress,
      calldata,
      fee,
      YieldManagerCombinedABI,
    );
    expect(logger.info).toHaveBeenCalledWith(
      `unstake succeeded, yieldProvider=${yieldProvider}, validatorCount=${withdrawalParams.pubkeys.length}, txHash=${txReceipt.transactionHash}`,
    );
    expect(logger.debug).toHaveBeenCalledWith(`unstake succeeded withdrawalParams`, {
      withdrawalParams,
    });
    expect(receipt).toBe(txReceipt);
  });

  it("adds to withdrawal reserve and logs success", async () => {
    const amount = 55n;
    const calldata = "0xreserve" as Hex;
    const txReceipt = { transactionHash: "0xreservehash" } as unknown as TransactionReceipt;
    mockedEncodeFunctionData.mockReturnValueOnce(calldata);
    blockchainClient.sendSignedTransaction.mockResolvedValueOnce(txReceipt);

    const client = createClient();
    const receipt = await client.safeAddToWithdrawalReserve(yieldProvider, amount);

    expect(mockedEncodeFunctionData).toHaveBeenCalledWith({
      abi: contractStub.abi,
      functionName: "safeAddToWithdrawalReserve",
      args: [yieldProvider, amount],
    });
    expect(blockchainClient.sendSignedTransaction).toHaveBeenCalledWith(
      contractAddress,
      calldata,
      undefined,
      YieldManagerCombinedABI,
    );
    expect(logger.info).toHaveBeenCalledWith(
      `safeAddToWithdrawalReserve succeeded, yieldProvider=${yieldProvider}, amount=${amount.toString()}, txHash=${txReceipt.transactionHash}`,
    );
    expect(receipt).toBe(txReceipt);
  });

  it("withdraws from yield provider and logs success", async () => {
    const amount = 75n;
    const calldata = "0xwithdraw" as Hex;
    const txReceipt = { transactionHash: "0xwithdrawhash" } as unknown as TransactionReceipt;
    mockedEncodeFunctionData.mockReturnValueOnce(calldata);
    blockchainClient.sendSignedTransaction.mockResolvedValueOnce(txReceipt);

    const client = createClient();
    const receipt = await client.safeWithdrawFromYieldProvider(yieldProvider, amount);

    expect(logger.info).toHaveBeenCalledWith(
      `safeWithdrawFromYieldProvider started, yieldProvider=${yieldProvider}, amount=${amount.toString()}`,
    );
    expect(mockedEncodeFunctionData).toHaveBeenCalledWith({
      abi: contractStub.abi,
      functionName: "safeWithdrawFromYieldProvider",
      args: [yieldProvider, amount],
    });
    expect(blockchainClient.sendSignedTransaction).toHaveBeenCalledWith(
      contractAddress,
      calldata,
      undefined,
      YieldManagerCombinedABI,
    );
    expect(logger.info).toHaveBeenCalledWith(
      `safeWithdrawFromYieldProvider succeeded, yieldProvider=${yieldProvider}, amount=${amount.toString()}, txHash=${txReceipt.transactionHash}`,
    );
    expect(receipt).toBe(txReceipt);
  });

  it("pauses and unpauses staking through encoded calls", async () => {
    const pauseCalldata = "0xpause" as Hex;
    const unpauseCalldata = "0xunpause" as Hex;
    const txReceipt = { transactionHash: "0xhash" } as unknown as TransactionReceipt;

    mockedEncodeFunctionData.mockReturnValueOnce(pauseCalldata).mockReturnValueOnce(unpauseCalldata);
    blockchainClient.sendSignedTransaction.mockResolvedValue(txReceipt);

    const client = createClient();

    await client.pauseStaking(yieldProvider);
    expect(mockedEncodeFunctionData).toHaveBeenNthCalledWith(1, {
      abi: contractStub.abi,
      functionName: "pauseStaking",
      args: [yieldProvider],
    });
    expect(logger.info).toHaveBeenCalledWith(
      `pauseStaking succeeded, yieldProvider=${yieldProvider}, txHash=${txReceipt.transactionHash}`,
    );

    await client.unpauseStaking(yieldProvider);
    expect(mockedEncodeFunctionData).toHaveBeenNthCalledWith(2, {
      abi: contractStub.abi,
      functionName: "unpauseStaking",
      args: [yieldProvider],
    });
    expect(logger.info).toHaveBeenCalledWith(
      `unpauseStaking succeeded, yieldProvider=${yieldProvider}, txHash=${txReceipt.transactionHash}`,
    );
  });

  it("progresses pending ossification", async () => {
    const calldata = "0xprogress" as Hex;
    mockedEncodeFunctionData.mockReturnValueOnce(calldata);

    const client = createClient();
    await client.progressPendingOssification(yieldProvider);

    expect(mockedEncodeFunctionData).toHaveBeenCalledWith({
      abi: contractStub.abi,
      functionName: "progressPendingOssification",
      args: [yieldProvider],
    });
    expect(blockchainClient.sendSignedTransaction).toHaveBeenCalledWith(
      contractAddress,
      calldata,
      undefined,
      YieldManagerCombinedABI,
    );
  });

  // ⚠️ N.B. — WARNING: Below describe block tests were intentionally handwritten because
  // the rebalance logic is too critical (and too fragile) to trust to vibe-coding.
  // Do NOT vibe-code away unless you fully understand the consequences.
  describe("getRebalanceRequirements", () => {
    /**
     * Helper to setup all mocks for a rebalance test scenario.
     *
     * @param {object} opts
     * @param {bigint} opts.l1MessageServiceBalance
     * @param {bigint} opts.totalSystemBalance
     * @param {bigint} opts.effectiveTarget
     * @param {bigint} opts.dashboardTotalValue
     * @param {bigint} opts.userFunds
     * @param {bigint} opts.peekedYieldAmount
     * @param {bigint} opts.peekedOutstandingNegativeYield
     * @param {bigint} opts.rebalanceToleranceAmountWei
     */
    function setupRebalanceTest(
      l1MessageServiceBalance: bigint,
      totalSystemBalance: bigint,
      effectiveTarget: bigint,
      dashboardTotalValue: bigint,
      userFunds: bigint,
      peekedYieldAmount: bigint,
      peekedOutstandingNegativeYield: bigint,
      rebalanceToleranceAmountWei: bigint = 1000000000000000000n, // 1 ETH default
      rebalanceQuotaService?: IRebalanceQuotaService,
      metricsUpdater?: Partial<INativeYieldAutomationMetricsUpdater>,
    ) {
      // --- Contract stubs ---
      contractStub.read.getTotalSystemBalance.mockResolvedValue(totalSystemBalance);
      contractStub.read.getEffectiveTargetWithdrawalReserve.mockResolvedValue(effectiveTarget);
      contractStub.read.userFunds.mockResolvedValue(userFunds);

      contractStub.simulate.reportYield.mockResolvedValue({
        result: [peekedYieldAmount, peekedOutstandingNegativeYield],
      });

      // --- Dashboard stub ---
      const mockDashboardClient = {
        totalValue: jest.fn().mockResolvedValue(dashboardTotalValue),
      };
      mockedDashboardContractClientGetOrCreate.mockReturnValue(mockDashboardClient as any);

      // --- Blockchain client stub ---
      blockchainClient.getBalance.mockResolvedValueOnce(l1MessageServiceBalance);

      // --- Instantiate client ---
      const client = createClient({
        rebalanceToleranceAmountWei,
        rebalanceQuotaService,
        metricsUpdater: metricsUpdater as INativeYieldAutomationMetricsUpdater | undefined,
      });

      return {
        client,
        mockDashboardClient,
        contractStub,
        blockchainClient,
      };
    }

    it("throws error when peekYieldReport returns undefined", async () => {
      contractStub.read.getTotalSystemBalance.mockResolvedValue(1_000_000n);
      contractStub.read.getEffectiveTargetWithdrawalReserve.mockResolvedValue(500_000n);
      contractStub.read.userFunds.mockResolvedValue(100_000n);
      contractStub.simulate.reportYield.mockResolvedValue(undefined);

      const mockDashboardClient = {
        totalValue: jest.fn().mockResolvedValue(600_000n),
      };
      mockedDashboardContractClientGetOrCreate.mockReturnValue(mockDashboardClient as any);

      const client = createClient();

      await expect(client.getRebalanceRequirements(yieldProvider, l2Recipient)).rejects.toThrow(
        "peekYieldReport returned undefined, cannot determine rebalance requirements",
      );
    });

    it("returns NONE when absRebalanceRequirement is within tolerance band - case 1", async () => {
      // Arrange - Inputs
      const l1MessageServiceBalance = 490_001n; // Just barely within tolerance band
      const totalSystemBalance = 1_000_000n;
      const effectiveTarget = 500_000n;
      const dashboardTotalValue = 500_000n;
      const userFunds = 500_000n;
      const peekedYieldAmount = 0n;
      const peekedOutstandingNegativeYield = 0n;
      const rebalanceToleranceAmountWei = 10_000n; // 100 bps of 1M = 10K wei
      const { client } = setupRebalanceTest(
        l1MessageServiceBalance,
        totalSystemBalance,
        effectiveTarget,
        dashboardTotalValue,
        userFunds,
        peekedYieldAmount,
        peekedOutstandingNegativeYield,
        rebalanceToleranceAmountWei,
      );
      // Expect
      await expect(client.getRebalanceRequirements(yieldProvider, l2Recipient)).resolves.toEqual({
        rebalanceDirection: RebalanceDirection.NONE,
        rebalanceAmount: 0n,
      });
    });

    it("returns NONE when absRebalanceRequirement is within tolerance band - case 2", async () => {
      // Arrange - Inputs
      const l1MessageServiceBalance = 509_999n; // Just barely within tolerance band
      const totalSystemBalance = 1_000_000n;
      const effectiveTarget = 500_000n;
      const dashboardTotalValue = 500_000n;
      const userFunds = 500_000n;
      const peekedYieldAmount = 0n;
      const peekedOutstandingNegativeYield = 0n;
      const rebalanceToleranceAmountWei = 10_000n; // 100 bps of 1M = 10K wei
      const { client } = setupRebalanceTest(
        l1MessageServiceBalance,
        totalSystemBalance,
        effectiveTarget,
        dashboardTotalValue,
        userFunds,
        peekedYieldAmount,
        peekedOutstandingNegativeYield,
        rebalanceToleranceAmountWei,
      );
      // Expect
      await expect(client.getRebalanceRequirements(yieldProvider, l2Recipient)).resolves.toEqual({
        rebalanceDirection: RebalanceDirection.NONE,
        rebalanceAmount: 0n,
      });
    });

    it("correctly evaluates rebalance requirements for negative yield scenario - case 1", async () => {
      // Arrange - 100K of system obligations in negative yield scenario
      const l1MessageServiceBalance = 500_000n;
      const totalSystemBalance = 1_000_000n;
      const effectiveTarget = 500_000n;
      const dashboardTotalValue = 500_000n;
      const userFunds = 500_000n;
      const peekedYieldAmount = 0n;
      const peekedOutstandingNegativeYield = 100_000n;
      const rebalanceToleranceAmountWei = 10_000n; // 100 bps of 1M = 10K wei
      const { client } = setupRebalanceTest(
        l1MessageServiceBalance,
        totalSystemBalance,
        effectiveTarget,
        dashboardTotalValue,
        userFunds,
        peekedYieldAmount,
        peekedOutstandingNegativeYield,
        rebalanceToleranceAmountWei,
      );
      // Expectations
      // adjustedSystemTotalBalance = 900K
      // adjustedEffectiveTarget = 450K
      // rebalanceRequirement = 50K
      // systemObligation = 100K
      await expect(client.getRebalanceRequirements(yieldProvider, l2Recipient)).resolves.toEqual({
        rebalanceDirection: RebalanceDirection.STAKE,
        rebalanceAmount: 150_000n,
      });
    });

    it("correctly evaluates rebalance requirements for negative yield scenario - case 2", async () => {
      // Arrange - 100K of system obligations in negative yield scenario
      const l1MessageServiceBalance = 450_000n;
      const totalSystemBalance = 1_000_000n;
      const effectiveTarget = 500_000n;
      const dashboardTotalValue = 500_000n;
      const userFunds = 500_000n;
      const peekedYieldAmount = 0n;
      const peekedOutstandingNegativeYield = 100_000n;
      const rebalanceToleranceAmountWei = 10_000n; // 100 bps of 1M = 10K wei
      const { client } = setupRebalanceTest(
        l1MessageServiceBalance,
        totalSystemBalance,
        effectiveTarget,
        dashboardTotalValue,
        userFunds,
        peekedYieldAmount,
        peekedOutstandingNegativeYield,
        rebalanceToleranceAmountWei,
      );
      // Expectations
      // adjustedSystemTotalBalance = 900K
      // adjustedEffectiveTarget = 450K
      // systemObligation = 100K
      await expect(client.getRebalanceRequirements(yieldProvider, l2Recipient)).resolves.toEqual({
        rebalanceDirection: RebalanceDirection.STAKE,
        rebalanceAmount: 100_000n,
      });
    });

    it("correctly evaluates rebalance requirements for positive yield scenario - case 1", async () => {
      // Setup positive yield with 50_000 system obligations
      const totalSystemBalance = 1_000_000n;
      const effectiveTarget = 200_000n;
      const dashboardTotalValue = 600_000n;
      const userFunds = 500_000n;
      const peekedYieldAmount = 50_000n;
      const peekedOutstandingNegativeYield = 0n;
      const rebalanceToleranceAmountWei = 10_000n; // 100 bps of 1M = 10K wei
      const l1MessageServiceBalance = 50_000n;
      const { client } = setupRebalanceTest(
        l1MessageServiceBalance,
        totalSystemBalance,
        effectiveTarget,
        dashboardTotalValue,
        userFunds,
        peekedYieldAmount,
        peekedOutstandingNegativeYield,
        rebalanceToleranceAmountWei,
      );
      // Expectations
      // systemObligation = 50K
      // adjustedSystemTotalBalance = 950K
      // adjustedEffectiveTarget = 95% of 200_000 = 190K
      await expect(client.getRebalanceRequirements(yieldProvider, l2Recipient)).resolves.toEqual({
        rebalanceDirection: RebalanceDirection.UNSTAKE,
        rebalanceAmount: 190_000n,
      });
    });

    it("correctly evaluates rebalance requirements for positive yield scenario - case 2", async () => {
      // Arrange - whale has depleted the reserve and racked up large LST liability
      const totalSystemBalance = 10_000_000n;
      const effectiveTarget = 4_000_000n;
      const dashboardTotalValue = 10_000_000n;
      const userFunds = 10_000_000n;
      const peekedYieldAmount = 0n;
      const peekedOutstandingNegativeYield = 5_000_000n; // Large LST liability
      const rebalanceToleranceAmountWei = 250_000n; // 500 bps of 5M (adjusted balance) = 250K wei
      const l1MessageServiceBalance = 0n;
      const { client } = setupRebalanceTest(
        l1MessageServiceBalance,
        totalSystemBalance,
        effectiveTarget,
        dashboardTotalValue,
        userFunds,
        peekedYieldAmount,
        peekedOutstandingNegativeYield,
        rebalanceToleranceAmountWei,
      );
      // Expectations
      // systemObligation = 5M
      // adjustedSystemTotalBalance = 5M
      // adjustedEffectiveTarget = 2M
      await expect(client.getRebalanceRequirements(yieldProvider, l2Recipient)).resolves.toEqual({
        rebalanceDirection: RebalanceDirection.UNSTAKE,
        rebalanceAmount: 7_000_000n,
      });
    });

    it("caps staking rebalance amount when quota service returns partial amount", async () => {
      const totalSystemBalance = 1_000_000n;
      const effectiveTarget = 400_000n;
      const dashboardTotalValue = 100_000n;
      const userFunds = 100_000n;
      const peekedYieldAmount = 0n;
      const peekedOutstandingNegativeYield = 0n;
      const rebalanceToleranceAmountWei = 10_000n; // 100 bps of 1M = 10K wei
      const l1MessageServiceBalance = 900_000n;
      const absRebalanceRequirement = l1MessageServiceBalance - effectiveTarget; // 500_000n
      const cappedAmount = 100_000n;
      const mockQuotaService = createDefaultQuotaService();
      (mockQuotaService.getRebalanceAmountAfterQuota as jest.Mock).mockReturnValue(cappedAmount);
      const { client } = setupRebalanceTest(
        l1MessageServiceBalance,
        totalSystemBalance,
        effectiveTarget,
        dashboardTotalValue,
        userFunds,
        peekedYieldAmount,
        peekedOutstandingNegativeYield,
        rebalanceToleranceAmountWei,
        mockQuotaService,
      );
      // Expectations
      await expect(client.getRebalanceRequirements(yieldProvider, l2Recipient)).resolves.toEqual({
        rebalanceDirection: RebalanceDirection.STAKE,
        rebalanceAmount: cappedAmount,
      });
      expect(mockQuotaService.getRebalanceAmountAfterQuota).toHaveBeenCalledWith(
        stakingVaultAddress,
        totalSystemBalance,
        absRebalanceRequirement,
      );
    });

    it("correctly applies quota when quota service returns 0n (quota exceeded)", async () => {
      const totalSystemBalance = 1_000_000n;
      const effectiveTarget = 400_000n;
      const dashboardTotalValue = 100_000n;
      const userFunds = 100_000n;
      const peekedYieldAmount = 0n;
      const peekedOutstandingNegativeYield = 0n;
      const rebalanceToleranceAmountWei = 10_000n; // 100 bps of 1M = 10K wei
      const l1MessageServiceBalance = 900_000n;
      const absRebalanceRequirement = l1MessageServiceBalance - effectiveTarget; // 500_000n
      const mockQuotaService = createDefaultQuotaService();
      (mockQuotaService.getRebalanceAmountAfterQuota as jest.Mock).mockReturnValue(0n);
      const { client } = setupRebalanceTest(
        l1MessageServiceBalance,
        totalSystemBalance,
        effectiveTarget,
        dashboardTotalValue,
        userFunds,
        peekedYieldAmount,
        peekedOutstandingNegativeYield,
        rebalanceToleranceAmountWei,
        mockQuotaService,
      );
      // Expectations
      await expect(client.getRebalanceRequirements(yieldProvider, l2Recipient)).resolves.toEqual({
        rebalanceDirection: RebalanceDirection.NONE,
        rebalanceAmount: 0n,
      });
      expect(mockQuotaService.getRebalanceAmountAfterQuota).toHaveBeenCalledWith(
        stakingVaultAddress,
        totalSystemBalance,
        absRebalanceRequirement,
      );
    });

    it("caps staking amount when absRebalanceAmountAfterQuota exceeds stakingRebalanceCeiling", async () => {
      // Arrange - absRebalanceAmountAfterQuota > stakingRebalanceCeiling
      const totalSystemBalance = 1_000_000n;
      const effectiveTarget = 400_000n;
      const dashboardTotalValue = 100_000n;
      const userFunds = 100_000n;
      const peekedYieldAmount = 0n;
      const peekedOutstandingNegativeYield = 0n;
      const rebalanceToleranceAmountWei = 10_000n;
      const l1MessageServiceBalance = 1_000_000n; // Above target, STAKE direction
      // stakingRebalanceCeiling = l1MessageServiceBalance - effectiveTarget = 1_000_000n - 400_000n = 600_000n
      const absRebalanceAmountAfterQuota = 800_000n; // Exceeds ceiling
      const mockQuotaService = createDefaultQuotaService();
      (mockQuotaService.getRebalanceAmountAfterQuota as jest.Mock).mockReturnValue(absRebalanceAmountAfterQuota);
      const { client } = setupRebalanceTest(
        l1MessageServiceBalance,
        totalSystemBalance,
        effectiveTarget,
        dashboardTotalValue,
        userFunds,
        peekedYieldAmount,
        peekedOutstandingNegativeYield,
        rebalanceToleranceAmountWei,
        mockQuotaService,
      );
      // Act & Assert
      await expect(client.getRebalanceRequirements(yieldProvider, l2Recipient)).resolves.toEqual({
        rebalanceDirection: RebalanceDirection.STAKE,
        rebalanceAmount: 600_000n, // Capped to stakingRebalanceCeiling
      });
    });

    it("uses full absRebalanceAmountAfterQuota when below stakingRebalanceCeiling", async () => {
      // Arrange - absRebalanceAmountAfterQuota <= stakingRebalanceCeiling
      const totalSystemBalance = 1_000_000n;
      const effectiveTarget = 400_000n;
      const dashboardTotalValue = 100_000n;
      const userFunds = 100_000n;
      const peekedYieldAmount = 0n;
      const peekedOutstandingNegativeYield = 0n;
      const rebalanceToleranceAmountWei = 10_000n;
      const l1MessageServiceBalance = 1_000_000n;
      // stakingRebalanceCeiling = 600_000n
      const absRebalanceAmountAfterQuota = 300_000n; // Below ceiling
      const mockQuotaService = createDefaultQuotaService();
      (mockQuotaService.getRebalanceAmountAfterQuota as jest.Mock).mockReturnValue(absRebalanceAmountAfterQuota);
      const { client } = setupRebalanceTest(
        l1MessageServiceBalance,
        totalSystemBalance,
        effectiveTarget,
        dashboardTotalValue,
        userFunds,
        peekedYieldAmount,
        peekedOutstandingNegativeYield,
        rebalanceToleranceAmountWei,
        mockQuotaService,
      );
      // Act & Assert
      await expect(client.getRebalanceRequirements(yieldProvider, l2Recipient)).resolves.toEqual({
        rebalanceDirection: RebalanceDirection.STAKE,
        rebalanceAmount: 300_000n, // Not capped, uses full amount
      });
    });

    it("handles edge case where stakingRebalanceCeiling equals absRebalanceAmountAfterQuota", async () => {
      // Arrange - Boundary condition
      const totalSystemBalance = 1_000_000n;
      const effectiveTarget = 400_000n;
      const dashboardTotalValue = 100_000n;
      const userFunds = 100_000n;
      const peekedYieldAmount = 0n;
      const peekedOutstandingNegativeYield = 0n;
      const rebalanceToleranceAmountWei = 10_000n;
      const l1MessageServiceBalance = 1_000_000n;
      // stakingRebalanceCeiling = 600_000n
      const absRebalanceAmountAfterQuota = 600_000n; // Exactly equals ceiling
      const mockQuotaService = createDefaultQuotaService();
      (mockQuotaService.getRebalanceAmountAfterQuota as jest.Mock).mockReturnValue(absRebalanceAmountAfterQuota);
      const { client } = setupRebalanceTest(
        l1MessageServiceBalance,
        totalSystemBalance,
        effectiveTarget,
        dashboardTotalValue,
        userFunds,
        peekedYieldAmount,
        peekedOutstandingNegativeYield,
        rebalanceToleranceAmountWei,
        mockQuotaService,
      );
      // Act & Assert
      await expect(client.getRebalanceRequirements(yieldProvider, l2Recipient)).resolves.toEqual({
        rebalanceDirection: RebalanceDirection.STAKE,
        rebalanceAmount: 600_000n, // Capped to ceiling (equal values)
      });
    });

    it("handles small stakingRebalanceCeiling edge case", async () => {
      // Arrange - When balance is just above target reserve
      const totalSystemBalance = 1_000_000n;
      const effectiveTarget = 400_000n;
      const dashboardTotalValue = 100_000n;
      const userFunds = 100_000n;
      const peekedYieldAmount = 0n;
      const peekedOutstandingNegativeYield = 0n;
      const rebalanceToleranceAmountWei = 10_000n;
      const l1MessageServiceBalance = 400_100n; // Just above target
      // stakingRebalanceCeiling = 400_100n - 400_000n = 100n (very small)
      const absRebalanceAmountAfterQuota = 50_000n; // Much larger than ceiling
      const mockQuotaService = createDefaultQuotaService();
      (mockQuotaService.getRebalanceAmountAfterQuota as jest.Mock).mockReturnValue(absRebalanceAmountAfterQuota);
      const { client } = setupRebalanceTest(
        l1MessageServiceBalance,
        totalSystemBalance,
        effectiveTarget,
        dashboardTotalValue,
        userFunds,
        peekedYieldAmount,
        peekedOutstandingNegativeYield,
        rebalanceToleranceAmountWei,
        mockQuotaService,
      );
      // Act & Assert
      await expect(client.getRebalanceRequirements(yieldProvider, l2Recipient)).resolves.toEqual({
        rebalanceDirection: RebalanceDirection.STAKE,
        rebalanceAmount: 100n, // Capped to small ceiling
      });
    });

    it("tracks metrics when metricsUpdater is provided and quota service returns 0n", async () => {
      const metricsUpdater = {
        incrementStakingDepositQuotaExceeded: jest.fn(),
        setActualRebalanceRequirement: jest.fn(),
        setReportedRebalanceRequirement: jest.fn(),
        incrementContractEstimateGasError: jest.fn(),
      };
      const totalSystemBalance = 1_000_000n;
      const effectiveTarget = 400_000n;
      const dashboardTotalValue = 100_000n;
      const userFunds = 100_000n;
      const peekedYieldAmount = 0n;
      const peekedOutstandingNegativeYield = 0n;
      const rebalanceToleranceAmountWei = 10_000n; // 100 bps of 1M = 10K wei
      const l1MessageServiceBalance = 900_000n;
      const absRebalanceRequirement = l1MessageServiceBalance - effectiveTarget; // 500_000n
      const mockQuotaService = createDefaultQuotaService();
      (mockQuotaService.getRebalanceAmountAfterQuota as jest.Mock).mockReturnValue(0n);
      const { client } = setupRebalanceTest(
        l1MessageServiceBalance,
        totalSystemBalance,
        effectiveTarget,
        dashboardTotalValue,
        userFunds,
        peekedYieldAmount,
        peekedOutstandingNegativeYield,
        rebalanceToleranceAmountWei,
        mockQuotaService,
        metricsUpdater,
      );

      await client.getRebalanceRequirements(yieldProvider, l2Recipient);

      // Actual requirement should be set with original requirement (converted to gwei)
      // Direction is based on l1MessageServiceBalance vs effectiveTarget (not tolerance band)
      expect(metricsUpdater.setActualRebalanceRequirement).toHaveBeenCalledWith(
        stakingVaultAddress,
        weiToGweiNumber(absRebalanceRequirement),
        RebalanceDirection.STAKE,
      );
      // Reported requirement should be 0 when quota service returns 0n
      expect(metricsUpdater.setReportedRebalanceRequirement).toHaveBeenCalledWith(stakingVaultAddress, 0, RebalanceDirection.NONE);
      // Note: incrementStakingDepositQuotaExceeded is called inside RebalanceQuotaService, not in YieldManagerContractClient.
      // Testing that the quota service calls the metrics updater is done in RebalanceQuotaService tests, not here.
    });

    it("tracks gauge metric for all rebalance paths when metricsUpdater is provided", async () => {
      const metricsUpdater = {
        incrementStakingDepositQuotaExceeded: jest.fn(),
        setActualRebalanceRequirement: jest.fn(),
        setReportedRebalanceRequirement: jest.fn(),
        incrementContractEstimateGasError: jest.fn(),
      };
      const totalSystemBalance = 1_000_000n;
      const effectiveTarget = 400_000n;
      const dashboardTotalValue = 100_000n;
      const userFunds = 100_000n;
      const peekedYieldAmount = 0n;
      const peekedOutstandingNegativeYield = 0n;
      const rebalanceToleranceAmountWei = 10_000n; // 100 bps of 1M = 10K wei
      const l1MessageServiceBalance = 450_000n; // Within tolerance, should return NONE
      const absRebalanceRequirement = absDiff(l1MessageServiceBalance, effectiveTarget); // 50_000n
      const { client } = setupRebalanceTest(
        l1MessageServiceBalance,
        totalSystemBalance,
        effectiveTarget,
        dashboardTotalValue,
        userFunds,
        peekedYieldAmount,
        peekedOutstandingNegativeYield,
        rebalanceToleranceAmountWei,
        undefined, // Use default quota service (returns full amount)
        metricsUpdater,
      );

      await client.getRebalanceRequirements(yieldProvider, l2Recipient);

      // Actual requirement should be set even when within tolerance band (NONE path, converted to gwei)
      // Direction is based on l1MessageServiceBalance vs effectiveTarget (not tolerance band)
      expect(metricsUpdater.setActualRebalanceRequirement).toHaveBeenCalledWith(
        stakingVaultAddress,
        weiToGweiNumber(absRebalanceRequirement),
        RebalanceDirection.STAKE,
      );
      // Reported requirement: absRebalanceRequirement (50_000) is not < toleranceBand (10_000), so result is STAKING, not NONE
      expect(metricsUpdater.setReportedRebalanceRequirement).toHaveBeenCalledWith(stakingVaultAddress, 0, RebalanceDirection.STAKE);
      // Counter should not be incremented when quota service returns full amount (within quota)
      expect(metricsUpdater.incrementStakingDepositQuotaExceeded).not.toHaveBeenCalled();
    });

    it("tracks reported requirement for STAKE path when metricsUpdater is provided", async () => {
      const metricsUpdater = {
        incrementStakingDepositQuotaExceeded: jest.fn(),
        setActualRebalanceRequirement: jest.fn(),
        setReportedRebalanceRequirement: jest.fn(),
        incrementContractEstimateGasError: jest.fn(),
      };
      const totalSystemBalance = 1_000_000n;
      const effectiveTarget = 400_000n;
      const dashboardTotalValue = 100_000n;
      const userFunds = 100_000n;
      const peekedYieldAmount = 0n;
      const peekedOutstandingNegativeYield = 0n;
      const rebalanceToleranceAmountWei = 10_000n; // 100 bps of 1M = 10K wei
      const l1MessageServiceBalance = 300_000n; // Below target, goes to UNSTAKE path (logic: if balance < target, UNSTAKE)
      const absRebalanceRequirement = effectiveTarget - l1MessageServiceBalance; // 100_000n
      const { client } = setupRebalanceTest(
        l1MessageServiceBalance,
        totalSystemBalance,
        effectiveTarget,
        dashboardTotalValue,
        userFunds,
        peekedYieldAmount,
        peekedOutstandingNegativeYield,
        rebalanceToleranceAmountWei,
        undefined, // Use default quota service (quota service not called for UNSTAKE direction)
        metricsUpdater,
      );

      await client.getRebalanceRequirements(yieldProvider, l2Recipient);

      // Actual requirement should be set with original requirement
      // Note: When l1MessageServiceBalance < effectiveTarget, direction is UNSTAKE
      expect(metricsUpdater.setActualRebalanceRequirement).toHaveBeenCalledWith(
        stakingVaultAddress,
        weiToGweiNumber(absRebalanceRequirement),
        RebalanceDirection.UNSTAKE,
      );
      // Reported requirement should be the full amount (quota service not called for UNSTAKE direction)
      expect(metricsUpdater.setReportedRebalanceRequirement).toHaveBeenCalledWith(
        stakingVaultAddress,
        weiToGweiNumber(absRebalanceRequirement),
        RebalanceDirection.UNSTAKE,
      );
    });

    it("tracks reported requirement for UNSTAKE path when metricsUpdater is provided", async () => {
      const metricsUpdater = {
        incrementStakingDepositQuotaExceeded: jest.fn(),
        setActualRebalanceRequirement: jest.fn(),
        setReportedRebalanceRequirement: jest.fn(),
        incrementContractEstimateGasError: jest.fn(),
      };
      const totalSystemBalance = 1_000_000n;
      const effectiveTarget = 400_000n;
      const dashboardTotalValue = 100_000n;
      const userFunds = 100_000n;
      const peekedYieldAmount = 0n;
      const peekedOutstandingNegativeYield = 0n;
      const rebalanceToleranceAmountWei = 10_000n; // 100 bps of 1M = 10K wei
      const l1MessageServiceBalance = 500_000n; // Above target, needs STAKE
      const absRebalanceRequirement = l1MessageServiceBalance - effectiveTarget; // 100_000n
      const { client } = setupRebalanceTest(
        l1MessageServiceBalance,
        totalSystemBalance,
        effectiveTarget,
        dashboardTotalValue,
        userFunds,
        peekedYieldAmount,
        peekedOutstandingNegativeYield,
        rebalanceToleranceAmountWei,
        undefined, // Use default quota service (returns full amount)
        metricsUpdater,
      );

      await client.getRebalanceRequirements(yieldProvider, l2Recipient);

      // Actual requirement should be set with original requirement
      // Note: When l1MessageServiceBalance > effectiveTargetWithdrawalReserveExcludingObligations, direction is STAKE
      expect(metricsUpdater.setActualRebalanceRequirement).toHaveBeenCalledWith(
        stakingVaultAddress,
        weiToGweiNumber(absRebalanceRequirement),
        RebalanceDirection.STAKE,
      );
      // Reported requirement: result direction is STAKE (l1MessageServiceBalance > effectiveTargetWithdrawalReserveExcludingObligations)
      expect(metricsUpdater.setReportedRebalanceRequirement).toHaveBeenCalledWith(
        stakingVaultAddress,
        weiToGweiNumber(absRebalanceRequirement),
        RebalanceDirection.STAKE,
      );
    });

    it("handles zero totalSystemBalance without division by zero", async () => {
      const totalSystemBalance = 0n;
      const effectiveTarget = 0n;
      const dashboardTotalValue = 0n;
      const userFunds = 0n;
      const peekedYieldAmount = 0n;
      const peekedOutstandingNegativeYield = 0n;
      const rebalanceToleranceAmountWei = 10_000n; // 100 bps of 1M = 10K wei
      const l1MessageServiceBalance = 0n;
      const { client } = setupRebalanceTest(
        l1MessageServiceBalance,
        totalSystemBalance,
        effectiveTarget,
        dashboardTotalValue,
        userFunds,
        peekedYieldAmount,
        peekedOutstandingNegativeYield,
        rebalanceToleranceAmountWei,
      );
      // Expectations
      await expect(client.getRebalanceRequirements(yieldProvider, l2Recipient)).resolves.toEqual({
        rebalanceDirection: RebalanceDirection.NONE,
        rebalanceAmount: 0n,
      });
    });
  });

  it("returns staking vault and dashboard addresses from yield provider data", async () => {
    contractStub.read.getYieldProviderData.mockResolvedValueOnce({
      ...defaultYieldProviderData,
      primaryEntrypoint: l2Recipient,
      ossifiedEntrypoint: stakingVaultAddress,
    });

    const client = createClient();

    await expect(client.getLidoStakingVaultAddress(yieldProvider)).resolves.toBe(stakingVaultAddress);
    await expect(client.getLidoDashboardAddress(yieldProvider)).resolves.toBe(l2Recipient);
  });

  it("pauses staking only when not already paused", async () => {
    const txReceipt = { transactionHash: "0xpause" } as unknown as TransactionReceipt;
    const client = createClient();
    const pauseSpy = jest.spyOn(client, "pauseStaking").mockResolvedValue(txReceipt);

    contractStub.read.isStakingPaused.mockResolvedValueOnce(false).mockResolvedValueOnce(true);

    await expect(client.pauseStakingIfNotAlready(yieldProvider)).resolves.toBe(txReceipt);
    await expect(client.pauseStakingIfNotAlready(yieldProvider)).resolves.toBeUndefined();
    expect(logger.info).toHaveBeenCalledWith(`Already paused staking for yieldProvider=${yieldProvider}`);
    expect(pauseSpy).toHaveBeenCalledTimes(1);
  });

  it("unpauses staking only when currently paused", async () => {
    const txReceipt = { transactionHash: "0xunpause" } as unknown as TransactionReceipt;

    const client = createClient();
    const unpauseSpy = jest.spyOn(client, "unpauseStaking").mockResolvedValue(txReceipt);

    contractStub.read.isStakingPaused.mockResolvedValueOnce(true).mockResolvedValueOnce(false);

    await expect(client.unpauseStakingIfNotAlready(yieldProvider)).resolves.toBe(txReceipt);
    expect(unpauseSpy).toHaveBeenCalledWith(yieldProvider);

    await expect(client.unpauseStakingIfNotAlready(yieldProvider)).resolves.toBeUndefined();
    expect(logger.info).toHaveBeenCalledWith(`Already resumed staking for yieldProvider=${yieldProvider}`);
    expect(unpauseSpy).toHaveBeenCalledTimes(1);
  });

  it("computes available unstaking rebalance balance", async () => {
    blockchainClient.getBalance.mockResolvedValueOnce(1_000n);
    contractStub.simulate.withdrawableValue.mockResolvedValueOnce({ result: 500n });

    const client = createClient();
    await expect(client.getAvailableUnstakingRebalanceBalance(yieldProvider)).resolves.toBe(1_500n);

    expect(blockchainClient.getBalance).toHaveBeenCalledWith(contractAddress);
    expect(contractStub.simulate.withdrawableValue).toHaveBeenCalledWith([yieldProvider]);
  });

  it("adds to withdrawal reserve only when above threshold", async () => {
    const client = createClient({ minWithdrawalThresholdEth: 1n });
    const addSpy = jest.spyOn(client, "safeAddToWithdrawalReserve").mockResolvedValue(undefined as any);
    const belowThresholdBalance = ONE_ETHER - 1n;
    const minThreshold = 1n * ONE_ETHER;
    jest
      .spyOn(client, "getAvailableUnstakingRebalanceBalance")
      .mockResolvedValueOnce(belowThresholdBalance)
      .mockResolvedValueOnce(ONE_ETHER + 100n);

    await expect(client.safeAddToWithdrawalReserveIfAboveThreshold(yieldProvider, 5n)).resolves.toBeUndefined();
    expect(addSpy).not.toHaveBeenCalled();
    expect(logger.info).toHaveBeenCalledWith(
      `safeAddToWithdrawalReserveIfAboveThreshold - skipping as availableWithdrawalBalance=${belowThresholdBalance} is below the minimum withdrawal threshold of ${minThreshold}`,
    );

    await client.safeAddToWithdrawalReserveIfAboveThreshold(yieldProvider, 7n);
    expect(addSpy).toHaveBeenCalledWith(yieldProvider, 7n);
  });

  it("adds the full available balance when calling safeMaxAddToWithdrawalReserve", async () => {
    const client = createClient({ minWithdrawalThresholdEth: 1n });
    const addSpy = jest.spyOn(client, "safeAddToWithdrawalReserve").mockResolvedValue(undefined as any);
    const available = ONE_ETHER + 50n;
    jest.spyOn(client, "getAvailableUnstakingRebalanceBalance").mockResolvedValue(available);

    await client.safeMaxAddToWithdrawalReserve(yieldProvider);

    expect(addSpy).toHaveBeenCalledWith(yieldProvider, available);
  });

  it("skips safeMaxAddToWithdrawalReserve when below the threshold", async () => {
    const client = createClient({ minWithdrawalThresholdEth: 2n });
    const addSpy = jest.spyOn(client, "safeAddToWithdrawalReserve").mockResolvedValue(undefined as any);
    const belowThresholdBalance = 2n * ONE_ETHER - 1n;
    const minThreshold = 2n * ONE_ETHER;
    jest.spyOn(client, "getAvailableUnstakingRebalanceBalance").mockResolvedValue(belowThresholdBalance);

    await expect(client.safeMaxAddToWithdrawalReserve(yieldProvider)).resolves.toBeUndefined();
    expect(addSpy).not.toHaveBeenCalled();
    expect(logger.info).toHaveBeenCalledWith(
      `safeMaxAddToWithdrawalReserve - skipping as availableWithdrawalBalance=${belowThresholdBalance} is below the minimum withdrawal threshold of ${minThreshold}`,
    );
  });

  it("extracts withdrawal events from receipts emitted by the contract", () => {
    const client = createClient();
    const log = { address: contractAddress, data: "0xdata", topics: ["0x01"] };
    mockedParseEventLogs.mockReturnValueOnce([
      {
        eventName: "WithdrawalReserveAugmented",
        args: { reserveIncrementAmount: 10n, yieldProvider },
        address: contractAddress,
      } as any,
    ]);

    const event = client.getWithdrawalEventFromTxReceipt(buildReceipt([log]));

    expect(event).toEqual({ reserveIncrementAmount: 10n, yieldProvider });
    expect(mockedParseEventLogs).toHaveBeenCalledWith({
      abi: contractStub.abi,
      eventName: "WithdrawalReserveAugmented",
      logs: [log],
    });
  });

  it("returns undefined when withdrawal events are absent or decoding fails", () => {
    const client = createClient();
    mockedParseEventLogs.mockReturnValueOnce([]);

    const event = client.getWithdrawalEventFromTxReceipt(
      buildReceipt([{ address: contractAddress.toUpperCase(), data: "0x", topics: [] }]),
    );

    expect(event).toBeUndefined();
    expect(logger.debug).toHaveBeenCalledWith(
      "getWithdrawalEventFromTxReceipt - WithdrawalReserveAugmented event not found in receipt",
    );
    expect(mockedParseEventLogs).toHaveBeenCalledTimes(1);
  });

  it("ignores withdrawal events from other contracts", () => {
    const client = createClient();
    const foreignLog = { address: "0x1234567890123456789012345678901234567890", data: "0x", topics: [] };

    mockedParseEventLogs.mockReturnValueOnce([
      {
        eventName: "WithdrawalReserveAugmented",
        args: { reserveIncrementAmount: 10n, yieldProvider },
        address: "0x1234567890123456789012345678901234567890",
      } as any,
    ]);

    const event = client.getWithdrawalEventFromTxReceipt(buildReceipt([foreignLog]));

    expect(event).toBeUndefined();
    expect(mockedParseEventLogs).toHaveBeenCalledTimes(1);
  });

  it("extracts yield reports from receipts emitted by the contract", () => {
    const client = createClient();
    const log = { address: contractAddress, data: "0xfeed", topics: ["0x1111"] };
    mockedParseEventLogs.mockReturnValueOnce([
      {
        eventName: "NativeYieldReported",
        args: { yieldAmount: 12n, outstandingNegativeYield: 5n, yieldProvider },
        address: contractAddress,
      } as any,
    ]);

    const report = client.getYieldReportFromTxReceipt(buildReceipt([log]));

    expect(report).toEqual({ yieldAmount: 12n, outstandingNegativeYield: 5n, yieldProvider });
    expect(mockedParseEventLogs).toHaveBeenCalledWith({
      abi: contractStub.abi,
      eventName: "NativeYieldReported",
      logs: [log],
    });
  });

  it("returns undefined when yield report events are absent", () => {
    const client = createClient();
    mockedParseEventLogs.mockReturnValueOnce([]);

    const report = client.getYieldReportFromTxReceipt(
      buildReceipt([{ address: contractAddress, data: "0x0", topics: [] }]),
    );

    expect(report).toBeUndefined();
    expect(logger.debug).toHaveBeenCalledWith(
      "getYieldReportFromTxReceipt - NativeYieldReported event not found in receipt",
    );
  });

  it("ignores yield report logs from other contracts", () => {
    const client = createClient();
    const foreignLog = { address: "0x1234567890123456789012345678901234567890", data: "0x", topics: [] };

    mockedParseEventLogs.mockReturnValueOnce([
      {
        eventName: "NativeYieldReported",
        args: { yieldAmount: 12n, outstandingNegativeYield: 5n, yieldProvider },
        address: "0x1234567890123456789012345678901234567890",
      } as any,
    ]);

    const report = client.getYieldReportFromTxReceipt(buildReceipt([foreignLog]));

    expect(report).toBeUndefined();
    expect(mockedParseEventLogs).toHaveBeenCalledTimes(1);
  });
});
