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
import java.io.OutputStream;
import java.nio.ByteBuffer;
import java.util.List;
import java.util.Map;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.ObjectWriter;
import net.consensys.linea.zktracer.container.module.Module;

/** Provides a basic API for writing an LT trace file. */
public class LtTraceFile {
  private final Map<String, Object> metadata;
  private final Trace.ColumnHeader[] headers;
  private final Column[] columns;

  public LtTraceFile(Map<String, Object> metadata, Trace.ColumnHeader[] headers, Column[] columns) {
    this.metadata = metadata;
    this.headers = headers;
    this.columns = columns;
  }

  public static class Column implements Trace.Column {
    private final String name;
    private final int bitWidth;
    private final int byteWidth;
    private final long longMax;
    private final ByteBuffer buffer;

    public Column(String name, int bitwidth, int length) {
      this.name = name;
      this.bitWidth = bitwidth;
      this.byteWidth = byteWidth(bitwidth);
      this.longMax = 1L << bitwidth;
      this.buffer = ByteBuffer.allocate(length * byteWidth);
    }

    @Override
    public void write(boolean value) {
      this.buffer.put((byte) (value ? 1 : 0));
    }

    @Override
    public void write(long value) {
      // Sanity check
      if (longMax <= value) {
        throw new IllegalArgumentException(name + " has invalid value (" + value + ")");
      }
      //
      switch (byteWidth) {
        case 8:
          this.buffer.put((byte) (value >> 56));
        case 7:
          this.buffer.put((byte) (value >> 48));
        case 6:
          this.buffer.put((byte) (value >> 40));
        case 5:
          this.buffer.put((byte) (value >> 32));
        case 4:
          this.buffer.put((byte) (value >> 24));
        case 3:
          this.buffer.put((byte) (value >> 16));
        case 2:
          this.buffer.put((byte) (value >> 8));
        case 1:
          this.buffer.put((byte) value);
          break;
        default:
          throw new IllegalArgumentException(name + " has invalid width (" + byteWidth + "bytes)");
      }
    }

    /**
     * Write element bytes
     *
     * @param bytes stored in big-endian form and already trimmed.
     */
    @Override
    public void write(byte[] bytes) {
      final int n = bitLengthOf(bytes);
      // Sanity check
      if (n > bitWidth || bytes.length > byteWidth) {
        throw new IllegalArgumentException(name + " has invalid width (" + n + " bits)");
      }
      // Write padding (if necessary)
      for (int i = bytes.length; i < byteWidth; i++) {
        buffer.put((byte) 0);
      }
      // Write data
      for (int i = 0; i != bytes.length; i++) {
        buffer.put(bytes[i]);
      }
    }
  }

  /** Object writer is used for generating JSON byte strings. */
  private static final ObjectWriter objectWriter = new ObjectMapper().writer();

  /**
   * Construct a full in-memory trace of all columns across all modules.
   *
   * @param metadata
   * @param trace
   * @param modules
   * @return
   */
  public static LtTraceFile of(Map<String, Object> metadata, Trace trace, List<Module> modules) {
    // Determine set of all columns headers
    final List<Trace.ColumnHeader> rawHeaders =
        modules.stream().flatMap(m -> m.columnHeaders(trace).stream()).toList();
    // Align column headers
    final Trace.ColumnHeader[] headers = alignHeaders(rawHeaders);
    // Initialise all trace columns
    final Column[] columns = initialiseTraceColumns(trace, headers);
    // Open trace for writing
    trace.open(columns);
    // Commit each module
    for (Module m : modules) {
      m.commit(trace);
    }
    //
    return new LtTraceFile(metadata, headers, columns);
  }

  /**
   * Write a binary representation of the trace to a given output stream.
   *
   * @param out
   * @return
   */
  public void write(OutputStream out) throws IOException {
    // Write header for LTv1 file
    out.write(getHeaderBytes(getMetadataBytes(metadata)));
    out.write(getColumnHeaderBytes(headers));
    // Write column data
    for (int i = 0; i != columns.length; i++) {
      if (columns[i] != null) {
        out.write(columns[i].buffer.array());
      }
    }
    // Done
    out.flush();
  }

  /**
   * Align headers ensures that the order in which columns are seen matches the order found in the
   * trace schema.
   *
   * @param headers The headers to be aligned.
   * @return The aligned headers.
   */
  private static Trace.ColumnHeader[] alignHeaders(List<Trace.ColumnHeader> headers) {
    int maxRegister = 0;
    // Determine largest register
    //
    for (Trace.ColumnHeader header : headers) {
      maxRegister = Math.max(header.register(), maxRegister);
    }
    //
    Trace.ColumnHeader[] alignedHeaders = new Trace.ColumnHeader[maxRegister + 1];
    //
    for (Trace.ColumnHeader header : headers) {
      alignedHeaders[header.register()] = header;
    }
    //
    return alignedHeaders;
  }

  /**
   * Initialise the column buffers where all data will be written.
   *
   * @param trace
   * @param headers
   * @return
   */
  private static Column[] initialiseTraceColumns(Trace trace, Trace.ColumnHeader[] headers) {
    final Column[] columns = new Column[headers.length];

    for (int i = 0; i < headers.length; i++) {
      Trace.ColumnHeader header = headers[i];
      if (header != null) {
        columns[i] = new Column(header.name(), header.bitwidth(), header.length());
      }
    }

    return columns;
  }

  public static byte[] getMetadataBytes(Map<String, Object> metadata) throws IOException {
    return objectWriter.writeValueAsBytes(metadata);
  }

  /**
   * Construct trace file header containing the given metadata bytes.
   *
   * @param metadata Metadata bytes to be embedded in the trace file.
   * @return bytes making up the header.
   */
  private static byte[] getHeaderBytes(byte[] metadata) {
    ByteBuffer buffer = ByteBuffer.allocate(16 + metadata.length);
    // File identifier
    buffer.put(new byte[] {'z', 'k', 't', 'r', 'a', 'c', 'e', 'r'});
    // Major version
    buffer.putShort((short) 1);
    // Minor version
    buffer.putShort((short) 0);
    // Metadata length
    buffer.putInt(metadata.length);
    // Metadata
    buffer.put(metadata);
    // Done
    return buffer.array();
  }

  /**
   * Write header information for the trace file.
   *
   * @param headers Column headers.*
   */
  private static byte[] getColumnHeaderBytes(Trace.ColumnHeader[] headers) throws IOException {
    ByteBuffer buffer = ByteBuffer.allocate(getColumnHeadersSize(headers));
    // Write column count as uint32
    buffer.putInt(countHeaders(headers));
    // Write column headers one-by-one
    for (Trace.ColumnHeader h : headers) {
      if (h != null) {
        buffer.putShort((short) h.name().length());
        buffer.put(h.name().getBytes());
        buffer.put((byte) byteWidth(h.bitwidth()));
        buffer.putInt(h.length());
      }
    }
    //
    return buffer.array();
  }

  /**
   * Precompute the size of the trace file in order to memory map the buffers.
   *
   * @param headers Set of headers for the columns being written.
   * @return Number of bytes requires for the trace file header.
   */
  private static int getColumnHeadersSize(Trace.ColumnHeader[] headers) {
    int nBytes = 4; // column count

    for (Trace.ColumnHeader header : headers) {
      if (header != null) {
        nBytes += 2; // name length
        nBytes += header.name().length();
        nBytes += 1; // byte per element
        nBytes += 4; // element count
      }
    }

    return nBytes;
  }

  /**
   * Counter number of active (i.e. non-null) headers. A header can be null if it represents a
   * column in a module which is not activated for this trace.
   */
  private static int countHeaders(Trace.ColumnHeader[] headers) {
    int count = 0;
    for (Trace.ColumnHeader h : headers) {
      if (h != null) {
        count++;
      }
    }
    return count;
  }

  /**
   * Convert a given bitwidth into a bytewidth. For example, a bitwidth of 1 becomes a bytewidth of
   * 1 whilst a bitwidth of 9 becomes a bytewidth of 2, etc.
   *
   * @param bitwidth
   * @return
   */
  private static int byteWidth(int bitwidth) {
    int byteWidth = bitwidth / 8;
    //
    if ((bitwidth % 8) != 0) {
      byteWidth++;
    }
    //
    return byteWidth;
  }

  public static int bitLengthOf(byte[] bytes) {
    int n = 0;
    // Skip forward
    while (n < bytes.length && bytes[n] == 0) {
      n++;
    }
    // Determine width
    if (n == bytes.length) {
      return 0;
    } else {
      // n > 0
      int m = bytes.length - n - 1;
      return bitLengthOf(bytes[n]) + (m * 8);
    }
  }

  private static int bitLengthOf(byte b) {
    // Convert into unsigned representation
    int val = b & 0xff;
    // NOTE: we could further improve performance by turning this into one big lookup table.
    if (val >= 16) {
      return bits[val >> 4] + 4;
    } else {
      return bits[val];
    }
  }

  private static final int[] bits = {0, 1, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4, 4, 4, 4, 4};
}
