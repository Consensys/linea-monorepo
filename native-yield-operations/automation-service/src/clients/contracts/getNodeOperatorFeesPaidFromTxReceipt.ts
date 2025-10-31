import { Address, decodeEventLog, TransactionReceipt } from "viem";
import { DashboardABI } from "../../core/abis/Dashboard.js";

// Functions that would be in a DashboardClient if we had one
// But DashboardClient cannot have a fixed address - we can have multiple Dashboard.sol contracts

/**
 * Extracts the node operator fee amount from a transaction receipt by decoding FeeDisbursed events.
 * Functions that would be in a DashboardClient if we had one, but DashboardClient cannot have a fixed address
 * since we can have multiple Dashboard.sol contracts.
 * Only decodes logs emitted by the specified dashboard contract. Skips unrelated logs (from the same contract or different ABIs).
 * If event not found, returns 0n.
 *
 * @param {TransactionReceipt} txReceipt - The transaction receipt to search for FeeDisbursed events.
 * @param {Address} dashboardAddress - The address of the Dashboard contract to filter logs by.
 * @returns {bigint} The fee amount from the FeeDisbursed event, or 0n if the event is not found.
 */
export function getNodeOperatorFeesPaidFromTxReceipt(txReceipt: TransactionReceipt, dashboardAddress: Address): bigint {
  for (const log of txReceipt.logs) {
    // Only decode logs emitted by this contract
    if (log.address.toLowerCase() !== dashboardAddress.toLowerCase()) continue;

    try {
      const decoded = decodeEventLog({
        abi: DashboardABI,
        data: log.data,
        topics: log.topics,
      });

      if (decoded.eventName === "FeeDisbursed") {
        const { fee } = decoded.args;
        return fee as bigint;
      }
    } catch {
      // skip unrelated logs (from the same contract or different ABIs)
    }
  }

  // If event not found
  return 0n;
}
