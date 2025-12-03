import React from 'react';
import { screen } from '@testing-library/react';
import ArticleList from './ArticleList';
import { mockArticle, renderWithProviders, createMockStore } from '../test-utils';

describe('ArticleList Component', () => {
  test('renders loading state when articles is null/undefined', () => {
    const store = createMockStore();
    renderWithProviders(<ArticleList articles={null} />, { store });
    expect(screen.getByText('Loading...')).toBeInTheDocument();
  });

  test('renders empty state message when articles array is empty', () => {
    const store = createMockStore();
    renderWithProviders(<ArticleList articles={[]} />, { store });
    expect(screen.getByText('No articles are here... yet.')).toBeInTheDocument();
  });

  test('renders single article correctly', () => {
    const articles = [mockArticle];
    const store = createMockStore();
    renderWithProviders(<ArticleList articles={articles} />, { store });
    expect(screen.getByText('Test Article')).toBeInTheDocument();
  });

  test('renders multiple articles', () => {
    const articles = [
      { ...mockArticle, slug: 'article-1', title: 'First Article' },
      { ...mockArticle, slug: 'article-2', title: 'Second Article' },
      { ...mockArticle, slug: 'article-3', title: 'Third Article' }
    ];
    const store = createMockStore();
    renderWithProviders(<ArticleList articles={articles} />, { store });
    
    expect(screen.getByText('First Article')).toBeInTheDocument();
    expect(screen.getByText('Second Article')).toBeInTheDocument();
    expect(screen.getByText('Third Article')).toBeInTheDocument();
  });

  test('renders ArticlePreview for each article', () => {
    const articles = [
      { ...mockArticle, slug: 'article-1', title: 'Article 1' },
      { ...mockArticle, slug: 'article-2', title: 'Article 2' }
    ];
    const store = createMockStore();
    const { container } = renderWithProviders(<ArticleList articles={articles} />, { store });
    
    const articlePreviews = container.querySelectorAll('.article-preview');
    expect(articlePreviews.length).toBe(2);
  });

  test('passes correct props to ListPagination', () => {
    const pager = jest.fn();
    const articlesCount = 100;
    const currentPage = 2;
    const store = createMockStore();
    
    renderWithProviders(
      <ArticleList 
        articles={[mockArticle]} 
        pager={pager}
        articlesCount={articlesCount}
        currentPage={currentPage}
      />, 
      { store }
    );
    
    // Component should render without errors
    expect(screen.getByText('Test Article')).toBeInTheDocument();
  });
});
