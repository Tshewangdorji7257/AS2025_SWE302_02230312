import React from 'react';
import { render } from '@testing-library/react';
import { Provider } from 'react-redux';
import { BrowserRouter } from 'react-router-dom';
import configureStore from 'redux-mock-store';
import { promiseMiddleware, localStorageMiddleware } from './middleware';

const middlewares = [promiseMiddleware, localStorageMiddleware];
const mockStore = configureStore(middlewares);

/**
 * Custom render function that includes Redux Provider and Router
 */
export function renderWithProviders(
  ui,
  {
    initialState = {},
    store = mockStore(initialState),
    ...renderOptions
  } = {}
) {
  function Wrapper({ children }) {
    return (
      <Provider store={store}>
        <BrowserRouter>
          {children}
        </BrowserRouter>
      </Provider>
    );
  }

  return {
    store,
    ...render(ui, { wrapper: Wrapper, ...renderOptions })
  };
}

/**
 * Create a mock store with default state
 */
export function createMockStore(state = {}) {
  const defaultState = {
    common: {
      appName: 'Conduit',
      token: null,
      currentUser: null
    },
    auth: {},
    home: {
      tags: [],
      articles: []
    },
    articleList: {
      articles: [],
      articlesCount: 0,
      currentPage: 0
    },
    article: {
      article: null,
      comments: []
    },
    editor: {},
    profile: {},
    settings: {},
    viewChangeCounter: 0,
    ...state
  };

  return mockStore(defaultState);
}

/**
 * Mock article data for testing
 */
export const mockArticle = {
  slug: 'test-article',
  title: 'Test Article',
  description: 'Test description',
  body: 'Test body content',
  tagList: ['test', 'article'],
  createdAt: '2023-01-01T00:00:00.000Z',
  updatedAt: '2023-01-01T00:00:00.000Z',
  favorited: false,
  favoritesCount: 0,
  author: {
    username: 'testuser',
    bio: 'Test bio',
    image: 'https://api.realworld.io/images/demo-avatar.png',
    following: false
  }
};

/**
 * Mock user data for testing
 */
export const mockUser = {
  email: 'test@example.com',
  token: 'test-jwt-token',
  username: 'testuser',
  bio: 'Test bio',
  image: 'https://api.realworld.io/images/demo-avatar.png'
};

/**
 * Mock comment data for testing
 */
export const mockComment = {
  id: 1,
  createdAt: '2023-01-01T00:00:00.000Z',
  updatedAt: '2023-01-01T00:00:00.000Z',
  body: 'Test comment',
  author: {
    username: 'testuser',
    bio: 'Test bio',
    image: 'https://api.realworld.io/images/demo-avatar.png',
    following: false
  }
};
