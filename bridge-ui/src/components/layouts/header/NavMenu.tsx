import { Chains, Wallet } from "./dropdowns";

type NavMenuProps = {
  isConnected: boolean;
};

export const NavMenu: React.FC<NavMenuProps> = ({ isConnected }) => (
  <div className="flex-none">
    <ul className="menu menu-horizontal gap-2 px-1">
      {isConnected && (
        <li>
          <Chains />
        </li>
      )}
      <li>
        <Wallet />
      </li>
    </ul>
  </div>
);
