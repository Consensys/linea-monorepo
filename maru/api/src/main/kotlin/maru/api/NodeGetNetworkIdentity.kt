/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.api/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
import com.fasterxml.jackson.annotation.JsonProperty
import io.javalin.http.Context
import io.javalin.http.Handler
import io.javalin.http.HttpStatus

/**
 * https://ethereum.github.io/beacon-APIs/#/Node/getNetworkIdentity
 */
data class GetNetworkIdentityResponse(
  val data: NetworkIdentity,
)

data class NetworkIdentity(
  @JsonProperty("peer_id") val peerId: String,
  val enr: String,
  @JsonProperty("p2p_addresses") val p2pAddresses: List<String>,
  @JsonProperty("discovery_addresses") val discoveryAddresses: List<String>,
  val metadata: Metadata,
)

/**
 * https://github.com/ethereum/consensus-specs/blob/v1.3.0/specs/phase0/p2p-interface.md#metadata
 */
data class Metadata(
  @JsonProperty("seq_number") val seqNumber: String,
  val attnets: List<String>,
  val syncnets: List<String>,
)

class NodeGetNetworkIdentity(
  val networkDataProvider: NetworkDataProvider,
) : Handler {
  override fun handle(ctx: Context) {
    val networkIdentity =
      NetworkIdentity(
        peerId = networkDataProvider.getNodeId(),
        enr = networkDataProvider.getEnr() ?: "",
        p2pAddresses = networkDataProvider.getNodeAddresses(),
        discoveryAddresses = networkDataProvider.getDiscoveryAddresses(),
        metadata =
          Metadata(
            seqNumber = "0",
            attnets = emptyList(),
            syncnets = emptyList(),
          ),
      )
    ctx.status(HttpStatus.OK).json(GetNetworkIdentityResponse(networkIdentity))
  }

  companion object {
    const val ROUTE = "/eth/v1/node/identity"
  }
}
