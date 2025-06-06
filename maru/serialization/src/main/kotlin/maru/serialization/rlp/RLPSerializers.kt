/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.serialization.rlp

import maru.core.HashUtil

object RLPSerializers {
  val ValidatorSerializer = ValidatorSerializer()

  val BeaconBlockHeaderSerializer =
    BeaconBlockHeaderSerializer(
      validatorSerializer = ValidatorSerializer,
      hasher = KeccakHasher,
      headerHashFunction = HashUtil::headerHash,
    )
  val SealSerializer = SealSerializer()
  val ExecutionPayloadSerializer = ExecutionPayloadSerializer()
  val BeaconBlockBodySerializer =
    BeaconBlockBodySerializer(
      sealSerializer = SealSerializer,
      executionPayloadSerializer = ExecutionPayloadSerializer,
    )
  val BeaconBlockSerializer =
    BeaconBlockSerializer(
      beaconBlockHeaderSerializer = BeaconBlockHeaderSerializer,
      beaconBlockBodySerializer = BeaconBlockBodySerializer,
    )
  val SealedBeaconBlockSerializer =
    SealedBeaconBlockSerializer(
      beaconBlockSerializer = BeaconBlockSerializer,
      sealSerializer = SealSerializer,
    )
  val BeaconStateSerializer =
    BeaconStateSerializer(
      beaconBlockHeaderSerializer = BeaconBlockHeaderSerializer,
      validatorSerializer = ValidatorSerializer,
    )
  val DefaultHeaderHashFunction = HashUtil.headerHash(BeaconBlockHeaderSerializer, KeccakHasher)
}
