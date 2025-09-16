/*
 * Copyright ConsenSys Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package net.consensys.linea.zktracer;

import static net.consensys.linea.zktracer.Trace.*;

import net.consensys.linea.plugins.BesuServiceProvider;
import org.hyperledger.besu.datatypes.HardforkId;
import org.hyperledger.besu.datatypes.HardforkId.MainnetHardforkId;
import org.hyperledger.besu.plugin.ServiceManager;
import org.hyperledger.besu.plugin.services.BlockchainService;

/**
 * release numbers of forks are defined from the <b>Ethereum Protocol Releases</b> table in <a
 * href="https://github.com/ethereum/execution-specs">execution specs</a> repo. We start counting at
 * 1 and include all named releases, including aborted ones (such as "DAO Wars").
 */
public enum Fork {
  LONDON(EVM_LONDON),
  PARIS(EVM_PARIS),
  SHANGHAI(EVM_SHANGHAI),
  CANCUN(EVM_CANCUN),
  PRAGUE(EVM_PRAGUE),
  OSAKA(EVM_OSAKA) // not yet live on L1
;

  private final int releaseNumber;

  Fork(int releaseNumber) {
    this.releaseNumber = releaseNumber;
  }

  public int getReleaseNumber() {
    return releaseNumber;
  }

  public static String toString(Fork fork) {
    return switch (fork) {
      case LONDON -> "london";
      case PARIS -> "paris";
      case SHANGHAI -> "shanghai";
      case CANCUN -> "cancun";
      case PRAGUE -> "prague";
      case OSAKA -> "osaka";
      default -> throw new IllegalArgumentException("Unknown fork: " + fork);
    };
  }

  /**
   * Construct a fork instance from the name of a fork (e.g. "London", "Shanghai", etc). Observe
   * that case does not matter here. Hence, "LONDON", "London", "london", "lonDon" are all suitable
   * aliases for the LONDON instance.
   *
   * @param fork
   * @return
   */
  public static Fork fromString(String fork) {
    return Fork.valueOf(fork.toUpperCase());
  }

  private static boolean forkIsAtLeast(Fork fork, Fork threshold) {
    return fork.getReleaseNumber() >= threshold.getReleaseNumber();
  }

  public static boolean isPostParis(Fork fork) {
    return forkIsAtLeast(fork, PARIS);
  }

  public static boolean isPostShanghai(Fork fork) {
    return forkIsAtLeast(fork, SHANGHAI);
  }

  public static boolean isPostCancun(Fork fork) {
    return forkIsAtLeast(fork, CANCUN);
  }

  public static boolean isPostPrague(Fork fork) {
    return forkIsAtLeast(fork, PRAGUE);
  }

  /**
   * Map MainnetHardforkId, datatype from Besu, to Fork enum instance
   *
   * @param hardForkId the hardfork id retrieved from Besu API
   * @return Fork
   */
  private static Fork fromMainnetHardforkId(MainnetHardforkId hardForkId) {
    return switch (hardForkId) {
      case MainnetHardforkId.LONDON -> LONDON;
      case MainnetHardforkId.PARIS -> PARIS;
      case MainnetHardforkId.SHANGHAI -> SHANGHAI;
      case MainnetHardforkId.CANCUN -> CANCUN;
      case MainnetHardforkId.PRAGUE -> PRAGUE;
      case MainnetHardforkId.OSAKA -> OSAKA;
      default -> throw new IllegalArgumentException("Unknown hardfork id: " + hardForkId);
    };
  }

  /**
   * Start a Besu Blockchain service and retrieve the hardfork id for a given block range
   *
   * @param context the context on which to start the service
   * @param fromBlock the block number at which to retrieve the hardfork id
   * @param toBlock the block number at which to retrieve the hardfork id
   * @return Fork corresponding Fork instance if the hardfork id is the same between fromBlock and
   *     toBlock, else throw
   */
  public static Fork getForkFromBesuBlockchainService(
      ServiceManager context, long fromBlock, long toBlock) {
    HardforkId hardforkIdFromBlock =
        BesuServiceProvider.getBesuService(context, BlockchainService.class)
            .getHardforkId(fromBlock);
    if (fromBlock != toBlock) {
      HardforkId hardforkIdToBlock =
          BesuServiceProvider.getBesuService(context, BlockchainService.class)
              .getHardforkId(toBlock);
      if (!hardforkIdFromBlock.equals(hardforkIdToBlock)) {
        throw new IllegalArgumentException(
            "Fork change between blocks " + fromBlock + " and " + toBlock);
      }
    }
    return fromMainnetHardforkId((MainnetHardforkId) hardforkIdFromBlock);
  }

  /**
   * Start a Besu Blockchain service and retrieve the hardfork id for a given block number
   *
   * @param context the context on which to start the service
   * @param blockNumber the block number at which to retrieve the hardfork id
   * @return Fork corresponding Fork instance
   */
  public static Fork getForkFromBesuBlockchainService(ServiceManager context, long blockNumber) {
    return getForkFromBesuBlockchainService(context, blockNumber, blockNumber);
  }
}
