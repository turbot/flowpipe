import { classNames } from "@flowpipe/utils/style";
import { ReactNode } from "react";

interface ButtonProps {
  children: ReactNode;
  disabled?: boolean;
  size?: "sm" | "md" | "lg";
  style?: "default" | "ok" | "alert";
  type?: "button" | "submit";
  onClick: () => void;
}

const Button = ({
  children,
  disabled = false,
  size = "lg",
  style = "default",
  type = "button",
  onClick,
}: ButtonProps) => {
  return (
    <button
      type={type}
      disabled={disabled}
      className={classNames(
        "rounded-md font-semibold shadow-sm hover:bg-opacity-80 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-info",
        "disabled:bg-opacity-50 disabled:cursor-not-allowed",
        size === "sm" ? "px-2 py-1 text-xs" : null,
        size === "md" ? "px-2.5 py-1.5 text-sm" : null,
        size === "lg" ? "px-3 py-2.5 text-sm" : null,
        style === "ok" ? "bg-ok text-white" : null,
        style === "alert" ? "bg-alert text-white" : null,
        // Default style
        style !== "ok" && style !== "alert"
          ? "bg-flowpipe-blue-dark text-white"
          : null,
      )}
      onClick={disabled ? undefined : onClick}
    >
      {children}
    </button>
  );
};

export default Button;
