import { useWeb3Modal } from "@web3modal/wagmi/react";
import { cn } from "@/utils/cn";
import Button from "./bridge/Button";

type ConnectButtonProps = {
  fullWidth?: boolean;
};

export default function ConnectButton({ fullWidth }: ConnectButtonProps) {
  const { open } = useWeb3Modal();
  return (
    <Button
      id="wallet-connect-btn"
      variant="primary"
      size="md"
      className={cn({
        "w-full": fullWidth,
      })}
      onClick={() => open()}
    >
      Connect Wallet
    </Button>
  );
}
