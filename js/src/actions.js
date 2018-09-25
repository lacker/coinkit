// Redux actions
import KeyPair from "./KeyPair";

export const LOAD_STATE = "LOAD_STATE";
export const LOG_OUT = "LOG_OUT";
export const NEW_KEY_PAIR = "NEW_KEY_PAIR";
export const SET_PASSWORD = "SET_PASSWORD";

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
    kp = KeyPair.fromSerialized(this.storage.data.keyPair);
  } catch (e) {
    console.log("invalid keypair from storage:", this.storage.data);
    return logOut();
  }

  return {
    type: LOAD_STATE,
    password: storage.password,
    keyPair: kp
  };
}

export function newKeyPair(kp) {
  return {
    type: LOAD_STATE,
    keyPair: kp
  };
}

export function setPassword(password) {
  return {
    type: SET_PASSWORD,
    password: password
  };
}
