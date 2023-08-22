package net.consensys.zkevm.ethereum.settlement.persistence

import io.vertx.core.AsyncResult
import io.vertx.core.Vertx
import io.vertx.sqlclient.Pool
import io.vertx.sqlclient.SqlClient
import net.consensys.zkevm.ethereum.settlement.persistence.DbHelper.dropAllTables
import net.consensys.zkevm.ethereum.settlement.persistence.DbHelper.resetAllConnections
import org.junit.jupiter.api.AfterAll
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.BeforeEach
import org.postgresql.ds.PGSimpleDataSource

abstract class CleanDbTestSuite {
  private val host = "localhost"
  private val port = 5432
  private val databaseName = "coordinator_tests"
  private val username = "coordinator"
  private val password = "coordinator_tests"
  private var dataSource =
    PGSimpleDataSource().also {
      it.portNumbers = intArrayOf(port)
      it.databaseName = databaseName
      it.user = username
      it.password = password
    }
  lateinit var pool: Pool
  lateinit var sqlClient: SqlClient

  @BeforeAll
  open fun beforeAll(vertx: Vertx) {
    pool = Db.vertxConnectionPool(vertx, host, port, databaseName, username, password)
    sqlClient = Db.vertxSqlClient(vertx, host, port, databaseName, username, password)
  }

  @BeforeEach
  open fun setUp(vertx: Vertx) {
    // drop flyway db metadata table as well
    // to recreate new db tables;
    dropAllTables(dataSource)
    Db.applyDbMigrations(dataSource)
  }

  @AfterEach
  @Throws(Exception::class)
  open fun tearDown() {
    resetAllConnections(dataSource, "coordinator_tests")
  }

  @AfterAll
  open fun afterAll() {
    pool.close { ar: AsyncResult<Void?> ->
      if (ar.failed()) {
        System.err.println("Error closing connection pool: " + ar.cause().message)
      }
    }
    sqlClient.close { ar: AsyncResult<Void?> ->
      if (ar.failed()) {
        System.err.println("Error closing sqlclient " + ar.cause().message)
      }
    }
  }
}
