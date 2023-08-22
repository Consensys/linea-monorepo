package net.consensys.zkevm.ethereum.settlement.persistence

import javax.sql.DataSource

object DbHelper {
  fun dropTables(dataSource: DataSource, vararg tableNames: String?) {
    val sql = "DROP TABLE IF EXISTS " + java.lang.String.join(",", *tableNames)
    dataSource.connection.use { con -> con.prepareStatement(sql).execute() }
  }

  fun resetAllConnections(dataSource: DataSource, db: String?) {
    val sql = "select pg_terminate_backend(pid) from pg_stat_activity where datname=?"
    try {
      dataSource.connection.use { con ->
        val ps = con.prepareStatement(sql)
        ps.setString(1, db)
        ps.execute()
      }
    } catch (ignored: Exception) {
      // The current connection will be dropped as well and the exception will be thrown
    }
  }

  fun dropAllTablesForce(dataSource: DataSource, databaseName: String?) {
    resetAllConnections(dataSource, databaseName)
    val sql = "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
    dataSource.connection.use { con -> con.prepareStatement(sql).execute() }
  }

  fun dropAllTables(dataSource: DataSource) {
    val sql = "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
    dataSource.connection.use { con -> con.prepareStatement(sql).execute() }
  }
}
