import Image from "next/image";
import styles from "./no-fees.module.scss";

type Props = {
  iconPath: string;
};

export default function NoFees({ iconPath }: Props) {
  return (
    <button type="button" className={styles["no-fees"]}>
      <Image src={iconPath} width={12} height={12} alt="no-fees-chain-icon" />
      <p className={styles["text"]}>No Fees</p>
    </button>
  );
}
