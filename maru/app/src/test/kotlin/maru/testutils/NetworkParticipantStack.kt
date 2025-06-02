/*
   Copyright 2025 Consensys Software Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
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
