import { classNames } from "utils/style.ts";
import { ReactNode } from "react";

interface ButtonProps {
  children: ReactNode;
  disabled?: boolean;
  type?: "button" | "submit";
  onClick: () => void;
}

const Button = ({
  children,
  disabled = false,
  type = "button",
  onClick,
}: ButtonProps) => {
  return (
    <button
      type={type}
      disabled={disabled}
      className={classNames(
        "rounded-md bg-indigo-600 px-2.5 py-1.5 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600",
        "disabled:bg-gray-300 disabled:text-gray-400 disabled:cursor-not-allowed",
      )}
      onClick={disabled ? undefined : onClick}
    >
      {children}
    </button>
  );
};

export default Button;
