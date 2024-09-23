import { MENU_ITEMS } from "@/utils/constants";
import { MenuItem } from "./MenuItem";

type MenuProps = {
  toggleMenu?: () => void;
  border?: boolean;
};

export const Menu: React.FC<MenuProps> = ({ toggleMenu, border }) => {
  return (
    <ul className="space-y-2 font-medium">
      {MENU_ITEMS.map((item) => (
        <MenuItem key={item.title} {...item} toggleMenu={toggleMenu} border={border} />
      ))}
    </ul>
  );
};
