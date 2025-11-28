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

import java.io.IOException;
import java.nio.ByteBuffer;
import java.nio.channels.FileChannel;
import java.nio.file.Path;
import java.util.HashMap;
import java.util.Map;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

/** Provides a basic API for reading the metadata from an LT trace file. */
public class LtTraceFile implements AutoCloseable {

  /** Represents the header component of an LtTraceFile. */
  public static class Header {
    /** Identifier should be "zktracer". */
    byte[] identifier;

    /** Major version for LT tracefile format. */
    short majorVersion;

    /** Minor version for LT tracefile format. */
    short minorVersion;

    /** Metadata bytes for LT tracefile (encoded as JSON by default). */
    byte[] metadata;

    public Header(byte[] identifier, short majorVersion, short minorVersion, byte[] metadata) {
      this.identifier = identifier;
      this.majorVersion = majorVersion;
      this.minorVersion = minorVersion;
      this.metadata = metadata;
    }

    /**
     * Parse the metadata bytes into a map.
     *
     * @return Map representing the metadata
     */
    public Map<String, Object> getMetaData() throws IOException {
      JsonNode node = objectMapper.readTree(this.metadata);
      return parseJsonNode(node);
    }

    private static Map<String, Object> parseJsonNode(JsonNode node) {
      HashMap<String, Object> metadata = new HashMap<>();

      for (java.util.Map.Entry<String, JsonNode> entry : node.properties()) {
        JsonNode val = entry.getValue();
        Object obj;
        //
        if (val.isObject()) {
          obj = parseJsonNode(val);
        } else if (val.isTextual()) {
          obj = val.asText();
        } else if (val.isBoolean()) {
          obj = val.asBoolean();
        } else if (val.isInt()) {
          obj = val.asInt();
        } else if (val.isLong()) {
          obj = val.asLong();
        } else {
          // Fall back, including for null.
          obj = val.textValue();
        }
        //
        metadata.put(entry.getKey(), obj);
      }
      //
      return metadata;
    }
  }

  private final FileChannel ch;

  public LtTraceFile(Path path) throws IOException {
    this.ch = FileChannel.open(path);
  }

  public Header getHeader() throws IOException {
    return parseHeader(ch);
  }

  @Override
  public void close() throws IOException {
    this.ch.close();
  }

  private static Header parseHeader(FileChannel ch) throws IOException {
    ByteBuffer header = ByteBuffer.allocate(16);
    byte[] identifier = new byte[8];
    // Reset channel position to beginning of file.
    ch.position(0);
    // Reader header bytes
    int nBytes = ch.read(header);
    //
    if (nBytes != 16) {
      return null;
    }
    // Reset buffer position
    header.position(0);
    // Read identifier bytes
    header.get(identifier);
    // Read major version
    short majorVersion = header.getShort();
    // Read minor version
    short minorVersion = header.getShort();
    // Read metadata length
    int metadataLength = header.getInt();
    // read metadata bytes
    byte[] metadata = parseMetaDataBytes(ch, metadataLength);
    // Sanity check
    if (metadata == null) {
      return null;
    }
    //
    return new Header(identifier, majorVersion, minorVersion, metadata);
  }

  private static byte[] parseMetaDataBytes(FileChannel ch, int metadataLength) throws IOException {
    // read metadata bytes
    ByteBuffer metadata = ByteBuffer.allocate(metadataLength);
    byte[] metadataBytes = new byte[metadataLength];
    int nBytes = ch.read(metadata);
    metadata.position(0);
    // Sanity check
    if (nBytes != metadataLength) {
      return null;
    }
    // Looks ok.
    metadata.get(metadataBytes);
    // Done
    return metadataBytes;
  }

  /** Object mapper is used for parsing JSON byte strings. */
  private static final ObjectMapper objectMapper = new ObjectMapper();
}
