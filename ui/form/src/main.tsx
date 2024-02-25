import AppWrapper from "@flowpipe/components/layout/AppWrapper";
import InputForm from "@flowpipe/components/InputForm";
import React from "react";
import ReactDOM from "react-dom/client";
import { API_FETCHER } from "@flowpipe/api";
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import { SWRConfig } from "swr";
import { ThemeProvider } from "@flowpipe/components/layout/ThemeProvider";
import "@flowpipe/styles/index.css";

const router = createBrowserRouter(
  [
    {
      path: "/",
      element: <AppWrapper />,
      children: [{ path: ":id/:hash", element: <InputForm /> }],
    },
  ],
  { basename: "/form" },
);

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <SWRConfig value={{ fetcher: API_FETCHER, shouldRetryOnError: false }}>
      <ThemeProvider>
        <RouterProvider router={router} />
      </ThemeProvider>
    </SWRConfig>
  </React.StrictMode>,
);
