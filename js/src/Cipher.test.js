import Cipher from "./Cipher";

test("Cipher basics", () => {
  let enc = Cipher.encrypt("password", "data");
  expect(Cipher.decrypt("wrong-password", enc)).toBe(null);
  expect(Cipher.decrypt("password", enc)).toBe("data");
});
