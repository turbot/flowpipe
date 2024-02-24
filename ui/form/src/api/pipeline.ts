import axios from "axios";
import { API_BASE_PATH, asyncRequest, useFlowpipeSWR } from "src/api/index.ts";
import { PipelineInput, PipelineInputResponse } from "src/types/input.ts";

export const useInputAPI = (
  id: string | null | undefined,
  hash: string | null | undefined,
) => {
  const { data, loading, error, mutate } = useFlowpipeSWR<PipelineInput>(
    !!id && !!hash ? `${API_BASE_PATH}/input/${id}/${hash}` : null,
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

  const postInput = async (
    response_url: string,
    input_result: PipelineInputResponse,
  ) => {
    const { result, error } = await asyncRequest<PipelineInput>(
      axios.post,
      response_url,
      input_result,
      //withIfMatchHeaderConfig(pipeline), // TODO: raise RE IfMatch header
    );
    console.log({ response_url, input_result, result, error });
    return {
      input: result ? result.data : null,
      error: error ? error : null,
    };
  };

  return {
    input: data,
    error,
    loading,
    reload: mutate,
    postInput,
  };
};
