import InputForm from "components/InputForm";
import { PipelineInput } from "src/types/input.ts";
import { useEffect, useState } from "react";
import { useInputAPI } from "src/api/pipeline.ts";
import { useParams } from "react-router-dom";

const InputApp = () => {
  const { id, hash } = useParams();
  const { getInput, postInput } = useInputAPI(id, hash);
  const [loading, setLoading] = useState<boolean>(true);
  const [input, setInput] = useState<PipelineInput | null>(null);
  // const location = useLocation();
  // const { id, hash } = useMemo(() => {
  //   const searchParams = new URLSearchParams(location.search);
  //   return {
  //     id: searchParams.get("id"),
  //     hash: searchParams.get("hash"),
  //   };
  // }, [location.search]);
  // const { getInput } = useInputAPI(id, hash);

  useEffect(() => {
    const fetchInputInfo = async () => {
      if (!id || !hash || !!input) {
        return;
      }
      const { input: fetchedInput, error } = await getInput();
      if (error) {
        console.error(error);
      } else {
        console.log(fetchedInput);
        setInput(fetchedInput);
        setLoading(false);
      }
    };
    fetchInputInfo();
  }, [id, hash, input, setInput]);

  return (
    <div className="mx-auto my-auto">
      {loading && <span className="italic">Loading...</span>}
      {!loading && <InputForm input={input} onSubmit={postInput} />}
    </div>
  );
};

export default InputApp;
