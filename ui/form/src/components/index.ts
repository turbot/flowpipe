import ErrorIcon from "@material-symbols/svg-300/rounded/error-fill.svg?react";
import SuccessIcon from "@material-symbols/svg-300/rounded/check_circle-fill.svg?react";
import { ComponentsMap } from "@flowpipe/types/component";
import { ReactNode } from "react";

const componentsMap = {};

const getComponent = (key: string) => componentsMap[key];

const registerComponent = (
  key: string,
  component: (props: any) => ReactNode,
) => {
  componentsMap[key] = component;
};

const buildComponentsMap = (overrides = {}): ComponentsMap => {
  return {
    ErrorIcon,
    SuccessIcon,
    ...overrides,
  };
};

export { buildComponentsMap, getComponent, registerComponent };
