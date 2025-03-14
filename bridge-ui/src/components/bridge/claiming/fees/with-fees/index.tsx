import Image from "next/image";
import { formatUnits } from "viem";
import styles from "./with-fees.module.scss";
import { useState } from "react";
import GasFees from "../../../modal/gas-fees";
import { useFees } from "@/hooks";
import { useConfigStore, useFormStore } from "@/stores";
import { BridgeType } from "@/types";

type Props = {
  iconPath: string;
};

export default function WithFees({ iconPath }: Props) {
  const [showGasFeesModal, setShowGasFeesModal] = useState<boolean>(false);
  const currency = useConfigStore.useCurrency();

  const mode = useFormStore((state) => state.mode);
  const token = useFormStore((state) => state.token);

  const { total, fees, isLoading } = useFees();

  const handleShowFees = () => {
    if (mode === BridgeType.NATIVE) {
      setShowGasFeesModal(true);
    }
  };

  if (isLoading) {
    return null;
  }

  return (
    <>
      {total && (
        <button type="button" className={styles["gas-fees"]} onClick={handleShowFees}>
          <Image src={iconPath} width={12} height={12} alt="fee-chain-icon" />
          <p
            className={styles["estimate-crypto"]}
          >{`${Number(formatUnits(total.fees, token.decimals)).toFixed(8)} ${token.symbol}`}</p>
          {total.fiatValue && (
            <p className={styles["estimate-amount"]}>{`(${total.fiatValue.toLocaleString("en-US", {
              style: "currency",
              currency: currency.label,
              maximumFractionDigits: 2,
            })})`}</p>
          )}
        </button>
      )}
      <GasFees isModalOpen={showGasFeesModal} onCloseModal={() => setShowGasFeesModal(false)} fees={fees} />
    </>
  );
}
