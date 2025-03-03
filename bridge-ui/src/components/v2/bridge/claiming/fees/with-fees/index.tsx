import Image from "next/image";
import styles from "./with-fees.module.scss";
import { useState } from "react";
import { BridgeType } from "@/config/config";
import GasFees from "../../../modal/gas-fees";
import AcrossFees from "../../../modal/across-fees";
import { formatUnits } from "viem";
import useFees from "@/hooks/fees/useFees";
import { useConfigStore } from "@/stores/configStore";
import { useFormStore } from "@/stores/formStoreProvider";

type Props = {
  iconPath: string;
};

export default function WithFees({ iconPath }: Props) {
  const [showGasFeesModal, setShowGasFeesModal] = useState<boolean>(false);
  const [showAcrossFeesModal, setShowAcrossFeesModal] = useState<boolean>(false);
  const currency = useConfigStore.useCurrency();

  const mode = useFormStore((state) => state.mode);
  const token = useFormStore((state) => state.token);

  const { total, fees } = useFees();

  const handleShowFees = () => {
    if (mode === BridgeType.NATIVE) {
      setShowGasFeesModal(true);
    } else if (mode === BridgeType.ACROSS) {
      setShowAcrossFeesModal(true);
    }
  };

  return (
    <>
      {total && (
        <button type="button" className={styles["gas-fees"]} onClick={handleShowFees}>
          <Image src={iconPath} width={12} height={12} alt="fee-chain-icon" />
          <p className={styles["estimate-crypto"]}>{`${formatUnits(total.fees, token.decimals)} ${token.symbol}`}</p>
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
      <AcrossFees isModalOpen={showAcrossFeesModal} onCloseModal={() => setShowAcrossFeesModal(false)} />
    </>
  );
}
