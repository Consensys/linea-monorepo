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

import kotlin.system.exitProcess
import org.apache.logging.log4j.LogManager
import picocli.CommandLine

object CliEntrypoint {
  private val log = LogManager.getLogger(this.javaClass)

  @JvmStatic
  fun main(args: Array<String>) {
    val cmd = CommandLine(MaruAppCli())
    cmd.setExecutionExceptionHandler { ex, _, _ ->
      log.error("Execution failure: ", ex)
      1
    }
    cmd.setParameterExceptionHandler { ex, _ ->
      log.error("Invalid args!: ", ex)
      1
    }
    val exitCode = cmd.execute(*args)
    if (exitCode != 0) {
      exitProcess(exitCode)
    }
  }
}
