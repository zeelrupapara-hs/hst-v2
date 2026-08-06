import { lazy } from "react";
import { Navigate } from "react-router-dom";

const Login = lazy(() => import("../pages/auth/Login"));
const Register = lazy(() => import("../pages/auth/Register"));
const ForgotPassword = lazy(() => import("../pages/auth/ForgotPassword"));

const Dashboard = lazy(() => import("../pages/dashboard/Dashboard"));

const routes = [
  // Auth Routes
  { path: "/login", type: "auth", component: Login },
  { path: "/register", type: "auth", component: Register },
  { path: "/forgot_password", type: "auth", component: ForgotPassword },

  // Private Routes
  { path: "/", type: "private", component: Dashboard, title: "Dashboard" },
  { path: "/dashboard", type: "private", component: Dashboard, title: "Dashboard" },

  // Not Found Route
  { path: "*", type: "private", component: () => <Navigate to="/" replace /> },
];

export default routes;
