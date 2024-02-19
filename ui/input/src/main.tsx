import AppWrapper from "components/layout/AppWrapper";
import InputForm from "components/InputForm";
import React from "react";
import ReactDOM from "react-dom/client";
import { API_FETCHER } from "src/api";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import { SWRConfig } from "swr";
import { ThemeProvider } from "components/layout/ThemeProvider";
import "styles/index.css";

const router = createBrowserRouter([
  {
    path: "/",
    element: <AppWrapper />,
    children: [{ path: "input/:id/:hash", element: <InputForm /> }],
  },
]);

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <SWRConfig value={{ fetcher: API_FETCHER, shouldRetryOnError: false }}>
      <ThemeProvider>
        <RouterProvider router={router} />
      </ThemeProvider>
    </SWRConfig>
  </React.StrictMode>,
);
