import { classNames } from "@flowpipe/utils/style";
import { PipelingError } from "@flowpipe/api/error";
import { useInputForm } from "@flowpipe/components/InputForm";

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
  const {
    components: { ErrorIcon },
  } = useInputForm();
  return (
    <span className={classNames(withIcon ? "flex space-x-2 items-start" : "")}>
      {withIcon ? (
        <ErrorIcon className="inline-block h-5 w-5 fill-alert mt-0.5" />
      ) : null}
      <span className={classNames("text-alert break-word", className)}>
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
