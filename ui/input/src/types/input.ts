export interface PipelineInputOption {
  label?: string;
  value: string;
  selected?: boolean;
}

export interface PipelineInput {
  execution_id: string;
  pipeline_execution_id: string;
  step_execution_id: string;
  string: string;
  prompt?: string;
  input_type:
    | "button"
    | "text"
    | "password"
    | "textarea"
    | "select"
    | "multiselect"
    | "checkbox"
    | "radio";
  options: PipelineInputOption[];
  response_url: string;
}
