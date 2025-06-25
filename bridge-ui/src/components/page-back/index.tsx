import Link from "next/link";
import ArrowLeftIcon from "@/assets/icons/arrow-left.svg";
import styles from "./page-back.module.scss";

type Props = {
  label?: string;
  href?: string;
};

export default function PageBack({ label = "Back", href = "/" }: Props) {
  return (
    <div className={styles.wrapper}>
      <Link className={styles.back} href={href}>
        <ArrowLeftIcon />
        <span>{label}</span>
      </Link>
    </div>
  );
}
