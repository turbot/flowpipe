import useLocalStorage from "@flowpipe/hooks/useLocalStorage";
import useMediaQuery from "@flowpipe/hooks/useMediaQuery";
import { classNames } from "@flowpipe/utils/style";
import { createContext, Ref, useContext, useState } from "react";

type Theme = {
  name: string;
  label: string;
};

export type IThemeContext = {
  localStorageTheme: string | null;
  theme: Theme;
  wrapperRef: Ref<null>;
  setTheme(theme: string): void;
  setWrapperRef(element: any): void;
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
  const [wrapperRef, setWrapperRef] = useState(null);
  const doSetWrapperRef = (element) => setWrapperRef(() => element);

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
        wrapperRef,
        setTheme: setLocalStorageTheme,
        setWrapperRef: doSetWrapperRef,
      }}
    >
      {children}
    </ThemeContext.Provider>
  );
};

const FullHeightThemeWrapper = ({ children }) => {
  const { theme, setWrapperRef } = useTheme();
  return (
    <div
      ref={setWrapperRef}
      className={classNames(
        `min-h-screen flex flex-col theme-${theme.name} bg-background print:bg-white print:theme-steampipe-default text-foreground print:text-black`,
      )}
    >
      {children}
    </div>
  );
};

export { FullHeightThemeWrapper, Themes, ThemeNames, ThemeProvider, useTheme };
