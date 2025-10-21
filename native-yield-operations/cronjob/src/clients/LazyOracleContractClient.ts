import { IContractClientLibrary } from "ts-libs/linea-shared-utils/src/core/client/IContractClientLibrary";
import {
  Address,
  encodeFunctionData,
  getContract,
  GetContractReturnType,
  Hex,
  PublicClient,
  TransactionReceipt,
} from "viem";
import { LazyOracleABI } from "../core/abis/LazyOracle";
import { ILazyOracle } from "../core/services/contracts/ILazyOracle";
import { LazyOracleReportData } from "../core/entities";

export class LazyOracleContractClient implements ILazyOracle<TransactionReceipt> {
  private readonly contract: GetContractReturnType<typeof LazyOracleABI, PublicClient, Address>;
  constructor(
    private readonly contractClientLibrary: IContractClientLibrary<PublicClient, TransactionReceipt>,
    private readonly contractAddress: Address,
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
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "updateVaultData",
      args: [vault, totalValue, cumulativeLidoFees, liabilityShares, maxLiabilityShares, slashingReserve, proof],
    });
    return await this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
  }
}
