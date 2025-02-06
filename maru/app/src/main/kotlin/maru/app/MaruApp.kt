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
package maru.app

import maru.app.config.MaruConfig
import maru.consensus.dummy.DummyConsensusConfig
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger

class MaruApp(
  private val config: MaruConfig,
  private val beaconGenesisConfig: DummyConsensusConfig,
) {
  private val log: Logger = LogManager.getLogger(this::class.java)

  fun start() {
    if (config.p2pConfig == null) {
      log.warn("P2P is disabled!")
    }
    if (config.validator == null) {
      log.info("Maru is running in follower-only node")
    }
    log.info("Maru is up")
  }

  fun stop() {}
}
