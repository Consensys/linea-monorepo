import { useAppKit } from "@reown/appkit/react";
import styles from "./connect-button.module.scss";
import clsx from "clsx";
import Button from "@/components/v2/ui/button";

type ConnectButtonProps = {
  text: string;
  fullWidth?: boolean;
};

export default function ConnectButton({ text, fullWidth }: ConnectButtonProps) {
  const { open } = useAppKit();

  return (
    <Button
      className={clsx(styles["connect-btn"], {
        [styles["full-width"]]: fullWidth,
      })}
      onClick={() => open()}
    >
      {text}
    </Button>
  );
}
