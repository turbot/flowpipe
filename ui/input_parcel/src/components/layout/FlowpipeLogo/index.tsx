import Logo from "jsx:./flowpipe-logo.svg";
// // @ts-ignore
// import { ReactComponent as LogoDarkmode } from "assets/flowpipe-logo-darkmode.svg";
// // @ts-ignore
// import { ReactComponent as LogoWordmark } from "assets/flowpipe-logo-wordmark.svg";
// // @ts-ignore
// import { ReactComponent as LogoWordmarkDarkmode } from "assets/flowpipe-logo-wordmark-darkmode.svg";

const FlowpipeLogo = () => {
  return (
    <div className="mr-1 md:mr-4">
        <div className="block md:hidden w-8">
          <Logo />
          {/*{theme.name === ThemeNames.STEAMPIPE_DEFAULT && <Logo />}*/}
          {/*{theme.name === ThemeNames.STEAMPIPE_DARK && <LogoDarkmode />}*/}
        </div>
        <div className="hidden md:block w-48">
          <LogoWordmark />
          {/*{theme.name === ThemeNames.STEAMPIPE_DEFAULT && <LogoWordmark />}*/}
          {/*{theme.name === ThemeNames.STEAMPIPE_DARK && <LogoWordmarkDarkmode />}*/}
        </div>
    </div>
  );
};

export default FlowpipeLogo;