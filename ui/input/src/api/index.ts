import axios, { AxiosHeaders } from "axios";
import { errorAwareRequest, PipelingError } from "./error";

axios.defaults.headers.post["Content-Type"] = "application/json";

const API_ROOT = "/api";
const API_VERSION = "v0";
export const API_BASE_PATH = `${API_ROOT}/${API_VERSION}`;

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
