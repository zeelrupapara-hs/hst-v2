import { Navigate } from "react-router-dom";
import useAuthStore from "../store/useAuthStore";
import PublicLayout from "../layouts/PublicLayout";
import PrivateLayout from "../layouts/PrivateLayout";

const ProtectedRoute = ({ type, title, parent, accessRole, children }) => {
  const { user, isLogin } = useAuthStore((state) => state);

  if (type === "auth") {
    if (!isLogin) {
      return <PublicLayout>{children}</PublicLayout>;
    } else {
      return <Navigate to="/" replace />;
    }
  }

  if (type === "private") {
    if (isLogin) {
      if (parent) return children;

      if (accessRole) {
        if (accessRole === user?.roleName?.toLowerCase()) {
          return <PrivateLayout title={title}>{children}</PrivateLayout>;
        } else {
          return <Navigate to="/" replace />;
        }
      } else {
        return <PrivateLayout title={title}>{children}</PrivateLayout>;
      }
    } else {
      return <Navigate to="/login" replace />;
    }
  }
};

export default ProtectedRoute;
