import axios, { AxiosHeaders } from "axios";
import { errorAwareRequest, PipelingError } from "./error";
import useSWR from "swr";

axios.defaults.headers.post["Content-Type"] = "application/json";

const API_ROOT = "/api";
const API_VERSION = "v0";
export const API_BASE_PATH = `${API_ROOT}/${API_VERSION}`;

export const API_FETCHER = (url) =>
  errorAwareRequest(axios.get, url).then((res) => res.data);

export type AsyncRequestResult<T = any> = {
  result: { data: T | null; headers: AxiosHeaders } | null;
  error?: PipelingError | null;
};

export const asyncRequest = async <T = any>(
  ...request
): Promise<AsyncRequestResult<T>> => {
  try {
    // @ts-ignore
    const result = await errorAwareRequest(...request);
    return { result, error: null };
  } catch (err) {
    return { result: null, error: err as PipelingError };
  }
};

export const useFlowpipeSWR = <TData = any>(
  path: string | [any, ...unknown[]] | readonly [any, ...unknown[]] | null,
  config?: any,
) => {
  const { data, isLoading, isValidating, ...rest } = useSWR<
    TData,
    PipelingError
  >(path, API_FETCHER, {
    ...config,
    revalidateOnFocus: false,
  });
  return {
    data: data || null,
    loading: !data && (isLoading || isValidating),
    isValidating,
    ...rest,
  };
};
