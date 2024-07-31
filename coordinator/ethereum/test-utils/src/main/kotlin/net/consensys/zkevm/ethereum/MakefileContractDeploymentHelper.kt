package net.consensys.zkevm.ethereum

import net.consensys.linea.async.toSafeFuture
import net.consensys.zkevm.coordinator.clients.smartcontract.LineaContractVersion
import org.apache.logging.log4j.LogManager
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.io.BufferedReader
import java.io.File
import java.io.InputStreamReader
import java.nio.file.Path
import java.nio.file.Paths
import java.util.regex.Matcher
import java.util.regex.Pattern

fun findFile(target: String): Path {
  var current = Paths.get("").toAbsolutePath()
  while (current != Paths.get("/")) {
    val targetFile = current.resolve(target).toFile()
    if (targetFile.exists()) {
      return targetFile.toPath()
    }
    current = current.parent
  }
  throw IllegalStateException("Couldn't find file $target")
}

data class CommandResult(
  val exitCode: Int,
  val stdOut: List<String>,
  val stdErr: List<String>
)

fun executeCommand(
  command: String,
  envVars: Map<String, String> = emptyMap(),
  executionDir: File = findFile("Makefile").parent.toFile()
): SafeFuture<CommandResult> {
  val log = LogManager.getLogger("net.consensys.zkevm.ethereum.CommandExecutor")
  val processBuilder = ProcessBuilder("/bin/sh", "-c", command)
  processBuilder.directory(executionDir)

  // Set environment variables
  val env = processBuilder.environment()
  for ((key, value) in envVars) {
    env[key] = value
  }

  val process = processBuilder.start()
  val stdOutReader = BufferedReader(InputStreamReader(process.inputStream))
  val stdErrorReader = BufferedReader(InputStreamReader(process.errorStream))

  // Read the standard output
  log.debug(
    "going to execute command: dir='{}', command='{}', envVars={} commandProcessId={} processInfo={}",
    executionDir,
    command,
    envVars,
    process.pid(),
    process.info()
  )
  process.waitFor(60, java.util.concurrent.TimeUnit.SECONDS)
  val futureResult = process
    .onExit()
    .thenApply { processResult ->
      val stdOutLines = stdOutReader.lines().toList()
      val stdErrLines = stdErrorReader.lines().toList()
      log.debug(
        "command finished: dir='{}', command='{}', exitCode={} envVars={} processId={} threadId={}",
        executionDir,
        command,
        processResult.exitValue(),
        envVars,
        ProcessHandle.current().pid(),
        Thread.currentThread().threadId()
      )
      log.debug(
        "stdout: {}",
        stdOutLines.joinToString("\n")
      )
      log.debug(
        "stderr: {}",
        stdErrLines.joinToString("\n")
      )
      CommandResult(processResult.exitValue(), stdOutLines, stdErrLines)
    }

  return futureResult.toSafeFuture()
}

private val lineaRollupAddressPattern = Pattern.compile(
  "^LineaRollup(?:AlphaV3)? deployed: address=(0x[0-9a-fA-F]{40}) blockNumber=(\\d+)"
)
private val l2MessageServiceAddressPattern = Pattern.compile(
  "^L2MessageService deployed: address=(0x[0-9a-fA-F]{40}) blockNumber=(\\d+)"
)

data class DeployedContract(
  val address: String,
  val blockNumber: Long
)

fun getDeployedAddress(
  commandResult: CommandResult,
  addressPattern: Pattern
): DeployedContract {
  val lines = commandResult.stdOut.toList().asReversed()
  val matcher: Matcher? = lines
    .firstOrNull { line -> addressPattern.matcher(line).find() }
    ?.let { addressPattern.matcher(it).also { it.find() } }

  return matcher
    ?.let { DeployedContract(it.group(1), it.group(2).toLong()) }
    ?: throw IllegalStateException("Couldn't extract contract address. Expecting pattern: $addressPattern")
}

private fun deployContract(
  command: String,
  env: Map<String, String> = emptyMap(),
  addressPattern: Pattern
): SafeFuture<DeployedContract> {
  return executeCommand(
    command = command,
    envVars = env
  )
    .thenApply { result ->
      if (result.exitCode != 0) {
        logCommand(result)
        throw IllegalStateException(
          "Command $command failed: " +
            "\nexitCode=${result.exitCode} " +
            "\nSTD_OUT: \n${result.stdOut.joinToString("\n")}" +
            "\nSTD_ERROR: \n${result.stdErr.joinToString("\n")}"
        )
      } else {
        runCatching { getDeployedAddress(result, addressPattern) }
          .onFailure { logCommand(result) }
          .getOrThrow()
      }
    }
}

fun makeDeployLineaRollup(
  deploymentPrivateKey: String? = null,
  operatorsAddresses: List<String>,
  contractVersion: LineaContractVersion
): SafeFuture<DeployedContract> {
  val env = mutableMapOf(
    "LINEA_ROLLUP_OPERATORS" to operatorsAddresses.joinToString(",")
    // "HARDHAT_DISABLE_CACHE" to "true"
  )
  deploymentPrivateKey?.let { env["DEPLOYMENT_PRIVATE_KEY"] = it }
  val command = if (contractVersion == LineaContractVersion.V5) {
    "make deploy-linea-rollup"
  } else {
    throw IllegalArgumentException("Unsupported contract version: $contractVersion")
  }

  return deployContract(
    command = command,
    env = env,
    addressPattern = lineaRollupAddressPattern
  )
}

fun makeDeployL2MessageService(
  deploymentPrivateKey: String? = null,
  anchorOperatorAddresses: String
): SafeFuture<DeployedContract> {
  val env = mutableMapOf(
    "L2MSGSERVICE_L1L2_MESSAGE_SETTER" to anchorOperatorAddresses
  )
  deploymentPrivateKey?.let { env["DEPLOYMENT_PRIVATE_KEY"] = it }

  return deployContract(
    command = "make deploy-l2messageservice",
    env = env,
    addressPattern = l2MessageServiceAddressPattern
  )
}

fun logCommand(commandResult: CommandResult) {
  println("stdout:")
  commandResult.stdOut.forEach { println(it) }
  println("stderr:")
  commandResult.stdErr.forEach { println(it) }
  println("exit code: ${commandResult.exitCode}")
}

fun main() {
  SafeFuture.collectAll(
    makeDeployLineaRollup(
      L1AccountManager.generateAccount().privateKey,
      listOf("03dfa322A95039BB679771346Ee2dBfEa0e2B773"),
      LineaContractVersion.V5
    ),
    makeDeployL2MessageService(
      L2AccountManager.generateAccount().privateKey,
      "03dfa322A95039BB679771346Ee2dBfEa0e2B773"
    )
  ).thenApply { addresses ->
    println("LineaRollup address: ${addresses[0]}")
    println("L2MessageService address: ${addresses[1]}")
  }.join()
}
