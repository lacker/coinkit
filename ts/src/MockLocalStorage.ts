export default class MockLocalStorage {
  constructor() {
    this.data = {};
  }

  numKeys() {
    let answer = 0;
    for (let key in this.data) {
      answer++;
    }
    return answer;
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
