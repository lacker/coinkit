import MockLocalStorage from "./MockLocalStorage";
import Storage from "./Storage";

test("basic redux flow", () => {
  let storage = new Storage(MockLocalStorage());
});
