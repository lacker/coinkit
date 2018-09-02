// The root to display in the sample app.

import React, { Component } from "react";

export default class App extends Component {
  constructor(props) {
    super(props);

    this.state = {
      user: null
    };
  }

  render(props) {
    return (
      <div>
        <h1>this is the sample app</h1>
        <h1>{this.state.user || "nobody"} is logged in</h1>
      </div>
    );
  }
}
