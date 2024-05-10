import { useEffect, useState } from 'react';
import { useAccount } from 'wagmi';

const useIsConnected = () => {
  const { isConnected } = useAccount();
  const [isMounted, setIsMounted] = useState(false);
  const [isConnectedState, setIsConnectedState] = useState(isConnected);

  useEffect(() => {
    setIsMounted(true);
    setIsConnectedState(isConnected);
  }, [isConnected]);

  return isMounted ? isConnectedState : undefined;
};

export default useIsConnected;
