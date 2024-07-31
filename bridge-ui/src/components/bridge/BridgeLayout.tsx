"use client";

import BridgeUI from "@/components/bridge/BridgeUI";
import Hero from "@/components/hero/Hero";
import { UIContext } from "@/contexts/ui.context";
import { Shortcut } from "@/models/shortcut";
import { createContext, useContext } from "react";

const ShortcutContext = createContext<Shortcut[]>([]);

export function useShortcuts() {
  return useContext(ShortcutContext);
}

export default function BridgeLayout({ shortcuts }: { shortcuts: Shortcut[] }) {
  const { showBridge } = useContext(UIContext);

  if (showBridge) {
    return <BridgeUI />;
  }

  return (
    <ShortcutContext.Provider value={shortcuts}>
      <Hero />
    </ShortcutContext.Provider>
  );
}
