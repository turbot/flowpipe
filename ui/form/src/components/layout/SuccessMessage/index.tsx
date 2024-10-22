import { classNames } from "@flowpipe/utils/style";
import { useInputForm } from "@flowpipe/components/InputForm";

type ErrorMessageProps = {
  className?: string;
  message: string;
  withIcon?: boolean;
};

const SuccessMessage = ({
  className,
  message,
  withIcon = true,
}: ErrorMessageProps) => {
  const {
    components: { SuccessIcon },
  } = useInputForm();
  return (
    <span className={classNames(withIcon ? "flex space-x-1 items-center" : "")}>
      {withIcon ? (
        <SuccessIcon className="inline-block h-5 w-5 fill-ok" />
      ) : null}
      <span className={classNames("text-ok break-word", className)}>
        {message}
      </span>
    </span>
  );
};

export default SuccessMessage;
