import { IContractClientLibrary } from "ts-libs/linea-shared-utils/src/core/client/IContractClientLibrary";
import {
  Address,
  encodeFunctionData,
  getContract,
  GetContractReturnType,
  PublicClient,
  TransactionReceipt,
} from "viem";
import { LineaRollupYieldExtensionABI } from "../core/abis/LineaRollupYieldExtension";
import { ILineaRollup } from "../core/services/contracts/ILineaRollup";
import { IBaseContractClient } from "../core/clients/IBaseContractClient";

export class LineaRollupYieldExtensionContractClient implements ILineaRollup<TransactionReceipt>, IBaseContractClient {
  private readonly contract: GetContractReturnType<typeof LineaRollupYieldExtensionABI, PublicClient, Address>;

  constructor(
    private readonly contractClientLibrary: IContractClientLibrary<PublicClient, TransactionReceipt>,
    private readonly contractAddress: Address,
  ) {
    this.contract = getContract({
      abi: LineaRollupYieldExtensionABI,
      address: contractAddress,
      client: contractClientLibrary.getBlockchainClient(),
    });
  }

  getAddress(): Address {
    return this.contractAddress;
  }

  getContract(): GetContractReturnType {
    return this.contract;
  }

  async transferFundsForNativeYield(amount: bigint): Promise<TransactionReceipt | null> {
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "transferFundsForNativeYield",
      args: [amount],
    });

    return this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
  }
}
