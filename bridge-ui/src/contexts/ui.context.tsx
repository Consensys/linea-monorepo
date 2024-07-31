import React, { createContext, useState, useCallback } from "react";

interface UIContextProps {
  showBridge: boolean;
  toggleShowBridge: (value?: boolean) => void;
}

export const UIContext = createContext<UIContextProps>({
  showBridge: false,
  toggleShowBridge: () => {
    return;
  },
});

type Props = {
  children: React.ReactNode;
};

export const UIProvider: React.FC<Props> = ({ children }) => {
  const [showBridge, setShowBridge] = useState(false);

  const toggleShowBridge = useCallback((value?: boolean) => {
    if (typeof value === "boolean") {
      return setShowBridge(value);
    }
    setShowBridge((prevShowBridge) => !prevShowBridge);
  }, []);

  return <UIContext.Provider value={{ showBridge, toggleShowBridge }}>{children}</UIContext.Provider>;
};
