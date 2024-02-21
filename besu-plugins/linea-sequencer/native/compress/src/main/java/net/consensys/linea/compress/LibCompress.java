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

import com.sun.jna.Native;

import java.io.File;
import java.io.InputStream;
import java.nio.file.Files;
import java.nio.file.Path;

/**
 * Java interface to compress
 */
public class LibCompress {

    @SuppressWarnings("WeakerAccess")
    public static final boolean ENABLED;

    static {
        boolean enabled;
        try {
            final File compressJni = Native.extractFromResourcePath("compress_jni");
            Native.register(LibCompress.class, compressJni.getAbsolutePath());

            Path dictFilePath = Files.createTempFile("tempCompressor_dict", "bin");
            try (InputStream stream = LibCompress.class.getClassLoader().getResourceAsStream("compressor_dict.bin")) {
                Files.copy(stream, dictFilePath, java.nio.file.StandardCopyOption.REPLACE_EXISTING);
                dictFilePath.toFile().deleteOnExit();
            } catch (Exception e) {
                System.out.println("Problem creating temp file from compressor_dict.bin resource: " + e.getMessage());
                dictFilePath.toFile().delete();
                System.exit(-1);
            }

            final String dictPath = dictFilePath.toAbsolutePath().toString();
            if (!Init(dictPath)) {
                throw new RuntimeException(Error());
            }
            enabled = true;
        } catch (final Throwable t) {
            t.printStackTrace();
            enabled = false;
        }
        ENABLED = enabled;
    }

    public static native boolean Init(String dictPath);
    
    public static native int CompressedSize(
            byte[] i, int i_len);

    public static native String Error();
}