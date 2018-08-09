// The root to display in the extension popup.

import React, { Component } from "react";

export default class Popup extends Component {
  constructor(props) {
    super(props);

    this.state = { n: 0 };

    this.click = this.click.bind(this);
  }

  click(newN) {
    this.setState({ n: newN });
  }

  render(props) {
    return (
      <div>
        <button onClick={() => this.click(this.state.n + 1)} />
        <h1>hello world: {this.state.n}</h1>
      </div>
    );
  }
}
