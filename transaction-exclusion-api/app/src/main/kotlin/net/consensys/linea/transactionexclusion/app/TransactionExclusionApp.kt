package net.consensys.linea.transactionexclusion.app

import com.sksamuel.hoplite.Masked
import io.micrometer.core.instrument.MeterRegistry
import io.vertx.core.Future
import io.vertx.core.Vertx
import io.vertx.micrometer.backends.BackendRegistries
import io.vertx.sqlclient.SqlClient
import net.consensys.linea.async.toSafeFuture
import net.consensys.linea.async.toVertxFuture
import net.consensys.linea.transactionexclusion.RejectedTransactionsRepository
import net.consensys.linea.transactionexclusion.TransactionExclusionServiceV1
import net.consensys.linea.transactionexclusion.app.api.Api
import net.consensys.linea.transactionexclusion.app.api.ApiConfig
import net.consensys.linea.transactionexclusion.repository.RejectedTransactionsRepositoryImpl
import net.consensys.linea.transactionexclusion.service.RejectedTransactionCleanupService
import net.consensys.linea.transactionexclusion.service.TransactionExclusionServiceV1Impl
import net.consensys.linea.vertx.loadVertxConfig
import net.consensys.zkevm.persistence.dao.rejectedtransaction.RejectedTransactionsPostgresDao
import net.consensys.zkevm.persistence.db.Db
import org.apache.logging.log4j.LogManager
import java.time.Duration
import kotlin.time.toKotlinDuration

data class DbConnectionConfig(
  val host: String,
  val port: Int,
  val username: String,
  val password: Masked
)

data class DatabaseConfig(
  val read: DbConnectionConfig,
  val write: DbConnectionConfig,
  val schema: String,
  val readPoolSize: Int,
  val readPipeliningLimit: Int,
  val transactionalPoolSize: Int,
  val migrationDirLocation: String
)

data class AppConfig(
  val api: ApiConfig,
  val database: DatabaseConfig,
  val dbStoragePeriod: Duration,
  val dbCleanupPollingInterval: Duration,
  val dbQueryableWindowSinceRejectTimestamp: Duration
)

class TransactionExclusionApp(config: AppConfig) {
  private val log = LogManager.getLogger(TransactionExclusionApp::class.java)
  private val meterRegistry: MeterRegistry
  private val vertx: Vertx
  private val api: Api
  private val sqlReadClient: SqlClient
  private val sqlWriteClient: SqlClient
  private val rejectedTransactionsRepository: RejectedTransactionsRepository
  private val transactionExclusionService: TransactionExclusionServiceV1
  private val rejectedTransactionCleanupService: RejectedTransactionCleanupService

  init {
    log.debug("System properties: {}", System.getProperties())
    val vertxConfig = loadVertxConfig()
    log.debug("Vertx full configs: {}", vertxConfig)
    log.info("App configs: {}", config)
    this.vertx = Vertx.vertx(vertxConfig)
    this.meterRegistry = BackendRegistries.getDefaultNow()
    this.sqlReadClient = initDb(
      connectionConfig = config.database.read,
      schema = config.database.schema,
      migrationDirLocation = config.database.migrationDirLocation,
      transactionalPoolSize = config.database.transactionalPoolSize,
      readPipeliningLimit = config.database.readPipeliningLimit
    )
    this.sqlWriteClient = initDb(
      connectionConfig = config.database.write,
      schema = config.database.schema,
      migrationDirLocation = config.database.migrationDirLocation,
      transactionalPoolSize = config.database.transactionalPoolSize,
      readPipeliningLimit = config.database.readPipeliningLimit
    )
    this.rejectedTransactionsRepository = RejectedTransactionsRepositoryImpl(
      rejectedTransactionsDao = RejectedTransactionsPostgresDao(
        readConnection = this.sqlReadClient,
        writeConnection = this.sqlWriteClient,
        config = RejectedTransactionsPostgresDao.Config(
          queryableWindowSinceRejectTimestamp = config.dbQueryableWindowSinceRejectTimestamp.toKotlinDuration()
        )
      )
    )
    this.transactionExclusionService = TransactionExclusionServiceV1Impl(
      repository = this.rejectedTransactionsRepository
    )
    this.rejectedTransactionCleanupService = RejectedTransactionCleanupService(
      config = RejectedTransactionCleanupService.Config(
        pollingInterval = config.dbCleanupPollingInterval.toKotlinDuration(),
        storagePeriod = config.dbStoragePeriod.toKotlinDuration()
      ),
      repository = this.rejectedTransactionsRepository,
      vertx = this.vertx
    )
    this.api =
      Api(
        configs = config.api,
        vertx = vertx,
        meterRegistry = meterRegistry,
        transactionExclusionService = transactionExclusionService
      )
  }

  fun start(): Future<*> {
    log.info("Starting up app..")
    return api.start().toSafeFuture()
      .thenCompose { rejectedTransactionCleanupService.start() }
      .thenPeek {
        log.info("App successfully started")
      }.toVertxFuture()
  }

  fun stop(): Future<*> {
    log.info("Shooting down app..")
    return api.stop().toSafeFuture()
      .thenCompose { rejectedTransactionCleanupService.stop() }
      .thenPeek {
        log.info("App successfully stopped")
      }.toVertxFuture()
  }

  private fun initDb(
    connectionConfig: DbConnectionConfig,
    schema: String,
    migrationDirLocation: String,
    transactionalPoolSize: Int,
    readPipeliningLimit: Int,
    version: String? = "1"
  ): SqlClient {
    Db.applyDbMigrations(
      host = connectionConfig.host,
      port = connectionConfig.port,
      database = schema,
      target = version!!,
      username = connectionConfig.username,
      password = connectionConfig.password.value,
      migrationLocations = "filesystem:$migrationDirLocation"
    )
    return Db.vertxSqlClient(
      vertx = vertx,
      host = connectionConfig.host,
      port = connectionConfig.port,
      database = schema,
      username = connectionConfig.username,
      password = connectionConfig.password.value,
      maxPoolSize = transactionalPoolSize,
      pipeliningLimit = readPipeliningLimit
    )
  }
}
