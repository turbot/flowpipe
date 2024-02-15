import ErrorIcon from "@material-symbols/svg-300/rounded/error-fill.svg?react";
import { classNames } from "utils/style";
import { PipelingError } from "api/error";

type ErrorMessageProps = {
  as?: "error" | "string";
  className?: string;
  error: string | Error | PipelingError | null;
  prefix?: string;
  withIcon?: boolean;
};

const ErrorMessage = ({
  as = "error",
  className,
  error,
  prefix = "",
  withIcon = false,
}: ErrorMessageProps) => {
  if (!error) {
    return null;
  }
  return (
    <span className={classNames(withIcon ? "flex space-x-1 items-center" : "")}>
      {withIcon ? (
        <ErrorIcon className="inline-block h-5 w-5 fill-red-600" />
      ) : null}
      <span className={classNames("text-red-600 break-word", className)}>
        {prefix ? `${prefix}: ` : ""}
        {as === "string"
          ? error
          : /*@ts-ignore*/
            error?.message || error?.detail || error?.title}
      </span>
    </span>
  );
};

export default ErrorMessage;
