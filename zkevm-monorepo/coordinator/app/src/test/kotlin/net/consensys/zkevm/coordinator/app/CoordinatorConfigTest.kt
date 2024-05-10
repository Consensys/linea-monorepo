package net.consensys.zkevm.coordinator.app

import com.github.michaelbull.result.get
import com.github.michaelbull.result.getError
import com.github.michaelbull.result.onFailure
import com.github.michaelbull.result.onSuccess
import com.sksamuel.hoplite.Masked
import net.consensys.linea.traces.TracingModule
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.fail
import org.junit.jupiter.api.Test
import java.io.File
import java.math.BigInteger
import java.net.URL
import java.nio.file.Path
import java.time.Duration
import kotlin.time.Duration.Companion.seconds
import kotlin.time.toJavaDuration

class CoordinatorConfigTest {
  companion object {

    private val apiConfig = ApiConfig(9545U)

    private val conflationConfig = ConflationConfig(
      consistentNumberOfBlocksOnL1ToWait = 1,
      conflationCalculatorVersion = "1.0.0",
      conflationDeadline = Duration.parse("PT6S"),
      conflationDeadlineCheckInterval = Duration.parse("PT3S"),
      conflationDeadlineLastBlockConfirmationDelay = Duration.parse("PT2S"),
      blocksLimit = null,
      _tracesLimits = mapOf(
        TracingModule.ADD to 524288U,
        TracingModule.BIN to 262144U,
        TracingModule.BIN_RT to 262144U,
        TracingModule.EC_DATA to 4096U,
        TracingModule.EXT to 131072U,
        TracingModule.HUB to 2097152U,
        TracingModule.INSTRUCTION_DECODER to 512U,
        TracingModule.MMIO to 131072U,
        TracingModule.MMU to 131072U,
        TracingModule.MMU_ID to 131072U,
        TracingModule.MOD to 131072U,
        TracingModule.MUL to 65536U,
        TracingModule.MXP to 524288U,
        TracingModule.PHONEY_RLP to 32768U,
        TracingModule.PUB_HASH to 32768U,
        TracingModule.PUB_HASH_INFO to 32768U,
        TracingModule.PUB_LOG to 16384U,
        TracingModule.PUB_LOG_INFO to 16384U,
        TracingModule.RLP to 512U,
        TracingModule.ROM to 4194304U,
        TracingModule.SHF to 65536U,
        TracingModule.SHF_RT to 4096U,
        TracingModule.TX_RLP to 131072U,
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
      ),
      smartContractErrors = mapOf(
        "0F06CD15" to "DataAlreadySubmitted",
        "D5AA5AD6" to "StateRootHashInvalid",
        "C01EAB56" to "EmptySubmissionData",
        "DCB23885" to "FinalizationDataMissing",
        "ABEFA5E8" to "DataStartingBlockDoesNotMatch",
        "9F6ADC69" to "DataEndingBlockDoesNotMatch",
        "5548C6B3" to "DataParentHasEmptyShnarf",
        "36459FA0" to "L1RollingHashDoesNotExistOnL1",
        "3211D9BE" to "TimestampsNotInSequence",
        "9C80BDEA" to "FinalBlockNumberInvalid",
        "9A89A758" to "ParentHashesDoesNotMatch",
        "E1CB6E60" to "FinalStateRootHashDoesNotMatch",
        "710CD580" to "DataHashesNotInSequence",
        "CBBD7953" to "FirstBlockGreaterThanFinalBlock",
        "A386ED70" to "FirstBlockLessThanOrEqualToLastFinalizedBlock",
        "70614405" to "FinalBlockNumberLessThanOrEqualToLastFinalizedBlock",
        "2898482A" to "FinalBlockStateEqualsZeroHash",
        "BF81C6E0" to "FinalizationInTheFuture",
        "0C256592" to "MissingMessageNumberForRollingHash",
        "5228F4C8" to "MissingRollingHashForMessageNumber",
        "729EEBCE" to "FirstByteIsNotZero",
        "6426C6C5" to "BytesLengthNotMultipleOf32",
        "32BDD925" to "YPointGreaterThanCurveModulus",
        "68DCAD5F" to "PointEvaluationResponseInvalid",
        "F75DB381" to "PrecompileReturnDataLengthWrong",
        "A71194AF" to "PointEvaluationFailed",
        "A0FEAE8D" to "EmptyBlobData",
        "2F22B98A" to "LastFinalizedShnarfWrong",
        "4E686675" to "L2MerkleRootDoesNotExist",
        "B05E92FA" to "InvalidMerkleProof",
        "5E3FD6AD" to "ProofLengthDifferentThanMerkleDepth",
        "9B0F0C28" to "SystemMigrationBlockZero",
        "E5D14425" to "L2MerkleRootAlreadyAnchored",
        "335A4A90" to "MessageAlreadyClaimed",
        "0C91D776" to "BytesLengthNotMultipleOfTwo",
        "D39E75F9" to "L1MessageNumberSynchronizationWrong",
        "AC1775CF" to "ServiceHasMigratedToRollingHashes",
        "7557A60A" to "L1RollingHashSynchronizationWrong",
        "36A4BB94" to "FinalRollingHashIsZero",
        "A75D20DB" to "BlockTimestampError",
        "EAD4C30E" to "StartingRootHashDoesNotMatch",
        "E98CFD2F" to "EmptyBlockDataArray",
        "8999649C" to "EmptyBlock",
        "7907D79B" to "ProofIsEmpty",
        "69ED70AB" to "InvalidProofType",
        "09BDE339" to "InvalidProof",
        "8579BEFE" to "ZeroAddressNotAllowed",
        "4124301E" to "MessageAlreadySent",
        "992D87C3" to "MessageDoesNotExistOrHasAlreadyBeenClaimed",
        "EE49E001" to "MessageAlreadyReceived",
        "62A064C5" to "L1L2MessageNotSent",
        "3B174434" to "MessageHashesListLengthHigherThanOneHundred",
        "52E9BB24" to "EmptyMessageHashesArray",
        "DB246DDE" to "IsPaused",
        "B015579F" to "IsNotPaused",
        "BAC5BF1B" to "TransactionShort",
        "E95A14A6" to "UnknownTransactionType",
        "06007832" to "NotList",
        "FCED5668" to "NoNext",
        "78268BBB" to "MemoryOutOfBounds"
      ),
      fetchBlocksLimit = 4000
    )

    private val zkGethTracesConfig = ZkGethTraces(
      URL("http://traces-node:8545"),
      Duration.parse("PT1S")
    )

    private val proverConfig = ProverConfig(
      version = "v2.0.0",
      fsRequestsDirectory = Path.of("/data/prover-execution/requests"),
      fsResponsesDirectory = Path.of("/data/prover-execution/responses"),
      fsInprogessRequestWritingSuffix = ".inprogress_coordinator_writing",
      fsInprogessProvingSuffixPattern = "\\.inprogress\\.prover.*",
      fsPollingInterval = Duration.parse("PT1S"),
      fsPollingTimeout = Duration.parse("PT10M")
    )

    private val blobCompressionConfig = BlobCompressionConfig(
      blobSizeLimit = 100 * 1024,
      handlerPollingInterval = Duration.parse("PT1S"),
      prover = ProverConfig(
        version = "v2.0.0",
        fsRequestsDirectory = Path.of("/data/prover-compression/requests"),
        fsResponsesDirectory = Path.of("/data/prover-compression/responses"),
        fsPollingInterval = Duration.parse("PT1S"),
        fsPollingTimeout = Duration.parse("PT10M"),
        fsInprogessRequestWritingSuffix = ".inprogress_coordinator_writing",
        fsInprogessProvingSuffixPattern = "\\.inprogress\\.prover.*"
      )
    )

    private val aggregationConfig = AggregationConfig(
      aggregationCalculatorVersion = "1.0.0",
      aggregationProofsLimit = 10,
      aggregationDeadline = Duration.parse("PT1M"),
      pollingInterval = Duration.parse("PT2S"),
      prover = ProverConfig(
        version = "v2.0.0",
        fsRequestsDirectory = Path.of("/data/prover-aggregation/requests"),
        fsResponsesDirectory = Path.of("/data/prover-aggregation/responses"),
        fsPollingInterval = Duration.parse("PT20S"),
        fsPollingTimeout = Duration.parse("PT20M"),
        fsInprogessRequestWritingSuffix = ".inprogress_coordinator_writing",
        fsInprogessProvingSuffixPattern = "\\.inprogress\\.prover.*"
      )
    )

    private val tracesConfig = TracesConfig(
      rawExecutionTracesVersion = "0.2.0",
      expectedTracesApiVersion = "0.2.0",
      counters = TracesConfig.FunctionalityEndpoint(
        listOf(
          URL("http://traces-api:8080/")
        ),
        requestLimitPerEndpoint = 20U,
        requestRetry = RequestRetryConfigTomlFriendly(
          maxAttempts = 4,
          backoffDelay = Duration.parse("PT1S"),
          failuresWarningThreshold = 2
        )
      ),
      conflation = TracesConfig.FunctionalityEndpoint(
        endpoints = listOf(
          URL("http://traces-api:8080/")
        ),
        requestLimitPerEndpoint = 2U,
        requestRetry = RequestRetryConfigTomlFriendly(
          maxAttempts = 4,
          backoffDelay = Duration.parse("PT1S"),
          failuresWarningThreshold = 2
        )
      ),
      fileManager = TracesConfig.FileManager(
        tracesFileExtension = "json.gz",
        rawTracesDirectory = Path.of("/data/traces/raw"),
        nonCanonicalRawTracesDirectory = Path.of("/data/traces/raw-non-canonical"),
        createNonCanonicalDirectory = true,
        pollingInterval = Duration.parse("PT1S"),
        tracesFileCreationWaitTimeout = Duration.parse("PT2M")
      )
    )

    private val type2StateProofProviderConfig = Type2StateProofProviderConfig(
      endpoints = listOf(URL("http://shomei-frontend:8888/")),
      requestRetry = RequestRetryConfigTomlFriendly(
        maxAttempts = 3,
        backoffDelay = Duration.parse("PT1S"),
        failuresWarningThreshold = 2
      )
    )
    private val stateManagerConfig = StateManagerClientConfig(
      version = "2.1.1",
      endpoints = listOf(URL("http://shomei:8888/")),
      requestLimitPerEndpoint = 3U,
      requestRetry = RequestRetryConfigTomlFriendly(
        maxAttempts = 5,
        backoffDelay = Duration.parse("PT2S"),
        failuresWarningThreshold = 2
      )
    )

    private val blobSubmissionConfig = BlobSubmissionConfig(
      dbPollingInterval = Duration.parse("PT10S"),
      maxBlobsToReturn = 100,
      maxBlobsToSubmitPerTick = 10,
      priorityFeePerGasUpperBound = BigInteger.valueOf(20000000000),
      priorityFeePerGasLowerBound = BigInteger.valueOf(2000000000),
      proofSubmissionDelay = Duration.parse("PT1S"),
      targetBlobsToSendPerTransaction = 6,
      disabled = true
    )

    private val aggregationFinalizationConfig = AggregationFinalizationConfig(
      dbPollingInterval = Duration.parse("PT10S"),
      maxAggregationsToFinalizePerTick = 1,
      proofSubmissionDelay = Duration.parse("PT1S"),
      disabled = false
    )

    private val databaseConfig = DatabaseConfig(
      host = "postgres",
      port = 5432,
      username = "postgres",
      password = Masked("postgres"),
      schema = "linea_coordinator",
      readPoolSize = 10,
      readPipeliningLimit = 10,
      transactionalPoolSize = 10
    )

    private val l1Config = L1Config(
      zkEvmContractAddress = "0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9",
      rpcEndpoint = URL("http://l1-el-node:8545"),
      finalizationPollingInterval = Duration.parse("PT6S"),
      newBatchPollingInterval = Duration.parse("PT6S"),
      blocksToFinalization = 2U,
      gasLimit = BigInteger.valueOf(10000000),
      feeHistoryBlockCount = 10,
      feeHistoryRewardPercentile = 15.0,
      maxFeePerGasCap = BigInteger.valueOf(100000000000),
      maxFeePerBlobGasCap = BigInteger.valueOf(100000000000),
      gasPriceCapMultiplierForFinalization = 2.0,
      earliestBlock = BigInteger.ZERO,
      sendMessageEventPollingInterval = Duration.parse("PT1S"),
      maxEventScrapingTime = Duration.parse("PT5S"),
      maxMessagesToCollect = 1000U,
      finalizedBlockTag = "latest",
      blockRangeLoopLimit = 100U,
      _ethFeeHistoryEndpoint = null
    )

    private val l2Config = L2Config(
      messageServiceAddress = "0xe537D669CA013d86EBeF1D64e40fC74CADC91987",
      rpcEndpoint = URL("http://sequencer:8545"),
      gasLimit = BigInteger.valueOf(10000000),
      maxFeePerGasCap = BigInteger.valueOf(100000000000),
      feeHistoryBlockCount = 4U,
      feeHistoryRewardPercentile = 15.0,
      blocksToFinalization = 2U,
      lastHashSearchWindow = 25U,
      lastHashSearchMaxBlocksBack = 1000U,
      anchoringReceiptPollingInterval = Duration.parse("PT01S"),
      maxReceiptRetries = 120U
    )

    private val finalizationSigner = SignerConfig(
      type = SignerConfig.Type.Web3j,
      web3signer = Web3SignerConfig(
        endpoint = "http://web3signer:9000",
        maxPoolSize = 10U,
        keepAlive = true,
        publicKey =
        "ba5734d8f7091719471e7f7ed6b9df170dc70cc661ca05e688601ad984f068b0d67351e5f06073092499336ab0839ef8a521afd334e5" +
          "3807205fa2f08eec74f4"
      ),
      web3j = Web3jConfig(Masked("0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"))
    )

    private val dataSubmissionSigner = SignerConfig(
      type = SignerConfig.Type.Web3j,
      web3signer = Web3SignerConfig(
        endpoint = "http://web3signer:9000",
        maxPoolSize = 10U,
        keepAlive = true,
        publicKey =
        "9d9031e97dd78ff8c15aa86939de9b1e791066a0224e331bc962a2099a7b1f0464b8bbafe1535f2301c72c2cb3535b172da30b02686a" +
          "b0393d348614f157fbdb"
      ),
      web3j = Web3jConfig(Masked("0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a"))
    )
    private val l2SignerConfig = SignerConfig(
      type = SignerConfig.Type.Web3j,
      web3signer = Web3SignerConfig(
        endpoint = "http://web3signer:9000",
        maxPoolSize = 10U,
        keepAlive = true,
        publicKey =
        "4a788ad6fa008beed58de6418369717d7492f37d173d70e2c26d9737e2c6eeae929452ef8602a19410844db3e200a0e73f5208fd7625" +
          "9a8766b73953fc3e7023"
      ),
      web3j = Web3jConfig(Masked("0x4d01ae6487860981699236a58b68f807ee5f17b12df5740b85cf4c4653be0f55"))
    )

    private val messageAnchoringServiceConfig = MessageAnchoringServiceConfig(
      pollingInterval = Duration.parse("PT10S"),
      maxMessagesToAnchor = 100U
    )

    private val dynamicGasPriceServiceConfig = DynamicGasPriceServiceConfig(
      priceUpdateInterval = Duration.parse("PT12S"),
      feeHistoryBlockCount = 50,
      feeHistoryRewardPercentile = 15.0,
      baseFeeCoefficient = 0.1.toBigDecimal(),
      priorityFeeCoefficient = 1.0.toBigDecimal(),
      baseFeeBlobCoefficient = 0.1.toBigDecimal(),
      blobSubmissionExpectedExecutionGas = 213_000.0.toBigDecimal(),
      expectedBlobGas = 131072.0.toBigDecimal(),
      gasPriceUpperBound = BigInteger.valueOf(10_000_000_000),
      gasPriceLowerBound = BigInteger.valueOf(90_000_000),
      gasPriceFixedCost = BigInteger.ZERO,
      gethGasPriceUpdateRecipients = listOf(
        URL("http://traces-node:8545/"),
        URL("http://l2-node:8545/")
      ),
      besuGasPriceUpdateRecipients = listOf(
        URL("http://sequencer:8545/")
      ),
      requestRetry = RequestRetryConfigTomlFriendly(
        maxAttempts = 3,
        timeout = 6.seconds.toJavaDuration(),
        backoffDelay = 1.seconds.toJavaDuration(),
        failuresWarningThreshold = 2
      )
    )

    private val l1DynamicGasPriceCapServiceConfig = L1DynamicGasPriceCapServiceConfig(
      disabled = true,
      gasPriceCapCalculation = L1DynamicGasPriceCapServiceConfig.GasPriceCapCalculation(
        adjustmentConstant = 25U,
        blobAdjustmentConstant = 25U,
        finalizationTargetMaxDelay = Duration.parse("PT32H"),
        baseFeePerGasPercentileWindow = Duration.parse("P7D"),
        baseFeePerGasPercentileWindowLeeway = Duration.parse("PT10M"),
        baseFeePerGasPercentile = 10.0,
        timeOfDayMultipliers = mapOf(
          "SUNDAY_0" to 1.7489178377946066,
          "SUNDAY_1" to 1.7494632175198737,
          "SUNDAY_2" to 1.75,
          "SUNDAY_3" to 1.733166295438555,
          "SUNDAY_4" to 1.6993775444542885,
          "SUNDAY_5" to 1.6350086618091364,
          "SUNDAY_6" to 1.5627740860151331,
          "SUNDAY_7" to 1.4831149222064164,
          "SUNDAY_8" to 1.4101476768256929,
          "SUNDAY_9" to 1.370085278922007,
          "SUNDAY_10" to 1.3516015544068651,
          "SUNDAY_11" to 1.3482404546676368,
          "SUNDAY_12" to 1.3580905751578942,
          "SUNDAY_13" to 1.3775497419563296,
          "SUNDAY_14" to 1.3700255667542938,
          "SUNDAY_15" to 1.2642948506461285,
          "SUNDAY_16" to 1.2794806131912935,
          "SUNDAY_17" to 1.2750892256476676,
          "SUNDAY_18" to 1.2919720208955585,
          "SUNDAY_19" to 1.317984990098603,
          "SUNDAY_20" to 1.4433501639513178,
          "SUNDAY_21" to 1.4705921238901998,
          "SUNDAY_22" to 1.515043370430801,
          "SUNDAY_23" to 1.5556742617266397,
          "MONDAY_0" to 1.5381562278760164,
          "MONDAY_1" to 1.5423761828433993,
          "MONDAY_2" to 1.539015963719092,
          "MONDAY_3" to 1.487676153648977,
          "MONDAY_4" to 1.430973985132037,
          "MONDAY_5" to 1.4656765439056292,
          "MONDAY_6" to 1.4484298622828233,
          "MONDAY_7" to 1.4459076216659752,
          "MONDAY_8" to 1.4899061835032241,
          "MONDAY_9" to 1.5249733712852067,
          "MONDAY_10" to 1.511367489481033,
          "MONDAY_11" to 1.4225695658047797,
          "MONDAY_12" to 1.2887291896624584,
          "MONDAY_13" to 1.1460926897291355,
          "MONDAY_14" to 1.0004897955233254,
          "MONDAY_15" to 0.8694664537368378,
          "MONDAY_16" to 0.8270273375962802,
          "MONDAY_17" to 0.7868289022833883,
          "MONDAY_18" to 0.7780303121746551,
          "MONDAY_19" to 0.7756215256634205,
          "MONDAY_20" to 0.7984895728860915,
          "MONDAY_21" to 0.8918589268832423,
          "MONDAY_22" to 0.9967716668541272,
          "MONDAY_23" to 1.0973334887144106,
          "TUESDAY_0" to 1.2233064209957951,
          "TUESDAY_1" to 1.3238883432855082,
          "TUESDAY_2" to 1.3874518307497257,
          "TUESDAY_3" to 1.463621147171298,
          "TUESDAY_4" to 1.4975989065490154,
          "TUESDAY_5" to 1.481679186141442,
          "TUESDAY_6" to 1.452778387763161,
          "TUESDAY_7" to 1.3414858185569951,
          "TUESDAY_8" to 1.2869454637983988,
          "TUESDAY_9" to 1.249347290389873,
          "TUESDAY_10" to 1.196488297386161,
          "TUESDAY_11" to 1.1136140507034202,
          "TUESDAY_12" to 0.9867528660797885,
          "TUESDAY_13" to 0.8018989158195754,
          "TUESDAY_14" to 0.6173048748109258,
          "TUESDAY_15" to 0.46718586671750373,
          "TUESDAY_16" to 0.4103633833041902,
          "TUESDAY_17" to 0.4871260756989506,
          "TUESDAY_18" to 0.5667378483016126,
          "TUESDAY_19" to 0.6464203510900723,
          "TUESDAY_20" to 0.7780268325299871,
          "TUESDAY_21" to 0.8995921101255763,
          "TUESDAY_22" to 1.0077600114996088,
          "TUESDAY_23" to 1.1109769960680498,
          "WEDNESDAY_0" to 1.2097668746150059,
          "WEDNESDAY_1" to 1.2631002319009361,
          "WEDNESDAY_2" to 1.2912775191940549,
          "WEDNESDAY_3" to 1.3229785939630059,
          "WEDNESDAY_4" to 1.3428607301494424,
          "WEDNESDAY_5" to 1.3750788517823973,
          "WEDNESDAY_6" to 1.3752344527256497,
          "WEDNESDAY_7" to 1.3505490078766218,
          "WEDNESDAY_8" to 1.2598503219367945,
          "WEDNESDAY_9" to 1.2051668977452374,
          "WEDNESDAY_10" to 1.0320896222195326,
          "WEDNESDAY_11" to 0.8900138031631949,
          "WEDNESDAY_12" to 0.6341155208698448,
          "WEDNESDAY_13" to 0.48337590254714624,
          "WEDNESDAY_14" to 0.2903189399226416,
          "WEDNESDAY_15" to 0.25,
          "WEDNESDAY_16" to 0.25711039485046006,
          "WEDNESDAY_17" to 0.37307641907591793,
          "WEDNESDAY_18" to 0.45280799454961196,
          "WEDNESDAY_19" to 0.5631397823847637,
          "WEDNESDAY_20" to 0.6285005244224133,
          "WEDNESDAY_21" to 0.6671897537279405,
          "WEDNESDAY_22" to 0.7268406397452634,
          "WEDNESDAY_23" to 0.8068904097486369,
          "THURSDAY_0" to 0.9021601102971811,
          "THURSDAY_1" to 1.023741688964238,
          "THURSDAY_2" to 1.1340689935096755,
          "THURSDAY_3" to 1.2530130345819006,
          "THURSDAY_4" to 1.3163421664973542,
          "THURSDAY_5" to 1.3536343767230727,
          "THURSDAY_6" to 1.3432290485306728,
          "THURSDAY_7" to 1.2864983218982178,
          "THURSDAY_8" to 1.2320488534113174,
          "THURSDAY_9" to 1.1984530721079034,
          "THURSDAY_10" to 1.0877338251341975,
          "THURSDAY_11" to 0.9999324929016475,
          "THURSDAY_12" to 0.87536726762619,
          "THURSDAY_13" to 0.6560822412167919,
          "THURSDAY_14" to 0.44836474861432074,
          "THURSDAY_15" to 0.36145134935025247,
          "THURSDAY_16" to 0.2695997829759713,
          "THURSDAY_17" to 0.2898426312618241,
          "THURSDAY_18" to 0.3970093434340387,
          "THURSDAY_19" to 0.5193273246848977,
          "THURSDAY_20" to 0.6426415257034419,
          "THURSDAY_21" to 0.800685718218497,
          "THURSDAY_22" to 0.9215516833839711,
          "THURSDAY_23" to 1.053701659160912,
          "FRIDAY_0" to 1.149649788723893,
          "FRIDAY_1" to 1.2046315447861193,
          "FRIDAY_2" to 1.2724031281576726,
          "FRIDAY_3" to 1.3525693456352732,
          "FRIDAY_4" to 1.3746126314960814,
          "FRIDAY_5" to 1.3744591862592468,
          "FRIDAY_6" to 1.3297812543035683,
          "FRIDAY_7" to 1.2762064429631657,
          "FRIDAY_8" to 1.235662409263294,
          "FRIDAY_9" to 1.2171558028785991,
          "FRIDAY_10" to 1.182722399785398,
          "FRIDAY_11" to 1.137345538963285,
          "FRIDAY_12" to 0.9999308422620752,
          "FRIDAY_13" to 0.8055000309055653,
          "FRIDAY_14" to 0.5667135273493851,
          "FRIDAY_15" to 0.4081529603000651,
          "FRIDAY_16" to 0.3987031354907009,
          "FRIDAY_17" to 0.5030075499003412,
          "FRIDAY_18" to 0.6518159532641841,
          "FRIDAY_19" to 0.8733483414970974,
          "FRIDAY_20" to 1.0496224913080463,
          "FRIDAY_21" to 1.1820684558591705,
          "FRIDAY_22" to 1.2561688567574458,
          "FRIDAY_23" to 1.3204704912328773,
          "SATURDAY_0" to 1.3832230236620218,
          "SATURDAY_1" to 1.4632908341022142,
          "SATURDAY_2" to 1.5019230781315296,
          "SATURDAY_3" to 1.5437332506007084,
          "SATURDAY_4" to 1.5934153179751855,
          "SATURDAY_5" to 1.6245578072557723,
          "SATURDAY_6" to 1.6294919789890665,
          "SATURDAY_7" to 1.6027665451672717,
          "SATURDAY_8" to 1.6068061069158674,
          "SATURDAY_9" to 1.624257927970777,
          "SATURDAY_10" to 1.5996112411089,
          "SATURDAY_11" to 1.5659672993092648,
          "SATURDAY_12" to 1.5333537902522736,
          "SATURDAY_13" to 1.445292929996356,
          "SATURDAY_14" to 1.2966021477035259,
          "SATURDAY_15" to 1.250999408961155,
          "SATURDAY_16" to 1.2535364828163025,
          "SATURDAY_17" to 1.2736456128871074,
          "SATURDAY_18" to 1.3348268054897328,
          "SATURDAY_19" to 1.4571388900094875,
          "SATURDAY_20" to 1.5073787902995706,
          "SATURDAY_21" to 1.5605139580010123,
          "SATURDAY_22" to 1.5885303316932382,
          "SATURDAY_23" to 1.6169891066719597
        )
      ),
      feeHistoryFetcher = L1DynamicGasPriceCapServiceConfig.FeeHistoryFetcher(
        fetchInterval = Duration.parse("PT1S"),
        maxBlockCount = 1000U,
        rewardPercentiles = listOf(10.0, 20.0, 30.0, 40.0, 50.0, 60.0, 70.0, 80.0, 90.0, 100.0),
        endpoint = null
      ),
      feeHistoryStorage = L1DynamicGasPriceCapServiceConfig.FeeHistoryStorage(
        storagePeriod = Duration.parse("P10D")
      )
    )

    private val coordinatorConfig = CoordinatorConfig(
      duplicatedLogsDebounceTime = Duration.parse("PT15S"),
      zkGethTraces = zkGethTracesConfig,
      prover = proverConfig,
      blobCompression = blobCompressionConfig,
      proofAggregation = aggregationConfig,
      traces = tracesConfig,
      type2StateProofProvider = type2StateProofProviderConfig,
      l1 = l1Config,
      l2 = l2Config,
      finalizationSigner = finalizationSigner,
      dataSubmissionSigner = dataSubmissionSigner,
      blobSubmission = blobSubmissionConfig,
      aggregationFinalization = aggregationFinalizationConfig,
      database = databaseConfig,
      stateManager = stateManagerConfig,
      conflation = conflationConfig,
      api = apiConfig,
      l2Signer = l2SignerConfig,
      messageAnchoringService = messageAnchoringServiceConfig,
      dynamicGasPriceService = dynamicGasPriceServiceConfig,
      l1DynamicGasPriceCapService = l1DynamicGasPriceCapServiceConfig,
      eip4844SwitchL2BlockNumber = 0
    )
  }

  private data class TestConfig(val extraField: String)

  @Test
  fun parsesValidConfig() {
    val smartContractErrorConfig =
      CoordinatorAppCli.loadConfigsOrError<SmartContractErrorCodesConfig>(
        listOf(File("../../config/common/smart-contract-errors.toml"))
      )
    val timeOfDayMultipliers =
      CoordinatorAppCli.loadConfigsOrError<GasPriceCapTimeOfDayMultipliersConfig>(
        listOf(File("../../config/common/gas-price-cap-time-of-day-multipliers.toml"))
      )
    val tracesLimitsConfigs =
      CoordinatorAppCli.loadConfigsOrError<TracesLimitsConfigFile>(
        listOf(File("../../config/common/traces-limits-v1.toml"))
      )
    CoordinatorAppCli.loadConfigsOrError<CoordinatorConfig>(
      listOf(File("../../config/coordinator/coordinator-docker.config.toml"))
    )
      .onFailure { error: String -> fail(error) }
      .onSuccess { config: CoordinatorConfig ->
        val configs = config.copy(
          conflation = config.conflation.copy(
            _tracesLimits = tracesLimitsConfigs.get()?.tracesLimits,
            smartContractErrors = smartContractErrorConfig.get()!!.smartContractErrors
          ),
          l1DynamicGasPriceCapService = config.l1DynamicGasPriceCapService.copy(
            gasPriceCapCalculation = config.l1DynamicGasPriceCapService.gasPriceCapCalculation.copy(
              timeOfDayMultipliers = timeOfDayMultipliers.get()?.gasPriceCapTimeOfDayMultipliers
            )
          )
        )
        assertEquals(coordinatorConfig, configs)
        assertEquals(coordinatorConfig.l1.rpcEndpoint, coordinatorConfig.l1.ethFeeHistoryEndpoint)
      }
  }

  @Test
  fun parsesValidWeb3SignerConfigOverride() {
    val smartContractErrorCodes =
      CoordinatorAppCli.loadConfigsOrError<SmartContractErrorCodesConfig>(
        listOf(File("../../config/common/smart-contract-errors.toml"))
      )
    val timeOfDayMultipliers =
      CoordinatorAppCli.loadConfigsOrError<GasPriceCapTimeOfDayMultipliersConfig>(
        listOf(File("../../config/common/gas-price-cap-time-of-day-multipliers.toml"))
      )
    val tracesLimitsConfigs =
      CoordinatorAppCli.loadConfigsOrError<TracesLimitsConfigFile>(
        listOf(File("../../config/common/traces-limits-v1.toml"))
      )

    CoordinatorAppCli.loadConfigsOrError<CoordinatorConfig>(
      listOf(
        File("../../config/coordinator/coordinator-docker.config.toml"),
        File("../../config/coordinator/coordinator-docker-web3signer-override.config.toml")
      )
    )
      .onFailure { error: String -> fail(error) }
      .onSuccess {
        val configs = it.copy(
          conflation = it.conflation.copy(
            _tracesLimits = tracesLimitsConfigs.get()?.tracesLimits,
            smartContractErrors = smartContractErrorCodes.get()!!.smartContractErrors
          ),
          l1DynamicGasPriceCapService = it.l1DynamicGasPriceCapService.copy(
            gasPriceCapCalculation = it.l1DynamicGasPriceCapService.gasPriceCapCalculation.copy(
              timeOfDayMultipliers = timeOfDayMultipliers.get()?.gasPriceCapTimeOfDayMultipliers
            )
          )
        )

        val expectedConfig =
          coordinatorConfig.copy(
            finalizationSigner = finalizationSigner.copy(type = SignerConfig.Type.Web3Signer),
            dataSubmissionSigner = dataSubmissionSigner.copy(type = SignerConfig.Type.Web3Signer),
            l2Signer = l2SignerConfig.copy(type = SignerConfig.Type.Web3Signer)
          )

        assertEquals(expectedConfig, configs)
      }
  }

  @Test
  fun invalidConfigReturnsErrorResult() {
    val configs =
      CoordinatorAppCli.loadConfigsOrError<TestConfig>(
        listOf(
          File("../../config/coordinator/coordinator-docker.config.toml"),
          File("../../config/coordinator/coordinator-docker-web3signer-override.config.toml")
        )
      )

    assertThat(configs.getError()).contains("'extraField': Missing from config")
  }
}
