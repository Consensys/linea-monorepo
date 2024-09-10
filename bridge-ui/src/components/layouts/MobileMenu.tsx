import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { MENU_ITEMS } from "@/utils/constants";
import { MdOutlineClose } from "react-icons/md";

type MobileMenuProps = {
  toggleMenu: () => void;
};

export default function MobileMenu({ toggleMenu }: MobileMenuProps) {
  const pathname = usePathname();

  return (
    <div className="fixed inset-0 z-50 bg-[#121212] p-8 md:hidden">
      <div className="flex items-center justify-between">
        <Image src={"/images/logo/linea.svg"} alt="Linea logo" width={95} height={45} />
        <button onClick={toggleMenu} className="btn btn-circle btn-ghost">
          <MdOutlineClose size="2em" />
        </button>
      </div>
      <div className="mt-5 flex h-full flex-col justify-between overflow-y-auto py-4">
        <div>
          <ul className="space-y-2 font-medium">
            {MENU_ITEMS.map(({ title, href, external, Icon }) => (
              <li key={title}>
                {external ? (
                  <Link
                    href={href}
                    passHref
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center gap-2 py-4"
                    onClick={toggleMenu}
                  >
                    <Icon alt={title} width={30} height={30} />
                    <span>{title}</span>
                  </Link>
                ) : (
                  <Link
                    href={href}
                    className={`flex items-center gap-2 py-3 ${pathname === href ? "text-primary" : ""}`}
                    onClick={toggleMenu}
                  >
                    <Icon alt={title} width={30} height={30} />
                    <span>{title}</span>
                  </Link>
                )}
              </li>
            ))}
          </ul>
        </div>
        <div className="space-y-2 py-8 font-medium">
          <Link
            className="flex items-center hover:text-primary"
            href="#"
            passHref
            target="_blank"
            rel="noopener noreferrer"
            onClick={toggleMenu}
          >
            Contact Support
          </Link>
          <Link
            className="flex items-center hover:text-primary"
            href="https://linea.build/terms-of-service"
            passHref
            target="_blank"
            rel="noopener noreferrer"
            onClick={toggleMenu}
          >
            Terms of service
          </Link>
        </div>
      </div>
    </div>
  );
}
