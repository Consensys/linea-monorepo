package net.consensys.zkevm.persistence.test

import io.vertx.core.Vertx
import io.vertx.sqlclient.Pool
import io.vertx.sqlclient.SqlClient
import net.consensys.zkevm.persistence.db.Db
import net.consensys.zkevm.persistence.db.Db.vertxConnectionPool
import net.consensys.zkevm.persistence.db.Db.vertxSqlClient
import net.consensys.zkevm.persistence.db.DbHelper
import org.junit.jupiter.api.AfterAll
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.BeforeAll
import org.junit.jupiter.api.BeforeEach
import org.postgresql.ds.PGSimpleDataSource
import javax.sql.DataSource

abstract class CleanDbTestSuite {
  private val host = "localhost"
  private val port = 5432
  abstract val databaseName: String
  private val username = "postgres"
  private val password = "postgres"
  private lateinit var dataSource: DataSource
  lateinit var pool: Pool
  lateinit var sqlClient: SqlClient
  val target: String = "3"

  @BeforeAll
  open fun beforeAll(vertx: Vertx) {
    val dbCreationDataSource = PGSimpleDataSource().also {
      it.serverNames = arrayOf(host)
      it.portNumbers = intArrayOf(port)
      it.databaseName = "postgres"
      it.user = username
      it.password = password
    }
    DbHelper.createDataBase(dbCreationDataSource, databaseName)
    dataSource = PGSimpleDataSource().also {
      it.serverNames = arrayOf(host)
      it.portNumbers = intArrayOf(port)
      it.databaseName = databaseName
      it.user = username
      it.password = password
    }
    pool = vertxConnectionPool(vertx, host, port, databaseName, username, password)
    sqlClient = vertxSqlClient(vertx, host, port, databaseName, username, password)
  }

  @BeforeEach
  open fun setUp(vertx: Vertx) {
    // drop flyway db metadata table as well
    // to recreate new db tables;
    DbHelper.dropAllTables(dataSource)
    Db.applyDbMigrations(dataSource, target)
  }

  @AfterEach
  open fun tearDown() {
    DbHelper.resetAllConnections(dataSource, databaseName)
  }

  @AfterAll
  open fun afterAll() {
    pool.close { ar: io.vertx.core.AsyncResult<Void?> ->
      if (ar.failed()) {
        System.err.println("Error closing connection pool: " + ar.cause().message)
      }
    }
    sqlClient.close { ar: io.vertx.core.AsyncResult<Void?> ->
      if (ar.failed()) {
        System.err.println("Error closing sqlclient " + ar.cause().message)
      }
    }
  }
}
