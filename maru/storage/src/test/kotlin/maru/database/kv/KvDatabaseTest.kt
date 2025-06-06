/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.database.kv

import java.nio.file.Path
import java.util.Optional
import kotlin.random.Random
import maru.core.ext.DataGenerators
import maru.database.BeaconChain
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.metrics.noop.NoOpMetricsSystem
import org.hyperledger.besu.plugin.services.metrics.MetricCategory
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.io.TempDir

class KvDatabaseTest {
  private object KvDatabaseTestMetricCategory : MetricCategory {
    override fun getName(): String = KvDatabaseTest::class.simpleName!!

    override fun getApplicationPrefix(): Optional<String> = Optional.empty()
  }

  private fun createDatabase(databasePath: Path): BeaconChain =
    KvDatabaseFactory.createRocksDbDatabase(
      databasePath = databasePath,
      metricsSystem = NoOpMetricsSystem(),
      metricCategory = KvDatabaseTestMetricCategory,
    )

  @Test
  fun `test read and write beacon state`(
    @TempDir databasePath: Path,
  ) {
    val testBeaconStates = (1..10).map { DataGenerators.randomBeaconState(it.toULong()) }
    createDatabase(databasePath).use { db ->
      testBeaconStates.forEach { testBeaconState ->
        db.newUpdater().use {
          it.putBeaconState(testBeaconState).commit()
        }
        assertThat(db.getBeaconState(testBeaconState.latestBeaconBlockHeader.hash))
          .isEqualTo(testBeaconState)
      }
    }

    createDatabase(databasePath).use { db ->
      testBeaconStates.forEach { testBeaconState ->
        assertThat(db.getBeaconState(testBeaconState.latestBeaconBlockHeader.hash))
          .isEqualTo(testBeaconState)
      }
    }
  }

  @Test
  fun `test read and write latest beacon state`(
    @TempDir databasePath: Path,
  ) {
    val testBeaconStates = (1..10).map { DataGenerators.randomBeaconState(it.toULong()) }
    createDatabase(databasePath).use { db ->
      testBeaconStates.forEach { testBeaconState ->
        db.newUpdater().use {
          it.putBeaconState(testBeaconState).commit()
        }
        assertThat(db.getLatestBeaconState())
          .isEqualTo(testBeaconState)
      }
    }

    createDatabase(databasePath).use { db ->
      assertThat(db.getLatestBeaconState())
        .isEqualTo(testBeaconStates.last())
    }
  }

  @Test
  fun `test invalid key read`(
    @TempDir databasePath: Path,
  ) {
    val randomKey = Random.nextBytes(32)
    createDatabase(databasePath).use { db ->
      assertThat(db.getBeaconState(randomKey)).isNull()
      assertThat(db.getSealedBeaconBlock(randomKey)).isNull()
      assertThat(db.getSealedBeaconBlock(100uL)).isNull()
    }
  }

  @Test
  fun `test read and write beacon blocks`(
    @TempDir databasePath: Path,
  ) {
    val testBeaconBlocks =
      (1..10).map { DataGenerators.randomSealedBeaconBlock(it.toULong()) }
    createDatabase(databasePath).use { db ->
      testBeaconBlocks.forEach { testBeaconBlock ->
        db.newUpdater().use {
          it.putSealedBeaconBlock(testBeaconBlock).commit()
        }
        assertThat(db.getSealedBeaconBlock(testBeaconBlock.beaconBlock.beaconBlockHeader.hash))
          .isEqualTo(testBeaconBlock)

        assertThat(db.getSealedBeaconBlock(testBeaconBlock.beaconBlock.beaconBlockHeader.number))
          .isEqualTo(testBeaconBlock)
      }
    }

    createDatabase(databasePath).use { db ->
      testBeaconBlocks.forEach { testBeaconBlock ->
        assertThat(db.getSealedBeaconBlock(testBeaconBlock.beaconBlock.beaconBlockHeader.hash))
          .isEqualTo(testBeaconBlock)

        assertThat(db.getSealedBeaconBlock(testBeaconBlock.beaconBlock.beaconBlockHeader.number))
          .isEqualTo(testBeaconBlock)
      }
    }
  }

  @Test
  fun `test repeated write`(
    @TempDir databasePath: Path,
  ) {
    val testBeaconBlock = DataGenerators.randomSealedBeaconBlock(1uL)
    createDatabase(databasePath).use { db ->
      db.newUpdater().use {
        it.putSealedBeaconBlock(testBeaconBlock).commit()
      }
      db.newUpdater().use {
        it.putSealedBeaconBlock(testBeaconBlock).commit()
      }
    }

    createDatabase(databasePath).use { db ->
      assertThat(db.getSealedBeaconBlock(testBeaconBlock.beaconBlock.beaconBlockHeader.hash))
        .isEqualTo(testBeaconBlock)

      assertThat(db.getSealedBeaconBlock(testBeaconBlock.beaconBlock.beaconBlockHeader.number))
        .isEqualTo(testBeaconBlock)
    }
  }

  @Test
  fun `test update rollback`(
    @TempDir databasePath: Path,
  ) {
    val testBeaconBlock1 = DataGenerators.randomSealedBeaconBlock(1uL)
    val testBeaconBlock1Number = testBeaconBlock1.beaconBlock.beaconBlockHeader.number
    val testBeaconBlock1Root = testBeaconBlock1.beaconBlock.beaconBlockHeader.hash
    val testBeaconBlock2 = DataGenerators.randomSealedBeaconBlock(2uL)
    val testBeaconBlock2Number = testBeaconBlock2.beaconBlock.beaconBlockHeader.number
    val testBeaconBlock2Root = testBeaconBlock2.beaconBlock.beaconBlockHeader.hash
    createDatabase(databasePath).use { db ->
      db.newUpdater().use {
        it.putSealedBeaconBlock(testBeaconBlock1).commit()
      }
      assertThat(db.getSealedBeaconBlock(testBeaconBlock1Root)).isEqualTo(testBeaconBlock1)
      assertThat(db.getSealedBeaconBlock(testBeaconBlock1Number)).isEqualTo(testBeaconBlock1)

      assertThat(db.getSealedBeaconBlock(testBeaconBlock2Root)).isNull()
      assertThat(db.getSealedBeaconBlock(testBeaconBlock2Number)).isNull()

      db.newUpdater().use { it.putSealedBeaconBlock(testBeaconBlock2).rollback() }

      assertThat(db.getSealedBeaconBlock(testBeaconBlock1Root)).isEqualTo(testBeaconBlock1)
      assertThat(db.getSealedBeaconBlock(testBeaconBlock1Number)).isEqualTo(testBeaconBlock1)

      assertThat(db.getSealedBeaconBlock(testBeaconBlock2Root)).isNull()
      assertThat(db.getSealedBeaconBlock(testBeaconBlock2Number)).isNull()
    }

    createDatabase(databasePath).use { db ->
      assertThat(db.getSealedBeaconBlock(testBeaconBlock1Root)).isEqualTo(testBeaconBlock1)
      assertThat(db.getSealedBeaconBlock(testBeaconBlock1Number)).isEqualTo(testBeaconBlock1)

      assertThat(db.getSealedBeaconBlock(testBeaconBlock2Root)).isNull()
      assertThat(db.getSealedBeaconBlock(testBeaconBlock2Number)).isNull()
    }
  }
}
