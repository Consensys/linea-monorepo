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
package maru.database.kv

import java.nio.file.Path
import org.hyperledger.besu.plugin.services.MetricsSystem
import org.hyperledger.besu.plugin.services.metrics.MetricCategory
import tech.pegasys.teku.storage.server.kvstore.KvStoreConfiguration
import tech.pegasys.teku.storage.server.rocksdb.RocksDbInstanceFactory

object KvDatabaseFactory {
  fun createRocksDbDatabase(
    databasePath: Path,
    metricsSystem: MetricsSystem,
    metricCategory: MetricCategory,
  ): KvDatabase {
    val rocksDbInstance =
      RocksDbInstanceFactory.create(
        metricsSystem,
        metricCategory,
        KvStoreConfiguration().withDatabaseDir(databasePath),
        listOf(
          KvDatabase.Companion.Schema.BeaconBlockByBlockRoot,
          KvDatabase.Companion.Schema.BeaconStateByBlockRoot,
        ),
        emptyList(),
      )
    return KvDatabase(rocksDbInstance)
  }
}
