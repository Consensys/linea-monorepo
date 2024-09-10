import { usePathname } from "next/navigation";
import { useAccount } from "wagmi";
import Image from "next/image";
import { useEffect, useState } from "react";
import { MdMenu } from "react-icons/md";
import Wallet from "./Wallet";
import Chains from "./Chains";
import MobileMenu from "../MobileMenu";

function formatPath(pathname: string): string {
  switch (pathname) {
    case "/":
    case "":
      return "Bridge";
    case "/transactions":
      return "Transactions";
    default:
      return "";
  }
}

export default function Header() {
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
    <header className="flex items-center justify-between gap-3 p-10">
      <Image src={"/images/logo/linea.svg"} alt="Linea logo" width={95} height={45} className="md:hidden" />
      <h1 className="hidden text-4xl md:flex">{formatPath(pathname)}</h1>
      <div className="flex items-center">
        <ul className="menu menu-horizontal m-0">
          {isConnected && (
            <li>
              <Chains />
            </li>
          )}
          <li>
            <Wallet />
          </li>
        </ul>
        <button onClick={toggleMenu} className="btn btn-circle btn-ghost md:hidden">
          <MdMenu size="2em" />
        </button>
        {isMenuOpen && <MobileMenu toggleMenu={toggleMenu} />}
      </div>
    </header>
  );
}
