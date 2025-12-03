import React from 'react';
import { render, screen } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import Header from './Header';
import { mockUser } from '../test-utils';

describe('Header Component', () => {
  test('renders app name in navbar brand', () => {
    render(
      <BrowserRouter>
        <Header appName="conduit" currentUser={null} />
      </BrowserRouter>
    );
    
    expect(screen.getByText('conduit')).toBeInTheDocument();
  });

  test('renders logged out view for guest users', () => {
    render(
      <BrowserRouter>
        <Header appName="conduit" currentUser={null} />
      </BrowserRouter>
    );
    
    expect(screen.getByText('Sign in')).toBeInTheDocument();
    expect(screen.getByText('Sign up')).toBeInTheDocument();
  });

  test('renders home link for guest users', () => {
    render(
      <BrowserRouter>
        <Header appName="conduit" currentUser={null} />
      </BrowserRouter>
    );
    
    const homeLinks = screen.getAllByText('Home');
    expect(homeLinks.length).toBeGreaterThan(0);
  });

  test('renders sign in link with correct href', () => {
    render(
      <BrowserRouter>
        <Header appName="conduit" currentUser={null} />
      </BrowserRouter>
    );
    
    const signInLink = screen.getByText('Sign in').closest('a');
    expect(signInLink).toHaveAttribute('href', '/login');
  });

  test('renders sign up link with correct href', () => {
    render(
      <BrowserRouter>
        <Header appName="conduit" currentUser={null} />
      </BrowserRouter>
    );
    
    const signUpLink = screen.getByText('Sign up').closest('a');
    expect(signUpLink).toHaveAttribute('href', '/register');
  });

  test('renders logged in view for authenticated users', () => {
    render(
      <BrowserRouter>
        <Header appName="conduit" currentUser={mockUser} />
      </BrowserRouter>
    );
    
    expect(screen.getByText('New Post')).toBeInTheDocument();
    expect(screen.getByText('Settings')).toBeInTheDocument();
    expect(screen.getByText(mockUser.username)).toBeInTheDocument();
  });

  test('renders new post link for logged in users', () => {
    render(
      <BrowserRouter>
        <Header appName="conduit" currentUser={mockUser} />
      </BrowserRouter>
    );
    
    const newPostLink = screen.getByText('New Post').closest('a');
    expect(newPostLink).toHaveAttribute('href', '/editor');
  });

  test('renders settings link for logged in users', () => {
    render(
      <BrowserRouter>
        <Header appName="conduit" currentUser={mockUser} />
      </BrowserRouter>
    );
    
    const settingsLink = screen.getByText('Settings').closest('a');
    expect(settingsLink).toHaveAttribute('href', '/settings');
  });

  test('renders user profile link with username', () => {
    render(
      <BrowserRouter>
        <Header appName="conduit" currentUser={mockUser} />
      </BrowserRouter>
    );
    
    const profileLink = screen.getByText(mockUser.username).closest('a');
    expect(profileLink).toHaveAttribute('href', `/@${mockUser.username}`);
  });

  test('renders user profile image', () => {
    render(
      <BrowserRouter>
        <Header appName="conduit" currentUser={mockUser} />
      </BrowserRouter>
    );
    
    const userImage = screen.getByAltText(mockUser.username);
    expect(userImage).toHaveAttribute('src', mockUser.image);
    expect(userImage).toHaveClass('user-pic');
  });

  test('renders fallback image when user image is missing', () => {
    const userWithoutImage = { ...mockUser, image: null };
    render(
      <BrowserRouter>
        <Header appName="conduit" currentUser={userWithoutImage} />
      </BrowserRouter>
    );
    
    const userImage = screen.getByAltText(mockUser.username);
    expect(userImage).toHaveAttribute('src', 'https://static.productionready.io/images/smiley-cyrus.jpg');
  });

  test('does not render logged in view for guest users', () => {
    render(
      <BrowserRouter>
        <Header appName="conduit" currentUser={null} />
      </BrowserRouter>
    );
    
    expect(screen.queryByText('New Post')).not.toBeInTheDocument();
    expect(screen.queryByText('Settings')).not.toBeInTheDocument();
  });

  test('does not render logged out view for authenticated users', () => {
    render(
      <BrowserRouter>
        <Header appName="conduit" currentUser={mockUser} />
      </BrowserRouter>
    );
    
    expect(screen.queryByText('Sign in')).not.toBeInTheDocument();
    expect(screen.queryByText('Sign up')).not.toBeInTheDocument();
  });

  test('navbar brand links to home page', () => {
    render(
      <BrowserRouter>
        <Header appName="conduit" currentUser={null} />
      </BrowserRouter>
    );
    
    const brandLink = screen.getByText('conduit').closest('a');
    expect(brandLink).toHaveAttribute('href', '/');
  });
});
