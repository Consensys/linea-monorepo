/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.api

import io.javalin.Javalin

class ApiServer(
  val config: Config,
  val networkDataProvider: NetworkDataProvider,
) {
  data class Config(
    val port: UInt,
  )

  var app: Javalin? = null

  fun start() {
    if (app != null) {
      app!!.start(config.port.toInt())
    } else {
      // To support apiserver restarts after stop, we need to create a new Javalin instance
      // https://github.com/javalin/javalin/issues/941
      app =
        Javalin
          .create()
          .get(NodeGetNetworkIdentity.ROUTE, NodeGetNetworkIdentity(networkDataProvider))
          .start(config.port.toInt())
    }
  }

  fun stop() {
    app?.stop()
    app = null
  }

  fun port(): Int = app!!.port()
}
