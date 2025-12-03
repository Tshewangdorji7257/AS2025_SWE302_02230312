// jest-dom adds custom jest matchers for asserting on DOM nodes.
// allows you to do things like:
// expect(element).toHaveTextContent(/react/i)
// learn more: https://github.com/testing-library/jest-dom
import '@testing-library/jest-dom';

// Suppress React lifecycle warnings in tests
const originalError = console.error;
const originalWarn = console.warn;

beforeAll(() => {
  console.error = (...args) => {
    if (
      typeof args[0] === 'string' &&
      (args[0].includes('Warning: ReactDOM.render') ||
       args[0].includes('not wrapped in act'))
    ) {
      return;
    }
    originalError.call(console, ...args);
  };

  console.warn = (...args) => {
    if (
      typeof args[0] === 'string' &&
      (args[0].includes('componentWillMount') ||
       args[0].includes('componentWillReceiveProps') ||
       args[0].includes('componentWillUpdate'))
    ) {
      return;
    }
    originalWarn.call(console, ...args);
  };
});

afterAll(() => {
  console.error = originalError;
  console.warn = originalWarn;
});

// Mock localStorage
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
};
globalThis.localStorage = localStorageMock;

// Mock agent module to prevent real API calls
jest.mock('./agent', () => ({
  Auth: {
    current: jest.fn(() => Promise.resolve({ user: { username: 'testuser', email: 'test@example.com' } })),
    login: jest.fn((email, password) => Promise.resolve({ user: { username: 'testuser', email, token: 'fake-jwt-token' } })),
    register: jest.fn((username, email, password) => Promise.resolve({ user: { username, email, token: 'fake-jwt-token' } })),
    save: jest.fn((user) => Promise.resolve({ user })),
  },
  Articles: {
    all: jest.fn(() => Promise.resolve({ articles: [], articlesCount: 0 })),
    byAuthor: jest.fn(() => Promise.resolve({ articles: [], articlesCount: 0 })),
    byTag: jest.fn(() => Promise.resolve({ articles: [], articlesCount: 0 })),
    del: jest.fn(() => Promise.resolve()),
    favorite: jest.fn((slug) => Promise.resolve({ article: { slug, favorited: true, favoritesCount: 1 } })),
    favoritedBy: jest.fn(() => Promise.resolve({ articles: [], articlesCount: 0 })),
    feed: jest.fn(() => Promise.resolve({ articles: [], articlesCount: 0 })),
    get: jest.fn((slug) => Promise.resolve({ article: { slug, title: 'Test Article' } })),
    unfavorite: jest.fn((slug) => Promise.resolve({ article: { slug, favorited: false, favoritesCount: 0 } })),
    update: jest.fn((article) => Promise.resolve({ article })),
    create: jest.fn((article) => Promise.resolve({ article: { ...article, slug: 'test-article' } })),
  },
  Comments: {
    create: jest.fn((slug, comment) => Promise.resolve({ comment: { id: 1, body: comment.body } })),
    delete: jest.fn(() => Promise.resolve()),
    forArticle: jest.fn(() => Promise.resolve({ comments: [] })),
  },
  Profile: {
    follow: jest.fn((username) => Promise.resolve({ profile: { username, following: true } })),
    get: jest.fn((username) => Promise.resolve({ profile: { username, following: false } })),
    unfollow: jest.fn((username) => Promise.resolve({ profile: { username, following: false } })),
  },
  Tags: {
    getAll: jest.fn(() => Promise.resolve({ tags: [] })),
  },
  setToken: jest.fn(),
}));
