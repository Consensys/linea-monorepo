import UnionIcon from "@/assets/icons/union.svg";
import LeftIllustration from "./illustration/left.svg";
import RightIllustration from "./illustration/right.svg";
import CloseIcon from "@/assets/icons/x-circle.svg";
import { useState } from "react";
import styles from "./top-banner.module.scss";

type Props = {
  text: string;
  href: string;
};

export default function TopBanner({ text, href }: Props) {
  const [showBanner, setShowBanner] = useState<boolean>(true);

  const handleClose = () => {
    setShowBanner(false);
  };

  if (!showBanner) return null;

  return (
    <div className={styles["banner-wrapper"]}>
      <LeftIllustration className={styles["left-illustration"]} />
      <div className={styles["banner"]}>
        <a href={href} target="_blank" rel="noopener noreferrer" className={styles["inner"]}>
          <span>{text}</span>
          <UnionIcon className={styles["external-icon"]} />
        </a>
      </div>
      <CloseIcon onClick={handleClose} className={styles["close-icon"]} />
      <RightIllustration className={styles["right-illustration"]} />
    </div>
  );
}
