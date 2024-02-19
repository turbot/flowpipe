import Select from "react-select";
import { PipelineInputOption } from "types/input.ts";
import useSelectInputStyles from "components/forms/SelectInput/useSelectInputStyles.ts";

type SelectInputProps = {
  disabled: boolean;
  multi?: boolean;
  label?: string;
  name: string;
  options: PipelineInputOption[];
  value: string[];
  onChange: (value: string[]) => void;
};

// const getValueForState = (multi, option) => {
//   if (multi) {
//     // @ts-ignore
//     return option.map((v) => v.value).join(",");
//   } else {
//     return option.value;
//   }
// };
//
const findOptions = (options, value) => {
  return options.filter((option) =>
    option.value ? value.indexOf(option.value.toString()) >= 0 : false,
  );
};

const SelectInput = ({
  disabled,
  multi = false,
  label,
  name,
  options,
  value,
  onChange,
}: SelectInputProps) => {
  const styles = useSelectInputStyles();

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
      <Select
        aria-labelledby={name}
        className="basic-single"
        classNamePrefix="select"
        // components={{
        //   // @ts-ignore
        //   MultiValueLabel: MultiValueLabelWithTags,
        //   // @ts-ignore
        //   Option: OptionWithTags,
        //   // @ts-ignore
        //   SingleValue: SingleValueWithTags,
        // }}
        inputId={name}
        isDisabled={disabled}
        isSearchable
        isMulti={multi}
        menuPortalTarget={document.getElementById("portals")}
        name={name}
        // @ts-ignore
        onChange={(v) => {
          multi
            ? onChange(v.map((newValue) => newValue.value || newValue.label))
            : onChange([v.value || v.label]);
        }}
        options={options}
        placeholder={multi ? "Select one or more values" : "Select a value"}
        // @ts-ignore
        styles={styles}
        value={findOptions(options, value)}
      />
    </div>
  );
};

export default SelectInput;
