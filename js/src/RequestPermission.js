// A screen for an app to request permissions from the user.

import React, { Component } from "react";
import Button from "@material-ui/core/Button";

import Styles from "./Styles";

export default class RequestPermission extends Component {
  // props.host is the entity requesting permissions
  // props.permissions is the permissions we are requesting
  // props.accept is what we call when the user accepts
  // props.deny is what we call when the user denies
  constructor(props) {
    super(props);
  }

  // Return a list of human-readable strings for the permissions we want
  permissionList() {
    let answer = [];
    if (this.props.permissions.publicKey) {
      answer.push("to know your public identity");
    }
    return answer;
  }

  render() {
    return (
      <div style={Styles.popup}>
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            justifyContent: "center",
            flex: 3
          }}
        >
          <h1>Grant Permission?</h1>
          <h2>{this.props.host} requests:</h2>
          <ol>
            {this.permissionList().map(x => (
              <li>{x}</li>
            ))}
          </ol>
        </div>

        <form
          style={{
            flex: 2,
            display: "flex",
            flexDirection: "column",
            width: "100%",
            justifyContent: "space-evenly"
          }}
          onSubmit={event => {
            event.preventDefault();
            this.handleInput();
          }}
        >
          <Button
            variant="contained"
            color="primary"
            type="submit"
            onClick={() => this.props.accept()}
          >
            Accept
          </Button>
          <Button
            variant="contained"
            color="default"
            onClick={() => this.props.deny()}
          >
            Deny
          </Button>
        </form>
      </div>
    );
  }
}
