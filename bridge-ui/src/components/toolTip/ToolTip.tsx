import classNames from "classnames";
import React from "react";

export type ToolTipProps = {
  text: string;
  position?: "top" | "bottom";
  align?: "left" | "right" | "center";
  children: React.ReactNode;
  className?: string;
};

const ToolTip: React.FC<ToolTipProps> = ({ text, children, position = "top", align = "right", className }) => {
  return (
    <div className="group relative">
      {children}
      <div
        className={classNames(
          "absolute px-2.5 py-2 bg-[#1D1D1D] text-[#C0C0C0] normal-case font-normal text-xs md:text-[0.8125rem] border rounded-sm border-primary min-w-40",
          "group-hover:scale-100 scale-0 transition-all duration-200 ease-in-out z-10 opacity-0 group-hover:opacity-100",
          {
            "bottom-full": position === "top",
            "top-full": position === "bottom",
            "right-0": align === "left",
            "left-full": align === "right",
            "left-1/2 -translate-x-1/2": align === "center",
            "origin-bottom-left": position === "top" && align === "right",
            "origin-bottom-right": position === "top" && align === "left",
            "origin-top-left": position === "bottom" && align === "right",
            "origin-top-right": position === "bottom" && align === "left",
            "origin-bottom": position === "top" && align === "center",
            "origin-top": position === "bottom" && align === "center",
          },
          className,
        )}
      >
        {text}
      </div>
    </div>
  );
};

export default ToolTip;
