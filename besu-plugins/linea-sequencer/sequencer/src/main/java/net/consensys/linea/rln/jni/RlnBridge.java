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
 */
package net.consensys.linea.rln.jni;

import java.nio.file.Path;
import java.nio.file.Paths;

public class RlnBridge {

  private static final String LIBRARY_NAME = "rln_bridge";

  static {
    boolean loaded = false;
    try {
      // Try loading using the library name (requires library path setup via JVM args or env vars)
      System.loadLibrary(LIBRARY_NAME);
      System.out.println(
          "Native library '" + LIBRARY_NAME + "' loaded successfully via System.loadLibrary.");
      loaded = true;
    } catch (UnsatisfiedLinkError e) {
      System.err.println(
          "Failed to load '" + LIBRARY_NAME + "' using System.loadLibrary: " + e.getMessage());
      System.err.println(
          "Attempting to load using calculated relative path (ensure Rust code is compiled and in the expected path)...");

      String libDirGuess1 =
          Paths.get("src", "main", "rust", "rln_bridge", "target", "release")
              .toAbsolutePath()
              .toString();
      String libDirGuess2 =
          Paths.get("build", "native", "release")
              .toAbsolutePath()
              .toString(); // Common for Gradle native builds

      String mappedLibName = System.mapLibraryName(LIBRARY_NAME);
      Path libPath1 = Paths.get(libDirGuess1, mappedLibName);
      Path libPath2 = Paths.get(libDirGuess2, mappedLibName);

      try {
        System.load(libPath1.toString());
        System.out.println("Native library loaded successfully via calculated path: " + libPath1);
        loaded = true;
      } catch (UnsatisfiedLinkError e2) {
        System.err.println("Failed to load native library using path: " + libPath1);
        System.err.println("  Trying alternative path: " + libPath2);
        try {
          System.load(libPath2.toString());
          System.out.println("Native library loaded successfully via calculated path: " + libPath2);
          loaded = true;
        } catch (UnsatisfiedLinkError e3) {
          System.err.println("FATAL: Failed to load native library using path: " + libPath2);
          System.err.println("  System.loadLibrary error: " + e.getMessage());
          System.err.println("  System.load error (path 1): " + e2.getMessage());
          System.err.println("  System.load error (path 2): " + e3.getMessage());
          System.err.println("\nPlease ensure:");
          System.err.println(
              "  1. You have compiled the Rust code (e.g., `cd sequencer/src/main/rust/rln_bridge && cargo build --release`)");
          System.err.println(
              "  2. The library file ('"
                  + mappedLibName
                  + "') exists at one of the attempted paths or is in java.library.path.");
          System.err.println("  3. The library architecture matches your JVM architecture.");
          System.err.println(
              "  4. If using a build system like Gradle, ensure it correctly packages/locates the native library.");
          // Re-throw the last error to prevent execution if not loaded
          throw e3;
        }
      }
    }
    if (!loaded) {
      // Should be unreachable if throw e3 is active
      throw new UnsatisfiedLinkError(
          "Could not load native library '" + LIBRARY_NAME + "' by any method.");
    }
  }

  /**
   * Verifies an RLN proof using the native Rust implementation.
   *
   * @param verifyingKeyBytes The serialized verifying key.
   * @param proofBytes The serialized proof.
   * @param publicInputsHex Array of public input hex strings in order: [share_x, share_y, epoch,
   *     root, nullifier].
   * @return true if the proof is valid, false otherwise.
   * @throws RuntimeException if the native code encounters an error or panic (relayed from Rust),
   *     or if there's an issue with JNI argument marshalling.
   */
  public static native boolean verifyRlnProof(
      byte[] verifyingKeyBytes,
      byte[] proofBytes,
      String[] publicInputsHex // [share_x_hex, share_y_hex, epoch_hex, root_hex, nullifier_hex]
      );

  /**
   * Parses and verifies an RLN proof from the combined format used by the RLN prover service.
   *
   * @param verifyingKeyBytes The serialized verifying key.
   * @param combinedProofBytes The combined proof data (proof + proof values serialized together).
   * @param currentEpochHex The current epoch identifier as hex string.
   * @return String array containing [share_x, share_y, epoch, root, nullifier, verification_result]
   *     if successful, null if verification fails.
   * @throws RuntimeException if the native code encounters an error or panic (relayed from Rust),
   *     or if there's an issue with JNI argument marshalling.
   */
  public static native String[] parseAndVerifyRlnProof(
      byte[] verifyingKeyBytes, byte[] combinedProofBytes, String currentEpochHex);
}
