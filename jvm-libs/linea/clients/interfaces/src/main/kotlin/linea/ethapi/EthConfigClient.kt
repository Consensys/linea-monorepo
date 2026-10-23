package linea.ethapi

import tech.pegasys.teku.infrastructure.async.SafeFuture

/**
 * Client for the EIP-7910 `eth_config` JSON-RPC method.
 */
interface EthConfigClient {
  fun ethConfig(): SafeFuture<EthConfig>
}

/**
 * `eth_config` result: the fork configuration currently in effect and, when a future fork is scheduled, the next
 * and the last configured ones.
 */
data class EthConfig(
  val current: ForkConfig,
  val next: ForkConfig?,
  val last: ForkConfig?,
)

data class ForkConfig(
  val activationTime: ULong,
  val chainId: ULong,
  /** eth_config `systemContracts`: EIP name (e.g. `BEACON_ROOTS_ADDRESS`) to 0x-prefixed lowercase address. */
  val systemContracts: Map<String, String>,
)

/**
 * System-contract addresses the execution engine uses for the chain, as 0x-prefixed lowercase hex.
 * Field order matches the RISC-V guest input and the `dynamicChainConfigHash` preimage.
 */
data class SystemContracts(
  val beaconRoots: String,
  val historyStorage: String,
  val deposit: String,
  val withdrawalRequest: String,
  val consolidationRequest: String,
  val builderDepositRequest: String,
  val builderExitRequest: String,
) {
  companion object {
    /** eth_config `systemContracts` keys, in field order. */
    val ETH_CONFIG_KEYS: List<String> = listOf(
      "BEACON_ROOTS_ADDRESS",
      "HISTORY_STORAGE_ADDRESS",
      "DEPOSIT_CONTRACT_ADDRESS",
      "WITHDRAWAL_REQUEST_PREDEPLOY_ADDRESS",
      "CONSOLIDATION_REQUEST_PREDEPLOY_ADDRESS",
      "BUILDER_DEPOSIT_CONTRACT_ADDRESS",
      "BUILDER_EXIT_CONTRACT_ADDRESS",
    )

    /** Returns null when [systemContracts] does not define every key in [ETH_CONFIG_KEYS] (e.g. a pre-Amsterdam fork). */
    fun fromEthConfig(systemContracts: Map<String, String>): SystemContracts? {
      val addresses = ETH_CONFIG_KEYS.map { key -> systemContracts[key]?.lowercase() ?: return null }
      return SystemContracts(
        beaconRoots = addresses[0],
        historyStorage = addresses[1],
        deposit = addresses[2],
        withdrawalRequest = addresses[3],
        consolidationRequest = addresses[4],
        builderDepositRequest = addresses[5],
        builderExitRequest = addresses[6],
      )
    }
  }
}

/**
 * The system contracts for proving [expectedChainId]: taken from every reported fork (current, next, last) that
 * defines all of them, which must agree. A change of system contracts across scheduled forks is rejected rather than
 * guessed at, since every value is committed to by the proofs (`dynamicChainConfigHash`).
 */
fun EthConfig.systemContracts(expectedChainId: ULong): SystemContracts {
  val forks = listOfNotNull(current, next, last)
  forks.forEach { fork ->
    require(fork.chainId == expectedChainId) {
      "eth_config chainId=${fork.chainId} does not match expected chainId=$expectedChainId"
    }
  }
  val candidates = forks.mapNotNull { SystemContracts.fromEthConfig(it.systemContracts) }.distinct()
  require(candidates.isNotEmpty()) {
    "eth_config reports no fork defining all of ${SystemContracts.ETH_CONFIG_KEYS}"
  }
  require(candidates.size == 1) {
    "eth_config reports different system contracts across forks, which is not supported: $candidates"
  }
  return candidates.single()
}
