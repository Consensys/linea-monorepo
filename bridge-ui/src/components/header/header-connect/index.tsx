"use client";
import Button from "@/components/ui/button";
import { useModal } from "@/contexts/ModalProvider";
import { useAccount, useConnect, useConnectors } from "wagmi";
import UserAvatar from "@/components/user-avatar";
import { usePrefetchPoh } from "@/hooks/useCheckPoh";
import { useEffect } from "react";
import { useEnsInfo } from "@/hooks/user/useEnsInfo";

export default function HeaderConnect() {
  const { isConnected, address } = useAccount();
  const { connectAsync, isPending: isConnecting } = useConnect();
  const { ensAvatar } = useEnsInfo();
  const { updateModal } = useModal();
  const connectors = useConnectors();
  const prefetchPoh = usePrefetchPoh();

  const handleConnectionToggle = async () => {
    if (isConnected) {
      updateModal(true, "user-wallet");
    } else {
      const web3authConnector = connectors.find((c) => c.id === "web3auth");
      if (web3authConnector) {
        await connectAsync({ connector: web3authConnector });
      }
    }
  };

  useEffect(() => {
    if (address) {
      prefetchPoh(address);
    }
  }, [address, prefetchPoh]);

  if (isConnected) return <UserAvatar src={ensAvatar ?? ""} size="small" onClick={handleConnectionToggle} />;

  return (
    <Button variant="primary" fullWidth disabled={isConnecting} onClick={handleConnectionToggle}>
      Connect
    </Button>
  );
}
