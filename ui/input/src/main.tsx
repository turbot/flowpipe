import AppWrapper from "components/layout/AppWrapper";
import InputApp from "components/InputApp";
import React from "react";
import ReactDOM from "react-dom/client";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import { ThemeProvider } from "components/layout/ThemeProvider";
import "./index.css";

const router = createBrowserRouter([
  {
    path: "/",
    element: <AppWrapper />,
    children: [{ path: "input/:id/:hash", element: <InputApp /> }],
  },
]);

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <ThemeProvider>
      <RouterProvider router={router} />
    </ThemeProvider>
  </React.StrictMode>,
);
