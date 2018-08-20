import KeyPair from "./KeyPair";
import SignedMessage from "./SignedMessage";

test("SignedMessage", () => {
  let m = { Number: 4 };
  let kp = KeyPair.fromSecretPhrase("foo");
  let sm = SignedMessage.fromSigning(m, kp);
  let serialized = sm.serialize();
  let sm2 = SignedMessage.fromSerialized(serialized);
  expect(sm2).toEqual(sm);
});
