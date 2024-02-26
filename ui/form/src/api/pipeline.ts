import axios from "axios";
import {
  API_BASE_PATH,
  asyncRequest,
  useFlowpipeSWR,
} from "@flowpipe/api/index";
import { InputFormValues, PipelineForm } from "@flowpipe/types/input";

export const useFormAPI = (
  id: string | null | undefined,
  hash: string | null | undefined,
) => {
  const { data, loading, error, mutate } = useFlowpipeSWR<PipelineForm>(
    !!id && !!hash ? `${API_BASE_PATH}/form/${id}/${hash}` : null,
  );

  // const getInput = async () => {
  //   const { result, error } = await asyncRequest<PipelineInput>(
  //     axios.get,
  //     `${API_BASE_PATH}/input/${id}/${hash}`,
  //   );
  //   return {
  //     input: result ? result.data : null,
  //     error: error ? error : null,
  //   };
  // };

  const postForm = async (form_result: InputFormValues) => {
    const { result, error } = await asyncRequest<PipelineForm>(
      axios.post,
      `${API_BASE_PATH}/form/${id}/${hash}/submit`,
      form_result,
      //withIfMatchHeaderConfig(pipeline), // TODO: raise RE IfMatch header
    );
    return {
      form: result ? result.data : null,
      error: error ? error : null,
    };
  };

  return {
    form: data,
    error,
    loading,
    reload: mutate,
    postForm,
  };
};
