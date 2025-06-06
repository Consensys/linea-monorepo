/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.serialization.rlp

import maru.core.BeaconBlockBody
import maru.core.BeaconBlockHeader
import maru.core.BeaconState
import maru.core.HashUtil

fun HashUtil.headerHash(beaconBlockHeader: BeaconBlockHeader): ByteArray =
  rootHash(beaconBlockHeader, RLPSerializers.BeaconBlockHeaderSerializer, KeccakHasher)

fun HashUtil.bodyRoot(beaconBlockBody: BeaconBlockBody): ByteArray =
  rootHash(beaconBlockBody, RLPSerializers.BeaconBlockBodySerializer, KeccakHasher)

fun HashUtil.stateRoot(beaconState: BeaconState): ByteArray =
  rootHash(beaconState, RLPSerializers.BeaconStateSerializer, KeccakHasher)
