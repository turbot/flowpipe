import Header from "components/layout/Header";
import SiteWrapper from "components/layout/SiteWrapper";
import { FullHeightThemeWrapper } from "components/layout/ThemeProvider";
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
