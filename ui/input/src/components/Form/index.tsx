import Button from "components/forms/Button";
import {
  PipelineInput,
  PipelineInputOption,
  PipelineInputType,
} from "src/types/input.ts";

interface FormProps {
  input: PipelineInput;
}

interface InputOptionsProps {
  input_type: PipelineInputType;
  options: PipelineInputOption[];
}

const InputOptions = ({ input_type, options }: InputOptionsProps) => {
  switch (input_type) {
    case "button":
      return (
        <div className="flex justify-end space-x-2">
          {options?.map((o) => (
            <Button
              value={o.value}
              label={o.label}
              onClick={(v: string) => {
                console.log(v);
              }}
            />
          ))}
        </div>
      );
    // case "text":
    //   return (
    //     <div className="flex justify-end space-x-2">
    //       {options?.map((o) => (
    //         <Button
    //           value={o.value}
    //           label={o.label}
    //           onClick={(v: string) => {
    //             console.log(v);
    //           }}
    //         />
    //       ))}
    //     </div>
    //   );
    default:
      return null;
  }
};

const Form = ({ input }: FormProps) => {
  return (
    <>
      {/*<div className="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />*/}
      {/*<div className="fixed inset-0 z-10 w-screen overflow-y-auto">*/}
      <div className="flex flex-col divide-y divide-gray-200 overflow-hidden rounded-lg bg-white shadow w-screen max-w-3xl">
        <div className="px-4 py-5 sm:px-6">
          <h3 className="text-base font-semibold leading-6 text-gray-900">
            {input.prompt}
          </h3>
        </div>
        <div className="px-4 py-5 sm:p-6">
          <InputOptions input_type={input.input_type} options={input.options} />
        </div>
      </div>
      {/*</div>*/}
    </>
  );
};

export default Form;
