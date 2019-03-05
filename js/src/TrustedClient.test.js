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
        type: "Create",
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
    'e:0x5b8f312caed13ac35805c69e889d24bbd3df7d6285fbca173cce47e7402a5d0bddf3:D97rQTtkUet8Ph24vm+ZkzJhULzEqI8dX6NhK8M6ivv7tAywLsIUW8OKn1fpqyLNmLRbndzIPdvE/hV01v9xDw:{"message":{"operations":[{"operation":{"data":{"foo":"bar"},"fee":1,"sequence":1,"signer":"0x5b8f312caed13ac35805c69e889d24bbd3df7d6285fbca173cce47e7402a5d0bddf3"},"signature":"wIS9/HZQQn8exsAZT2mmhPPC95UBBSqSxFmCknymwRozxe//emT0vscf8eq55n4fZ0JO+4NiDpknlCi4UKYmDA","type":"Create"}]},"type":"Operation"}'
  );
});
