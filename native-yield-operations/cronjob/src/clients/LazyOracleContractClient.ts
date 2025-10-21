import { IContractClientLibrary } from "ts-libs/linea-shared-utils/src/core/client/IContractClientLibrary";
import {
  Address,
  encodeFunctionData,
  getContract,
  GetContractReturnType,
  Hex,
  parseSignature,
  PublicClient,
  serializeTransaction,
  TransactionReceipt,
  TransactionSerializableEIP1559,
} from "viem";
import { LazyOracleABI } from "../core/abis/LazyOracle";
import { ILazyOracle } from "../core/services/contracts/ILazyOracle";
import { LazyOracleReportData } from "../core/entities";
import { IContractSignerService } from "ts-libs/linea-shared-utils/src/core/services/IContractSignerService";

export class LazyOracleContractClient implements ILazyOracle<TransactionReceipt> {
  private readonly contract: GetContractReturnType<typeof LazyOracleABI, PublicClient, Address>;
  constructor(
    private readonly contractClientLibrary: IContractClientLibrary<PublicClient, TransactionReceipt>,
    private readonly contractAddress: Address,
    private readonly contractSignerService: IContractSignerService,
  ) {
    this.contract = getContract({
      abi: LazyOracleABI,
      address: contractAddress,
      client: contractClientLibrary.getBlockchainClient(),
    });
  }

  async latestReportData(): Promise<LazyOracleReportData> {
    const resp = await this.contract.read.latestReportData();
    return {
      timestamp: resp[0],
      refSlot: resp[1],
      treeRoot: resp[2],
      reportCid: resp[3],
    };
  }

  async updateVaultData(
    vault: Address,
    totalValue: bigint,
    cumulativeLidoFees: bigint,
    liabilityShares: bigint,
    maxLiabilityShares: bigint,
    slashingReserve: bigint,
    proof: Hex[],
  ): Promise<TransactionReceipt | null> {
    const args = [
      vault,
      totalValue,
      cumulativeLidoFees,
      liabilityShares,
      maxLiabilityShares,
      slashingReserve,
      proof,
    ] as const;
    const { maxFeePerGas, maxPriorityFeePerGas } = await this.contractClientLibrary.estimateGasFees();
    const gasLimit = await this.contract.estimateGas.updateVaultData(args, {});
    const tx: TransactionSerializableEIP1559 = {
      to: this.contractAddress,
      type: "eip1559",
      data: encodeFunctionData({
        abi: this.contract.abi,
        functionName: "updateVaultData",
        args,
      }),
      chainId: await this.contractClientLibrary.getChainId(),
      gas: gasLimit,
      maxFeePerGas,
      maxPriorityFeePerGas,
    };
    const signature = await this.contractSignerService.sign(tx);
    const serializeSignedTransaction = serializeTransaction(tx, parseSignature(signature));
    return await this.contractClientLibrary.sendSerializedTransaction(serializeSignedTransaction);
  }
}
