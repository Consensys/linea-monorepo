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
package maru.executionlayer.manager

enum class ExecutionPayloadStatus(
  private val validity: Validity,
) {
  VALID(Validity.VALID),
  INVALID(Validity.INVALID),
  SYNCING(Validity.NOT_VALIDATED),
  ACCEPTED(Validity.NOT_VALIDATED),
  ;

  fun isValid(): Boolean = validity == Validity.VALID

  enum class Validity {
    VALID,
    NOT_VALIDATED,
    INVALID,
  }
}

data class PayloadStatus(
  val status: ExecutionPayloadStatus,
  val latestValidHash: ByteArray?,
  val validationError: String?,
) {
  override fun equals(other: Any?): Boolean {
    if (this === other) return true
    if (javaClass != other?.javaClass) return false

    other as PayloadStatus

    if (status != other.status) return false
    if (latestValidHash != null) {
      if (other.latestValidHash == null) return false
      if (!latestValidHash.contentEquals(other.latestValidHash)) return false
    } else if (other.latestValidHash != null) {
      return false
    }
    if (validationError != other.validationError) return false

    return true
  }

  override fun hashCode(): Int {
    var result = status.hashCode()
    result = 31 * result + (latestValidHash?.contentHashCode() ?: 0)
    result = 31 * result + (validationError?.hashCode() ?: 0)
    return result
  }
}
