import Link from "next/link";
import styles from "./item.module.scss";

import BridgeIcon from "@/assets/icons/bridge.svg";
import TransactionIcon from "@/assets/icons/transaction.svg";
import clsx from "clsx";

type Props = {
  title: string;
  href: string;
  active: boolean;
};

const iconMap: Record<string, React.FC<React.SVGProps<SVGSVGElement>>> = {
  "/": BridgeIcon,
  "/transactions": TransactionIcon,
};

export default function NavItem({ title, href, active }: Props) {
  const Title = active ? "h1" : "span";
  const Icon = iconMap[href];

  return (
    <Link
      href={href}
      className={clsx(styles["nav"], {
        [styles["active"]]: active,
      })}
    >
      {Icon && <Icon className={styles["icon"]} />}
      <Title>{title}</Title>
    </Link>
  );
}
