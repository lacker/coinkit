// The Storage class wraps Chrome storage to handle encryption.
// Anything kept in Chrome storage is encrypted, because other processes on the user's
// machine may be able to read Chrome storage.
// A Storage object should only be created from the background page, because it
// stores encryption keys in memory, and thus should be as persistent as possible.
import Cipher from "./Cipher";

export default class Storage {
  constructor() {
    this.password = null;
    this.encrypted = null;
    this.data = null;
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
    this.password = null;
    this.data = null;
    this.initialized = true;
  }

  // Returns whether this password is a valid password for our encrypted data.
  // If it is valid, sets both password and data.
  async checkPassword(password) {
    await this.init();

    if (!this.encrypted) {
      return false;
    }
    let json = Cipher.decrypt(password, this.encrypted);
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

  async setPasswordAndData(password, data) {
    await this.init();

    let json = JSON.stringify(data);
    this.encrypted = Cipher.encrypt(password, json);
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
