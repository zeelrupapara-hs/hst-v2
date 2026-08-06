import React, { Suspense } from "react";
import { BrowserRouter as Router, useRoutes } from "react-router-dom";
import GlobalLoader from "../components/loader/GlobalLoader";
import routes from "./routes";
import ProtectedRoute from "./ProtectedRoute";

const generateRoutes = (routes) => {
  return routes.map(({ type, component: Component, children, ...rest }) => {
    return {
      element: (
        <ProtectedRoute type={type} {...rest}>
          <Component />
        </ProtectedRoute>
      ),
      children: children?.length ? generateRoutes(children) : null,
      ...rest,
    };
  });
};

const RoutesWrapper = () => {
  return useRoutes(generateRoutes(routes));
};

const AppRoutes = () => {
  return (
    <Router>
      <Suspense fallback={<GlobalLoader isSuspense={true} />}>
        <RoutesWrapper />
      </Suspense>
    </Router>
  );
};

export default AppRoutes;
