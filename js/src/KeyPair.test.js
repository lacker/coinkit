import KeyPair from "./KeyPair";

test("KeyPair has a public key", () => {
  let kp = new KeyPair();
  expect(kp.publicKey).toBeDefined();
});
