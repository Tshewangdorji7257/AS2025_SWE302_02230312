import editorReducer from './editor';
import {
  EDITOR_PAGE_LOADED,
  EDITOR_PAGE_UNLOADED,
  ARTICLE_SUBMITTED,
  ASYNC_START,
  ADD_TAG,
  REMOVE_TAG,
  UPDATE_FIELD_EDITOR
} from '../constants/actionTypes';

describe('editor reducer', () => {
  test('returns initial empty state', () => {
    const state = editorReducer(undefined, {});
    expect(state).toEqual({});
  });

  test('handles EDITOR_PAGE_LOADED with new article (null payload)', () => {
    const action = {
      type: EDITOR_PAGE_LOADED,
      payload: null
    };
    const state = editorReducer({}, action);
    
    expect(state.articleSlug).toBe('');
    expect(state.title).toBe('');
    expect(state.description).toBe('');
    expect(state.body).toBe('');
    expect(state.tagInput).toBe('');
    expect(state.tagList).toEqual([]);
  });

  test('handles EDITOR_PAGE_LOADED with existing article', () => {
    const action = {
      type: EDITOR_PAGE_LOADED,
      payload: {
        article: {
          slug: 'test-article',
          title: 'Test Article',
          description: 'Test description',
          body: 'Test body content',
          tagList: ['react', 'testing']
        }
      }
    };
    const state = editorReducer({}, action);
    
    expect(state.articleSlug).toBe('test-article');
    expect(state.title).toBe('Test Article');
    expect(state.description).toBe('Test description');
    expect(state.body).toBe('Test body content');
    expect(state.tagList).toEqual(['react', 'testing']);
    expect(state.tagInput).toBe('');
  });

  test('handles EDITOR_PAGE_UNLOADED clears state', () => {
    const initialState = {
      title: 'Test Article',
      description: 'Test',
      body: 'Content',
      tagList: ['tag1']
    };
    const action = { type: EDITOR_PAGE_UNLOADED };
    const state = editorReducer(initialState, action);
    
    expect(state).toEqual({});
  });

  test('handles UPDATE_FIELD_EDITOR for title', () => {
    const action = {
      type: UPDATE_FIELD_EDITOR,
      key: 'title',
      value: 'New Article Title'
    };
    const state = editorReducer({}, action);
    
    expect(state.title).toBe('New Article Title');
  });

  test('handles UPDATE_FIELD_EDITOR for description', () => {
    const action = {
      type: UPDATE_FIELD_EDITOR,
      key: 'description',
      value: 'Article description'
    };
    const state = editorReducer({}, action);
    
    expect(state.description).toBe('Article description');
  });

  test('handles UPDATE_FIELD_EDITOR for body', () => {
    const action = {
      type: UPDATE_FIELD_EDITOR,
      key: 'body',
      value: 'Article body content'
    };
    const state = editorReducer({}, action);
    
    expect(state.body).toBe('Article body content');
  });

  test('handles UPDATE_FIELD_EDITOR for tagInput', () => {
    const action = {
      type: UPDATE_FIELD_EDITOR,
      key: 'tagInput',
      value: 'newtag'
    };
    const state = editorReducer({}, action);
    
    expect(state.tagInput).toBe('newtag');
  });

  test('handles ADD_TAG adds tag from tagInput', () => {
    const initialState = {
      tagList: ['existing-tag'],
      tagInput: 'new-tag'
    };
    const action = { type: ADD_TAG };
    const state = editorReducer(initialState, action);
    
    expect(state.tagList).toEqual(['existing-tag', 'new-tag']);
    expect(state.tagInput).toBe('');
  });

  test('handles ADD_TAG with empty tagList', () => {
    const initialState = {
      tagList: [],
      tagInput: 'first-tag'
    };
    const action = { type: ADD_TAG };
    const state = editorReducer(initialState, action);
    
    expect(state.tagList).toEqual(['first-tag']);
    expect(state.tagInput).toBe('');
  });

  test('handles REMOVE_TAG removes specific tag', () => {
    const initialState = {
      tagList: ['react', 'javascript', 'testing']
    };
    const action = {
      type: REMOVE_TAG,
      tag: 'javascript'
    };
    const state = editorReducer(initialState, action);
    
    expect(state.tagList).toEqual(['react', 'testing']);
  });

  test('handles REMOVE_TAG when tag not in list', () => {
    const initialState = {
      tagList: ['react', 'testing']
    };
    const action = {
      type: REMOVE_TAG,
      tag: 'nonexistent'
    };
    const state = editorReducer(initialState, action);
    
    expect(state.tagList).toEqual(['react', 'testing']);
  });

  test('handles ARTICLE_SUBMITTED success', () => {
    const action = {
      type: ARTICLE_SUBMITTED,
      payload: {
        article: { slug: 'new-article' }
      }
    };
    const state = editorReducer({ inProgress: true }, action);
    
    expect(state.inProgress).toBeNull();
    expect(state.errors).toBeNull();
  });

  test('handles ARTICLE_SUBMITTED error', () => {
    const action = {
      type: ARTICLE_SUBMITTED,
      error: true,
      payload: {
        errors: {
          title: ["can't be blank"],
          body: ["can't be blank"]
        }
      }
    };
    const state = editorReducer({ inProgress: true }, action);
    
    expect(state.inProgress).toBeNull();
    expect(state.errors).toEqual({
      title: ["can't be blank"],
      body: ["can't be blank"]
    });
  });

  test('handles ASYNC_START for ARTICLE_SUBMITTED', () => {
    const action = {
      type: ASYNC_START,
      subtype: ARTICLE_SUBMITTED
    };
    const state = editorReducer({}, action);
    
    expect(state.inProgress).toBe(true);
  });

  test('preserves existing fields when updating one field', () => {
    const initialState = {
      title: 'Existing Title',
      description: 'Existing Description',
      body: 'Existing Body'
    };
    const action = {
      type: UPDATE_FIELD_EDITOR,
      key: 'body',
      value: 'Updated Body'
    };
    const state = editorReducer(initialState, action);
    
    expect(state.title).toBe('Existing Title');
    expect(state.description).toBe('Existing Description');
    expect(state.body).toBe('Updated Body');
  });

  test('tagList persists when updating other fields', () => {
    const initialState = {
      tagList: ['important-tag'],
      title: 'Title'
    };
    const action = {
      type: UPDATE_FIELD_EDITOR,
      key: 'title',
      value: 'New Title'
    };
    const state = editorReducer(initialState, action);
    
    expect(state.tagList).toEqual(['important-tag']);
    expect(state.title).toBe('New Title');
  });
});
