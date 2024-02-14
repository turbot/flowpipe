export interface PipelineInputOption {
  label?: string;
  value: string;
  selected?: boolean;
}

export type PipelineInputType =
  | "button"
  | "text"
  | "select"
  | "multiselect"
  | "combo"
  | "multicombo";

export interface PipelineInput {
  execution_id: string;
  pipeline_execution_id: string;
  step_execution_id: string;
  status: string;
  prompt?: string;
  input_type: PipelineInputType;
  options: PipelineInputOption[];
  response_url: string;
}

export interface PipelineInputResponse {
  execution_id: string;
  pipeline_execution_id: string;
  step_execution_id: string;
  values: string[];
}
