import Header from "@flowpipe/components/layout/Header";
import SiteWrapper from "@flowpipe/components/layout/SiteWrapper";
import { FullHeightThemeWrapper } from "@flowpipe/components/layout/ThemeProvider";
import { Outlet } from "react-router-dom";

const AppWrapper = () => {
  return (
    <FullHeightThemeWrapper>
      <SiteWrapper>
        <Header />
        <Outlet />
      </SiteWrapper>
    </FullHeightThemeWrapper>
  );
};

export default AppWrapper;
