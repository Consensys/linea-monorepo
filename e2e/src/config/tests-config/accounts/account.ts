export default class Account {
  public readonly privateKey: string;
  public readonly address: string;

  constructor(privateKey: string, address: string) {
    this.privateKey = privateKey;
    this.address = address;
  }
}
