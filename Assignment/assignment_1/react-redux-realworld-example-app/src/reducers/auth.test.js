import authReducer from './auth';
import {
  LOGIN,
  REGISTER,
  LOGIN_PAGE_UNLOADED,
  REGISTER_PAGE_UNLOADED,
  ASYNC_START,
  UPDATE_FIELD_AUTH
} from '../constants/actionTypes';

describe('auth reducer', () => {
  test('returns initial empty state', () => {
    const state = authReducer(undefined, {});
    expect(state).toEqual({});
  });

  test('handles UPDATE_FIELD_AUTH for email', () => {
    const action = {
      type: UPDATE_FIELD_AUTH,
      key: 'email',
      value: 'test@example.com'
    };
    const state = authReducer({}, action);
    expect(state.email).toBe('test@example.com');
  });

  test('handles UPDATE_FIELD_AUTH for password', () => {
    const action = {
      type: UPDATE_FIELD_AUTH,
      key: 'password',
      value: 'password123'
    };
    const state = authReducer({}, action);
    expect(state.password).toBe('password123');
  });

  test('handles LOGIN success', () => {
    const action = {
      type: LOGIN,
      payload: {
        user: {
          email: 'test@example.com',
          token: 'jwt-token',
          username: 'testuser'
        }
      }
    };
    const state = authReducer({}, action);
    expect(state.inProgress).toBe(false);
    expect(state.errors).toBeNull();
  });

  test('handles LOGIN error', () => {
    const action = {
      type: LOGIN,
      error: true,
      payload: {
        errors: {
          'email or password': ['is invalid']
        }
      }
    };
    const state = authReducer({}, action);
    expect(state.inProgress).toBe(false);
    expect(state.errors).toEqual({ 'email or password': ['is invalid'] });
  });

  test('handles REGISTER success', () => {
    const action = {
      type: REGISTER,
      payload: {
        user: {
          email: 'newuser@example.com',
          token: 'jwt-token',
          username: 'newuser'
        }
      }
    };
    const state = authReducer({}, action);
    expect(state.inProgress).toBe(false);
    expect(state.errors).toBeNull();
  });

  test('handles REGISTER error', () => {
    const action = {
      type: REGISTER,
      error: true,
      payload: {
        errors: {
          email: ['has already been taken'],
          username: ['has already been taken']
        }
      }
    };
    const state = authReducer({}, action);
    expect(state.inProgress).toBe(false);
    expect(state.errors).toEqual({
      email: ['has already been taken'],
      username: ['has already been taken']
    });
  });

  test('handles ASYNC_START for LOGIN', () => {
    const action = {
      type: ASYNC_START,
      subtype: LOGIN
    };
    const state = authReducer({}, action);
    expect(state.inProgress).toBe(true);
  });

  test('handles ASYNC_START for REGISTER', () => {
    const action = {
      type: ASYNC_START,
      subtype: REGISTER
    };
    const state = authReducer({}, action);
    expect(state.inProgress).toBe(true);
  });

  test('handles LOGIN_PAGE_UNLOADED clears state', () => {
    const initialState = {
      email: 'test@example.com',
      password: 'password',
      inProgress: true
    };
    const action = { type: LOGIN_PAGE_UNLOADED };
    const state = authReducer(initialState, action);
    expect(state).toEqual({});
  });

  test('handles REGISTER_PAGE_UNLOADED clears state', () => {
    const initialState = {
      email: 'test@example.com',
      username: 'testuser',
      password: 'password'
    };
    const action = { type: REGISTER_PAGE_UNLOADED };
    const state = authReducer(initialState, action);
    expect(state).toEqual({});
  });

  test('preserves existing state when updating field', () => {
    const initialState = {
      email: 'test@example.com',
      username: 'testuser'
    };
    const action = {
      type: UPDATE_FIELD_AUTH,
      key: 'password',
      value: 'newpassword'
    };
    const state = authReducer(initialState, action);
    expect(state).toEqual({
      email: 'test@example.com',
      username: 'testuser',
      password: 'newpassword'
    });
  });

  test('handles LOGIN error without payload.errors', () => {
    const action = {
      type: LOGIN,
      error: true,
      payload: null
    };
    const state = authReducer({}, action);
    expect(state.inProgress).toBe(false);
    expect(state.errors).toBeNull();
  });
});
