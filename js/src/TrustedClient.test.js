import Message from "./Message";
import MockLocalStorage from "./MockLocalStorage";
import Storage from "./Storage";
import TrustedClient from "./TrustedClient";

test("Operation message signing", () => {
  let local = new MockLocalStorage();
  let storage = new Storage(local);
  let client = new TrustedClient(storage);
  let unsigned = new Message("Operation", {
    operations: [
      {
        type: "Create",
        operation: {
          signer: client.getKeyPair().getPublicKey(),
          sequence: 1,
          fee: 1,
          data: {
            foo: "bar"
          }
        }
      }
    ]
  });
  let signed = client.signOperationMessage(unsigned);

  // TODO: check various things about signed
});
