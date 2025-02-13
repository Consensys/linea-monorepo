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
package maru.serialization.rlp

object RLPSerializers {
  val ValidatorSerializer = ValidatorSerializer()
  val BeaconBlockHeaderSerializer =
    BeaconBlockHeaderSerializer(
      validatorSerializer = ValidatorSerializer,
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
  val BeaconStateSerializer =
    BeaconStateSerializer(
      beaconBlockHeaderSerializer = BeaconBlockHeaderSerializer,
      validatorSerializer = ValidatorSerializer,
    )
}
