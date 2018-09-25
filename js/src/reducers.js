import { combineReducers } from "redux";
import { LOAD_STATE, LOG_OUT, NEW_KEY_PAIR, SET_PASSWORD } from "./actions";

function password(state = null, action) {
  switch (action.type) {
    case LOAD_STATE:
      return action.password;

    case LOG_OUT:
      return null;

    case NEW_KEY_PAIR:
      return null;

    case SET_PASSWORD:
      return action.password;

    default:
      return state;
  }
}

function keyPair(state = null, action) {
  switch (action.type) {
    case LOAD_STATE:
      return action.keyPair;

    case LOG_OUT:
      return null;

    case NEW_KEY_PAIR:
      return action.keyPair;

    default:
      return state;
  }
}

export default combineReducers({ password, keyPair });
