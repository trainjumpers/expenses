import axios from "axios";

export const fetchSomething = async () => {
  const response = await axios.get(
    "https://jsonplaceholder.typicode.com/todos/1"
  );
  return response.data;
};
