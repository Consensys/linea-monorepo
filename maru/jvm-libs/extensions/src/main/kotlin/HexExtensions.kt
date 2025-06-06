/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.extensions
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
import java.util.HexFormat

fun ByteArray.encodeHex(prefix: Boolean = true): String {
  val hexStr = HexFormat.of().formatHex(this)
  return if (prefix) {
    "0x$hexStr"
  } else {
    hexStr
  }
}

fun String.fromHexToByteArray(): ByteArray = HexFormat.of().parseHex(removePrefix("0x"))
