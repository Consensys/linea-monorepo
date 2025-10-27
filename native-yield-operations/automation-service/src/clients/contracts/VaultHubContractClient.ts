import { IBlockchainClient } from "@consensys/linea-shared-utils";
import { Address, decodeEventLog, getContract, GetContractReturnType, PublicClient, TransactionReceipt } from "viem";
import { IVaultHub } from "../../core/clients/contracts/IVaultHub.js";
import { VaultHubABI } from "../../core/abis/VaultHub.js";

export class VaultHubContractClient implements IVaultHub<TransactionReceipt> {
  private readonly contract: GetContractReturnType<typeof VaultHubABI, PublicClient, Address>;

  constructor(
    private readonly contractClientLibrary: IBlockchainClient<PublicClient, TransactionReceipt>,
    private readonly contractAddress: Address,
  ) {
    this.contract = getContract({
      abi: VaultHubABI,
      address: contractAddress,
      client: this.contractClientLibrary.getBlockchainClient(),
    });
  }

  getAddress(): Address {
    return this.contractAddress;
  }

  getContract(): GetContractReturnType {
    return this.contract;
  }

  getLiabilityPaymentFromTxReceipt(txReceipt: TransactionReceipt): bigint {
    for (const log of txReceipt.logs) {
      // Only decode logs emitted by this contract
      if (log.address.toLowerCase() !== this.contractAddress.toLowerCase()) continue;

      try {
        const decoded = decodeEventLog({
          abi: this.contract.abi,
          data: log.data,
          topics: log.topics,
        });

        if (decoded.eventName === "VaultRebalanced") {
          const { etherWithdrawn } = decoded.args;
          return etherWithdrawn as bigint;
        }
      } catch {
        // skip unrelated logs (from the same contract or different ABIs)
      }
    }

    // If event not found
    return 0n;
  }

  getLidoFeePaymentFromTxReceipt(txReceipt: TransactionReceipt): bigint {
    for (const log of txReceipt.logs) {
      // Only decode logs emitted by this contract
      if (log.address.toLowerCase() !== this.contractAddress.toLowerCase()) continue;

      try {
        const decoded = decodeEventLog({
          abi: this.contract.abi,
          data: log.data,
          topics: log.topics,
        });

        if (decoded.eventName === "LidoFeesSettled") {
          const { transferred } = decoded.args;
          return transferred as bigint;
        }
      } catch {
        // skip unrelated logs (from the same contract or different ABIs)
      }
    }

    // If event not found
    return 0n;
  }
}
