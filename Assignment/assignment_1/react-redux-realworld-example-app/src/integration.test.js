import React from 'react';
import { screen, fireEvent, waitFor } from '@testing-library/react';
import { renderWithProviders, createMockStore, mockUser, mockArticle } from './test-utils';
import Login from './components/Login';
import ArticlePreview from './components/ArticlePreview';
import Editor from './components/Editor';
import { LOGIN, ARTICLE_FAVORITED, ARTICLE_SUBMITTED, UPDATE_FIELD_AUTH, UPDATE_FIELD_EDITOR } from './constants/actionTypes';

/**
 * Integration tests verify that components work correctly with Redux state management
 */

describe('Integration Tests', () => {
  
  describe('Login Flow Integration', () => {
    test('complete login flow updates Redux state', async () => {
      const store = createMockStore({ 
        auth: { email: '', password: '' } 
      });
      
      const { container } = renderWithProviders(<Login />, { store });
      
      // User enters email
      const emailInput = screen.getByPlaceholderText('Email');
      fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
      
      // User enters password
      const passwordInput = screen.getByPlaceholderText('Password');
      fireEvent.change(passwordInput, { target: { value: 'password123' } });
      
      // User submits form
      const form = container.querySelector('form');
      fireEvent.submit(form);
      
      // Verify actions were dispatched in correct order
      const actions = store.getActions();
      
      // Should have UPDATE_FIELD_AUTH actions for email and password
      const emailUpdate = actions.find(a => a.type === UPDATE_FIELD_AUTH && a.key === 'email');
      expect(emailUpdate).toBeDefined();
      expect(emailUpdate.value).toBe('test@example.com');
      
      const passwordUpdate = actions.find(a => a.type === UPDATE_FIELD_AUTH && a.key === 'password');
      expect(passwordUpdate).toBeDefined();
      expect(passwordUpdate.value).toBe('password123');
      
      // Should dispatch LOGIN action
      const hasLoginAction = actions.some(a => a.type === LOGIN);
      expect(hasLoginAction).toBe(true);
    });

    test('login with errors displays error messages', () => {
      const store = createMockStore({ 
        auth: { 
          email: 'invalid@example.com',
          password: 'wrongpassword',
          errors: { 
            'email or password': ['is invalid'] 
          }
        } 
      });
      
      renderWithProviders(<Login />, { store });
      
      // Error message should be displayed
      expect(screen.getByText(/email or password/)).toBeInTheDocument();
    });

    test('login form disables submit button during submission', () => {
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
  });

  describe('Article Favorite Flow Integration', () => {
    test('favoriting an article updates Redux state', () => {
      const store = createMockStore({ 
        common: { currentUser: mockUser }
      });
      
      const { container } = renderWithProviders(
        <ArticlePreview article={mockArticle} />, 
        { store }
      );
      
      // Click favorite button
      const favoriteButton = container.querySelector('.btn-outline-primary');
      fireEvent.click(favoriteButton);
      
      // Verify ARTICLE_FAVORITED action was dispatched
      const actions = store.getActions();
      const hasFavoriteAction = actions.some(a => a.type === ARTICLE_FAVORITED);
      expect(hasFavoriteAction).toBe(true);
    });

    test('unfavoriting an article dispatches correct action', () => {
      const favoritedArticle = { ...mockArticle, favorited: true };
      const store = createMockStore({ 
        common: { currentUser: mockUser }
      });
      
      const { container } = renderWithProviders(
        <ArticlePreview article={favoritedArticle} />, 
        { store }
      );
      
      // Click unfavorite button
      const unfavoriteButton = container.querySelector('.btn-primary');
      fireEvent.click(unfavoriteButton);
      
      // Verify action was dispatched
      const actions = store.getActions();
      const hasUnfavoriteAction = actions.some(a => a.type === 'ARTICLE_UNFAVORITED');
      expect(hasUnfavoriteAction).toBe(true);
    });

    test('favorite count is displayed correctly', () => {
      const articleWith10Favorites = { ...mockArticle, favoritesCount: 10 };
      const store = createMockStore({});
      
      renderWithProviders(
        <ArticlePreview article={articleWith10Favorites} />, 
        { store }
      );
      
      expect(screen.getByText('10')).toBeInTheDocument();
    });
  });

  describe('Article Creation Flow Integration', () => {
    const mockMatch = { params: {} };

    test('complete article creation flow', () => {
      const store = createMockStore({ 
        editor: { 
          title: '', 
          description: '', 
          body: '', 
          tagList: [],
          tagInput: ''
        },
        common: { currentUser: mockUser }
      });
      
      renderWithProviders(<Editor match={mockMatch} />, { store });
      
      // Fill in title
      const titleInput = screen.getByPlaceholderText('Article Title');
      fireEvent.change(titleInput, { target: { value: 'My New Article' } });
      
      // Fill in description
      const descInput = screen.getByPlaceholderText("What's this article about?");
      fireEvent.change(descInput, { target: { value: 'Article about testing' } });
      
      // Fill in body
      const bodyTextarea = screen.getByPlaceholderText('Write your article (in markdown)');
      fireEvent.change(bodyTextarea, { target: { value: '# Article Content\n\nThis is the body.' } });
      
      // Submit the form
      const publishButton = screen.getByRole('button', { name: 'Publish Article' });
      fireEvent.click(publishButton);
      
      // Verify all actions were dispatched
      const actions = store.getActions();
      
      const titleUpdate = actions.find(a => a.type === UPDATE_FIELD_EDITOR && a.key === 'title');
      expect(titleUpdate).toBeDefined();
      
      const hasSubmitAction = actions.some(a => a.type === ARTICLE_SUBMITTED);
      expect(hasSubmitAction).toBe(true);
    });

    test('article editor with validation errors', () => {
      const store = createMockStore({ 
        editor: { 
          title: '', 
          description: '', 
          body: '', 
          tagList: [],
          errors: {
            title: ["can't be blank"],
            body: ["can't be blank"]
          }
        }
      });
      
      renderWithProviders(<Editor match={mockMatch} />, { store });
      
      // Error messages should be displayed
      expect(screen.getByText(/title/)).toBeInTheDocument();
    });

    test('adding tags to article', () => {
      const store = createMockStore({ 
        editor: { 
          title: '', 
          description: '', 
          body: '', 
          tagList: [],
          tagInput: 'react'
        }
      });
      
      renderWithProviders(<Editor match={mockMatch} />, { store });
      
      const tagInput = screen.getByPlaceholderText('Enter tags');
      fireEvent.keyUp(tagInput, { keyCode: 13 }); // Enter key
      
      const actions = store.getActions();
      const addTagAction = actions.find(a => a.type === 'ADD_TAG');
      expect(addTagAction).toBeDefined();
    });

    test('removing tags from article', () => {
      const store = createMockStore({ 
        editor: { 
          title: 'Test', 
          description: 'Test', 
          body: 'Test', 
          tagList: ['react', 'javascript']
        }
      });
      
      const { container } = renderWithProviders(<Editor match={mockMatch} />, { store });
      
      // Click remove icon on first tag
      const removeIcon = container.querySelector('.ion-close-round');
      fireEvent.click(removeIcon);
      
      const actions = store.getActions();
      const removeTagAction = actions.find(a => a.type === 'REMOVE_TAG');
      expect(removeTagAction).toBeDefined();
      expect(removeTagAction.tag).toBe('react');
    });

    test('editor disables publish during submission', () => {
      const store = createMockStore({ 
        editor: { 
          title: 'Test', 
          description: 'Test', 
          body: 'Test', 
          tagList: [],
          inProgress: true
        }
      });
      
      renderWithProviders(<Editor match={mockMatch} />, { store });
      
      const publishButton = screen.getByRole('button', { name: 'Publish Article' });
      expect(publishButton).toBeDisabled();
    });
  });

  describe('Multi-Component State Synchronization', () => {
    test('article preview reflects favorited state from store', () => {
      const favoritedArticle = { ...mockArticle, favorited: true, favoritesCount: 15 };
      const store = createMockStore({});
      
      const { container } = renderWithProviders(
        <ArticlePreview article={favoritedArticle} />, 
        { store }
      );
      
      // Button should have favorited class
      const button = container.querySelector('.btn-primary');
      expect(button).toBeInTheDocument();
      expect(screen.getByText('15')).toBeInTheDocument();
    });

    test('login state changes are reflected in component', () => {
      const store = createMockStore({ 
        auth: { 
          email: 'prepopulated@example.com', 
          password: 'prepopulatedpass' 
        } 
      });
      
      renderWithProviders(<Login />, { store });
      
      const emailInput = screen.getByPlaceholderText('Email');
      const passwordInput = screen.getByPlaceholderText('Password');
      
      expect(emailInput).toHaveValue('prepopulated@example.com');
      expect(passwordInput).toHaveValue('prepopulatedpass');
    });

    test('editor reflects pre-loaded article data', () => {
      const store = createMockStore({ 
        editor: { 
          title: 'Existing Article', 
          description: 'Description here', 
          body: 'Body content here', 
          tagList: ['tag1', 'tag2']
        }
      });
      
      const mockMatch = { params: { slug: 'existing-article' } };
      renderWithProviders(<Editor match={mockMatch} />, { store });
      
      expect(screen.getByDisplayValue('Existing Article')).toBeInTheDocument();
      expect(screen.getByDisplayValue('Description here')).toBeInTheDocument();
      expect(screen.getByDisplayValue('Body content here')).toBeInTheDocument();
      expect(screen.getByText('tag1')).toBeInTheDocument();
      expect(screen.getByText('tag2')).toBeInTheDocument();
    });
  });

  describe('Error Handling Integration', () => {
    test('network errors are handled gracefully in login', () => {
      const store = createMockStore({ 
        auth: { 
          errors: { 
            'network': ['connection failed'] 
          } 
        } 
      });
      
      renderWithProviders(<Login />, { store });
      
      // Error should be displayed via ListErrors component
      expect(screen.getByText(/network/)).toBeInTheDocument();
    });

    test('validation errors are displayed in editor', () => {
      const store = createMockStore({ 
        editor: { 
          title: '', 
          description: '', 
          body: '', 
          tagList: [],
          errors: {
            title: ["can't be blank", "is too short"],
            body: ["can't be blank"]
          }
        }
      });
      
      const mockMatch = { params: {} };
      renderWithProviders(<Editor match={mockMatch} />, { store });
      
      // Both error keys should be visible
      expect(screen.getByText(/title/)).toBeInTheDocument();
    });
  });
});
