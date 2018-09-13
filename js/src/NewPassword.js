// A screen to let the user create a password to locally encrypt keys.

import React, { Component } from "react";
import Button from "@material-ui/core/Button";
import TextField from "@material-ui/core/TextField";

export default class NewPassword extends Component {
  // props.popup is a reference to the root popup
  constructor(props) {
    super(props);

    this.popup = props.popup;

    this.state = {
      password: "",
      repeatPassword: "",
      warning: false
    };
  }

  render() {
    let style = {
      display: "flex",
      flexDirection: "column",
      justifyContent: "flex-start",
      alignItems: "center",
      width: 360,
      padding: 30
    };

    return (
      <div style={style}>
        <h1>Choose a password</h1>
        <form
          onSubmit={event => {
            event.preventDefault();

            if (this.state.password == "") {
              this.passwordField.focus();
              return;
            }

            if (this.state.repeatPassword == "") {
              this.repeatPasswordField.focus();
              return;
            }

            if (this.state.password != this.state.repeatPassword) {
              this.setState({
                repeatPassword: "",
                warning: true
              });
              this.repeatPasswordField.focus();
              return;
            }

            // TODO: handle the case where we actually have a new password
            console.log("XXX", this.state.password);
          }}
        >
          <div>Password</div>
          <TextField
            type="password"
            autoFocus={true}
            value={this.state.password}
            onChange={event => {
              this.setState({
                password: event.target.value
              });
            }}
            inputRef={input => (this.passwordField = input)}
          />
          <div>
            {this.state.warning
              ? "Passwords must match"
              : "Repeat your password"}
          </div>
          <TextField
            type="password"
            value={this.state.repeatPassword}
            error={this.state.warning}
            onChange={event => {
              this.setState({
                repeatPassword: event.target.value
              });
            }}
            inputRef={input => (this.repeatPasswordField = input)}
          />
          <div />
          <Button variant="contained" color="primary" type="submit">
            Create Password
          </Button>
        </form>
      </div>
    );
  }
}
