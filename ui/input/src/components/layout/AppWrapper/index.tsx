import Header from "components/layout/Header";
import SiteWrapper from "components/layout/SiteWrapper";
import { Outlet } from "react-router-dom";

const AppWrapper = () => {
  return (
    <SiteWrapper>
      <Header />
      <Outlet />
    </SiteWrapper>
  );
};

export default AppWrapper;
