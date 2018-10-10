// The Storage class wraps Chrome storage to handle encryption.
// Anything kept in Chrome storage is encrypted, because other processes on the user's
// machine may be able to read Chrome storage.
// A Storage object should only be created from the background page, because it
// stores encryption keys in memory, and thus should be as persistent as possible.

import { createStore } from "redux";

import Cipher from "./Cipher";
import KeyPair from "./KeyPair";
import { loadFromStorage } from "./actions";
import reducers from "./reducers";

export default class Storage {
  constructor() {
    // encrypted should be an object holding iv, salt, and ciphertext keys.
    this.encrypted = null;

    // this.data is the decrypted version of encrypted.
    // it holds keyPair and permissions
    this.data = null;

    this.request = null;
    this.password = null;
    this.initialized = false;
  }

  // Once the Storage object is initialized, it will act as a write-through cache
  // for browser storage.
  // Before it is initialized, we shouldn't write to it.
  async init() {
    if (this.initialized) {
      return;
    }

    this.encrypted = await getLocalStorage("encrypted");
    if (
      typeof this.encrypted != "object" ||
      !this.encrypted.iv ||
      !this.encrypted.salt ||
      !this.encrypted.ciphertext
    ) {
      this.encrypted = null;
    }

    this.password = null;
    this.data = null;
    this.initialized = true;
  }

  static async get() {
    let storage = chrome.extension.getBackgroundPage().storage;
    if (!storage) {
      throw new Error("cannot find storage");
    }
    await storage.init();
    return storage;
  }

  async handleStoreUpdate(store) {
    let state = store.getState();
    this.request = state.request;

    if (state.password == null && state.keyPair == null) {
      this.logOut();
    } else if (state.password != null) {
      await this.setPasswordAndData(
        state.password,
        state.keyPair,
        state.permissions
      );
    }
  }

  // Makes a redux store that is persisted using this storage object
  static async makeStore() {
    // Each popup gets its own redux store object.
    // I tried to let them share one but ran into weird bugs.
    let store = createStore(reducers);
    let storage = await Storage.get();
    let action = loadFromStorage(storage);
    store.dispatch(action);

    // Save all state updates when there is a password set to retrieve them
    store.subscribe(() => {
      storage.handleStoreUpdate(store);
    });

    return store;
  }

  // Drops the password and decrypted data
  logOut() {
    this.password = null;
    this.data = null;
  }

  // Returns a nice form of the data, an object with:
  // keyPair: the keypair for the logged-in user
  // permissions: an object with domain -> permissions objects
  // this.data is jsonable, getData() returns something inflated with objects.
  // Returns null if there is no data.
  getData() {
    if (!this.data) {
      return null;
    }
    let kp;
    try {
      kp = KeyPair.fromSerialized(this.data.keyPair);
    } catch (e) {
      console.log("invalid keypair in data:", this.data);
      return null;
    }

    return {
      keyPair: kp,
      permissions: this.data.permissions
    };
  }

  // Returns whether this password is a valid password for our encrypted data.
  // If it is valid, sets both password and data.
  async checkPassword(password) {
    await this.init();

    if (!this.encrypted) {
      return false;
    }
    let json = Cipher.decrypt(
      password,
      this.encrypted.iv,
      this.encrypted.salt,
      this.encrypted.ciphertext
    );
    if (!json) {
      return false;
    }
    try {
      this.data = JSON.parse(json);
    } catch (e) {
      return false;
    }

    this.password = password;
    return true;
  }

  async setPasswordAndData(password, keyPair, permissions) {
    await this.init();

    let data = {
      keyPair: keyPair.serialize(),
      permissions: permissions
    };

    let json = JSON.stringify(data);
    let iv = Cipher.makeIV();
    let salt = Cipher.makeSalt();
    let ciphertext = Cipher.encrypt(password, iv, salt, json);
    this.encrypted = {
      iv: iv,
      salt: salt,
      ciphertext: ciphertext
    };
    this.data = data;
    this.password = password;

    await setLocalStorage("encrypted", this.encrypted);
  }
}

// A helper to fetch local storage and return a promise
// Resolves to null if there is no data but the fetch worked
async function getLocalStorage(key) {
  return new Promise((resolve, reject) => {
    chrome.storage.local.get([key], result => {
      resolve(result[key]);
    });
  });
}

async function setLocalStorage(key, value) {
  return new Promise((resolve, reject) => {
    chrome.storage.local.set({ key, value }, () => {
      resolve();
    });
  });
}
