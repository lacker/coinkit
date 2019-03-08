import KeyPair from "./KeyPair";
import Message from "./Message";
import MockLocalStorage from "./MockLocalStorage";
import SignedMessage from "./SignedMessage";
import Storage from "./Storage";
import TrustedClient from "./TrustedClient";

test("Operation message signing", async () => {
  let local = new MockLocalStorage();
  let storage = new Storage(local);
  let kp = KeyPair.fromSecretPhrase("blorp");
  await storage.setPasswordAndData("monkey", kp, {});
  let client = new TrustedClient(storage);
  let unsigned = new Message("Operation", {
    operations: [
      {
        type: "CreateDocument",
        operation: {
          sequence: 1,
          fee: 1,
          data: {
            foo: "bar"
          }
        }
      }
    ]
  });
  let message = client.signOperationMessage(unsigned);
  let signed = SignedMessage.fromSigning(message, kp);

  // See tests of this string in operation_message_test.go
  expect(signed.serialize()).toEqual(
    'e:0x5b8f312caed13ac35805c69e889d24bbd3df7d6285fbca173cce47e7402a5d0bddf3:a09g9sLYa7xUSdLvG1K1r4kaD9Iu2+bMjFhQtnWUKrvK/UPHh3GpcZPNLbDc1vibkZs1TqF1QNz9B2u7FEzjBA:{"message":{"operations":[{"operation":{"data":{"foo":"bar"},"fee":1,"sequence":1,"signer":"0x5b8f312caed13ac35805c69e889d24bbd3df7d6285fbca173cce47e7402a5d0bddf3"},"signature":"a6a+P+y0UkhFLXonoQXobIMFCXJFGyMiCc5yjQDOa0fz9jnaQf/LKYkzN2CII4nSQjIiromm/bzOaLJumQfpCg","type":"Create"}]},"type":"Operation"}'
  );
});
