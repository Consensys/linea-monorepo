import React from "react";
import { cn } from "@/utils/cn";

interface TooltipProps {
  text: string;
  position?: "top" | "right" | "bottom" | "left";
  children: React.ReactNode;
  className?: string;
}

const Tooltip: React.FC<TooltipProps> = ({ text, position = "top", children, className }) => (
  <div
    className={cn(
      "tooltip",
      {
        "tooltip-top": position === "top",
        "tooltip-right": position === "right",
        "tooltip-bottom": position === "bottom",
        "tooltip-left": position === "left",
      },
      className,
    )}
    data-tip={text}
  >
    {children}
  </div>
);

export default Tooltip;
