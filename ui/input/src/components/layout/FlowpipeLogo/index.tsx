/// <reference types="vite-plugin-svgr/client" />
import Logo from "assets/flowpipe-logo.svg?react";
// import LogoDarkmode from "assets/flowpipe-logo-darkmode.svg?react";
import LogoWordmark from "assets/flowpipe-logo-wordmark.svg?react";
// import LogoWordmarkDarkmode from "assets/flowpipe-logo-wordmark-darkmode.svg?react";

const FlowpipeLogo = () => {
  return (
    <div className="mr-1 md:mr-4">
      {/*<div className="block sm:hidden w-8">*/}
      {/*  <Logo />*/}
      {/*  /!*{theme.name === ThemeNames.STEAMPIPE_DEFAULT && <Logo />}*!/*/}
      {/*  /!*{theme.name === ThemeNames.STEAMPIPE_DARK && <LogoDarkmode />}*!/*/}
      {/*</div>*/}
      {/*<div className="hidden sm:block w-40">*/}
      <div className="w-40">
        <LogoWordmark />
        {/*{theme.name === ThemeNames.STEAMPIPE_DEFAULT && <LogoWordmark />}*/}
        {/*{theme.name === ThemeNames.STEAMPIPE_DARK && <LogoWordmarkDarkmode />}*/}
      </div>
    </div>
  );
};

export default FlowpipeLogo;
