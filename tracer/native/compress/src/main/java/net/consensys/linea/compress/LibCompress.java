/*
 * Copyright Consensys Software Inc.
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
 *
 */
package net.consensys.linea.compress;

import java.io.File;
import java.io.InputStream;
import java.nio.file.Files;
import java.nio.file.Path;

import com.sun.jna.Native;
import lombok.extern.slf4j.Slf4j;

/** Java interface to compress */
@Slf4j
public class LibCompress {

  @SuppressWarnings("WeakerAccess")
  public static final boolean ENABLED;

  static {
    try {
      final File compressJni =
          Native.extractFromResourcePath("compress_jni", LibCompress.class.getClassLoader());
      Native.register(LibCompress.class, compressJni.getAbsolutePath());

      Path dictFilePath = Files.createTempFile("tempCompressor_dict", "bin");
      try (InputStream stream =
          LibCompress.class.getClassLoader().getResourceAsStream("compressor_dict.bin")) {
        Files.copy(stream, dictFilePath, java.nio.file.StandardCopyOption.REPLACE_EXISTING);
        dictFilePath.toFile().deleteOnExit();
      } catch (Exception e) {
        log.error(
            "Problem creating temp file from compressor_dict.bin resource: " + e.getMessage());
        dictFilePath.toFile().delete();
        System.exit(1);
      }

      final String dictPath = dictFilePath.toAbsolutePath().toString();
      if (!Init(dictPath)) {
        throw new RuntimeException(Error());
      }
      log.info(
          "Loaded compress_jni native library from {} with compressor_dict {}",
          compressJni,
          dictPath);
    } catch (final Throwable t) {
      log.error("Error loading native compress_jni library", t);
      System.exit(1);
    }
    ENABLED = true;
  }

  public static native boolean Init(String dictPath);

  public static native int CompressedSize(byte[] i, int i_len);

  public static native String Error();
}
