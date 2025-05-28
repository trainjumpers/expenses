export const createResource = <T>(asyncFunction: () => Promise<T>) => {
  type State = "pending" | "success" | "error";
  let status: State = "pending";
  let result: T;
  const suspender = asyncFunction().then(
    (res) => {
      status = "success";
      result = res;
    },
    (err) => {
      status = "error";
      result = err;
    }
  );

  return {
    read() {
      if (status === "pending") {
        throw suspender;
      } else if (status === "error") {
        throw result;
      } else if (status === "success") {
        return result;
      }
    },
  };
};
