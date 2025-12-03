import articleListReducer from './articleList';
import {
  ARTICLE_FAVORITED,
  ARTICLE_UNFAVORITED,
  SET_PAGE,
  APPLY_TAG_FILTER,
  HOME_PAGE_LOADED,
  HOME_PAGE_UNLOADED,
  CHANGE_TAB,
  PROFILE_PAGE_LOADED,
  PROFILE_FAVORITES_PAGE_LOADED
} from '../constants/actionTypes';

describe('articleList reducer', () => {
  test('returns initial empty state', () => {
    const state = articleListReducer(undefined, {});
    expect(state).toEqual({});
  });

  test('handles ARTICLE_FAVORITED updates correct article', () => {
    const initialState = {
      articles: [
        { slug: 'article-1', favorited: false, favoritesCount: 5 },
        { slug: 'article-2', favorited: false, favoritesCount: 3 }
      ]
    };
    const action = {
      type: ARTICLE_FAVORITED,
      payload: {
        article: { slug: 'article-1', favorited: true, favoritesCount: 6 }
      }
    };
    const state = articleListReducer(initialState, action);
    
    expect(state.articles[0].favorited).toBe(true);
    expect(state.articles[0].favoritesCount).toBe(6);
    expect(state.articles[1].favorited).toBe(false);
    expect(state.articles[1].favoritesCount).toBe(3);
  });

  test('handles ARTICLE_UNFAVORITED updates correct article', () => {
    const initialState = {
      articles: [
        { slug: 'article-1', favorited: true, favoritesCount: 6 },
        { slug: 'article-2', favorited: false, favoritesCount: 3 }
      ]
    };
    const action = {
      type: ARTICLE_UNFAVORITED,
      payload: {
        article: { slug: 'article-1', favorited: false, favoritesCount: 5 }
      }
    };
    const state = articleListReducer(initialState, action);
    
    expect(state.articles[0].favorited).toBe(false);
    expect(state.articles[0].favoritesCount).toBe(5);
  });

  test('handles SET_PAGE updates pagination', () => {
    const action = {
      type: SET_PAGE,
      page: 2,
      payload: {
        articles: [{ slug: 'article-3' }, { slug: 'article-4' }],
        articlesCount: 100
      }
    };
    const state = articleListReducer({}, action);
    
    expect(state.articles).toHaveLength(2);
    expect(state.articlesCount).toBe(100);
    expect(state.currentPage).toBe(2);
  });

  test('handles APPLY_TAG_FILTER', () => {
    const mockPager = jest.fn();
    const action = {
      type: APPLY_TAG_FILTER,
      tag: 'react',
      pager: mockPager,
      payload: {
        articles: [{ slug: 'react-article' }],
        articlesCount: 15
      }
    };
    const state = articleListReducer({}, action);
    
    expect(state.tag).toBe('react');
    expect(state.tab).toBeNull();
    expect(state.currentPage).toBe(0);
    expect(state.articles).toHaveLength(1);
    expect(state.articlesCount).toBe(15);
  });

  test('handles HOME_PAGE_LOADED with valid payload', () => {
    const mockPager = jest.fn();
    const action = {
      type: HOME_PAGE_LOADED,
      tab: 'all',
      pager: mockPager,
      payload: [
        { tags: ['react', 'javascript', 'testing'] },
        { articles: [{ slug: 'article-1' }], articlesCount: 50 }
      ]
    };
    const state = articleListReducer({}, action);
    
    expect(state.tags).toEqual(['react', 'javascript', 'testing']);
    expect(state.articles).toHaveLength(1);
    expect(state.articlesCount).toBe(50);
    expect(state.currentPage).toBe(0);
    expect(state.tab).toBe('all');
  });

  test('handles HOME_PAGE_LOADED with null payload', () => {
    const action = {
      type: HOME_PAGE_LOADED,
      payload: null
    };
    const state = articleListReducer({}, action);
    
    expect(state.tags).toEqual([]);
    expect(state.articles).toEqual([]);
    expect(state.articlesCount).toBe(0);
  });

  test('handles HOME_PAGE_UNLOADED clears state', () => {
    const initialState = {
      articles: [{ slug: 'article-1' }],
      tags: ['react'],
      currentPage: 2
    };
    const action = { type: HOME_PAGE_UNLOADED };
    const state = articleListReducer(initialState, action);
    
    expect(state).toEqual({});
  });

  test('handles CHANGE_TAB', () => {
    const mockPager = jest.fn();
    const action = {
      type: CHANGE_TAB,
      tab: 'feed',
      pager: mockPager,
      payload: {
        articles: [{ slug: 'feed-article' }],
        articlesCount: 10
      }
    };
    const state = articleListReducer({ tag: 'react' }, action);
    
    expect(state.tab).toBe('feed');
    expect(state.tag).toBeNull();
    expect(state.currentPage).toBe(0);
    expect(state.articles).toHaveLength(1);
  });

  test('handles PROFILE_PAGE_LOADED', () => {
    const mockPager = jest.fn();
    const action = {
      type: PROFILE_PAGE_LOADED,
      pager: mockPager,
      payload: [
        { profile: { username: 'testuser' } },
        { articles: [{ slug: 'user-article' }], articlesCount: 25 }
      ]
    };
    const state = articleListReducer({}, action);
    
    expect(state.articles).toHaveLength(1);
    expect(state.articlesCount).toBe(25);
    expect(state.currentPage).toBe(0);
  });

  test('handles PROFILE_FAVORITES_PAGE_LOADED', () => {
    const mockPager = jest.fn();
    const action = {
      type: PROFILE_FAVORITES_PAGE_LOADED,
      pager: mockPager,
      payload: [
        { profile: { username: 'testuser' } },
        { articles: [{ slug: 'favorited-article' }], articlesCount: 12 }
      ]
    };
    const state = articleListReducer({}, action);
    
    expect(state.articles).toHaveLength(1);
    expect(state.articlesCount).toBe(12);
  });

  test('preserves other articles when favoriting one', () => {
    const initialState = {
      articles: [
        { slug: 'article-1', title: 'First', favorited: false, favoritesCount: 5 },
        { slug: 'article-2', title: 'Second', favorited: false, favoritesCount: 3 },
        { slug: 'article-3', title: 'Third', favorited: false, favoritesCount: 8 }
      ]
    };
    const action = {
      type: ARTICLE_FAVORITED,
      payload: {
        article: { slug: 'article-2', favorited: true, favoritesCount: 4 }
      }
    };
    const state = articleListReducer(initialState, action);
    
    expect(state.articles[0].title).toBe('First');
    expect(state.articles[1].favorited).toBe(true);
    expect(state.articles[2].title).toBe('Third');
  });
});
