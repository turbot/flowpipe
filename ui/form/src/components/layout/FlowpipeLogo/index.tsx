/// <reference types="vite-plugin-svgr/client" />
import LogoWordmark from "@flowpipe/assets/flowpipe-logo-wordmark.svg?react";
import LogoWordmarkDarkmode from "@flowpipe/assets/flowpipe-logo-wordmark-darkmode.svg?react";
import {
  ThemeNames,
  useTheme,
} from "@flowpipe/components/layout/ThemeProvider";

const FlowpipeLogo = () => {
  const { theme } = useTheme();
  return (
    <div className="mr-1 md:mr-4">
      <div className="w-24">
        {theme.name === ThemeNames.PIPELING_DEFAULT && <LogoWordmark />}
        {theme.name === ThemeNames.PIPELING_DARK && <LogoWordmarkDarkmode />}
      </div>
    </div>
  );
};

export default FlowpipeLogo;
