import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { renderWithProviders, createMockStore } from '../test-utils';
import Editor from './Editor';
import { 
  UPDATE_FIELD_EDITOR, 
  ADD_TAG, 
  REMOVE_TAG, 
  ARTICLE_SUBMITTED 
} from '../constants/actionTypes';

// Mock the match prop
const mockMatch = {
  params: {}
};

describe('Editor Component', () => {
  test('renders article title input field', () => {
    const store = createMockStore({ 
      editor: { title: '', description: '', body: '', tagList: [] } 
    });
    renderWithProviders(<Editor match={mockMatch} />, { store });
    
    const titleInput = screen.getByPlaceholderText('Article Title');
    expect(titleInput).toBeInTheDocument();
  });

  test('renders article description input field', () => {
    const store = createMockStore({ 
      editor: { title: '', description: '', body: '', tagList: [] } 
    });
    renderWithProviders(<Editor match={mockMatch} />, { store });
    
    const descInput = screen.getByPlaceholderText("What's this article about?");
    expect(descInput).toBeInTheDocument();
  });

  test('renders article body textarea', () => {
    const store = createMockStore({ 
      editor: { title: '', description: '', body: '', tagList: [] } 
    });
    renderWithProviders(<Editor match={mockMatch} />, { store });
    
    const bodyTextarea = screen.getByPlaceholderText('Write your article (in markdown)');
    expect(bodyTextarea).toBeInTheDocument();
    expect(bodyTextarea.tagName).toBe('TEXTAREA');
  });

  test('renders tag input field', () => {
    const store = createMockStore({ 
      editor: { title: '', description: '', body: '', tagList: [], tagInput: '' } 
    });
    renderWithProviders(<Editor match={mockMatch} />, { store });
    
    const tagInput = screen.getByPlaceholderText('Enter tags');
    expect(tagInput).toBeInTheDocument();
  });

  test('renders publish article button', () => {
    const store = createMockStore({ 
      editor: { title: '', description: '', body: '', tagList: [] } 
    });
    renderWithProviders(<Editor match={mockMatch} />, { store });
    
    const publishButton = screen.getByRole('button', { name: 'Publish Article' });
    expect(publishButton).toBeInTheDocument();
  });

  test('title input updates on change', () => {
    const store = createMockStore({ 
      editor: { title: '', description: '', body: '', tagList: [] } 
    });
    renderWithProviders(<Editor match={mockMatch} />, { store });
    
    const titleInput = screen.getByPlaceholderText('Article Title');
    fireEvent.change(titleInput, { target: { value: 'My New Article' } });
    
    const actions = store.getActions();
    const updateAction = actions.find(a => a.type === UPDATE_FIELD_EDITOR && a.key === 'title');
    expect(updateAction).toBeDefined();
    expect(updateAction.value).toBe('My New Article');
  });

  test('description input updates on change', () => {
    const store = createMockStore({ 
      editor: { title: '', description: '', body: '', tagList: [] } 
    });
    renderWithProviders(<Editor match={mockMatch} />, { store });
    
    const descInput = screen.getByPlaceholderText("What's this article about?");
    fireEvent.change(descInput, { target: { value: 'Article description' } });
    
    const actions = store.getActions();
    const updateAction = actions.find(a => a.type === UPDATE_FIELD_EDITOR && a.key === 'description');
    expect(updateAction).toBeDefined();
    expect(updateAction.value).toBe('Article description');
  });

  test('body textarea updates on change', () => {
    const store = createMockStore({ 
      editor: { title: '', description: '', body: '', tagList: [] } 
    });
    renderWithProviders(<Editor match={mockMatch} />, { store });
    
    const bodyTextarea = screen.getByPlaceholderText('Write your article (in markdown)');
    fireEvent.change(bodyTextarea, { target: { value: 'Article body content' } });
    
    const actions = store.getActions();
    const updateAction = actions.find(a => a.type === UPDATE_FIELD_EDITOR && a.key === 'body');
    expect(updateAction).toBeDefined();
    expect(updateAction.value).toBe('Article body content');
  });

  test('tag input updates on change', () => {
    const store = createMockStore({ 
      editor: { title: '', description: '', body: '', tagList: [], tagInput: '' } 
    });
    renderWithProviders(<Editor match={mockMatch} />, { store });
    
    const tagInput = screen.getByPlaceholderText('Enter tags');
    fireEvent.change(tagInput, { target: { value: 'react' } });
    
    const actions = store.getActions();
    const updateAction = actions.find(a => a.type === UPDATE_FIELD_EDITOR && a.key === 'tagInput');
    expect(updateAction).toBeDefined();
    expect(updateAction.value).toBe('react');
  });

  test('pressing Enter in tag input adds tag', () => {
    const store = createMockStore({ 
      editor: { title: '', description: '', body: '', tagList: [], tagInput: 'react' } 
    });
    renderWithProviders(<Editor match={mockMatch} />, { store });
    
    const tagInput = screen.getByPlaceholderText('Enter tags');
    fireEvent.keyUp(tagInput, { keyCode: 13 });
    
    const actions = store.getActions();
    const addTagAction = actions.find(a => a.type === ADD_TAG);
    expect(addTagAction).toBeDefined();
  });

  test('renders tags in tag list', () => {
    const store = createMockStore({ 
      editor: { 
        title: '', 
        description: '', 
        body: '', 
        tagList: ['react', 'javascript', 'testing'] 
      } 
    });
    renderWithProviders(<Editor match={mockMatch} />, { store });
    
    expect(screen.getByText('react')).toBeInTheDocument();
    expect(screen.getByText('javascript')).toBeInTheDocument();
    expect(screen.getByText('testing')).toBeInTheDocument();
  });

  test('clicking remove icon on tag dispatches remove action', () => {
    const store = createMockStore({ 
      editor: { 
        title: '', 
        description: '', 
        body: '', 
        tagList: ['react', 'javascript'] 
      } 
    });
    const { container } = renderWithProviders(<Editor match={mockMatch} />, { store });
    
    const removeIcons = container.querySelectorAll('.ion-close-round');
    fireEvent.click(removeIcons[0]);
    
    const actions = store.getActions();
    const removeAction = actions.find(a => a.type === REMOVE_TAG);
    expect(removeAction).toBeDefined();
    expect(removeAction.tag).toBe('react');
  });

  test('clicking publish button dispatches article submitted action', () => {
    const store = createMockStore({ 
      editor: { 
        title: 'Test Article', 
        description: 'Test desc', 
        body: 'Test body', 
        tagList: ['test'] 
      } 
    });
    renderWithProviders(<Editor match={mockMatch} />, { store });
    
    const publishButton = screen.getByRole('button', { name: 'Publish Article' });
    fireEvent.click(publishButton);
    
    const actions = store.getActions();
    const hasSubmitAction = actions.some(a => a.type === ARTICLE_SUBMITTED);
    expect(hasSubmitAction).toBe(true);
  });

  test('publish button is disabled when inProgress is true', () => {
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

  test('displays pre-filled form values', () => {
    const store = createMockStore({ 
      editor: { 
        title: 'Existing Article', 
        description: 'Existing description', 
        body: 'Existing body content', 
        tagList: ['existing-tag'] 
      } 
    });
    renderWithProviders(<Editor match={mockMatch} />, { store });
    
    expect(screen.getByDisplayValue('Existing Article')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Existing description')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Existing body content')).toBeInTheDocument();
    expect(screen.getByText('existing-tag')).toBeInTheDocument();
  });

  test('displays error messages when errors exist', () => {
    const store = createMockStore({ 
      editor: { 
        title: '', 
        description: '', 
        body: '', 
        tagList: [],
        errors: { 
          title: ["can't be blank"] 
        } 
      } 
    });
    renderWithProviders(<Editor match={mockMatch} />, { store });
    
    // ListErrors component should render errors
    expect(screen.getByText(/title/)).toBeInTheDocument();
  });
});
