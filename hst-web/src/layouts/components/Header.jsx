import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import {
  LuMenu,
  LuMoon,
  LuSun,
  LuLogOut,
  LuLanguages,
  LuPanelLeft,
  LuPanelBottom,
  LuPanelRight,
  LuLayoutDashboard,
  LuSquare,
  LuColumns2,
  LuRows2,
  LuLayoutGrid,
} from "react-icons/lu";
import { MdAdsClick } from "react-icons/md";
import { Dropdown } from "antd";
import useAuthStore from "../../store/useAuthStore";
import useGlobalStore from "../../store/useGlobalStore";
import { successToast, errorToast } from "../../utils/utils";
import FullScreenToggle from "../../components/common/FullScreenToggle";
import ConnectionStatus from "../../components/common/ConnectionStatus";
import { useTheme } from "../../context/ThemeContext";
import Icon from "../../components/common/Icon";
import Logo from "../../components/common/Logo";

const Header = () => {
  const { i18n } = useTranslation();
  const navigate = useNavigate();
  const { theme, toggleTheme } = useTheme();
  const logout = useAuthStore((state) => state.logout);
  const setGlobalStore = useGlobalStore((state) => state.setGlobalStore);
  const togglePanel = useGlobalStore((state) => state.togglePanel);
  const chartLayout = useGlobalStore((state) => state.chartLayout);
  const setChartLayout = useGlobalStore((state) => state.setChartLayout);
  const oneClick = useGlobalStore((state) => state.oneClick);

  const logoutHandler = () => {
    try {
      logout();
      successToast("Logout successful!");
      navigate("/login");
    } catch (error) {
      errorToast("Logout failed!");
    }
  };

  const changeLanguage = (lang) => {
    i18n.changeLanguage(lang);
  };

  const menuItems = [
    {
      key: 1,
      icon:
        theme === "light" ? (
          <Icon Icon={LuMoon} size={16} />
        ) : (
          <Icon Icon={LuSun} size={16} />
        ),
      label: theme === "light" ? "Dark Mode" : "Light Mode",
      onClick: () => toggleTheme(),
    },
    // {
    //   key: 2,
    //   icon: <Icon Icon={LuLanguages} size={16} />,
    //   label: "Language",
    //   children: [
    //     { key: "en", label: "English", onClick: () => changeLanguage("en") },
    //     { key: "hi", label: "हिंदी", onClick: () => changeLanguage("hi") },
    //     { key: "ar", label: "العربية", onClick: () => changeLanguage("ar") },
    //   ],
    // },
    {
      key: 3,
      icon: <Icon Icon={LuLogOut} size={16} />,
      label: "Log Out",
      onClick: () => logoutHandler(),
    },
  ];

  const chartItems = [
    {
      key: 1,
      icon: <Icon Icon={LuSquare} className="stroke-[1.5]" />,
      onClick: () => setChartLayout(1),
    },
    {
      key: 2,
      icon: <Icon Icon={LuColumns2} className="stroke-[1.5]" />,
      onClick: () => setChartLayout(2),
    },
    {
      key: 3,
      icon: <Icon Icon={LuRows2} className="stroke-[1.5]" />,
      onClick: () => setChartLayout(3),
    },
    {
      key: 4,
      icon: <Icon Icon={LuLayoutGrid} className="stroke-[1.5]" />,
      onClick: () => setChartLayout(4),
    },
  ];

  const toggleOneClick = () => setGlobalStore({ oneClick: !oneClick });

  return (
    <div className="h-full flex items-center justify-between border-b border-theme-border">
      <div className="h-full flex items-center divide-x divide-theme-border p-1">
        <Dropdown
          menu={{ items: menuItems }}
          trigger={["click"]}
          overlayClassName="custom-dropdown-menu"
        >
          <button className="h-full px-4">
            <Icon Icon={LuMenu} size={24} />
          </button>
        </Dropdown>

        <div className="px-4">
          <Logo className="h-8" />
        </div>

        <div className="flex items-center justify-center gap-1 px-2">
          <button
            title="Toggle Market Watch"
            className="btn-icon-hover"
            onClick={() => togglePanel("marketWatch")}
          >
            <Icon Icon={LuPanelLeft} className="stroke-[1.5]" />
          </button>

          <button
            title="Toggle Activity"
            className="btn-icon-hover"
            onClick={() => togglePanel("activity")}
          >
            <Icon Icon={LuPanelBottom} className="stroke-[1.5]" />
          </button>

          <button
            title="Toggle Order"
            className="btn-icon-hover"
            onClick={() => togglePanel("order")}
          >
            <Icon Icon={LuPanelRight} className="stroke-[1.5]" />
          </button>
        </div>

        <div className="px-2">
          <Dropdown
            menu={{ items: chartItems }}
            trigger={["click"]}
            overlayClassName="custom-dropdown-menu"
          >
            <button title="Chart Layout" className="btn-icon-hover">
              <Icon Icon={LuLayoutDashboard} className="stroke-[1.5]" />
            </button>
          </Dropdown>
        </div>
      </div>

      <div className="flex items-center gap-2">
        {chartLayout === 1 && (
          <button onClick={toggleOneClick} title="One Click Trade">
            <Icon Icon={MdAdsClick} isActive={oneClick} />
          </button>
        )}

        <ConnectionStatus />

        <FullScreenToggle />
      </div>
    </div>
  );
};

export default Header;
