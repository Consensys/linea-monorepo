import { useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import { useAccount } from "wagmi";
import { MdMenu } from "react-icons/md";
import MobileMenu from "../MobileMenu";
import Button from "@/components/bridge/Button";
import { HeaderLogo } from "./HeaderLogo";
import { NavMenu } from "./NavMenu";

export function Header() {
  // Hooks
  const { isConnected } = useAccount();
  const pathname = usePathname();
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  const toggleMenu = () => {
    setIsMenuOpen(!isMenuOpen);
  };

  useEffect(() => {
    if (isMenuOpen) {
      document.body.classList.add("overflow-hidden");
    } else {
      document.body.classList.remove("overflow-hidden");
    }
  }, [isMenuOpen]);

  return (
    <header className="navbar flex items-center justify-between gap-3 p-3 md:p-10">
      <HeaderLogo pathname={pathname} />
      <NavMenu isConnected={isConnected} />

      <div className="flex items-center">
        <div className="m-0"></div>
        <Button onClick={toggleMenu} className="btn-circle btn-ghost md:hidden">
          <MdMenu size="2em" />
        </Button>
        {isMenuOpen && <MobileMenu toggleMenu={toggleMenu} />}
      </div>
    </header>
  );
}
