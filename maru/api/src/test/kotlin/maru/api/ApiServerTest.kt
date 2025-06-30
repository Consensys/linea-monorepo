/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.api

import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import java.util.concurrent.TimeUnit
import okhttp3.HttpUrl.Companion.toHttpUrl
import okhttp3.OkHttpClient
import okhttp3.Request
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test

class ApiServerTest {
  private val defaultObjectMapper = jacksonObjectMapper()

  private lateinit var apiServer: ApiServer
  private lateinit var client: OkHttpClient
  private lateinit var apiServerUrl: String

  private val fakeNetworkDataProvider =
    object : NetworkDataProvider {
      override fun getNodeId(): String = "TEST_NODE_ID"

      override fun getEnr(): String = "TEST_ENR"

      override fun getNodeAddresses(): List<String> = listOf("TEST_NODE_ADDRESS")

      override fun getDiscoveryAddresses(): List<String> = listOf("TEST_DISCOVERY_ADDRESS")
    }

  @BeforeEach
  fun beforeEach() {
    apiServer =
      ApiServer(
        config = ApiServer.Config(port = 0u),
        networkDataProvider = fakeNetworkDataProvider,
      )
    apiServer.start()
    apiServerUrl = "http://localhost:${apiServer.port()}"
    client = OkHttpClient.Builder().readTimeout(0, TimeUnit.SECONDS).build()
  }

  @AfterEach
  fun afterEach() {
    apiServer.stop()
  }

  @Test
  fun `test NodeGetNetworkIdentityV1 method`() {
    val url = (apiServerUrl + NodeGetNetworkIdentity.ROUTE).toHttpUrl()
    val request = Request.Builder().url(url).build()
    val networkIdentity =
      NetworkIdentity(
        peerId = fakeNetworkDataProvider.getNodeId(),
        enr = fakeNetworkDataProvider.getEnr(),
        p2pAddresses = fakeNetworkDataProvider.getNodeAddresses(),
        discoveryAddresses = fakeNetworkDataProvider.getDiscoveryAddresses(),
        metadata =
          Metadata(
            seqNumber = "0",
            attnets = emptyList(),
            syncnets = emptyList(),
          ),
      )

    val expectedGetNetworkIdentityResponse = GetNetworkIdentityResponse(networkIdentity)
    val response = client.newCall(request).execute()
    assertThat(response).isNotNull
    assertThat(response.code).isEqualTo(200)
    val responseBody = response.body?.string()
    val getNetworkIdentityResponse =
      defaultObjectMapper.readValue(
        responseBody,
        GetNetworkIdentityResponse::class.java,
      )
    assertThat(getNetworkIdentityResponse).isEqualTo(expectedGetNetworkIdentityResponse)
  }
}
