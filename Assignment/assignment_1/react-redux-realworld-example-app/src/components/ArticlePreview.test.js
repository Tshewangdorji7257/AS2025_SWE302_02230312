import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { renderWithProviders, createMockStore, mockArticle } from '../test-utils';
import ArticlePreview from './ArticlePreview';
import { ARTICLE_FAVORITED, ARTICLE_UNFAVORITED } from '../constants/actionTypes';

describe('ArticlePreview Component', () => {
  test('renders article title correctly', () => {
    const store = createMockStore();
    renderWithProviders(<ArticlePreview article={mockArticle} />, { store });
    
    expect(screen.getByText('Test Article')).toBeInTheDocument();
  });

  test('renders article description', () => {
    const store = createMockStore();
    renderWithProviders(<ArticlePreview article={mockArticle} />, { store });
    
    expect(screen.getByText('Test description')).toBeInTheDocument();
  });

  test('renders author username correctly', () => {
    const store = createMockStore();
    renderWithProviders(<ArticlePreview article={mockArticle} />, { store });
    
    expect(screen.getByText('testuser')).toBeInTheDocument();
  });

  test('renders author image with correct src', () => {
    const store = createMockStore();
    renderWithProviders(<ArticlePreview article={mockArticle} />, { store });
    
    const authorImage = screen.getByAltText('testuser');
    expect(authorImage).toHaveAttribute('src', mockArticle.author.image);
  });

  test('renders fallback image when author image is missing', () => {
    const articleNoImage = {
      ...mockArticle,
      author: { ...mockArticle.author, image: null }
    };
    const store = createMockStore();
    renderWithProviders(<ArticlePreview article={articleNoImage} />, { store });
    
    const authorImage = screen.getByAltText('testuser');
    expect(authorImage).toHaveAttribute('src', 'https://static.productionready.io/images/smiley-cyrus.jpg');
  });

  test('renders formatted creation date', () => {
    const store = createMockStore();
    renderWithProviders(<ArticlePreview article={mockArticle} />, { store });
    
    const formattedDate = new Date(mockArticle.createdAt).toDateString();
    expect(screen.getByText(formattedDate)).toBeInTheDocument();
  });

  test('renders favorites count correctly', () => {
    const articleWithFavorites = {
      ...mockArticle,
      favoritesCount: 5
    };
    const store = createMockStore();
    renderWithProviders(<ArticlePreview article={articleWithFavorites} />, { store });
    
    expect(screen.getByText('5')).toBeInTheDocument();
  });

  test('renders all tags in tag list', () => {
    const store = createMockStore();
    renderWithProviders(<ArticlePreview article={mockArticle} />, { store });
    
    expect(screen.getByText('test')).toBeInTheDocument();
    expect(screen.getByText('article')).toBeInTheDocument();
  });

  test('favorite button has correct class when article is not favorited', () => {
    const store = createMockStore();
    const { container } = renderWithProviders(<ArticlePreview article={mockArticle} />, { store });
    
    const button = container.querySelector('.btn-outline-primary');
    expect(button).toBeInTheDocument();
  });

  test('favorite button has correct class when article is favorited', () => {
    const favoritedArticle = { ...mockArticle, favorited: true };
    const store = createMockStore();
    const { container } = renderWithProviders(<ArticlePreview article={favoritedArticle} />, { store });
    
    const button = container.querySelector('.btn-primary');
    expect(button).toBeInTheDocument();
  });

  test('clicking favorite button on non-favorited article dispatches favorite action', () => {
    const store = createMockStore();
    const { container } = renderWithProviders(<ArticlePreview article={mockArticle} />, { store });
    
    const favoriteButton = container.querySelector('.btn-outline-primary');
    fireEvent.click(favoriteButton);
    
    const actions = store.getActions();
    const hasFavoriteAction = actions.some(a => a.type === ARTICLE_FAVORITED);
    expect(hasFavoriteAction).toBe(true);
  });

  test('clicking favorite button on favorited article dispatches unfavorite action', () => {
    const favoritedArticle = { ...mockArticle, favorited: true };
    const store = createMockStore();
    const { container } = renderWithProviders(<ArticlePreview article={favoritedArticle} />, { store });
    
    const favoriteButton = container.querySelector('.btn-primary');
    fireEvent.click(favoriteButton);
    
    const actions = store.getActions();
    const hasUnfavoriteAction = actions.some(a => a.type === ARTICLE_UNFAVORITED);
    expect(hasUnfavoriteAction).toBe(true);
  });

  test('article title links to correct article page', () => {
    const store = createMockStore();
    renderWithProviders(<ArticlePreview article={mockArticle} />, { store });
    
    const articleLink = screen.getByText('Test Article').closest('a');
    expect(articleLink).toHaveAttribute('href', '/article/test-article');
  });

  test('author name links to correct profile page', () => {
    const store = createMockStore();
    const { container } = renderWithProviders(<ArticlePreview article={mockArticle} />, { store });
    
    const authorLinks = container.querySelectorAll('a[href="/@testuser"]');
    expect(authorLinks.length).toBeGreaterThan(0);
  });

  test('renders "Read more..." text', () => {
    const store = createMockStore();
    renderWithProviders(<ArticlePreview article={mockArticle} />, { store });
    
    expect(screen.getByText('Read more...')).toBeInTheDocument();
  });
});
