import { combineReducers } from "redux";
import {
  LOAD_FROM_STORAGE,
  LOG_OUT,
  NEW_KEY_PAIR,
  NEW_PASSWORD
} from "./actions";

function password(state = null, action) {
  switch (action.type) {
    case LOAD_FROM_STORAGE:
      return action.password;

    case LOG_OUT:
      return null;

    case NEW_KEY_PAIR:
      return null;

    case NEW_PASSWORD:
      return action.password;

    default:
      return state;
  }
}

function keyPair(state = null, action) {
  switch (action.type) {
    case LOAD_FROM_STORAGE:
      return action.keyPair;

    case LOG_OUT:
      return null;

    case NEW_KEY_PAIR:
      return action.keyPair;

    default:
      return state;
  }
}

function permissions(state = {}, action) {
  switch (action.type) {
    case LOAD_FROM_STORAGE:
      return action.permissions;

    case LOG_OUT:
      return {};

    case NEW_KEY_PAIR:
      return {};

    default:
      return state;
  }
}

export default combineReducers({ password, keyPair, permissions });
