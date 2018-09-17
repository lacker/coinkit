// Client is designed to be included in applications and run in untrusted application
// environment. It gets permissions by requesting them from the extension, whose code
// is trusted.
export default class Client {
  constructor() {
    // publicKey is null before permissions are acquired
    this.publicKey = null;
  }

  // Requests public key permission from the extension if we don't already have it.
  // Returns null if permission is denied.
  async getPublicKey() {
    // TODO
  }
}
