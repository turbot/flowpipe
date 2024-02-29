import Button from "@flowpipe/components/forms/Button";
import ErrorMessage from "@flowpipe/components/layout/ErrorMessage";
import FlowpipeLogo from "@flowpipe/components/layout/FlowpipeLogo";
import SelectInput from "@flowpipe/components/forms/SelectInput";
import SuccessMessage from "@flowpipe/components/layout/SuccessMessage";
import TextInput from "@flowpipe/components/forms/TextInput";
import { Form, Formik, FormikTouched } from "formik";
import {
  FormEvent,
  Fragment,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { FormikErrors } from "formik/dist/types";
import {
  InputFormValues,
  PipelineForm,
  PipelineFormStatus,
  PipelineInputOption,
  PipelineInputType,
} from "@flowpipe/types/input";
import { PipelingError } from "@flowpipe/api/error";
import { useFormAPI } from "@flowpipe/api/pipeline";
import { useLocation, useParams } from "react-router-dom";

interface InputFormProps {
  autoSubmit?: boolean;
}

interface InputFormState {
  status: "pending" | "responded" | "error";
  error?: PipelingError | string | null;
}

interface InputFormInnerProps {
  autoSubmit: boolean;
  form: PipelineForm;
  formState: InputFormState;
  errors: FormikErrors<InputFormValues>;
  submitting: boolean;
  valid: boolean;
  touched: FormikTouched<InputFormValues>;
  values: InputFormValues;
  setFieldTouched: (
    field: string,
    isTouched?: boolean | undefined,
    shouldValidate?: boolean | undefined,
  ) => Promise<void | FormikErrors<InputFormValues>>;
  setFieldValue: (
    field: string,
    value: any,
    shouldValidate?: boolean,
  ) => Promise<void | FormikErrors<InputFormValues>>;
  submit: (e?: FormEvent<HTMLFormElement> | undefined) => void;
}

interface InputOptionsProps {
  name: string;
  formState: InputFormState;
  errors: FormikErrors<InputFormValues>;
  inputType: PipelineInputType;
  submitting: boolean;
  options: PipelineInputOption[];
  touched: FormikTouched<InputFormValues>;
  values: InputFormValues;
  setFieldTouched: (
    field: string,
    isTouched?: boolean | undefined,
    shouldValidate?: boolean | undefined,
  ) => Promise<void | FormikErrors<InputFormValues>>;
  setFieldValue: (
    field: string,
    value: any,
    shouldValidate?: boolean,
  ) => Promise<void | FormikErrors<InputFormValues>>;
}

interface SubmitOptionsProps {
  name: string;
  autoSubmit: boolean;
  formState: InputFormState;
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

interface InputPromptProps {
  formStatus: PipelineFormStatus;
  prompt: string | undefined;
}

const Message = ({ children }) => (
  <span className="font-semibold leading-6 text-foreground">{children}</span>
);

const InputPrompt = ({ formStatus, prompt }: InputPromptProps) => {
  if (formStatus !== "starting" && formStatus !== "started") {
    return null;
  }
  return <Message>{prompt}</Message>;
};

const InputOptions = ({
  name,
  formState,
  errors,
  inputType,
  submitting,
  options,
  touched,
  values,
  setFieldTouched,
  setFieldValue,
}: InputOptionsProps) => {
  switch (inputType) {
    case "button":
      return null;
    case "select":
    case "multiselect":
      return (
        <div className="flex flex-col justify-end space-x-2">
          <SelectInput
            name={name}
            disabled={submitting || formState.status === "responded"}
            // @ts-ignore
            error={!!errors[name] ? errors[name] : null}
            multi={inputType === "multiselect"}
            options={options}
            value={
              (inputType === "select"
                ? [values[name] as string]
                : (values[name] as string[])) || []
            }
            onChange={async (v) => {
              await setFieldTouched(name, true);
              await setFieldValue(
                name,
                inputType === "select" ? v[0] : v,
                true,
              );
            }}
          />
        </div>
      );
    case "text":
      return (
        <div className="flex flex-col justify-end space-x-2">
          <TextInput
            name={name}
            disabled={submitting || formState.status === "responded"}
            // @ts-ignore
            error={!!errors[name] ? errors[name] : null}
            touched={!!touched[name]}
            value={(values[name] as string) || ""}
            onChange={async (v) => {
              await setFieldTouched(name, true);
              await setFieldValue(name, v, true);
            }}
          />
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

const SubmitOptions = ({
  name,
  autoSubmit,
  formState,
  inputType,
  submitting,
  options,
  valid,
  setFieldValue,
  onSubmit,
}: SubmitOptionsProps) => {
  if (autoSubmit) {
    return (
      <>
        {!valid && (
          <ErrorMessage as="string" error="Invalid form values" withIcon />
        )}
        {valid && formState.status === "pending" && (
          <Message>Submitting response...</Message>
        )}
        {valid && formState.status === "error" && formState.error && (
          <ErrorMessage error={formState.error} />
        )}
        {valid && formState.status === "responded" && (
          <SuccessMessage message="Input response sent" />
        )}
      </>
    );
  }

  return (
    <div className="flex flex-wrap items-center gap-2 justify-end">
      {formState.status === "error" && formState.error && (
        <ErrorMessage error={formState.error} />
      )}
      {formState.status === "responded" && (
        <SuccessMessage message="Input response sent" />
      )}
      {inputType === "button" &&
        options?.map((o) => (
          <Button
            key={o.value}
            disabled={submitting || formState.status === "responded"}
            style={!!o.style ? o.style : "default"}
            type="button"
            onClick={async () => {
              await setFieldValue(name, o.value, true);
              onSubmit();
            }}
          >
            {o.label || o.value}
          </Button>
        ))}
      {(inputType === "select" ||
        inputType === "multiselect" ||
        inputType === "text") && (
        <Button
          disabled={!valid || submitting || formState.status === "responded"}
          type="submit"
          onClick={onSubmit}
        >
          Submit
        </Button>
      )}
      {inputType !== "button" &&
        inputType !== "select" &&
        inputType !== "multiselect" &&
        inputType !== "text" && (
          <ErrorMessage
            as="string"
            error={`Unsupported input type ${inputType}`}
          />
        )}
    </div>
  );
};

const InputFormInner = ({
  autoSubmit = false,
  errors,
  form,
  formState,
  touched,
  values,
  submitting,
  valid,
  setFieldTouched,
  setFieldValue,
  submit,
}: InputFormInnerProps) => {
  const submittedRef = useRef<boolean>(false);
  useEffect(() => {
    const doSumit = async () => {
      await submit();
    };
    if (
      !autoSubmit ||
      !valid ||
      submitting ||
      submittedRef.current ||
      !!formState.error ||
      formState.status === "responded"
    ) {
      return;
    }
    submittedRef.current = true;
    doSumit();
  }, [autoSubmit, formState, submitting, valid]);

  return (
    <Form className="divide-y divide-modal-divide">
      {!autoSubmit && form.status === "finished" && (
        <div className="px-4 py-4">
          <span className="font-semibold leading-6 text-foreground">
            Input has already been responded to.
          </span>
        </div>
      )}
      {form.status === "error" && (
        <div className="px-4 py-4">
          <ErrorMessage withIcon error="Input is in a failed state." />
        </div>
      )}
      {Object.entries(form.inputs).map(([input_name, input]) => (
        <Fragment key={input_name}>
          {!autoSubmit && (
            <div className="px-4 py-4 space-y-2">
              <InputPrompt prompt={input.prompt} formStatus={form.status} />
              <InputOptions
                name={input_name}
                formState={formState}
                errors={errors}
                inputType={input.input_type}
                setFieldTouched={setFieldTouched}
                setFieldValue={setFieldValue}
                submitting={submitting}
                touched={touched}
                options={input.options}
                values={values}
              />
            </div>
          )}
          <div className="px-4 py-4">
            <SubmitOptions
              name={input_name}
              autoSubmit={autoSubmit}
              formState={formState}
              inputType={input.input_type}
              setFieldValue={setFieldValue}
              submitting={submitting}
              options={input.options}
              valid={valid}
              values={values}
              onSubmit={submit}
            />
          </div>
        </Fragment>
      ))}
    </Form>
  );
};

const InputForm = ({ autoSubmit = false }: InputFormProps) => {
  const { id, hash } = useParams();
  const { search } = useLocation();
  const { form, error, loading, postForm } = useFormAPI(id, hash);
  const initialValues = useMemo<InputFormValues>(() => {
    if (!form || !form.inputs) {
      return {};
    }
    const initial = {};
    const formValues = new URLSearchParams(search);
    for (const [input_name, input] of Object.entries(form.inputs)) {
      if (input.input_type === "multiselect" && formValues.has(input_name)) {
        const rawValues = formValues.get(input_name);
        // @ts-ignore this isn't null as the formValues.has check is truthy
        initial[input_name] = rawValues.split(",");
      } else if (input.input_type === "multiselect") {
        initial[input_name] = input.options
          .filter((o) => o.selected)
          .map((o) => o.value);
      } else if (input.input_type === "select" && formValues.has(input_name)) {
        initial[input_name] = formValues.get(input_name);
      } else if (input.input_type === "select") {
        initial[input_name] =
          input.options.find((o) => o.selected)?.value || "";
      } else if (formValues.has(input_name)) {
        initial[input_name] = formValues.get(input_name);
      } else {
        initial[input_name] = "";
      }
    }
    return initial;
  }, [form, search]);
  const [state, setState] = useState<InputFormState>({
    status: "pending",
    error: null,
  });

  useEffect(() => {
    setState((existing) => ({
      ...existing,
      value: form ? form.status : "pending",
      error: null,
    }));
  }, [form, setState]);

  return (
    <div className="mx-auto my-auto">
      <div className="flex flex-col overflow-hidden rounded-lg bg-modal shadow w-screen md:min-w-xl max-w-xl">
        {error && (
          <div className="px-4 py-4">
            <ErrorMessage withIcon error={error} />
          </div>
        )}
        {loading && <div className="px-4 py-4">Loading...</div>}
        {form && (
          <Formik
            initialValues={initialValues}
            validate={(values) => {
              const errors: { [input_name: string]: string } = {};
              for (const [input_name, input] of Object.entries(form.inputs)) {
                if (
                  (input.input_type === "multiselect" && !values[input_name]) ||
                  values[input_name]?.length === 0
                ) {
                  errors[input_name] = `Select a value.`;
                } else if (
                  input.input_type === "select" &&
                  !values[input_name]
                ) {
                  errors[input_name] = `Select a value.`;
                } else if (input.input_type === "text" && !values[input_name]) {
                  errors[input_name] = `Enter a value.`;
                }
              }
              return errors;
            }}
            onSubmit={async (values, { setSubmitting }) => {
              const { error } = await postForm(values);
              if (error) {
                setState({ status: "error", error });
              } else {
                setState({ status: "responded", error: null });
                // await reload();
              }
              setSubmitting(false);
            }}
            enableReinitialize
            validateOnMount
          >
            {({
              errors,
              isSubmitting,
              isValid,
              setFieldValue,
              setFieldTouched,
              handleSubmit,
              touched,
              values,
            }) => {
              return (
                <InputFormInner
                  autoSubmit={autoSubmit}
                  form={form}
                  formState={state}
                  valid={isValid}
                  submitting={isSubmitting}
                  values={values}
                  errors={errors}
                  touched={touched}
                  setFieldValue={setFieldValue}
                  setFieldTouched={setFieldTouched}
                  submit={handleSubmit}
                />
              );
            }}
          </Formik>
        )}
      </div>
      <div className="ml-4 mt-4">
        <FlowpipeLogo />
      </div>
    </div>
  );
};

export default InputForm;
