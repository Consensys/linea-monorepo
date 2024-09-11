import { usePathname } from "next/navigation";
import Link from "next/link";
import Image from "next/image";
import { MENU_ITEMS } from "@/utils/constants";

export default function Sidebar() {
  const pathname = usePathname();

  return (
    <aside id="sidebar" className="fixed left-0 top-0 z-40 hidden h-screen w-52 md:block" aria-label="Sidebar">
      <div className="flex h-full flex-col justify-between overflow-y-auto bg-cardBg py-4">
        <div>
          <div className="flex h-24 items-center p-4">
            <Link href="/">
              <Image src={"/images/logo/linea.svg"} alt="Linea logo" width={95} height={45} />
            </Link>
          </div>
          <ul className="space-y-2 font-medium">
            {MENU_ITEMS.map(({ title, href, external, Icon }) => (
              <li key={title}>
                {external ? (
                  <Link
                    href={href}
                    passHref
                    target="_blank"
                    rel="noopener noreferrer"
                    className="flex items-center gap-2 p-4"
                  >
                    <Icon alt={title} width={30} height={30} />
                    <span>{title}</span>
                  </Link>
                ) : (
                  <Link
                    href={href}
                    className={`flex items-center gap-2 p-3 ${pathname === href ? "border-r-2 border-primary text-primary" : ""}`}
                  >
                    <Icon alt={title} width={30} height={30} />
                    <span>{title}</span>
                  </Link>
                )}
              </li>
            ))}
          </ul>
        </div>
        <div className="space-y-2 p-4 font-medium">
          <Link
            className="flex items-center hover:text-primary"
            href="#"
            passHref
            target="_blank"
            rel="noopener noreferrer"
          >
            Contact Support
          </Link>
          <Link
            className="flex items-center hover:text-primary"
            href="https://linea.build/terms-of-service"
            passHref
            target="_blank"
            rel="noopener noreferrer"
          >
            Terms of service
          </Link>
        </div>
      </div>
    </aside>
  );
}
