import Button from "components/forms/Button";
import FlowpipeLogo from "components/layout/FlowpipeLogo";
import TextInput from "components/forms/TextInput";
import { Form, Formik } from "formik";
import { FormEvent, useState } from "react";
import { FormikErrors } from "formik/dist/types";
import { PipelingError } from "api/error.ts";
import {
  PipelineInput,
  PipelineInputOption,
  PipelineInputResponse,
  PipelineInputType,
} from "types/input.ts";

interface FormProps {
  input: PipelineInput;
  onSubmit: (
    response_url: string,
    input_result: PipelineInputResponse,
  ) => Promise<{ input: PipelineInput | null; error: PipelingError | null }>;
}

interface InputFormState {
  value: "pending" | "responded" | "error";
  error?: PipelingError | string | null;
}

interface InputFormValues {
  values: string[];
}

interface InputOptionsProps {
  errors: FormikErrors<InputFormValues>;
  inputType: PipelineInputType;
  state: InputFormState;
  submitting: boolean;
  options: PipelineInputOption[];
  valid: boolean;
  values: InputFormValues;
  setFieldValue: (
    field: string,
    value: any,
    shouldValidate?: boolean,
  ) => Promise<void | FormikErrors<InputFormValues>>;
  onSubmit: (e?: FormEvent<HTMLFormElement>) => void;
}

const InputOptions = ({
  errors,
  inputType,
  state,
  submitting,
  options,
  valid,
  values,
  setFieldValue,
  onSubmit,
}: InputOptionsProps) => {
  switch (inputType) {
    case "button":
      return (
        <div className="flex justify-end space-x-2">
          {options?.map((o) => (
            <Button
              key={o.value}
              disabled={submitting}
              type="button"
              onClick={async () => {
                await setFieldValue("values", [o.value], true);
                onSubmit();
              }}
            >
              {o.label || o.value}
            </Button>
          ))}
        </div>
      );
    case "text":
      return (
        <div className="flex flex-col justify-end space-x-2">
          <div>
            <TextInput
              name="values"
              // @ts-ignore
              error={!!errors.values ? errors.values : null}
              value={values.values[0] || ""}
              onChange={(v) => setFieldValue("values", [v], true)}
            />
          </div>
          <div className="flex justify-end space-x-2">
            <Button
              disabled={!valid || submitting}
              type="submit"
              onClick={onSubmit}
            >
              Submit
            </Button>
          </div>
        </div>
      );
    default:
      return null;
  }
};

const InputForm = ({ input, onSubmit }: FormProps) => {
  const initialValues: InputFormValues = { values: [] };
  const [state, setState] = useState<InputFormState>({
    value: "pending",
    error: null,
  });
  return (
    <div className="flex flex-col divide-y divide-gray-200 overflow-hidden rounded-lg bg-white shadow w-screen max-w-xl">
      <div className="px-4 py-4">
        <FlowpipeLogo />
      </div>
      <div className="px-4 py-5">
        <h3 className="text-base font-semibold leading-6 text-gray-900">
          {input.prompt}
        </h3>
      </div>
      <div className="px-4 py-4">
        <Formik
          initialValues={initialValues}
          validate={(values) => {
            console.log(values);
            const errors = {};
            if (!values.values || !values.values.every((v) => !!v)) {
              errors.values = `Please ${input.input_type === "text" ? "enter" : "select"} a value.`;
            }
            return errors;
          }}
          onSubmit={async (values, { setSubmitting }) => {
            console.log("Submitting...", values.values);
            setSubmitting(false);
            const { error } = await onSubmit(input.response_url, {
              execution_id: input.execution_id,
              pipeline_execution_id: input.pipeline_execution_id,
              step_execution_id: input.step_execution_id,
              values: values.values,
            });
            console.log(error);
          }}
        >
          {({
            errors,
            isSubmitting,
            isValid,
            setFieldValue,
            handleSubmit,
            values,
          }) => (
            <Form>
              <InputOptions
                errors={errors}
                inputType={input.input_type}
                setFieldValue={setFieldValue}
                state={state}
                submitting={isSubmitting}
                options={input.options}
                valid={isValid}
                values={values}
                onSubmit={handleSubmit}
              />
            </Form>
          )}
        </Formik>
      </div>
    </div>
  );
};

export default InputForm;
