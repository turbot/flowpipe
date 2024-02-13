import { useEffect } from "react";
import { useInputAPI } from "src/api/pipeline.ts";
import { useParams } from "react-router-dom";

const InputApp = () => {
  const { id, hash } = useParams();
  const { getInput } = useInputAPI(id, hash);
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
      if (!id || !hash) {
        return;
      }
      const { input, error } = await getInput();
      if (error) {
        console.error(error);
      } else {
        console.log(input);
      }
    };
    fetchInputInfo();
  }, [id, hash]);
  return <div></div>;
};

export default InputApp;
