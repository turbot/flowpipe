import LightMode from "@material-symbols/svg-300/rounded/light_mode-fill.svg?react";
import DarkMode from "@material-symbols/svg-300/rounded/dark_mode-fill.svg?react";
import { classNames } from "utils/style";
import { ThemeNames, useTheme } from "components/layout/ThemeProvider";

const ThemeToggle = () => {
  const { theme, setTheme } = useTheme();
  return (
    <button
      className={classNames("flex items-center h-5 w-5 text-gray-500")}
      onClick={() =>
        setTheme(
          theme.name === ThemeNames.PIPELING_DEFAULT
            ? ThemeNames.PIPELING_DARK
            : ThemeNames.PIPELING_DEFAULT,
        )
      }
    >
      {theme.name === ThemeNames.PIPELING_DARK ? (
        <LightMode title="Switch to light theme" />
      ) : (
        <DarkMode title="Switch to dark theme" />
      )}
    </button>
  );
};

export default ThemeToggle;
