import { useEffect, useState } from "react";
import { useInputForm } from "@flowpipe/components/InputForm";

const useSelectInputStyles = () => {
  const [, setRandomVal] = useState(0);
  const {
    themeContext: { theme, wrapperRef },
  } = useInputForm();

  // This is annoying, but unless I force a refresh the theme doesn't stay in sync when you switch
  useEffect(() => setRandomVal(Math.random()), [theme.name]);

  if (!wrapperRef) {
    return null;
  }

  // @ts-ignore
  const style = window.getComputedStyle(wrapperRef);
  const background = style.getPropertyValue("--color-background");
  const backgroundModal = style.getPropertyValue("--color-modal");
  const foreground = style.getPropertyValue("--color-foreground");
  const foregroundLight = style.getPropertyValue("--color-foreground-light");

  return {
    clearIndicator: (provided) => ({
      ...provided,
      cursor: "pointer",
    }),
    control: (provided) => {
      return {
        ...provided,
        backgroundColor: backgroundModal,
        borderColor: foregroundLight,
        boxShadow: "none",
      };
    },
    dropdownIndicator: (provided) => ({
      ...provided,
      cursor: "pointer",
    }),
    input: (provided) => {
      return {
        ...provided,
        color: foreground,
      };
    },
    multiValue: (provided) => {
      return {
        ...provided,
        backgroundColor: background,
        color: foreground,
      };
    },
    multiValueLabel: (provided) => {
      return {
        ...provided,
        backgroundColor: background,
        color: foreground,
      };
    },
    placeholder: (provided) => {
      return {
        ...provided,
        color: foreground,
      };
    },
    singleValue: (provided) => {
      return {
        ...provided,
        backgroundColor: backgroundModal,
        color: foreground,
      };
    },
    menu: (provided) => {
      return {
        ...provided,
        backgroundColor: backgroundModal,
        border: `1px solid ${foregroundLight}`,
        boxShadow: "none",
        marginTop: 0,
        marginBottom: 0,
      };
    },
    menuList: (provided) => {
      return {
        ...provided,
        paddingTop: 0,
        paddingBottom: 0,
      };
    },
    menuPortal: (base) => ({ ...base, zIndex: 9999 }),
    option: (provided, state) => {
      return {
        ...provided,
        backgroundColor: state.isFocused ? background : "none",
        color: foreground,
      };
    },
  };
};

export default useSelectInputStyles;
