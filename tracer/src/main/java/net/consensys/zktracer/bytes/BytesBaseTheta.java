package net.consensys.zktracer.bytes;

import net.consensys.zktracer.module.alu.mul.Res;
import org.apache.tuweni.bytes.Bytes32;

import java.math.BigInteger;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.util.Arrays;

public class BytesBaseTheta {

        private byte[][] bytes;

        public BytesBaseTheta(final Bytes32 arg) {
            bytes = new byte[4][8];
            byte[] argBytes = arg.toArray();

            for (int k = 0; k < 4; k++) {
                System.arraycopy(argBytes, 8 * k, bytes[3 - k], 0, 8);
            }
        }

    public BytesBaseTheta(final Res res) {
        bytes = new byte[4][8];
        byte[] argBytesHi = res.getResHi().toArray();
        byte[] argBytesLo = res.getResLo().toArray();

        for (int k = 0; k < 2; k++) {
            System.arraycopy(argBytesHi, 8 * k, bytes[3 - k], 0, 8);
        }
        for (int k = 2; k < 4; k++) {
            System.arraycopy(argBytesLo, 8 * k, bytes[3 - k], 0, 8);
        }
    }

    // TODO can Res become Pair<Bytes16, Bytes16> as below
        public Pair<byte[], byte[]> getHiLo() {
            byte[] hiBytes = new byte[16];
            byte[] loBytes = new byte[16];

            System.arraycopy(bytes[3], 0, hiBytes, 0, 8);
            System.arraycopy(bytes[2], 0, hiBytes, 8, 8);

            System.arraycopy(bytes[1], 0, loBytes, 0, 8);
            System.arraycopy(bytes[0], 0, loBytes, 8, 8);

            return new Pair<>(hiBytes, loBytes);
        }

    public byte get(final int i, final int j) {
            return bytes[i][j];
    }
    public byte[] getRange(final int i, final int start, final int end) {
        return Arrays.copyOfRange(bytes[i], start, end);
    }
}

@SuppressWarnings("UnusedVariable")
record Pair<A, B>(A first, B second) {
}

    class UInt256 {
        private byte[] bytes;

        public UInt256(byte[] bytes) {
            this.bytes = bytes;
        }

        public byte[] getBytes32() {
            ByteBuffer buf = ByteBuffer.allocate(32);
            buf.order(ByteOrder.BIG_ENDIAN);
            buf.put(bytes);
            return buf.array();
        }
    }

