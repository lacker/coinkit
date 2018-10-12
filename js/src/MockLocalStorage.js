export default class MockLocalStorage {
  constructor() {
    this.data = {};
  }

  async get(key) {
    if (!(key in this.data)) {
      return null;
    }
    return this.data[key];
  }

  async set(key, value) {
    this.data[key] = value;
  }
}
