import Button from "@flowpipe/components/forms/Button";
import ErrorMessage from "@flowpipe/components/layout/ErrorMessage";
import FlowpipeLogo from "@flowpipe/components/layout/FlowpipeLogo";
import SuccessMessage from "@flowpipe/components/layout/SuccessMessage";
import TextInput from "@flowpipe/components/forms/TextInput";
import { Form, Formik } from "formik";
import { FormEvent, useEffect, useMemo, useState } from "react";
import { FormikErrors } from "formik/dist/types";
import { PipelingError } from "@flowpipe/api/error";
import { PipelineInputOption, PipelineInputType } from "@flowpipe/types/input";
import { useInputAPI } from "@flowpipe/api/pipeline";
import { useParams } from "react-router-dom";
import SelectInput from "@flowpipe/components/forms/SelectInput";

interface InputFormState {
  status: "pending" | "responded" | "error";
  error?: PipelingError | string | null;
}

interface InputFormValues {
  values: string[];
}

interface InputOptionsProps {
  formState: InputFormState;
  errors: FormikErrors<InputFormValues>;
  inputType: PipelineInputType;
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
  formState,
  errors,
  inputType,
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
        <div className="flex items-center space-x-2 justify-end">
          {formState.status === "error" && formState.error && (
            <ErrorMessage error={formState.error} />
          )}
          {formState.status === "responded" && (
            <SuccessMessage message="Input response sent" />
          )}
          {options?.map((o) => (
            <Button
              key={o.value}
              disabled={submitting || formState.status === "responded"}
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
    case "select":
    case "multiselect":
      return (
        <div className="flex flex-col justify-end space-x-2 space-y-3">
          <div>
            <SelectInput
              name="values"
              disabled={submitting || formState.status === "responded"}
              // @ts-ignore
              error={!!errors.values ? errors.values : null}
              multi={inputType === "multiselect"}
              options={options}
              value={values.values || []}
              onChange={(v) => setFieldValue("values", v, true)}
            />
          </div>
          <div className="flex items-center justify-end space-x-2">
            {formState.status === "error" && formState.error && (
              <ErrorMessage error={formState.error} />
            )}
            {formState.status === "responded" && (
              <SuccessMessage message="Input response sent" />
            )}
            <Button
              disabled={
                !valid || submitting || formState.status === "responded"
              }
              type="submit"
              onClick={onSubmit}
            >
              Submit
            </Button>
          </div>
        </div>
      );
    case "text":
      return (
        <div className="flex flex-col justify-end space-x-2">
          <div>
            <TextInput
              name="values"
              disabled={submitting || formState.status === "responded"}
              // @ts-ignore
              error={!!errors.values ? errors.values : null}
              value={values.values[0] || ""}
              onChange={(v) => setFieldValue("values", [v], true)}
            />
          </div>
          <div className="flex items-center justify-end space-x-2">
            {formState.status === "error" && formState.error && (
              <ErrorMessage error={formState.error} />
            )}
            {formState.status === "responded" && (
              <SuccessMessage message="Input response sent" />
            )}
            <Button
              disabled={
                !valid || submitting || formState.status === "responded"
              }
              type="submit"
              onClick={onSubmit}
            >
              Submit
            </Button>
          </div>
        </div>
      );
    default:
      return (
        <ErrorMessage
          as="string"
          error={`Unsupported input type ${inputType}`}
        />
      );
  }
};

const InputForm = () => {
  const { id, hash } = useParams();
  const { input, error, loading, postInput } = useInputAPI(id, hash);
  const initialValues = useMemo<InputFormValues>(() => {
    if (
      !input ||
      !input.options ||
      (input.input_type !== "select" && input.input_type !== "multiselect")
    ) {
      return { values: [] };
    }
    return {
      values: input.options.filter((o) => o.selected).map((o) => o.value),
    };
  }, [input]);
  const [state, setState] = useState<InputFormState>({
    status: "pending",
    error: null,
  });

  useEffect(() => {
    setState((existing) => ({
      ...existing,
      value: input ? input.status : "pending",
      error: null,
    }));
  }, [input, setState]);

  const showFooter =
    state.status === "error" ||
    state.status === "responded" ||
    (state.status === "pending" &&
      !error &&
      (!input || input.status === "starting" || input.status === "started"));

  return (
    <div className="mx-auto my-auto">
      <div className="flex flex-col divide-y divide-modal-divide overflow-hidden rounded-lg bg-modal shadow w-screen md:min-w-xl max-w-xl">
        <div className="px-4 py-4">
          <FlowpipeLogo />
        </div>
        <div className="px-4 py-5">
          {error && <ErrorMessage withIcon error={error} />}
          {loading && "Loading..."}
          {input && (
            <>
              {(input.status === "starting" || input.status === "started") && (
                <h3 className="text-base font-semibold leading-6 text-foreground">
                  {input.prompt}
                </h3>
              )}
              {input.status === "finished" && (
                <h3 className="text-base font-semibold leading-6 text-foreground">
                  Input has already been responded to.
                </h3>
              )}
              {input.status === "error" && (
                <ErrorMessage withIcon error="Input is in a failed state." />
              )}
            </>
          )}
        </div>
        {showFooter && (
          <div className="px-4 py-4">
            {input && (
              <Formik
                initialValues={initialValues}
                validate={(values) => {
                  const errors: { values?: string } = {};
                  if (!values.values || !values.values.every((v) => !!v)) {
                    errors.values = `Please ${input?.input_type === "text" ? "enter" : "select"} a value.`;
                  }
                  return errors;
                }}
                onSubmit={async (values, { setSubmitting }) => {
                  setSubmitting(false);
                  // TODO remove
                  const response_url = new URL(input?.response_url);
                  const { error } = await postInput(response_url.pathname, {
                    execution_id: input.execution_id,
                    pipeline_execution_id: input.pipeline_execution_id,
                    step_execution_id: input.step_execution_id,
                    values: values.values,
                  });
                  if (error) {
                    setState({ status: "error", error });
                  } else {
                    setState({ status: "responded", error: null });
                    // await reload();
                  }
                }}
                enableReinitialize
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
                      formState={state}
                      errors={errors}
                      inputType={input.input_type}
                      setFieldValue={setFieldValue}
                      // state={state}
                      submitting={isSubmitting}
                      options={input.options}
                      valid={isValid}
                      values={values}
                      onSubmit={handleSubmit}
                    />
                  </Form>
                )}
              </Formik>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default InputForm;
