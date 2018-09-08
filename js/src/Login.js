// A screen to prompt a login.

import React, { Component } from "react";
import Button from "@material-ui/core/Button";
import TextField from "@material-ui/core/TextField";

export default class Login extends Component {
  // props.callback takes a keypair once the user has logged in
  constructor(props) {
    super(props);

    this.callback = props.callback;
  }

  render() {
    return (
      <div>
        <TextField>foo</TextField>
      </div>
    );
  }
}
