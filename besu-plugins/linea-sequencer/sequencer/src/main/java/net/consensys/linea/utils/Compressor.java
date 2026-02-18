/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */

package net.consensys.linea.utils;

import linea.blob.GoBackedTxCompressor;
import linea.blob.TxCompressor;
import linea.blob.TxCompressorVersion;
import org.apache.logging.log4j.LogManager;

public class Compressor {
  public static TxCompressor instance;

  static {
    try {
      instance = GoBackedTxCompressor.getInstance(TxCompressorVersion.V2, 128 * 1024);
    } catch (Throwable t) {
      LogManager.getLogger(Compressor.class)
          .error("Failed to load GoBackedBlobCompressor. errorMessage={}", t.getMessage(), t);
      throw new RuntimeException("Failed to load GoBackedBlobCompressor", t);
    }
  }
}
