package net.consensys.zkevm.coordinator.app

import com.sksamuel.hoplite.Masked
import net.consensys.linea.traces.TracingModule
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test
import java.io.File
import java.io.PrintWriter
import java.math.BigInteger
import java.net.URL
import java.nio.charset.Charset
import java.nio.file.Path
import java.time.Duration

class CoordinatorConfigTest {
  val errorWriter = PrintWriter(System.err, true, Charset.defaultCharset())

  companion object {

    val apiConfig = ApiConfig(9545U)

    val conflationConfig = ConflationConfig(
      1,
      Duration.parse("PT6S"),
      Duration.parse("PT3S"),
      Duration.parse("PT2S"),
      129072,
      40,
      1699,
      null,
      null,
      mapOf<TracingModule, UInt>(
        TracingModule.ADD to 262144U,
        TracingModule.BIN to 262144U,
        TracingModule.BIN_RT to 262144U,
        TracingModule.EC_DATA to 4096U,
        TracingModule.EXT to 16384U,
        TracingModule.HUB to 2097152U,
        TracingModule.INSTRUCTION_DECODER to 512U,
        TracingModule.MMIO to 1048576U,
        TracingModule.MMU to 524288U,
        TracingModule.MMU_ID to 256U,
        TracingModule.MOD to 131072U,
        TracingModule.MUL to 65536U,
        TracingModule.MXP to 524288U,
        TracingModule.PHONEY_RLP to 65536U,
        TracingModule.PUB_HASH to 32768U,
        TracingModule.PUB_HASH_INFO to 8192U,
        TracingModule.PUB_LOG to 16384U,
        TracingModule.PUB_LOG_INFO to 16384U,
        TracingModule.RLP to 128U,
        TracingModule.ROM to 1048576U,
        TracingModule.SHF to 65536U,
        TracingModule.SHF_RT to 262144U,
        TracingModule.TX_RLP to 65536U,
        TracingModule.WCP to 262144U,
        TracingModule.BLOCK_TX to 200U,
        TracingModule.BLOCK_L2L1LOGS to 16U,
        TracingModule.BLOCK_KECCAK to 8192U,
        TracingModule.PRECOMPILE_ECRECOVER to 10000U,
        TracingModule.PRECOMPILE_SHA2 to 10000U,
        TracingModule.PRECOMPILE_RIPEMD to 10000U,
        TracingModule.PRECOMPILE_IDENTITY to 10000U,
        TracingModule.PRECOMPILE_MODEXP to 10000U,
        TracingModule.PRECOMPILE_ECADD to 10000U,
        TracingModule.PRECOMPILE_ECMUL to 10000U,
        TracingModule.PRECOMPILE_ECPAIRING to 10000U,
        TracingModule.PRECOMPILE_BLAKE2F to 512U
      )
    )

    val zkGethTracesConfig = ZkGethTraces(
      URL("http://traces-node:8545"),
      Duration.parse("PT1S")
    )

    val sequencerConfig = SequencerConfig(
      URL("http://sequencer:8545")
    )

    val proverConfig = ProverConfig(
      version = "0.2.0",
      fsInputDirectory = Path.of("/data/prover/request"),
      fsOutputDirectory = Path.of("/data/prover/response"),
      fsPollingInterval = Duration.parse("PT20S"),
      fsInprogessProvingSuffixPattern = "\\.inprogress\\.prover.*",
      timeout = Duration.parse("PT10M")
    )

    val tracesConfig = TracesConfig(
      "0.2.0",
      TracesConfig.FunctionalityEndpoint(
        listOf(
          URL("http://traces-api:8080/")
        ),
        20U,
        2U,
        Duration.parse("PT1S")
      ),
      TracesConfig.FunctionalityEndpoint(
        listOf(
          URL("http://traces-api:8080/")
        ),
        2U,
        2U,
        Duration.parse("PT1S")
      ),
      TracesConfig.FileManager(
        "json.gz",
        Path.of("/data/traces/raw"),
        Path.of("/data/traces/raw-non-canonical"),
        true,
        Duration.parse("PT1S"),
        Duration.parse("PT30S")
      )
    )

    val stateManagerConfig = StateManagerClientConfig(
      "1.2.0",
      listOf(
        URL("http://shomei:8888/")
      ),
      3U
    )

    val batchSubmissionConfig = BatchSubmissionConfig(
      10,
      Duration.parse("PT1S")
    )

    val databaseConfig = DatabaseConfig(
      "postgres",
      5432,
      "postgres",
      Masked("postgres"),
      "linea_coordinator",
      10,
      10,
      10
    )

    val l1Config = L1Config(
      "0xC737F2334651ea85A72D8DA9d933c821A8377F9f",
      URL("http://l1-validator:8545"),
      Duration.parse("PT6S"),
      Duration.parse("PT6S"),
      2U,
      BigInteger.valueOf(10000000),
      10,
      15.0,
      BigInteger.valueOf(100000000000),
      BigInteger.ZERO,
      Duration.parse("PT6S"),
      Duration.parse("PT02M"),
      1000U,
      "latest",
      5U
    )

    val l2Config = L2Config(
      "0xe537D669CA013d86EBeF1D64e40fC74CADC91987",
      BigInteger.valueOf(10000000),
      BigInteger.valueOf(100000000000),
      4U,
      15.0,
      2U,
      25U,
      1000U,
      Duration.parse("PT01S"),
      120U
    )

    val l1SignerConfig = SignerConfig(
      SignerConfig.Type.Web3j,
      Web3SignerConfig(
        "http://127.0.0.1:9000",
        10U,
        true,
        "0x3c753c0c9db5ce651144dbba3da47476ddede5b98607272de97a347d72c2eac2be3f6aea33598f297a0cd0f12cf08fe0" +
          "e5198531393ea726a36efd07a7872382"
      ),
      Web3jConfig(Masked("0x202454d1b4e72c41ebf58150030f649648d3cf5590297fb6718e27039ed9c86d"))
    )

    val l1SignerWeb3Config = SignerConfig(
      SignerConfig.Type.Web3Signer,
      Web3SignerConfig(
        "http://127.0.0.1:9000",
        10U,
        true,
        "0x3c753c0c9db5ce651144dbba3da47476ddede5b98607272de97a347d72c2eac2be3f6aea33598f297a0cd0f12cf08fe0" +
          "e5198531393ea726a36efd07a7872382"
      ),
      Web3jConfig(Masked("0x202454d1b4e72c41ebf58150030f649648d3cf5590297fb6718e27039ed9c86d"))
    )

    val l2SignerConfig = SignerConfig(
      SignerConfig.Type.Web3j,
      Web3SignerConfig(
        "http://127.0.0.1:9000",
        10U,
        true,
        "0cdc73b5a30ecb3eb2028fdd2d5ab423221763fbda7127e38d664b033455e30e7c6833bff6b99197de4d2069de30f6aaa9" +
          "0626986b737b7e74e635f4f7cedfbf"
      ),
      Web3jConfig(Masked("0x1dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae"))
    )

    val l2Web3SignerConfig = SignerConfig(
      SignerConfig.Type.Web3Signer,
      Web3SignerConfig(
        "http://127.0.0.1:9000",
        10U,
        true,
        "0cdc73b5a30ecb3eb2028fdd2d5ab423221763fbda7127e38d664b033455e30e7c6833bff6b99197de4d2069de30f6aaa9" +
          "0626986b737b7e74e635f4f7cedfbf"
      ),
      Web3jConfig(Masked("0x1dd171cec7e2995408b5513004e8207fe88d6820aeff0d82463b3e41df251aae"))
    )

    val messageAnchoringServiceConfig = MessageAnchoringServiceConfig(
      Duration.parse("PT48S"),
      100U
    )

    val dynamicGasPriceServiceConfig = DynamicGasPriceServiceConfig(
      Duration.parse("PT12S"),
      50,
      15.0,
      0.1.toBigDecimal(),
      1.0.toBigDecimal(),
      BigInteger.valueOf(10_000_000_000),
      listOf(
        URL("http://sequencer:8545/"),
        URL("http://traces-node:8545/"),
        URL("http://l2-node:8545/")
      )
    )

    val coordinatorConfig = CoordinatorConfig(
      sequencerConfig,
      zkGethTracesConfig,
      proverConfig,
      tracesConfig,
      l1Config,
      l2Config,
      l1SignerConfig,
      batchSubmissionConfig,
      databaseConfig,
      stateManagerConfig,
      conflationConfig,
      apiConfig,
      l2SignerConfig,
      messageAnchoringServiceConfig,
      dynamicGasPriceServiceConfig
    )
  }

  @Test
  fun parsesValidConfig() {
    val tracesLimitsConfigs =
      CoordinatorAppCli.loadConfigs<TracesLimitsConfigFile>(
        listOf(File("../../config/common/traces-limits-v1.toml")),
        errorWriter
      )
    val configs = CoordinatorAppCli.loadConfigs<CoordinatorConfig>(
      listOf(File("../../config/coordinator/coordinator-docker.config.toml")),
      errorWriter
    )
      ?.let { config: CoordinatorConfig ->
        config.copy(
          conflation = config.conflation.copy(_tracesLimits = tracesLimitsConfigs?.tracesLimits)
        )
      }

    assertEquals(coordinatorConfig, configs)
  }

  @Test
  fun parsesValidWeb3SignerConfigOverride() {
    val tracesLimitsConfigs =
      CoordinatorAppCli.loadConfigs<TracesLimitsConfigFile>(
        listOf(File("../../config/common/traces-limits-v1.toml")),
        errorWriter
      )
    val configs =
      CoordinatorAppCli.loadConfigs<CoordinatorConfig>(
        listOf(
          File("../../config/coordinator/coordinator-docker.config.toml"),
          File("../../config/coordinator/coordinator-docker-web3signer-override.config.toml")
        ),
        errorWriter
      )?.let {
        it.copy(
          conflation = it.conflation.copy(_tracesLimits = tracesLimitsConfigs?.tracesLimits)
        )
      }

    val expectedConfig =
      coordinatorConfig.copy(
        l1Signer = l1SignerWeb3Config,
        l2Signer = l2Web3SignerConfig
      )

    assertEquals(expectedConfig, configs)
  }
}
