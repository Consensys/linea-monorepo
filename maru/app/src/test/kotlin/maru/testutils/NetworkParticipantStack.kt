/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.testutils

import java.nio.file.Files
import java.nio.file.Path
import maru.app.MaruApp
import maru.testutils.besu.BesuFactory
import org.hyperledger.besu.tests.acceptance.dsl.node.cluster.Cluster

class NetworkParticipantStack(
  cluster: Cluster,
  maruBuilder: (ethereumJsonRpcBaseUrl: String, engineRpcUrl: String, tmpDir: Path) -> MaruApp,
) {
  val besuNode = BesuFactory.buildTestBesu()
  val tmpDir: Path =
    Files.createTempDirectory("maru-app").also {
      it.toFile().deleteOnExit()
    }
  var maruApp: MaruApp =
    let {
      cluster.start(besuNode)
      val ethereumJsonRpcBaseUrl = besuNode.jsonRpcBaseUrl().get()
      val engineRpcUrl = besuNode.engineRpcUrl().get()
      maruBuilder(ethereumJsonRpcBaseUrl, engineRpcUrl, tmpDir)
    }

  fun stop() {
    maruApp.stop()
    besuNode.stop()
  }

  val p2pPort: UInt
    get() = maruApp.p2pPort()
}
