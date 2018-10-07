// Redux actions
import KeyPair from "./KeyPair";

export const GRANT_PERMISSION = "GRANT_PERMISSION";
export const LOAD_FROM_STORAGE = "LOAD_FROM_STORAGE";
export const LOG_OUT = "LOG_OUT";
export const NEW_KEY_PAIR = "NEW_KEY_PAIR";
export const NEW_PASSWORD = "NEW_PASSWORD";

export function grantPermission(permission) {
  return {
    type: GRANT_PERMISSION,
    permission: permission
  };
}

export function logOut() {
  return { type: LOG_OUT };
}

export function loadFromStorage(storage) {
  let data = storage.getData();
  if (!data) {
    return logOut();
  }

  return {
    type: LOAD_FROM_STORAGE,
    keyPair: data.keyPair,
    password: storage.password,
    permissions: data.permissions
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
