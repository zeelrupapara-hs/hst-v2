import { useEffect } from "react";
import { Toaster } from "react-hot-toast";
import "./App.css";
import "slick-carousel/slick/slick.css";
import "slick-carousel/slick/slick-theme.css";
import { ConfigProvider, theme as antdTheme } from "antd";
import AppRoutes from "./routes/AppRoutes";
import useAuthStore from "./store/useAuthStore";
import useAppStore from "./store/useAppStore";
import { useTheme } from "./context/ThemeContext";
import GlobalLoader from "./components/loader/GlobalLoader";
import { getLocalItem } from "./utils/utils";

const App = () => {
  const isLogin = useAuthStore((state) => state.isLogin);
  const logout = useAuthStore((state) => state.logout);
  const { loading, fetchAppData } = useAppStore();
  const { theme } = useTheme();

  useEffect(() => {
    if (!isLogin) return;

    if (!getLocalItem("token")) {
      logout();
      return;
    }

    fetchAppData();
  }, [isLogin]);

  if (loading) return <GlobalLoader isSuspense={true} />;

  return (
    <ConfigProvider
      theme={{
        token: {
          colorText: theme === "dark" ? "#89898B" : "#151924",
          colorPrimary: theme === "dark" ? "#32D486" : "#48A66B",
          fontFamily: "Inter, sans-serif",
          controlOutline: "none",
          borderRadius: "4px",
          colorBgBase: theme === "dark" ? "#101013" : "#fff",
          colorBgContainer: theme === "dark" ? "#101013" : "#fff",
          colorBgElevated: theme === "dark" ? "#101013" : "#fff",
          colorBgLayout: theme === "dark" ? "#101013" : "#fff",
        },
        components: {
          Dropdown: {
            colorBgElevated: theme === "dark" ? "#1e222d" : "#fff",
            colorText: theme === "dark" ? "#d1d4dc" : "#151924",
          },
        },
        algorithm:
          theme === "dark"
            ? antdTheme.darkAlgorithm
            : antdTheme.defaultAlgorithm,
      }}
      componentSize="large"
    >
      <Toaster position="top-right" toastOptions={{ duration: 3000 }} />
      <AppRoutes />
    </ConfigProvider>
  );
};

export default App;
