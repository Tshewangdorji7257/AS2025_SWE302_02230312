import { promiseMiddleware, localStorageMiddleware } from './middleware';
import { LOGIN, REGISTER, LOGOUT, ASYNC_START, ASYNC_END } from './constants/actionTypes';

describe('promiseMiddleware', () => {
  let store, next, invoke;

  beforeEach(() => {
    store = {
      getState: jest.fn(() => ({ viewChangeCounter: 0 })),
      dispatch: jest.fn()
    };
    next = jest.fn();
    invoke = (action) => promiseMiddleware(store)(next)(action);
  });

  test('passes non-promise actions to next', () => {
    const action = { type: 'TEST_ACTION', payload: 'test' };
    invoke(action);
    
    expect(next).toHaveBeenCalledWith(action);
    expect(store.dispatch).not.toHaveBeenCalled();
  });

  test('dispatches ASYNC_START for promise actions', () => {
    const promise = Promise.resolve({ data: 'test' });
    const action = { type: 'TEST_ACTION', payload: promise };
    
    invoke(action);
    
    expect(store.dispatch).toHaveBeenCalledWith({
      type: ASYNC_START,
      subtype: 'TEST_ACTION'
    });
  });

  test('dispatches ASYNC_END and action on promise resolution', async () => {
    const resolvedData = { data: 'test' };
    const promise = Promise.resolve(resolvedData);
    const action = { type: 'TEST_ACTION', payload: promise };
    
    invoke(action);
    
    await promise;
    await new Promise(resolve => setTimeout(resolve, 0));
    
    expect(store.dispatch).toHaveBeenCalledWith(
      expect.objectContaining({ type: ASYNC_END })
    );
    expect(store.dispatch).toHaveBeenCalledWith(
      expect.objectContaining({ type: 'TEST_ACTION', payload: resolvedData })
    );
  });

  test('dispatches error action on promise rejection', async () => {
    const errorResponse = { response: { body: { errors: { test: ['error'] } } } };
    const promise = Promise.reject(errorResponse);
    const action = { type: 'TEST_ACTION', payload: promise };
    
    invoke(action);
    
    await promise.catch(() => {});
    await new Promise(resolve => setTimeout(resolve, 0));
    
    expect(store.dispatch).toHaveBeenCalledWith(
      expect.objectContaining({ 
        type: 'TEST_ACTION', 
        error: true,
        payload: { errors: { test: ['error'] } }
      })
    );
  });

  test('handles promise rejection without response body', async () => {
    const promise = Promise.reject(new Error('Network error'));
    const action = { type: 'TEST_ACTION', payload: promise };
    
    invoke(action);
    
    await promise.catch(() => {});
    await new Promise(resolve => setTimeout(resolve, 0));
    
    expect(store.dispatch).toHaveBeenCalledWith(
      expect.objectContaining({ 
        error: true,
        payload: { errors: { body: ['Server error occurred'] } }
      })
    );
  });

  test('skips dispatch if view changed', async () => {
    const promise = Promise.resolve({ data: 'test' });
    const action = { type: 'TEST_ACTION', payload: promise };
    
    // Change view counter after action is invoked
    store.getState = jest.fn()
      .mockReturnValueOnce({ viewChangeCounter: 0 })
      .mockReturnValueOnce({ viewChangeCounter: 1 });
    
    invoke(action);
    
    await promise;
    await new Promise(resolve => setTimeout(resolve, 0));
    
    // Should dispatch ASYNC_START but not the resolved action
    const asyncStartCalls = store.dispatch.mock.calls.filter(
      call => call[0].type === ASYNC_START
    );
    expect(asyncStartCalls.length).toBe(1);
  });

  test('continues dispatch when skipTracking is true', async () => {
    const promise = Promise.resolve({ data: 'test' });
    const action = { type: 'TEST_ACTION', payload: promise, skipTracking: true };
    
    store.getState = jest.fn()
      .mockReturnValueOnce({ viewChangeCounter: 0 })
      .mockReturnValueOnce({ viewChangeCounter: 999 });
    
    invoke(action);
    
    await promise;
    await new Promise(resolve => setTimeout(resolve, 0));
    
    expect(store.dispatch).toHaveBeenCalledWith(
      expect.objectContaining({ type: 'TEST_ACTION' })
    );
  });
});

describe('localStorageMiddleware', () => {
  let store, next, invoke;

  beforeEach(() => {
    store = {
      getState: jest.fn(),
      dispatch: jest.fn()
    };
    next = jest.fn();
    invoke = (action) => localStorageMiddleware(store)(next)(action);
    
    // Mock localStorage
    Object.defineProperty(window, 'localStorage', {
      value: {
        setItem: jest.fn(),
        getItem: jest.fn(),
        removeItem: jest.fn()
      },
      writable: true
    });
  });

  test('saves JWT token to localStorage on successful LOGIN', () => {
    const action = {
      type: LOGIN,
      payload: {
        user: {
          email: 'test@example.com',
          token: 'jwt-token-123',
          username: 'testuser'
        }
      }
    };
    
    invoke(action);
    
    expect(window.localStorage.setItem).toHaveBeenCalledWith('jwt', 'jwt-token-123');
    expect(next).toHaveBeenCalledWith(action);
  });

  test('saves JWT token to localStorage on successful REGISTER', () => {
    const action = {
      type: REGISTER,
      payload: {
        user: {
          email: 'newuser@example.com',
          token: 'new-jwt-token',
          username: 'newuser'
        }
      }
    };
    
    invoke(action);
    
    expect(window.localStorage.setItem).toHaveBeenCalledWith('jwt', 'new-jwt-token');
    expect(next).toHaveBeenCalledWith(action);
  });

  test('does not save token on failed LOGIN', () => {
    const action = {
      type: LOGIN,
      error: true,
      payload: {
        errors: { 'email or password': ['is invalid'] }
      }
    };
    
    invoke(action);
    
    expect(window.localStorage.setItem).not.toHaveBeenCalled();
    expect(next).toHaveBeenCalledWith(action);
  });

  test('clears JWT token on LOGOUT', () => {
    const action = { type: LOGOUT };
    
    invoke(action);
    
    expect(window.localStorage.setItem).toHaveBeenCalledWith('jwt', '');
    expect(next).toHaveBeenCalledWith(action);
  });

  test('passes through non-auth actions unchanged', () => {
    const action = { type: 'OTHER_ACTION', payload: 'test' };
    
    invoke(action);
    
    expect(window.localStorage.setItem).not.toHaveBeenCalled();
    expect(next).toHaveBeenCalledWith(action);
  });

  test('always calls next middleware', () => {
    const loginAction = {
      type: LOGIN,
      payload: { user: { token: 'token' } }
    };
    const logoutAction = { type: LOGOUT };
    const otherAction = { type: 'OTHER' };
    
    invoke(loginAction);
    invoke(logoutAction);
    invoke(otherAction);
    
    expect(next).toHaveBeenCalledTimes(3);
  });
});
