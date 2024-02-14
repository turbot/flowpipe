import useLocalStorage from "hooks/useLocalStorage.ts";
import useMediaQuery from "hooks/useMediaQuery.ts";
import { classNames } from "src/utils/style.ts";
import { createContext, useContext, useEffect, useState } from "react";

type Theme = {
  name: string;
  label: string;
};

type IThemeContext = {
  localStorageTheme: string | null;
  theme: Theme;
  setTheme(theme: string): void;
};

const ThemeContext = createContext<IThemeContext | undefined>(undefined);

const ThemeNames = {
  PIPELING_DEFAULT: "pipeling-default",
  PIPELING_DARK: "pipeling-dark",
};

const Themes = {
  [ThemeNames.PIPELING_DEFAULT]: {
    label: "Light",
    name: ThemeNames.PIPELING_DEFAULT,
  },
  [ThemeNames.PIPELING_DARK]: {
    label: "Dark",
    name: ThemeNames.PIPELING_DARK,
  },
};

const useTheme = () => {
  const context = useContext(ThemeContext);
  if (context === undefined) {
    throw new Error("useTheme must be used within a ThemeContext");
  }
  return context;
};

const ThemeProvider = ({ children }) => {
  const [localStorageTheme, setLocalStorageTheme] =
    useLocalStorage("steampipe.ui.theme");
  const prefersDarkTheme = useMediaQuery("(prefers-color-scheme: dark)");

  let theme;

  if (
    localStorageTheme &&
    (localStorageTheme === ThemeNames.PIPELING_DEFAULT ||
      localStorageTheme === ThemeNames.PIPELING_DARK)
  ) {
    theme = Themes[localStorageTheme];
  } else if (prefersDarkTheme) {
    theme = Themes[ThemeNames.PIPELING_DARK];
  } else {
    theme = Themes[ThemeNames.PIPELING_DEFAULT];
  }

  return (
    <ThemeContext.Provider
      value={{
        localStorageTheme,
        theme,
        setTheme: setLocalStorageTheme,
      }}
    >
      {children}
    </ThemeContext.Provider>
  );
};

const FullHeightThemeWrapper = ({ children }) => {
  const { theme } = useTheme();
  return (
    <div
      className={classNames(
        `min-h-screen flex flex-col theme-${theme.name} bg-background print:bg-white print:theme-steampipe-default text-foreground print:text-black`,
      )}
    >
      {children}
    </div>
  );
};

export { FullHeightThemeWrapper, Themes, ThemeNames, ThemeProvider, useTheme };
