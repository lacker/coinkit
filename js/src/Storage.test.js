import MockLocalStorage from "./MockLocalStorage";
import Storage from "./Storage";

test("basic redux flow", () => {
  Storage.mock = new Storage(MockLocalStorage());
});
