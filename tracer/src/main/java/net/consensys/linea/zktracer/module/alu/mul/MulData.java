package net.consensys.linea.zktracer.module.alu.mul;

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.bytes.BytesBaseTheta;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;

import java.math.BigInteger;

@SuppressWarnings("UnusedVariable")
public class MulData {
    final OpCode opCode;
    final boolean tinyBase;
    final boolean tinyExponent;

    final BytesBaseTheta aBytes;
    final BytesBaseTheta bBytes;
    BytesBaseTheta cBytes;
    BytesBaseTheta hBytes;
    boolean snm = false;
    Res res;
    public MulData(OpCode opCode, Bytes32 arg1, Bytes32 arg2) {

        this.opCode = opCode;
        this.aBytes = new BytesBaseTheta(arg1);
        this.bBytes = new BytesBaseTheta(arg2);
        this.cBytes = null;
        this.hBytes = null;
        boolean snm = false;

        this.res = Res.create(opCode, arg1, arg2); // TODO can we get this from the EVM


        final UInt256 arg1Int = UInt256.fromBytes(arg1);
        final UInt256 arg2Int = UInt256.fromBytes(arg2);
        final BigInteger arg1BigInt = arg1Int.toUnsignedBigInteger();
        final BigInteger arg2BigInt = arg2Int.toUnsignedBigInteger();

        this.tinyBase = isTiny(arg1BigInt);
        this.tinyExponent = isTiny(arg2BigInt);

        final Regime regime = getRegime(opCode);
        System.out.println(regime);
        switch (regime) {
            case TRIVIAL_MUL:
                break;
            case NON_TRIVIAL_MUL:
                cBytes = new BytesBaseTheta(res);
                break;
            case EXPONENT_ZERO_RESULT:
                setArraysForZeroResultCase();
                break;
            case EXPONENT_NON_ZERO_RESULT:
                setExponentBit();
                snm = false;
                break;
            case IOTA:
                throw new RuntimeException("alu/mul regime was never set");
        }
    }

    private void setArraysForZeroResultCase() {
        // TODO
    }

    private boolean setExponentBit() {
        // TODO
        return false;
        //    return string(exponentBits[md.index]) == "1";
    }

    private enum Regime {
        IOTA,
        TRIVIAL_MUL,
        NON_TRIVIAL_MUL,
        EXPONENT_ZERO_RESULT,
        EXPONENT_NON_ZERO_RESULT
    }

    public boolean isOneLineInstruction() {
        return tinyBase || tinyExponent;
    }

    private Regime getRegime(
            final OpCode opCode) {

        if (isOneLineInstruction()) return Regime.TRIVIAL_MUL;

        if (OpCode.MUL.equals(opCode)) {
            return Regime.NON_TRIVIAL_MUL;
        }

        if (OpCode.EXP.equals(opCode)) {
            if (res.isZero()) {
                return Regime.EXPONENT_ZERO_RESULT;
            } else {
                return Regime.EXPONENT_NON_ZERO_RESULT;
            }
        }
        return Regime.IOTA;
    }
    public static boolean isTiny(BigInteger arg) {
        return arg.compareTo(BigInteger.valueOf(1)) <= 0;
    }
}
