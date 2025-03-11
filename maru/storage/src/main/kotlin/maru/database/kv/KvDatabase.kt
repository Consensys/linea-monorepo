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

import kotlin.jvm.optionals.getOrNull
import maru.core.BeaconState
import maru.core.SealedBeaconBlock
import maru.database.Database
import maru.database.Updater
import tech.pegasys.teku.storage.server.kvstore.KvStoreAccessor
import tech.pegasys.teku.storage.server.kvstore.KvStoreAccessor.KvStoreTransaction
import tech.pegasys.teku.storage.server.kvstore.schema.KvStoreColumn
import tech.pegasys.teku.storage.server.kvstore.schema.KvStoreVariable

class KvDatabase(
  private val kvStoreAccessor: KvStoreAccessor,
) : Database {
  companion object {
    object Schema {
      val BeaconStateByBlockRoot: KvStoreColumn<ByteArray, BeaconState> =
        KvStoreColumn.create(
          1,
          KvStoreSerializers.BytesSerializer,
          KvStoreSerializers.BeaconStateSerializer,
        )

      val SealedBeaconBlockByBlockRoot: KvStoreColumn<ByteArray, SealedBeaconBlock> =
        KvStoreColumn.create(
          2,
          KvStoreSerializers.BytesSerializer,
          KvStoreSerializers.SealedBeaconBlockSerializer,
        )

      val LatestBeaconState: KvStoreVariable<BeaconState> =
        KvStoreVariable.create(
          1,
          KvStoreSerializers.BeaconStateSerializer,
        )
    }
  }

  override fun getLatestBeaconState(): BeaconState? = kvStoreAccessor.get(Schema.LatestBeaconState).getOrNull()

  override fun getBeaconState(beaconBlockRoot: ByteArray): BeaconState? =
    kvStoreAccessor.get(Schema.BeaconStateByBlockRoot, beaconBlockRoot).getOrNull()

  override fun getSealedBeaconBlock(beaconBlockRoot: ByteArray): SealedBeaconBlock? =
    kvStoreAccessor.get(Schema.SealedBeaconBlockByBlockRoot, beaconBlockRoot).getOrNull()

  override fun newUpdater(): Updater = KvUpdater(this.kvStoreAccessor)

  override fun close() {
    kvStoreAccessor.close()
  }

  class KvUpdater(
    kvStoreAccessor: KvStoreAccessor,
  ) : Updater {
    private val transaction: KvStoreTransaction = kvStoreAccessor.startTransaction()

    override fun putBeaconState(beaconState: BeaconState): Updater {
      transaction.put(Schema.BeaconStateByBlockRoot, beaconState.latestBeaconBlockRoot, beaconState)
      transaction.put(Schema.LatestBeaconState, beaconState)
      return this
    }

    override fun putSealedBeaconBlock(
      sealedBeaconBlock: SealedBeaconBlock,
      beaconBlockRoot: ByteArray,
    ): Updater {
      transaction.put(Schema.SealedBeaconBlockByBlockRoot, beaconBlockRoot, sealedBeaconBlock)
      return this
    }

    override fun commit() {
      transaction.commit()
    }

    override fun rollback() {
      transaction.rollback()
    }

    override fun close() {
      transaction.close()
    }
  }
}
