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
  }

  // Returns the encrypted data, lazy-loading if needed.
  // Since encrypted data is lazy-loaded it may just have not been loaded yet.
  async getEncrypted() {
    if (!this.encrypted) {
      this.encrypted = await getFromLocalStorage("encrypted");
    }
    return this.encrypted;
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
