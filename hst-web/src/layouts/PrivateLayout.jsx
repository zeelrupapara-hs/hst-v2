import { Navigate } from "react-router-dom";
import useAuthStore from "../store/useAuthStore";
import Header from "./components/Header";
import { SocketProvider } from "../socket";

const PrivateLayout = ({ title, children }) => {
  const isLogin = useAuthStore((state) => state.isLogin);

  if (!isLogin) return <Navigate to="/login" />;

  return (
    <SocketProvider>
      <div className="min-h-screen bg-theme-bg text-theme-text transition-all duration-300">
        <div className="h-[45px]">
          <Header />
        </div>

        <div className="h-[calc(100vh-45px)]">{children}</div>
      </div>
    </SocketProvider>
  );
};

export default PrivateLayout;
