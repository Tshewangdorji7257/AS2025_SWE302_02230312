import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { renderWithProviders, createMockStore, mockUser } from '../test-utils';
import Login from './Login';
import { LOGIN, UPDATE_FIELD_AUTH } from '../constants/actionTypes';

describe('Login Component', () => {
  test('renders sign in heading', () => {
    const store = createMockStore({ auth: {} });
    renderWithProviders(<Login />, { store });
    
    expect(screen.getByText('Sign In')).toBeInTheDocument();
  });

  test('renders link to register page', () => {
    const store = createMockStore({ auth: {} });
    renderWithProviders(<Login />, { store });
    
    const registerLink = screen.getByText('Need an account?');
    expect(registerLink).toBeInTheDocument();
    expect(registerLink.closest('a')).toHaveAttribute('href', '/register');
  });

  test('renders email input field', () => {
    const store = createMockStore({ auth: {} });
    renderWithProviders(<Login />, { store });
    
    const emailInput = screen.getByPlaceholderText('Email');
    expect(emailInput).toBeInTheDocument();
    expect(emailInput).toHaveAttribute('type', 'email');
  });

  test('renders password input field', () => {
    const store = createMockStore({ auth: {} });
    renderWithProviders(<Login />, { store });
    
    const passwordInput = screen.getByPlaceholderText('Password');
    expect(passwordInput).toBeInTheDocument();
    expect(passwordInput).toHaveAttribute('type', 'password');
  });

  test('renders submit button', () => {
    const store = createMockStore({ auth: {} });
    renderWithProviders(<Login />, { store });
    
    const submitButton = screen.getByRole('button', { name: 'Sign in' });
    expect(submitButton).toBeInTheDocument();
  });

  test('email input updates on change', () => {
    const store = createMockStore({ auth: { email: '' } });
    renderWithProviders(<Login />, { store });
    
    const emailInput = screen.getByPlaceholderText('Email');
    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    
    const actions = store.getActions();
    expect(actions[0].type).toBe(UPDATE_FIELD_AUTH);
    expect(actions[0].key).toBe('email');
    expect(actions[0].value).toBe('test@example.com');
  });

  test('password input updates on change', () => {
    const store = createMockStore({ auth: { password: '' } });
    renderWithProviders(<Login />, { store });
    
    const passwordInput = screen.getByPlaceholderText('Password');
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    
    const actions = store.getActions();
    expect(actions[0].type).toBe(UPDATE_FIELD_AUTH);
    expect(actions[0].key).toBe('password');
    expect(actions[0].value).toBe('password123');
  });

  test('form submission dispatches login action', () => {
    const store = createMockStore({ 
      auth: { 
        email: 'test@example.com', 
        password: 'password123' 
      } 
    });
    const { container } = renderWithProviders(<Login />, { store });
    
    const form = container.querySelector('form');
    fireEvent.submit(form);
    
    const actions = store.getActions();
    // Check for LOGIN action (may be preceded by ASYNC_START)
    const hasLoginAction = actions.some(action => action.type === LOGIN);
    expect(hasLoginAction).toBe(true);
  });

  test('submit button is disabled when inProgress is true', () => {
    const store = createMockStore({ 
      auth: { 
        email: 'test@example.com', 
        password: 'password123',
        inProgress: true
      } 
    });
    renderWithProviders(<Login />, { store });
    
    const submitButton = screen.getByRole('button', { name: 'Sign in' });
    expect(submitButton).toBeDisabled();
  });

  test('submit button is enabled when inProgress is false', () => {
    const store = createMockStore({ 
      auth: { 
        email: 'test@example.com', 
        password: 'password123',
        inProgress: false
      } 
    });
    renderWithProviders(<Login />, { store });
    
    const submitButton = screen.getByRole('button', { name: 'Sign in' });
    expect(submitButton).not.toBeDisabled();
  });

  test('displays error messages when errors exist', () => {
    const store = createMockStore({ 
      auth: { 
        errors: { 
          'email or password': ['is invalid'] 
        } 
      } 
    });
    renderWithProviders(<Login />, { store });
    
    // ListErrors component should render errors
    expect(screen.getByText(/email or password/)).toBeInTheDocument();
  });

  test('form submission triggers login', () => {
    const store = createMockStore({ 
      auth: { 
        email: 'test@example.com', 
        password: 'password123' 
      } 
    });
    renderWithProviders(<Login />, { store });
    
    const submitButton = screen.getByRole('button', { name: 'Sign in' });
    fireEvent.click(submitButton);
    
    // Should dispatch action
    const actions = store.getActions();
    expect(actions.length).toBeGreaterThan(0);
  });

  test('displays pre-filled email and password values', () => {
    const store = createMockStore({ 
      auth: { 
        email: 'prefilled@example.com', 
        password: 'prefilledpassword' 
      } 
    });
    renderWithProviders(<Login />, { store });
    
    const emailInput = screen.getByPlaceholderText('Email');
    const passwordInput = screen.getByPlaceholderText('Password');
    
    expect(emailInput).toHaveValue('prefilled@example.com');
    expect(passwordInput).toHaveValue('prefilledpassword');
  });
});
