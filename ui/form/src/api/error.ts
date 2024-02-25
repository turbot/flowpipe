class PipelingError extends Error {
  private _code: number;
  private _id: string | null;
  private _title: string | null;

  constructor(
    code: number,
    id: string | null,
    message: string,
    title: string | null,
  ) {
    super(message);
    this._code = code;
    this._id = id;
    this._title = title;
  }

  public static fromMessage(message: string): PipelingError {
    return new PipelingError(500, null, message, null);
  }

  get code(): any {
    return this._code;
  }

  get id(): any {
    return this._id;
  }

  get title(): any {
    return this._title;
  }
}

const errorAwareRequest = (
  request,
  url,
  dataOrConfig: Object | null = null,
  config: Object | null = null,
) => {
  return request(url, dataOrConfig, config).catch((err) => {
    if (err.response) {
      // The request was made and the server responded with a status code
      // that falls out of the range of 2xx
      const {
        status: code,
        instance,
        detail: message,
        validation_errors: errors = [],
        title,
      } = err.response.data || {};
      let errorMessage = "";
      if (errors && errors.length > 0) {
        errorMessage = errors.map((error) => error.message).join(" ");
      }
      throw new PipelingError(
        code,
        instance,
        `${message}${errorMessage ? `: ${errorMessage}` : ""}`,
        title,
      );
    } else if (err.request) {
      // The request was made but no response was received
      // `error.request` is an instance of XMLHttpRequest in the browser and an instance of
      // http.ClientRequest in node.js
      throw new PipelingError(err.code, null, err.message, null);
    } else {
      // Something happened in setting up the request that triggered an Error
      throw new PipelingError(err.code, null, err.message, null);
    }
  });
};

const getErrorMessage = (error: PipelingError | string | null | undefined) => {
  if (!error) {
    return "";
  }
  if (typeof error === "string") {
    return error;
  }
  return error.message;
};

export { errorAwareRequest, getErrorMessage, PipelingError };
