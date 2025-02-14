import clsx from "clsx";
import styles from "./item.module.scss";
import { formatAddress } from "@/utils/format";
import { TransactionStatus } from "@/components/transactions/TransactionItem";
import { Transaction as TransactionType } from "@/components/v2/transaction/list-transaction";
import CheckIcon from "@/assets/icons/check.svg";
import ClockIcon from "@/assets/icons/clock.svg";
import BridgeTwoLogo from "@/components/v2/bridge/bridge-two-logo";

type Props = TransactionType & {
  onClick: (code: string) => void;
};

export default function Transaction({ code, value, unit, date, status, estimatedTime, onClick }: Props) {
  const formatedCode = formatAddress(code);

  const renderStatus = () => {
    switch (status) {
      case TransactionStatus.COMPLETED:
        return (
          <>
            <CheckIcon />
            <div>Completed</div>
          </>
        );
      case TransactionStatus.READY_TO_CLAIM:
        return (
          <>
            <CheckIcon />
            <div>Ready to claim</div>
          </>
        );
      case TransactionStatus.PENDING:
        return (
          <>
            <ClockIcon />
            <div>{estimatedTime}</div>
          </>
        );

      default:
        return null;
    }
  };

  return (
    <li
      className={clsx(styles["transaction-item"], {
        [styles["completed"]]: status === TransactionStatus.COMPLETED,
        [styles["ready"]]: status === TransactionStatus.READY_TO_CLAIM,
        [styles["pending"]]: status === TransactionStatus.PENDING,
      })}
      onClick={() => onClick(code)}
    >
      <div className={styles["left"]}>
        <div className={styles["image-wrapper"]}>
          <BridgeTwoLogo
            src1="/images/logo/ethereum-rounded.svg"
            src2="/images/logo/linea-rounded.svg"
            alt1="eth"
            alt2="linea"
          />
        </div>
        <div className={styles["info"]}>
          <span className={styles["code"]} data-original-code={code}>
            {formatedCode}
          </span>
          <span className={styles["date"]}>{date}</span>
        </div>
      </div>
      <div className={styles["right"]}>
        <div className={styles["value-wrapper"]}>
          <span className={styles["value"]}>{value} &nbsp;</span>
          <span className={styles["unit"]}>{unit}</span>
        </div>
        <div className={styles["status"]}>{renderStatus()}</div>
      </div>
    </li>
  );
}
