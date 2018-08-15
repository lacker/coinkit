import KeyPair from "./KeyPair";

test("KeyPair can be constructed from a private key", () => {
  let kp = KeyPair.fromPrivateKey(
    "1YBC5qpaM14DrVdsap5DtBWRv9IHf3Leyd95MOSSBV1cua0Uhxl/Y6afXFHIvFP+/m9V99AeVQndCtBV1E7/Tw"
  );
  expect(kp.publicKey).toBeDefined();
  expect(kp.privateKey).toBeDefined();
});

test("KeyPair.readPublicKey", () => {
  expect(() => {
    KeyPair.readPublicKey("blah");
  }).toThrow();
  expect(() => {
    KeyPair.readPublicKey("0xblahblahblah");
  }).toThrow();
  expect(() => {
    KeyPair.readPublicKey(
      "0x12345678901234567890123456789012345678901234567890123456789012345678"
    );
  }).toThrow();

  KeyPair.readPublicKey(
    "0x5cb9ad1487197f63a69f5c51c8bc53fefe6f55f7d01e5509dd0ad055d44eff4f9a86"
  );
});

/*
test("KeyPair's signatures match Go", () => {
  let serialized = `{
  "Public": "0x5cb9ad1487197f63a69f5c51c8bc53fefe6f55f7d01e5509dd0ad055d44eff4f9a86",
  "Private": "1YBC5qpaM14DrVdsap5DtBWRv9IHf3Leyd95MOSSBV1cua0Uhxl/Y6afXFHIvFP+/m9V99AeVQndCtBV1E7/Tw"
}
`;
  let kp = KeyPair.fromSerialized(serialized);

  // XXX
});
*/
