import { useMemo } from "react";
import { formatUnits } from "viem";

export function useFormattedDigit(value: bigint, decimals: number = 18): JSX.Element {
  return useMemo(() => {
    if (value <= 0n) return <span>0.00</span>;

    const valueStr = formatUnits(value, decimals);
    const num = Number(valueStr);

    if (!isFinite(num) || num === 0) return <span>0.00</span>;
    if (num >= 1e-8) return <span>{num.toFixed(8)}</span>;

    const match = /^0\.0+(?=\d)/.exec(valueStr);
    if (!match) return <span>{num.toFixed(8)}</span>;

    const zeroCount = match[0].length - 2;
    const remainderDigits = valueStr.slice(match[0].length);
    const rounded = Math.round(parseFloat("0." + remainderDigits) * 100)
      .toString()
      .padStart(2, "0");

    return (
      <span>
        0.0<sub>{zeroCount}</sub>
        {rounded}
      </span>
    );
  }, [value, decimals]);
}
