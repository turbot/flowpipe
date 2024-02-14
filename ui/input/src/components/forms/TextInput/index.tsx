import ErrorIcon from "@material-symbols/svg-300/rounded/error.svg?react";
import { classNames } from "utils/style.ts";

interface TextInputProps {
  name: string;
  label?: string;
  value: string;
  error?: string | null;
  onChange: (value: string) => void;
}

const TextInput = ({ name, label, value, error, onChange }: TextInputProps) => {
  return (
    <div>
      {label && (
        <label
          htmlFor={name}
          className="block text-sm font-medium leading-6 text-gray-900"
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
            !!error
              ? "pr-10 text-red-900 ring-red-300 placeholder:text-red-300 focus:ring-red-500"
              : null,
          )}
          aria-invalid={!!error ? "true" : "false"}
          aria-describedby={`${name}-error`}
          value={value}
          onChange={(e) => onChange(e.target.value)}
        />
        {!!error && (
          <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
            <ErrorIcon className="h-5 w-5 fill-red-500" aria-hidden="true" />
          </div>
        )}
      </div>
      <p className="mt-2 text-sm text-red-600" id={`${name}-error`}>
        {error || <span>&nbsp;</span>}
      </p>
    </div>
  );
};

export default TextInput;
