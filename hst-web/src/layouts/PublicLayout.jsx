import { Navigate } from "react-router-dom";
import useAuthStore from "../store/useAuthStore";

const PublicLayout = ({ children }) => {
  const isLogin = useAuthStore((state) => state.isLogin);

  if (isLogin) return <Navigate to="/" />;

  return (
    <div className="min-h-screen flex items-center justify-center bg-theme-bg text-theme-text">
      {children}
    </div>
  );
};

export default PublicLayout;
