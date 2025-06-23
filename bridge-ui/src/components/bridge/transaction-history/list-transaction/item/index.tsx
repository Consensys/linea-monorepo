import clsx from "clsx";
import { formatUnits } from "viem";
import styles from "./item.module.scss";
import CheckIcon from "@/assets/icons/check.svg";
import ClockIcon from "@/assets/icons/clock.svg";
import BridgeTwoLogo from "@/components/bridge/bridge-two-logo";
import { getChainLogoPath, formatHex, formatTimestamp, getEstimatedTimeText } from "@/utils";
import { BridgeTransaction, TransactionStatus } from "@/types";

type Props = BridgeTransaction & {
  onClick: (code: string) => void;
};

export default function Transaction({
  bridgingTx,
  status,
  fromChain,
  toChain,
  timestamp,
  message,
  token,
  onClick,
}: Props) {
  const formatedTxHash = formatHex(bridgingTx);
  const estimatedTimeText = getEstimatedTimeText(fromChain, token, {
    withSpaceAroundHyphen: true,
    isAbbreviatedTimeUnit: true,
  });

  const renderStatus = () => {
    switch (status) {
      case TransactionStatus.COMPLETED:
        return (
          <>
            <CheckIcon />
            <span>Completed</span>
          </>
        );
      case TransactionStatus.READY_TO_CLAIM:
        return (
          <>
            <CheckIcon />
            <span>Ready to claim</span>
          </>
        );
      case TransactionStatus.PENDING:
        return (
          <>
            <ClockIcon />
            <span>{estimatedTimeText}</span>
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
      onClick={() => onClick(bridgingTx)}
    >
      <div className={styles["left"]}>
        <div className={styles["image-wrapper"]}>
          <BridgeTwoLogo
            src1={getChainLogoPath(fromChain.id)}
            src2={getChainLogoPath(toChain.id)}
            alt1={fromChain.id.toString()}
            alt2={toChain.id.toString()}
          />
        </div>
        <div className={styles["info"]}>
          <span className={styles["code"]} data-original-code={formatedTxHash}>
            {formatedTxHash}
          </span>
          <span className={styles["date"]}>{formatTimestamp(Number(timestamp), "MMM, dd, yyyy")}</span>
        </div>
      </div>
      <div className={styles["right"]}>
        <div className={styles["value-wrapper"]}>
          <span className={styles["value"]}>{formatUnits(message.amountSent, token.decimals)} &nbsp;</span>
          <span className={styles["unit"]}>{token.symbol}</span>
        </div>
        <div className={styles["status"]}>{renderStatus()}</div>
      </div>
    </li>
  );
}
