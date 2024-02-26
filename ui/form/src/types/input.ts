export interface PipelineInputOption {
  label?: string;
  value: string;
  selected?: boolean;
}

export type PipelineFormStatus =
  | "pending"
  | "starting"
  | "started"
  | "finished"
  | "error";

export type PipelineInputType = "button" | "text" | "select" | "multiselect";

export interface PipelineFormInput {
  prompt?: string;
  input_type: PipelineInputType;
  options: PipelineInputOption[];
}

export interface PipelineFormInputs {
  [input_name: string]: PipelineFormInput;
}

export interface PipelineForm {
  execution_id: string;
  pipeline_execution_id: string;
  step_execution_id: string;
  status: PipelineFormStatus;
  inputs: PipelineFormInputs;
}

export interface InputFormValues {
  [input_name: string]: string | string[];
}
