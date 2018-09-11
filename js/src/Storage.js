// The Storage class wraps Chrome storage to handle encryption.
// Anything kept in Chrome storage is encrypted, because other processes on the user's
// machine may be able to read Chrome storage.
// A Storage object should only be created from the background page, because it
// stores encryption keys in memory, and thus should be as persistent as possible.
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

    this.encrypted = await getFromLocalStorage("encrypted");
    this.initialized = true;
  }

  // Returns whether this password is a valid password for our encrypted data.
  // If it is valid, sets both password and data.
  async checkPassword(password) {
    await this.init();

    // TODO
  }
}

// A helper to fetch local storage and return a promise
// Resolves to null if there is no data but the fetch worked
async function getFromLocalStorage(key) {
  return new Promise((resolve, reject) => {
    chrome.storage.local.get([key], result => {
      resolve(result[key]);
    });
  });
}
