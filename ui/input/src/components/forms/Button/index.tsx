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
      className="rounded-md bg-indigo-600 px-2.5 py-1.5 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
      onClick={disabled ? undefined : onClick}
    >
      {children}
    </button>
  );
};

export default Button;
