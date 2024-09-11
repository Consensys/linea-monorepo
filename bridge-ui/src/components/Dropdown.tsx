import Image from "next/image";
import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import { MdCallMade } from "react-icons/md";
import { cn } from "@/utils/cn";

export type DropdownItem = {
  title: string;
  iconPath: string;
  onClick?: () => Promise<void>;
  externalLink?: string;
};

type DropdownProps = {
  initialValue?: DropdownItem;
  outline?: boolean;
  showDropdownToggle?: boolean;
  items: DropdownItem[];
};

export default function Dropdown({ items, initialValue, outline = false, showDropdownToggle = true }: DropdownProps) {
  const ref = useRef<HTMLDetailsElement>(null);
  const [selectedItem, setSelectedItem] = useState<DropdownItem | null>(null);

  useEffect(() => {
    setSelectedItem(initialValue || null);
  }, [initialValue]);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        ref.current.removeAttribute("open");
      }
    };
    document.addEventListener("click", handleClickOutside);
    return () => {
      document.removeEventListener("click", handleClickOutside);
    };
  }, []);

  const handleItemClick = async (item: DropdownItem) => {
    setSelectedItem(item);
    if (item.onClick) {
      await item.onClick();
    }
    ref.current?.removeAttribute("open");
  };

  return (
    <details className="dropdown relative" ref={ref}>
      <summary
        className={cn("flex cursor-pointer items-center gap-2 rounded-full p-2 px-3", {
          "border-2 border-card": outline,
          "bg-[#2D2D2D] text-white": !outline,
        })}
      >
        {selectedItem ? (
          <>
            {selectedItem.iconPath && (
              <Image
                src={selectedItem.iconPath}
                alt="MetaMask"
                width={0}
                height={0}
                style={{ width: "18px", height: "auto" }}
              />
            )}
            <span>{selectedItem.title}</span>
          </>
        ) : (
          <span>Select an option</span>
        )}
        {showDropdownToggle && (
          <svg
            className="ml-1 size-4 text-card transition-transform"
            fill="none"
            stroke={"white"}
            viewBox="0 0 24 24"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="3" d="M19 9l-7 7-7-7"></path>
          </svg>
        )}
      </summary>
      <ul className="menu dropdown-content absolute right-0 z-10 mt-2 min-w-max border-2 border-card bg-cardBg p-0 shadow">
        {items.map((item, index) => (
          <li key={`dropdown-${item}-${index}`} className="w-full">
            {item.externalLink ? (
              <Link
                href={item.externalLink}
                passHref
                target="_blank"
                rel="noopener noreferrer"
                className="btn btn-md flex justify-start rounded-none border-none bg-cardBg text-white"
              >
                {item.iconPath && (
                  <Image
                    src={item.iconPath}
                    alt="MetaMask"
                    width={0}
                    height={0}
                    style={{ width: "18px", height: "auto" }}
                  />
                )}
                {item.title}
                <MdCallMade />
              </Link>
            ) : (
              <button
                className={cn("btn btn-md flex justify-start border-none bg-cardBg rounded-none", {
                  "text-white": !outline,
                })}
                onClick={() => handleItemClick(item)}
              >
                {item.iconPath && (
                  <Image
                    src={item.iconPath}
                    alt="MetaMask"
                    width={0}
                    height={0}
                    style={{ width: "18px", height: "auto" }}
                  />
                )}
                {item.title}
              </button>
            )}
          </li>
        ))}
      </ul>
      <style>{`
        details[open] summary svg {
          transform: rotate(180deg);
        }
      `}</style>
    </details>
  );
}
