import ErrorIcon from "@material-symbols/svg-300/rounded/error.svg?react";
import { classNames } from "@flowpipe/utils/style";
import {
  ThemeNames,
  useTheme,
} from "@flowpipe/components/layout/ThemeProvider";

interface TextInputProps {
  disabled: boolean;
  name: string;
  label?: string;
  touched: boolean;
  value: string;
  error?: string | null;
  onChange: (value: string) => void;
}

const TextInput = ({
  disabled,
  name,
  label,
  touched,
  value,
  error,
  onChange,
}: TextInputProps) => {
  const { theme } = useTheme();
  return (
    <div>
      {label && (
        <label
          htmlFor={name}
          className="block text-sm font-medium leading-6 text-foreground-light"
        >
          {label}
        </label>
      )}
      <div className="relative mt-2 rounded-md shadow-sm">
        <input
          type="text"
          name={name}
          id={name}
          className={classNames(
            "block w-full rounded-md border-0 pl-2 py-1.5 ring-1 ring-inset focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6",
            !!error && touched
              ? "pr-10 text-alert ring-red-300 focus:ring-error"
              : null,
            theme.name === ThemeNames.PIPELING_DARK
              ? "bg-gray-700"
              : "bg-white",
          )}
          aria-invalid={!!error ? "true" : "false"}
          aria-describedby={`${name}-error`}
          value={value}
          onChange={disabled ? undefined : (e) => onChange(e.target.value)}
        />
        {!!error && touched && (
          <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
            <ErrorIcon className="h-5 w-5 fill-alert" aria-hidden="true" />
          </div>
        )}
      </div>
      {error && touched && (
        <p className="mt-2 text-sm text-error" id={`${name}-error`}>
          {error || <span>&nbsp;</span>}
        </p>
      )}
    </div>
  );
};

export default TextInput;
