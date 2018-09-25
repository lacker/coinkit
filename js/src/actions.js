// Redux actions
import KeyPair from "./KeyPair";

export const LOAD_FROM_STORAGE = "LOAD_FROM_STORAGE";
export const LOG_OUT = "LOG_OUT";
export const NEW_KEY_PAIR = "NEW_KEY_PAIR";
export const NEW_PASSWORD = "NEW_PASSWORD";

export function logOut() {
  return { type: LOG_OUT };
}

export function loadFromStorage(storage) {
  if (!storage.data) {
    return logOut();
  }

  if (typeof storage.data != "object") {
    console.log("bad stored data:", storage.data);
    return logOut();
  }

  let kp;
  try {
    kp = KeyPair.fromSerialized(storage.data.keyPair);
  } catch (e) {
    console.log("invalid keypair from storage:", storage.data);
    return logOut();
  }

  return {
    type: LOAD_FROM_STORAGE,
    password: storage.password,
    keyPair: kp
  };
}

export function newKeyPair(kp) {
  return {
    type: NEW_KEY_PAIR,
    keyPair: kp
  };
}

export function newPassword(password) {
  return {
    type: NEW_PASSWORD,
    password: password
  };
}
